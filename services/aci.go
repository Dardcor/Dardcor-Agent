package services

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type ACIService struct {
	CommandService *CommandService
}

func NewACIService(cmdSvc *CommandService) *ACIService {
	return &ACIService{
		CommandService: cmdSvc,
	}
}

func (aci *ACIService) ViewFile(filePath string, startLine, endLine int) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var sb strings.Builder
	currentLine := 1

	for scanner.Scan() {
		if currentLine >= startLine && (endLine == -1 || currentLine <= endLine) {
			sb.WriteString(fmt.Sprintf("%d: %s\n", currentLine, scanner.Text()))
		}
		if endLine != -1 && currentLine > endLine {
			break
		}
		currentLine++
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return sb.String(), nil
}

func (aci *ACIService) EditFile(filePath, searchString, replacementString string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	text := string(content)
	if !strings.Contains(text, searchString) {
		return fmt.Errorf("search string not found in file")
	}

	occurrences := strings.Count(text, searchString)
	if occurrences > 1 {
		return fmt.Errorf("search string matches %d occurrences; must be highly specific payload", occurrences)
	}

	newText := strings.Replace(text, searchString, replacementString, 1)

	return os.WriteFile(filePath, []byte(newText), 0644)
}

func (aci *ACIService) RunCommandSandboxed(command, workDir string, timeoutSec int) (string, error) {
	if timeoutSec == 0 {
		timeoutSec = 30
	}

	cmd := exec.Command("cmd", "/C", command)
	cmd.Dir = workDir

	var outBuf bytes.Buffer
	var errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	cmd.Start()

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-time.After(time.Duration(timeoutSec) * time.Second):
		cmd.Process.Kill()
		return outBuf.String() + "\n" + errBuf.String(), fmt.Errorf("command timed out")
	case err := <-done:
		if err != nil {
			return outBuf.String() + "\n" + errBuf.String(), err
		}
	}

	return outBuf.String(), nil
}