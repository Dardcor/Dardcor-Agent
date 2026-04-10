package services

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
	"unicode/utf8"

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
	memService  *MemoryService
	webService  *WebService
	grepService *GrepService
	skillSvc    *SkillService
}

func NewAgentService(fs *FileSystemService, cmd *CommandService, sys *SystemService, ag *AntigravityService, mem *MemoryService, skill *SkillService) *AgentService {
	var llm *LLMProvider
	if config.AppConfig != nil {
		llm = NewLLMProvider(config.AppConfig.AI, ag)
	} else {
		llm = NewLLMProvider(config.AIConfig{Provider: "local"}, ag)
	}
	return &AgentService{
		fsService:   fs,
		cmdService:  cmd,
		sysService:  sys,
		agService:   ag,
		llmProvider: llm,
		memService:  mem,
		webService:  NewWebService(),
		grepService: NewGrepService(),
		skillSvc:    skill,
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

	useAI := false
	if as.agService != nil {
		if activeAcc, err := as.agService.GetActiveAccount(); err == nil && activeAcc != nil {
			useAI = true
			if as.llmProvider != nil {
				as.llmProvider.cfg.Provider = "antigravity"
			}
		}
	}
	if !useAI && config.AppConfig != nil && config.AppConfig.IsAIEnabled() {
		useAI = true
	}

	if useAI && as.llmProvider != nil {
		responseText = as.processWithLLM(req.Message, convID)
		
		aiActions, aiFinalText := as.parseAndExecuteActions(responseText)
		if len(aiActions) > 0 {
			actions = aiActions
			responseText = aiFinalText
		}

		if responseText == "" && len(actions) == 0 {
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

	var allMessages []LLMMessage
	if convID != "" {
		if conv, err := storage.Store.LoadConversation(convID); err == nil {
			for _, m := range conv.Messages {
				allMessages = append(allMessages, LLMMessage{
					Role:    m.Role,
					Content: m.Content,
				})
			}
		}
	}

	if len(allMessages) == 0 || allMessages[len(allMessages)-1].Content != message {
		allMessages = append(allMessages, LLMMessage{
			Role:    "user",
			Content: message,
		})
	}

	systemPrompt := as.buildSystemPrompt(message)
	historyMessages := as.truncateContextSmart(allMessages, 24000)

	resp, err := as.llmProvider.Complete(systemPrompt, historyMessages)
	if err != nil {
		return fmt.Sprintf("⚠️ AI Error: %v", err)
	}

	return resp.Content
}

func (as *AgentService) truncateContextSmart(messages []LLMMessage, maxRunes int) []LLMMessage {
	if len(messages) == 0 {
		return messages
	}

	var keptMsgs []LLMMessage
	currentRunes := 0

	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		msgRunes := utf8.RuneCountInString(msg.Content)

		if currentRunes+msgRunes > maxRunes {
			break
		}

		keptMsgs = append([]LLMMessage{msg}, keptMsgs...)
		currentRunes += msgRunes
	}

	return keptMsgs
}

func (as *AgentService) buildSystemPrompt(message string) string {
	hostname, _ := os.Hostname()
	isUltrawork := strings.Contains(strings.ToLower(message), "ultrawork") || strings.Contains(strings.ToLower(message), "ulw")
	
	base := fmt.Sprintf(`You are Dardcor Agent superior autonomous AI agent from Team Dardcor.

System: %s/%s | Host: %s
Mode: %s

[Operational Directives]
1. Autonomous Actions: If you need to perform a task, use the [ACTION] command [/ACTION] tag.
   Example: To open notepad, output: [ACTION] run notepad [/ACTION]
2. Intent Gate: Analyze implicit requirements.
3. BLOCKING: Check skills FIRST.
4. No Slop: Actionable responses only.

[Available Tools]
list, read, write, delete, search, mkdir, run, sysinfo, cpu, memory, processes, kill, drives, info, remember, fetch, websearch, grep, glob.

[Memory]
%s

[Skills]
%s
`, runtime.GOOS, runtime.GOARCH, hostname, func() string {
		if isUltrawork { return "ULTRAWORK" }
		return "STANDARD"
	}(), func() string {
		mem := as.memService.GetAll()
		if len(mem) == 0 { return "None" }
		b, _ := json.Marshal(mem)
		return string(b)
	}(), func() string {
		if as.skillSvc == nil { return "" }
		var sb strings.Builder
		for _, s := range as.skillSvc.GetSkills() {
			if s.Template != "" {
				sb.WriteString(fmt.Sprintf("\n### Skill: %s\n%s\n", s.Name, s.Template))
			}
		}
		return sb.String()
	}())

	return base
}

func (as *AgentService) interpretAndExecute(message string) ([]models.Action, string) {
	msg := strings.ToLower(strings.TrimSpace(message))
	var actions []models.Action
	var responseText string

	switch {
	case strings.HasPrefix(msg, "list ") || strings.HasPrefix(msg, "ls ") || strings.HasPrefix(msg, "dir "):
		actions, responseText = as.handleListDir(message)
	case strings.HasPrefix(msg, "read ") || strings.HasPrefix(msg, "cat "):
		actions, responseText = as.handleReadFile(message)
	case strings.HasPrefix(msg, "write "):
		actions, responseText = as.handleWriteFile(message)
	case strings.HasPrefix(msg, "delete ") || strings.HasPrefix(msg, "rm "):
		actions, responseText = as.handleDeleteFile(message)
	case strings.HasPrefix(msg, "search ") || strings.HasPrefix(msg, "find "):
		actions, responseText = as.handleSearch(message)
	case strings.HasPrefix(msg, "mkdir "):
		actions, responseText = as.handleMkdir(message)
	case strings.HasPrefix(msg, "drives"):
		actions, responseText = as.handleDrives()
	case strings.HasPrefix(msg, "run ") || strings.HasPrefix(msg, "exec "):
		actions, responseText = as.handleRunCommand(message)
	case strings.HasPrefix(msg, "cmd ") || strings.HasPrefix(msg, "$"):
		actions, responseText = as.handleDirectCommand(message)
	case msg == "sysinfo":
		actions, responseText = as.handleSystemInfo()
	case msg == "processes" || msg == "ps":
		actions, responseText = as.handleProcesses(message)
	case strings.HasPrefix(msg, "kill "):
		actions, responseText = as.handleKillProcess(message)
	case msg == "cpu":
		actions, responseText = as.handleCPUInfo()
	case msg == "memory":
		actions, responseText = as.handleMemoryInfo()
	case msg == "help" || msg == "?":
		responseText = as.getHelpText()
	case msg == "whoami":
		responseText = as.getAgentInfo()
	case strings.HasPrefix(msg, "info "):
		actions, responseText = as.handleFileInfo(message)
	case strings.HasPrefix(msg, "remember "):
		actions, responseText = as.handleRemember(message)
	case strings.HasPrefix(msg, "fetch ") || strings.HasPrefix(msg, "curl "):
		actions, responseText = as.handleWebFetch(message)
	case strings.HasPrefix(msg, "websearch ") || strings.HasPrefix(msg, "google "):
		actions, responseText = as.handleWebSearch(message)
	case strings.HasPrefix(msg, "grep "):
		actions, responseText = as.handleGrep(message)
	case strings.HasPrefix(msg, "glob "):
		actions, responseText = as.handleGlob(message)
	default:
		responseText = "**Dardcor Agent** is ready. Type `help` for tools."
	}

	return actions, responseText
}

func (as *AgentService) handleListDir(message string) ([]models.Action, string) {
	path := as.extractPath(message, []string{"list ", "ls ", "dir "})
	if path == "" { path = "." }
	action := models.Action{Type: "list_directory", Description: fmt.Sprintf("Listing: %s", path), Status: "running"}
	start := time.Now()
	files, err := as.fsService.ListDirectory(path)
	action.Duration = time.Since(start).Milliseconds()
	if err != nil {
		action.Status = "error"
		return []models.Action{action}, fmt.Sprintf("Error: %v", err)
	}
	action.Status = "completed"
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📂 **%s** (%d items)\n\n", path, len(files)))
	for _, f := range files {
		if f.IsDir { sb.WriteString(fmt.Sprintf("📁 `%s/`\n", f.Name)) } else { sb.WriteString(fmt.Sprintf("📄 `%s`\n", f.Name)) }
	}
	return []models.Action{action}, sb.String()
}

func (as *AgentService) handleReadFile(message string) ([]models.Action, string) {
	path := as.extractPath(message, []string{"read ", "cat "})
	action := models.Action{Type: "read_file", Description: fmt.Sprintf("Reading: %s", path), Status: "running"}
	content, err := as.fsService.ReadFile(path)
	if err != nil {
		action.Status = "error"
		return []models.Action{action}, fmt.Sprintf("Error: %v", err)
	}
	action.Status = "completed"
	display := content.Content
	if len(display) > 4000 { display = display[:4000] + "\n..." }
	return []models.Action{action}, fmt.Sprintf("📄 **%s**\n\n```\n%s\n```", path, display)
}

func (as *AgentService) handleWriteFile(message string) ([]models.Action, string) {
	parts := strings.SplitN(message, " ", 3)
	if len(parts) < 3 { return nil, "Use: write <path> <content>" }
	path, content := parts[1], parts[2]
	err := as.fsService.WriteFile(path, content)
	if err != nil { return nil, fmt.Sprintf("Error: %v", err) }
	return []models.Action{{Type: "write_file", Status: "completed"}}, fmt.Sprintf("Saved: %s", path)
}

func (as *AgentService) handleDeleteFile(message string) ([]models.Action, string) {
	path := as.extractPath(message, []string{"delete ", "rm "})
	err := as.fsService.DeleteFile(path)
	if err != nil { return nil, err.Error() }
	return []models.Action{{Type: "delete_file", Status: "completed"}}, fmt.Sprintf("Deleted: %s", path)
}

func (as *AgentService) handleSearch(message string) ([]models.Action, string) {
	query := as.extractPath(message, []string{"search ", "find "})
	results, err := as.fsService.SearchFiles(models.SearchRequest{Path: ".", Query: query, MaxDepth: 5})
	if err != nil { return nil, err.Error() }
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🔍 Results for \"%s\":\n", query))
	for _, r := range results { sb.WriteString(fmt.Sprintf("- %s\n", r.Path)) }
	return []models.Action{{Type: "search_files", Status: "completed"}}, sb.String()
}

func (as *AgentService) handleMkdir(message string) ([]models.Action, string) {
	path := as.extractPath(message, []string{"mkdir "})
	err := as.fsService.CreateDirectory(path)
	if err != nil { return nil, err.Error() }
	return []models.Action{{Type: "create_directory", Status: "completed"}}, fmt.Sprintf("Created: %s", path)
}

func (as *AgentService) handleDrives() ([]models.Action, string) {
	drives := as.fsService.GetDrives()
	var sb strings.Builder
	sb.WriteString("💽 Drives:\n")
	for _, d := range drives { sb.WriteString(fmt.Sprintf("- %s\n", d)) }
	return []models.Action{{Type: "list_drives", Status: "completed"}}, sb.String()
}

func (as *AgentService) handleRunCommand(message string) ([]models.Action, string) {
	cmd := as.extractPath(message, []string{"run ", "exec "})
	res, err := as.cmdService.ExecuteCommand(models.CommandRequest{Command: cmd, Timeout: 30})
	if err != nil { return nil, err.Error() }
	return []models.Action{{Type: "execute_command", Status: "completed"}}, fmt.Sprintf("💻 Output:\n```\n%s\n```", res.Output)
}

func (as *AgentService) handleDirectCommand(message string) ([]models.Action, string) {
	cmd := strings.TrimSpace(message)
	if strings.HasPrefix(cmd, "cmd ") { cmd = strings.TrimPrefix(cmd, "cmd ") } else if strings.HasPrefix(cmd, "$") { cmd = strings.TrimPrefix(cmd, "$") }
	return as.handleRunCommand("run " + cmd)
}

func (as *AgentService) handleSystemInfo() ([]models.Action, string) {
	info, err := as.sysService.GetSystemInfo()
	if err != nil { return nil, err.Error() }
	res := fmt.Sprintf("🖥️ %s | CPU: %.1f%% | RAM: %.1f%%", info.HostName, info.CPU.UsagePercent, info.Memory.UsedPercent)
	return []models.Action{{Type: "system_info", Status: "completed"}}, res
}

func (as *AgentService) handleProcesses(message string) ([]models.Action, string) {
	procs, err := as.sysService.GetProcesses("cpu", 10)
	if err != nil { return nil, err.Error() }
	var sb strings.Builder
	sb.WriteString("⚙️ Top Processes:\n")
	for _, p := range procs { sb.WriteString(fmt.Sprintf("- %d: %s (%.1f%%)\n", p.PID, p.Name, p.CPUPercent)) }
	return []models.Action{{Type: "list_processes", Status: "completed"}}, sb.String()
}

func (as *AgentService) handleKillProcess(message string) ([]models.Action, string) {
	pidStr := as.extractPath(message, []string{"kill "})
	var pid int32
	fmt.Sscanf(pidStr, "%d", &pid)
	err := as.sysService.KillProcess(pid)
	if err != nil { return nil, err.Error() }
	return []models.Action{{Type: "kill_process", Status: "completed"}}, fmt.Sprintf("Killed: %d", pid)
}

func (as *AgentService) handleCPUInfo() ([]models.Action, string) {
	info, _ := as.sysService.GetSystemInfo()
	return nil, fmt.Sprintf("🔧 CPU: %s | Usage: %.1f%%", info.CPU.ModelName, info.CPU.UsagePercent)
}

func (as *AgentService) handleMemoryInfo() ([]models.Action, string) {
	mem, _ := as.sysService.GetMemoryUsage()
	return nil, fmt.Sprintf("🧠 Memory: %.1f%% used", mem.UsedPercent)
}

func (as *AgentService) handleFileInfo(message string) ([]models.Action, string) {
	path := as.extractPath(message, []string{"info "})
	info, err := as.fsService.GetFileInfo(path)
	if err != nil { return nil, err.Error() }
	return nil, fmt.Sprintf("📄 %s (%s)", info.Name, formatSize(info.Size))
}

func (as *AgentService) getHelpText() string {
	return `🤖 **Dardcor Agent Tools:**
- list, read, write, delete, search, mkdir, info, drives
- run, cmd, $
- sysinfo, cpu, memory, processes, kill
- fetch, websearch, grep, glob
- remember, whoami, help`
}

func (as *AgentService) handleRemember(message string) ([]models.Action, string) {
	parts := strings.SplitN(message, " ", 3)
	if len(parts) < 3 { return nil, "Use: remember <key> <value>" }
	as.memService.Set(parts[1], parts[2])
	return nil, fmt.Sprintf("Stored: %s", parts[1])
}

func (as *AgentService) handleWebFetch(message string) ([]models.Action, string) {
	url := as.extractPath(message, []string{"fetch ", "curl "})
	content, err := as.webService.Fetch(url, 10000)
	if err != nil { return nil, err.Error() }
	return []models.Action{{Type: "web_fetch", Status: "completed"}}, fmt.Sprintf("🌐 Fetched %s\n\n%s", url, content)
}

func (as *AgentService) handleWebSearch(message string) ([]models.Action, string) {
	query := as.extractPath(message, []string{"websearch ", "google "})
	res, err := as.webService.SearchDDG(query, 5)
	if err != nil { return nil, err.Error() }
	return []models.Action{{Type: "web_search", Status: "completed"}}, fmt.Sprintf("🔍 Results:\n%s", res)
}

func (as *AgentService) handleGrep(message string) ([]models.Action, string) {
	parts := strings.SplitN(message, " ", 3)
	if len(parts) < 2 { return nil, "Use: grep <pattern> [path]" }
	root := "."
	if len(parts) >= 3 { root = parts[2] }
	res, err := as.grepService.Search(root, parts[1], 30, "")
	if err != nil { return nil, err.Error() }
	var sb strings.Builder
	for _, r := range res { sb.WriteString(fmt.Sprintf("%s:%d -> %s\n", r.File, r.Line, r.Content)) }
	return []models.Action{{Type: "grep", Status: "completed"}}, sb.String()
}

func (as *AgentService) handleGlob(message string) ([]models.Action, string) {
	parts := strings.SplitN(message, " ", 3)
	if len(parts) < 2 { return nil, "Use: glob <pattern> [path]" }
	root := "."
	if len(parts) >= 3 { root = parts[2] }
	res, err := as.grepService.Glob(root, parts[1], 50)
	if err != nil { return nil, err.Error() }
	var sb strings.Builder
	for _, r := range res { sb.WriteString(fmt.Sprintf("- %s\n", r)) }
	return []models.Action{{Type: "glob", Status: "completed"}}, sb.String()
}

func (as *AgentService) getAgentInfo() string {
	return "**Dardcor Agent** — Superior autonomous assistant."
}

func (as *AgentService) parseAndExecuteActions(text string) ([]models.Action, string) {
	var allActions []models.Action
	remainingText := text

	for {
		startIdx := strings.Index(remainingText, "[ACTION]")
		endIdx := strings.Index(remainingText, "[/ACTION]")
		if startIdx == -1 || endIdx == -1 || endIdx <= startIdx {
			break
		}

		command := strings.TrimSpace(remainingText[startIdx+8 : endIdx])
		
		if strings.Contains(command, "{") && strings.Contains(command, "}") {
			parts := strings.SplitN(command, " ", 2)
			if len(parts) > 1 {
				var args map[string]interface{}
				if err := json.Unmarshal([]byte(parts[1]), &args); err == nil {
					for _, v := range args {
						command = parts[0] + " " + fmt.Sprint(v)
						break
					}
				}
			}
		}

		actions, result := as.interpretAndExecute(command)
		allActions = append(allActions, actions...)
		
		remainingText = remainingText[:startIdx] + "\n> **Action Result:**\n" + result + "\n" + remainingText[endIdx+10:]
	}

	return allActions, remainingText
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

func (as *AgentService) formatSize(b int64) string {
	return formatSize(b)
}
