package services

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"dardcor-agent/config"
	"dardcor-agent/models"
	"dardcor-agent/storage"

	"github.com/google/uuid"
)

type AgentService struct {
	fsService   *FileSystemService
	cmdService  *CommandService
	sysService  *SystemService
	agService   *AntigravityService
	llmProvider *LLMProvider
}

func NewAgentService(fs *FileSystemService, cmd *CommandService, sys *SystemService, ag *AntigravityService) *AgentService {
	var llm *LLMProvider
	if config.AppConfig != nil {
		llm = NewLLMProvider(config.AppConfig.AI, ag)
	} else {
		// Create LLM provider even without full config, so antigravity can work
		llm = NewLLMProvider(config.AIConfig{Provider: "local"}, ag)
	}
	return &AgentService{
		fsService:   fs,
		cmdService:  cmd,
		sysService:  sys,
		agService:   ag,
		llmProvider: llm,
	}
}

func (as *AgentService) ProcessMessage(req models.AgentRequest) (*models.AgentResponse, error) {
	var convID string
	if req.ConversationID != "" {
		convID = req.ConversationID
	} else {
		conv, err := storage.Store.CreateConversation(as.generateTitle(req.Message))
		if err != nil {
			return nil, err
		}
		convID = conv.ID
	}

	userMsg := models.Message{
		Role:    "user",
		Content: req.Message,
	}
	storage.Store.AddMessage(convID, userMsg)

	var actions []models.Action
	var responseText string

	// Determine if we should use LLM
	// Priority 1: Active Antigravity account exists -> always route through it
	// Priority 2: Config says AI is enabled (e.g. OpenAI, Anthropic, etc)
	useAI := false
	if as.agService != nil {
		if activeAcc, err := as.agService.GetActiveAccount(); err == nil && activeAcc != nil {
			// Account is active - callAntigravity will handle token refresh automatically
			useAI = true
			if as.llmProvider != nil {
				as.llmProvider.cfg.Provider = "antigravity"
			}
			_ = activeAcc // used implicitly via agService in callAntigravity
		}
	}
	if !useAI && config.AppConfig != nil && config.AppConfig.IsAIEnabled() {
		useAI = true
	}

	if useAI && as.llmProvider != nil {
		responseText = as.processWithLLM(req.Message, convID)
		if responseText == "" {
			actions, responseText = as.interpretAndExecute(req.Message)
		}
	} else {
		actions, responseText = as.interpretAndExecute(req.Message)
	}

	response := &models.AgentResponse{
		ID:             uuid.New().String(),
		ConversationID: convID,
		Role:           "assistant",
		Content:        responseText,
		Actions:        actions,
		Timestamp:      time.Now(),
		Status:         "completed",
	}

	assistantMsg := models.Message{
		Role:    "assistant",
		Content: responseText,
		Actions: actions,
	}
	storage.Store.AddMessage(convID, assistantMsg)

	return response, nil
}

func (as *AgentService) processWithLLM(message string, convID string) string {
	if as.llmProvider == nil {
		return ""
	}

	var historyMessages []LLMMessage
	if convID != "" {
		if conv, err := storage.Store.LoadConversation(convID); err == nil {
			start := 0
			if len(conv.Messages) > 6 {
				start = len(conv.Messages) - 6
			}
			for _, m := range conv.Messages[start:] {
				historyMessages = append(historyMessages, LLMMessage{
					Role:    m.Role,
					Content: m.Content,
				})
			}
		}
	}

	if len(historyMessages) == 0 || historyMessages[len(historyMessages)-1].Content != message {
		historyMessages = append(historyMessages, LLMMessage{
			Role:    "user",
			Content: message,
		})
	}

	systemPrompt := as.buildSystemPrompt(message)

	resp, err := as.llmProvider.Complete(systemPrompt, historyMessages)
	if err != nil {
		// Return a meaningful error message instead of empty string
		return fmt.Sprintf("⚠️ AI Error: %v\n\nPastikan akun Antigravity sudah diaktifkan di halaman Model.", err)
	}

	return resp.Content
}

func (as *AgentService) buildSystemPrompt(message string) string {
	hostname, _ := os.Hostname()
	isUltrawork := strings.Contains(strings.ToLower(message), "ultrawork") ||
		strings.Contains(strings.ToLower(message), "ulw")
	isPlan := strings.Contains(strings.ToLower(message), "[read-only")

	base := fmt.Sprintf(`You are Dardcor Agent, a superior autonomous AI programming assistant.

System: %s/%s | Host: %s
Status: High-Efficiency Mode

Capabilities:
- Filesystem: read, write, delete, search, mkdir
- Shell: full command execution
- Monitoring: CPU, RAM, Disk, Processes
- Context: Optimized conversation memory

Operational Directives:
1. Analysis: Precision-driven reasoning before action.
2. Execution: Autonomous, error-resilient command use.
3. Efficiency: Minimal tokens, maximal results.
4. Response: Markdown-focused, concise, actionable.

Available Tools: list, read, write, delete, search, mkdir, run, sysinfo, cpu, memory, processes, kill, drives, info.
`, runtime.GOOS, runtime.GOARCH, hostname)

	if isUltrawork {
		base += `
ULTRAWORK MODE: Execute complex multi-step tasks autonomously. Adapt, retry, and achieve the goal without interruption.
`
	}

	if isPlan {
		base += `
PLAN MODE: Read-only analysis. Strategic planning without file or system modification.
`
	}

	return base
}

func (as *AgentService) interpretAndExecute(message string) ([]models.Action, string) {
	msg := strings.ToLower(strings.TrimSpace(message))
	var actions []models.Action
	var responseText string

	switch {
	case strings.HasPrefix(msg, "list ") || strings.HasPrefix(msg, "ls ") || strings.HasPrefix(msg, "dir "):
		actions, responseText = as.handleListDir(message)
	case strings.HasPrefix(msg, "read ") || strings.HasPrefix(msg, "cat ") || strings.HasPrefix(msg, "baca "):
		actions, responseText = as.handleReadFile(message)
	case strings.HasPrefix(msg, "write ") || strings.HasPrefix(msg, "tulis "):
		actions, responseText = as.handleWriteFile(message)
	case strings.HasPrefix(msg, "delete ") || strings.HasPrefix(msg, "hapus ") || strings.HasPrefix(msg, "rm "):
		actions, responseText = as.handleDeleteFile(message)
	case strings.HasPrefix(msg, "search ") || strings.HasPrefix(msg, "cari ") || strings.HasPrefix(msg, "find "):
		actions, responseText = as.handleSearch(message)
	case strings.HasPrefix(msg, "mkdir ") || strings.HasPrefix(msg, "buat folder "):
		actions, responseText = as.handleMkdir(message)
	case strings.HasPrefix(msg, "drives") || strings.HasPrefix(msg, "disk"):
		actions, responseText = as.handleDrives()
	case strings.HasPrefix(msg, "run ") || strings.HasPrefix(msg, "exec ") || strings.HasPrefix(msg, "jalankan "):
		actions, responseText = as.handleRunCommand(message)
	case strings.HasPrefix(msg, "cmd ") || strings.HasPrefix(msg, "$"):
		actions, responseText = as.handleDirectCommand(message)
	case msg == "sysinfo" || msg == "system info" || msg == "info sistem" || msg == "system":
		actions, responseText = as.handleSystemInfo()
	case msg == "processes" || msg == "proses" || msg == "ps" || strings.HasPrefix(msg, "top"):
		actions, responseText = as.handleProcesses(message)
	case strings.HasPrefix(msg, "kill ") || strings.HasPrefix(msg, "matikan "):
		actions, responseText = as.handleKillProcess(message)
	case msg == "cpu":
		actions, responseText = as.handleCPUInfo()
	case msg == "memory" || msg == "ram" || msg == "mem":
		actions, responseText = as.handleMemoryInfo()
	case msg == "help" || msg == "bantuan" || msg == "?":
		responseText = as.getHelpText()
	case msg == "whoami" || msg == "siapa":
		responseText = as.getAgentInfo()
	case strings.HasPrefix(msg, "info "):
		actions, responseText = as.handleFileInfo(message)
	default:
		responseText = fmt.Sprintf("**Dardcor Agent** is ready. Status: Online.\n\nType `help` for available tools.")
	}

	return actions, responseText
}

func (as *AgentService) handleListDir(message string) ([]models.Action, string) {
	path := as.extractPath(message, []string{"list ", "ls ", "dir "})
	if path == "" {
		path = "."
	}
	action := models.Action{
		Type:        "list_directory",
		Description: fmt.Sprintf("Listing: %s", path),
		Parameters:  map[string]interface{}{"path": path},
		Status:      "running",
	}
	start := time.Now()
	files, err := as.fsService.ListDirectory(path)
	action.Duration = time.Since(start).Milliseconds()
	if err != nil {
		action.Status = "error"
		action.Error = err.Error()
		return []models.Action{action}, fmt.Sprintf("Error: %s", err.Error())
	}
	action.Status = "completed"
	action.Result = files
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📂 **Directory:** `%s` (%d items)\n\n", path, len(files)))
	for _, f := range files {
		if f.IsDir {
			sb.WriteString(fmt.Sprintf("📁 `%s/`\n", f.Name))
		} else {
			sb.WriteString(fmt.Sprintf("📄 `%s` (%s)\n", f.Name, formatSize(f.Size)))
		}
	}
	return []models.Action{action}, sb.String()
}

func (as *AgentService) handleReadFile(message string) ([]models.Action, string) {
	path := as.extractPath(message, []string{"read ", "cat ", "baca "})
	action := models.Action{
		Type:        "read_file",
		Description: fmt.Sprintf("Reading: %s", path),
		Parameters:  map[string]interface{}{"path": path},
		Status:      "running",
	}
	start := time.Now()
	content, err := as.fsService.ReadFile(path)
	action.Duration = time.Since(start).Milliseconds()
	if err != nil {
		action.Status = "error"
		action.Error = err.Error()
		return []models.Action{action}, fmt.Sprintf("Error: %s", err.Error())
	}
	action.Status = "completed"
	display := content.Content
	if len(display) > 4000 {
		display = display[:4000] + "\n... (truncated)"
	}
	return []models.Action{action}, fmt.Sprintf("📄 **File:** `%s`\n\n```\n%s\n```", path, display)
}

func (as *AgentService) handleWriteFile(message string) ([]models.Action, string) {
	parts := strings.SplitN(message, " ", 3)
	if len(parts) < 3 {
		return nil, "Use: `write <path> <content>`"
	}
	path := parts[1]
	content := parts[2]
	action := models.Action{
		Type:        "write_file",
		Description: fmt.Sprintf("Writing: %s", path),
		Parameters:  map[string]interface{}{"path": path},
		Status:      "running",
	}
	start := time.Now()
	err := as.fsService.WriteFile(path, content)
	action.Duration = time.Since(start).Milliseconds()
	if err != nil {
		action.Status = "error"
		action.Error = err.Error()
		return []models.Action{action}, fmt.Sprintf("Error: %s", err.Error())
	}
	action.Status = "completed"
	return []models.Action{action}, fmt.Sprintf("File saved: `%s`", path)
}

func (as *AgentService) handleDeleteFile(message string) ([]models.Action, string) {
	path := as.extractPath(message, []string{"delete ", "hapus ", "rm "})
	action := models.Action{
		Type:        "delete_file",
		Description: fmt.Sprintf("Deleting: %s", path),
		Parameters:  map[string]interface{}{"path": path},
		Status:      "running",
	}
	start := time.Now()
	err := as.fsService.DeleteFile(path)
	action.Duration = time.Since(start).Milliseconds()
	if err != nil {
		action.Status = "error"
		action.Error = err.Error()
		return []models.Action{action}, fmt.Sprintf("Error: %s", err.Error())
	}
	action.Status = "completed"
	return []models.Action{action}, fmt.Sprintf("Deleted: `%s`", path)
}

func (as *AgentService) handleSearch(message string) ([]models.Action, string) {
	query := as.extractPath(message, []string{"search ", "cari ", "find "})
	action := models.Action{
		Type:        "search_files",
		Description: fmt.Sprintf("Searching: %s", query),
		Parameters:  map[string]interface{}{"query": query},
		Status:      "running",
	}
	start := time.Now()
	results, err := as.fsService.SearchFiles(models.SearchRequest{Path: ".", Query: query, MaxDepth: 5})
	action.Duration = time.Since(start).Milliseconds()
	if err != nil {
		action.Status = "error"
		action.Error = err.Error()
		return []models.Action{action}, fmt.Sprintf("Error: %s", err.Error())
	}
	action.Status = "completed"
	action.Result = results
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🔍 **Search for:** \"%s\" (%d results)\n\n", query, len(results)))
	for _, r := range results {
		icon := "📄"
		if r.IsDir { icon = "📁" }
		sb.WriteString(fmt.Sprintf("%s `%s`\n", icon, r.Path))
	}
	return []models.Action{action}, sb.String()
}

func (as *AgentService) handleMkdir(message string) ([]models.Action, string) {
	path := as.extractPath(message, []string{"mkdir ", "buat folder "})
	action := models.Action{
		Type:        "create_directory",
		Description: fmt.Sprintf("Mkdir: %s", path),
		Parameters:  map[string]interface{}{"path": path},
		Status:      "running",
	}
	start := time.Now()
	err := as.fsService.CreateDirectory(path)
	action.Duration = time.Since(start).Milliseconds()
	if err != nil {
		action.Status = "error"
		action.Error = err.Error()
		return []models.Action{action}, fmt.Sprintf("Error: %s", err.Error())
	}
	action.Status = "completed"
	return []models.Action{action}, fmt.Sprintf("Created: `%s`", path)
}

func (as *AgentService) handleDrives() ([]models.Action, string) {
	drives := as.fsService.GetDrives()
	action := models.Action{Type: "list_drives", Status: "completed", Result: drives}
	var sb strings.Builder
	sb.WriteString("💽 **Drives:**\n\n")
	for _, d := range drives { sb.WriteString(fmt.Sprintf("- `%s`\n", d)) }
	return []models.Action{action}, sb.String()
}

func (as *AgentService) handleRunCommand(message string) ([]models.Action, string) {
	cmd := as.extractPath(message, []string{"run ", "exec ", "jalankan "})
	action := models.Action{
		Type:        "execute_command",
		Description: fmt.Sprintf("Running: %s", cmd),
		Parameters:  map[string]interface{}{"command": cmd},
		Status:      "running",
	}
	start := time.Now()
	result, err := as.cmdService.ExecuteCommand(models.CommandRequest{Command: cmd, Timeout: 30})
	action.Duration = time.Since(start).Milliseconds()
	if err != nil {
		action.Status = "error"
		action.Error = err.Error()
		return []models.Action{action}, fmt.Sprintf("Error: %s", err.Error())
	}
	action.Status = "completed"
	output := result.Output
	if len(output) > 4000 { output = output[:4000] + "\n... (truncated)" }
	return []models.Action{action}, fmt.Sprintf("💻 **Command:** `%s`\n\n```\n%s\n```", cmd, output)
}

func (as *AgentService) handleDirectCommand(message string) ([]models.Action, string) {
	cmd := strings.TrimSpace(message)
	if strings.HasPrefix(cmd, "cmd ") { cmd = strings.TrimPrefix(cmd, "cmd ") } else if strings.HasPrefix(cmd, "$") { cmd = strings.TrimPrefix(cmd, "$") }
	return as.handleRunCommand("run " + strings.TrimSpace(cmd))
}

func (as *AgentService) handleSystemInfo() ([]models.Action, string) {
	info, err := as.sysService.GetSystemInfo()
	if err != nil { return nil, err.Error() }
	action := models.Action{Type: "system_info", Status: "completed", Result: info}
	res := fmt.Sprintf("🖥️ **System:** %s | **CPU:** %.1f%% | **RAM:** %.1f%%", info.HostName, info.CPU.UsagePercent, info.Memory.UsedPercent)
	return []models.Action{action}, res
}

func (as *AgentService) handleProcesses(message string) ([]models.Action, string) {
	procs, err := as.sysService.GetProcesses("cpu", 10)
	if err != nil { return nil, err.Error() }
	action := models.Action{Type: "list_processes", Status: "completed", Result: procs}
	var sb strings.Builder
	sb.WriteString("⚙️ **Processes:**\n\n")
	for _, p := range procs { sb.WriteString(fmt.Sprintf("- %d: %s (%.1f%%)\n", p.PID, p.Name, p.CPUPercent)) }
	return []models.Action{action}, sb.String()
}

func (as *AgentService) handleKillProcess(message string) ([]models.Action, string) {
	pidStr := as.extractPath(message, []string{"kill ", "matikan "})
	var pid int32
	fmt.Sscanf(pidStr, "%d", &pid)
	if pid == 0 { return nil, "Missing PID." }
	err := as.sysService.KillProcess(pid)
	if err != nil { return nil, err.Error() }
	return []models.Action{{Type: "kill_process", Status: "completed"}}, fmt.Sprintf("Killed: %d", pid)
}

func (as *AgentService) handleCPUInfo() ([]models.Action, string) {
	info, _ := as.sysService.GetSystemInfo()
	return nil, fmt.Sprintf("🔧 **CPU:** %s | Usage: %.1f%%", info.CPU.ModelName, info.CPU.UsagePercent)
}

func (as *AgentService) handleMemoryInfo() ([]models.Action, string) {
	mem, _ := as.sysService.GetMemoryUsage()
	return nil, fmt.Sprintf("🧠 **Memory:** %.1f%% used", mem.UsedPercent)
}

func (as *AgentService) handleFileInfo(message string) ([]models.Action, string) {
	path := as.extractPath(message, []string{"info "})
	info, err := as.fsService.GetFileInfo(path)
	if err != nil { return nil, err.Error() }
	return nil, fmt.Sprintf("📄 **Info:** %s (%s)", info.Name, formatSize(info.Size))
}

func (as *AgentService) getHelpText() string {
	return `🤖 **Dardcor Agent Help**

- list, read, write, delete, search, mkdir, info, drives
- run, cmd, $ (command execution)
- sysinfo, cpu, memory, processes, kill
- whoami, help`
}

func (as *AgentService) getAgentInfo() string {
	return "**Dardcor Agent** — Superior Autonomous System Assistant."
}

func (as *AgentService) generateTitle(message string) string {
	if len(message) > 40 { return message[:40] + "..." }
	return message
}

func (as *AgentService) extractPath(message string, prefixes []string) string {
	for _, p := range prefixes {
		if strings.HasPrefix(strings.ToLower(message), p) { return strings.TrimSpace(message[len(p):]) }
	}
	parts := strings.Fields(message)
	if len(parts) > 1 { return strings.Join(parts[1:], " ") }
	return ""
}

func formatSize(b int64) string {
	if b >= 1073741824 { return fmt.Sprintf("%.1f GB", float64(b)/1073741824) }
	if b >= 1048576 { return fmt.Sprintf("%.1f MB", float64(b)/1048576) }
	if b >= 1024 { return fmt.Sprintf("%.1f KB", float64(b)/1024) }
	return fmt.Sprintf("%d B", b)
}

func formatDuration(seconds uint64) string {
	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	mins := (seconds % 3600) / 60
	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return fmt.Sprintf("%dm", mins)
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n-3] + "..."
	}
	return s
}
