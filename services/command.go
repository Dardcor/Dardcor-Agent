package services

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"dardcor-agent/models"
	"dardcor-agent/storage"

	"github.com/google/uuid"
)

type CommandService struct {
	activeCommands sync.Map
}

type ActiveCommand struct {
	ID      string
	Cmd     *exec.Cmd
	Cancel  context.CancelFunc
	Output  *bytes.Buffer
	ErrBuf  *bytes.Buffer
	Started time.Time
}

func NewCommandService() *CommandService {
	return &CommandService{}
}

// ExecuteCommand executes a system command and returns the result
func (cs *CommandService) ExecuteCommand(req models.CommandRequest) (*models.CommandResponse, error) {
	if req.Command == "" {
		return nil, fmt.Errorf("command cannot be empty")
	}

	timeout := req.Timeout
	if timeout <= 0 {
		timeout = 30
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/C", req.Command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", req.Command)
	}

	if req.WorkingDir != "" {
		cmd.Dir = req.WorkingDir
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	startTime := time.Now()
	err := cmd.Run()
	endTime := time.Now()

	response := &models.CommandResponse{
		ID:         uuid.New().String(),
		Command:    req.Command,
		Output:     stdout.String(),
		ExitCode:   0,
		Duration:   endTime.Sub(startTime).Milliseconds(),
		StartedAt:  startTime,
		FinishedAt: endTime,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			response.ExitCode = exitErr.ExitCode()
		} else {
			response.ExitCode = -1
		}
		response.Error = stderr.String()
		if response.Error == "" {
			response.Error = err.Error()
		}
	}

	// Save to history
	storage.Store.SaveCommandHistory(*response)

	return response, nil
}

// ExecuteCommandStreaming executes a command with streaming output via callback
func (cs *CommandService) ExecuteCommandStreaming(req models.CommandRequest, onOutput func(string, bool)) (*models.CommandResponse, error) {
	if req.Command == "" {
		return nil, fmt.Errorf("command cannot be empty")
	}

	timeout := req.Timeout
	if timeout <= 0 {
		timeout = 60
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/C", req.Command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", req.Command)
	}

	if req.WorkingDir != "" {
		cmd.Dir = req.WorkingDir
	}

	cmdID := uuid.New().String()
	var outputBuf, errBuf bytes.Buffer

	// Set up pipes
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return nil, err
	}

	startTime := time.Now()
	if err := cmd.Start(); err != nil {
		cancel()
		return nil, err
	}

	// Store active command
	cs.activeCommands.Store(cmdID, &ActiveCommand{
		ID:      cmdID,
		Cmd:     cmd,
		Cancel:  cancel,
		Output:  &outputBuf,
		ErrBuf:  &errBuf,
		Started: startTime,
	})
	defer cs.activeCommands.Delete(cmdID)

	// Read stdout in goroutine
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		buf := make([]byte, 4096)
		for {
			n, err := stdoutPipe.Read(buf)
			if n > 0 {
				text := string(buf[:n])
				outputBuf.WriteString(text)
				if onOutput != nil {
					onOutput(text, false)
				}
			}
			if err != nil {
				break
			}
		}
	}()

	go func() {
		defer wg.Done()
		buf := make([]byte, 4096)
		for {
			n, err := stderrPipe.Read(buf)
			if n > 0 {
				text := string(buf[:n])
				errBuf.WriteString(text)
				if onOutput != nil {
					onOutput(text, true)
				}
			}
			if err != nil {
				break
			}
		}
	}()

	wg.Wait()
	cmdErr := cmd.Wait()
	endTime := time.Now()

	response := &models.CommandResponse{
		ID:         cmdID,
		Command:    req.Command,
		Output:     outputBuf.String(),
		ExitCode:   0,
		Duration:   endTime.Sub(startTime).Milliseconds(),
		StartedAt:  startTime,
		FinishedAt: endTime,
	}

	if cmdErr != nil {
		if exitErr, ok := cmdErr.(*exec.ExitError); ok {
			response.ExitCode = exitErr.ExitCode()
		} else {
			response.ExitCode = -1
		}
		response.Error = errBuf.String()
		if response.Error == "" {
			response.Error = cmdErr.Error()
		}
	}

	storage.Store.SaveCommandHistory(*response)
	return response, nil
}

// KillCommand kills an active command
func (cs *CommandService) KillCommand(id string) error {
	if val, ok := cs.activeCommands.Load(id); ok {
		ac := val.(*ActiveCommand)
		ac.Cancel()
		return nil
	}
	return fmt.Errorf("command not found: %s", id)
}

// GetCommandHistory returns recent command history
func (cs *CommandService) GetCommandHistory(limit int) ([]models.CommandResponse, error) {
	return storage.Store.GetCommandHistory(limit)
}

// GetShellInfo returns information about the default shell
func (cs *CommandService) GetShellInfo() map[string]string {
	info := map[string]string{
		"os":   runtime.GOOS,
		"arch": runtime.GOARCH,
	}

	if runtime.GOOS == "windows" {
		info["shell"] = "cmd.exe"
		info["shell_flag"] = "/C"

		// Check if PowerShell is available
		if _, err := exec.LookPath("powershell"); err == nil {
			info["powershell"] = "available"
		}
	} else {
		info["shell"] = "/bin/bash"
		info["shell_flag"] = "-c"
	}

	// Get PATH
	if path, err := exec.Command("echo", "%PATH%").Output(); err == nil {
		pathStr := strings.TrimSpace(string(path))
		if len(pathStr) > 500 {
			pathStr = pathStr[:500] + "..."
		}
		info["path"] = pathStr
	}

	return info
}
