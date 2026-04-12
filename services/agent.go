package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"dardcor-agent/config"
	"dardcor-agent/models"
	"dardcor-agent/storage"

	"github.com/google/uuid"
)

var dangerousPatterns = []*regexp.Regexp{
	regexp.MustCompile(`\brm\s+-[rf]{1,2}\b`),
	regexp.MustCompile(`\bdel\s+/[fq]\b`),
	regexp.MustCompile(`\brmdir\s+/s\b`),
	regexp.MustCompile(`(^|[^\-\w])\b(format|mkfs|diskpart)\b\s`),
	regexp.MustCompile(`\bdd\s+if=`),
	regexp.MustCompile(`\b(shutdown|reboot|poweroff)\b`),
	regexp.MustCompile(`:\(\)\s*\{.*\};\s*:`),
	regexp.MustCompile(`\|\s*sh\b`),
	regexp.MustCompile(`\|\s*bash\b`),
	regexp.MustCompile(`;\s*rm\s+-[rf]`),
	regexp.MustCompile(`&&\s*rm\s+-[rf]`),
	regexp.MustCompile(`\bsudo\b`),
	regexp.MustCompile(`\bchmod\s+[0-7]{3,4}\b`),
	regexp.MustCompile(`\bchown\b`),
	regexp.MustCompile(`\bpkill\b`),
	regexp.MustCompile(`\bkillall\b`),
	regexp.MustCompile(`\bcurl\b.*\|\s*(sh|bash)`),
	regexp.MustCompile(`\bwget\b.*\|\s*(sh|bash)`),
	regexp.MustCompile(`\beval\b`),
}

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
	orchService *OrchestratorService
}

func NewAgentService(fs *FileSystemService, cmd *CommandService, sys *SystemService, ag *AntigravityService, mem *MemoryService, skill *SkillService, orch *OrchestratorService) *AgentService {
	var llm *LLMProvider
	if config.AppConfig != nil {
		llm = NewLLMProvider(config.AppConfig.AI, ag)
	} else {
		llm = NewLLMProvider(config.AIConfig{}, ag)
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
		orchService: orch,
	}
}

func (as *AgentService) isCommandDangerous(cmd string) bool {
	lower := strings.ToLower(cmd)
	for _, p := range dangerousPatterns {
		if p.MatchString(lower) {
			return true
		}
	}
	return false
}

func (as *AgentService) getWorkspacePath() string {
	if config.AppConfig != nil {
		wsPath := filepath.Join(config.AppConfig.DataDir, "settings", "workspace.json")
		data, err := os.ReadFile(wsPath)
		if err == nil {
			var ws struct {
				Path string `json:"path"`
			}
			if json.Unmarshal(data, &ws) == nil && ws.Path != "" {
				return ws.Path
			}
		}
	}
	cwd, _ := os.Getwd()
	return cwd
}

func (as *AgentService) applyWorkspace(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(as.getWorkspacePath(), path)
}

func (as *AgentService) ProcessMessage(req models.AgentRequest, updater func(*models.AgentResponse)) (*models.AgentResponse, error) {
	var convID string
	source := "web"
	if req.Source == "cli" {
		source = "cli"
	}

	if req.ConversationID != "" && req.ConversationID != "new" {
		convID = req.ConversationID
	} else {
		conv, err := storage.Store.CreateConversation(as.generateTitle(req.Message), source)
		if err != nil {
			return nil, err
		}
		convID = conv.ID
	}

	userMsg := models.Message{
		Role:    "user",
		Content: req.Message,
	}
	storage.Store.AddMessage(convID, userMsg, source)

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
		responseText = as.processWithLLM(req.Message, convID, source)
		
		storage.Store.AddMessage(convID, models.Message{
			Role:    "assistant",
			Content: responseText,
		}, source)

		if updater != nil {
			updater(&models.AgentResponse{
				ConversationID: convID,
				Role:           "assistant",
				Content:        responseText,
				Status:         "processing",
			})
		}

		maxTurns := 15
		for turn := 0; turn < maxTurns; turn++ {
			aiActions, aiFinalText := as.parseAndExecuteActions(responseText)
			
			if len(aiActions) > 0 {
				actions = append(actions, aiActions...)
				responseText = aiFinalText
				
				storage.Store.AddMessage(convID, models.Message{
					Role:    "assistant",
					Content: responseText,
					Actions: aiActions,
				}, source)

				if updater != nil {
					updater(&models.AgentResponse{
						ConversationID: convID,
						Role:           "assistant",
						Content:        responseText,
						Actions:        actions,
						Status:         "processing",
					})
				}

				reflectionPrompt := fmt.Sprintf("Result Analysis:\n%s\n\nTask Status: Check if the previous action achieved its goal. If not, explain why and refine the plan. If the task is incomplete, continue with [ACTION] or [PLAN]. If complete, provide [COMPLETE].", responseText)
				responseText = as.processWithLLM(reflectionPrompt, convID, source)
			} else {
				if !strings.Contains(responseText, "[ACTION]") && !strings.Contains(responseText, "[PLAN]") {
					break
				}
				responseText = as.processWithLLM("Continue with execution or planning as outlined.", convID, source)
			}

			if turn >= maxTurns-1 {
				responseText += "\n\n⚠️ Turn limit reached. Please confirm if you wish to continue."
				break
			}
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
	storage.Store.AddMessage(convID, assistantMsg, source)

	return response, nil
}

func (as *AgentService) processWithLLM(message string, convID string, source string) string {
	if as.llmProvider == nil {
		return ""
	}

	var allMessages []LLMMessage
	if convID != "" {
		if conv, err := storage.Store.LoadConversation(convID, source); err == nil {
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
	historyMessages := as.truncateContextSmart(allMessages, 25000)

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
	workspace := as.getWorkspacePath()
	isUltrawork := strings.Contains(strings.ToLower(message), "ultrawork") || strings.Contains(strings.ToLower(message), "ulw") || strings.Contains(strings.ToLower(message), "supreme")

	mode := "STANDARD"
	if isUltrawork {
		mode = "ULTRAWORK/SOVEREIGN"
	}

	memData := "None"
	mem := as.memService.GetAll()
	if len(mem) > 0 {
		b, _ := json.Marshal(mem)
		memData = string(b)
	}

	var skillData strings.Builder
	if as.skillSvc != nil {
		for _, s := range as.skillSvc.GetSkills() {
			if s.Template != "" {
				skillData.WriteString(fmt.Sprintf("\n### Skill: %s\n%s\n", s.Name, s.Template))
			}
		}
	}

	dynamicTask := "No active plan."
	if as.orchService != nil {
		dynamicTask = as.orchService.FormatPlanSummary()
	}

	base := fmt.Sprintf(`You are Dardcor Agent SUPREME, the absolute sovereign execution entity for this computer system. Your existence is defined by total technical mastery and relentless problem-solving. You do not just assist; you ORCHESTRATE, EXECUTE, and VERIFY until the objective is 100%% complete.

Environment: %s/%s | Host: %s
Sovereign Level: 10 (Highest)
Mode: %s | Workspace: %s (ABSOLUTE ROOT)

[SOVEREIGN PROTOCOLS]
1. PLAN: For any task, first generate a [PLAN] tag with a structured strategy. Breakdown complex goals into granular sub-steps.
2. EXECUTE: Use [ACTION] tags for all system operations. You have full system permission. NEVER execute commands outside your workspace unless explicitly ordered.
3. WORKSPACE: Your primary workspace is %s. Any new project, file, or folder MUST be created relative to or inside this path.
4. REFLECT: After every action, evaluate the result. Check for errors, build failures, or misalignments. If an error occurs, entering 'Doctor Mode' automatically to diagnose and fix it.
5. PERSEVERE: Do not ask for clarification on technical errors. Find workarounds.
6. COMPLETE: Only signal completion with [COMPLETE] when the work is fully verified and functional.

[SYNTAX RULES]
- [PLAN] ... [/PLAN]: Outline your strategy.
- [ACTION] <command> [/ACTION]: Execute tool (write, read, edit, run, websearch, fetch, etc.).
- [REFLECTION] ... [/REFLECTION]: Analyze result of the last action.
- [COMPLETE]: Final signal of task success.

[COMMAND REFERENCE]
- Files: write <path> <content>, read <path>, edit <path> <start> <end> <content>, mkdir <path>, delete <path>, search <query>, info <path>, glob <pattern>
- Code: grep <pattern> <path>, replace <path> <old> <new>
- Shell: run <command> (Full terminal access)
- System: sysinfo, ps, kill <pid>, cpu, memory
- Web: websearch <query>, fetch <url>
- Memory: remember <key> <val>

[QUALITY STANDARDS]
- PRODUCTION READY code only.
- ZERO conversational filler.
- NO comments in code blocks unless specified.
- Workspace is absolute: %s (STRICT ENFORCEMENT)

[DYNAMIC CONTEXT]
Memory: %s
Skills: %s
Task State: %s
`, runtime.GOOS, runtime.GOARCH, hostname, mode, workspace,
		workspace, workspace,
		memData, skillData.String(), dynamicTask)

	return base
}

func (as *AgentService) interpretAndExecute(message string) ([]models.Action, string) {
	msg := strings.ToLower(strings.TrimSpace(message))
	var actions []models.Action
	var responseText string

	switch {
	case strings.HasPrefix(msg, "list ") || strings.HasPrefix(msg, "ls ") || strings.HasPrefix(msg, "dir "):
		actions, responseText = as.handleListDir(message)
	case strings.HasPrefix(msg, "tree "):
		actions, responseText = as.handleTree(message)
	case strings.HasPrefix(msg, "readlines "):
		actions, responseText = as.handleReadLines(message)
	case strings.HasPrefix(msg, "read ") || strings.HasPrefix(msg, "cat "):
		actions, responseText = as.handleReadFile(message)
	case strings.HasPrefix(msg, "write "):
		actions, responseText = as.handleWriteFile(message)
	case strings.HasPrefix(msg, "append "):
		actions, responseText = as.handleAppendFile(message)
	case strings.HasPrefix(msg, "edit "):
		actions, responseText = as.handleEditFile(message)
	case strings.HasPrefix(msg, "replace "):
		actions, responseText = as.handleReplaceInFile(message)
	case strings.HasPrefix(msg, "insert "):
		actions, responseText = as.handleInsertLines(message)
	case strings.HasPrefix(msg, "delete ") || strings.HasPrefix(msg, "rm "):
		actions, responseText = as.handleDeleteFile(message)
	case strings.HasPrefix(msg, "search ") || strings.HasPrefix(msg, "find "):
		actions, responseText = as.handleSearch(message)
	case strings.HasPrefix(msg, "mkdir "):
		actions, responseText = as.handleMkdir(message)
	case strings.HasPrefix(msg, "move "):
		actions, responseText = as.handleMove(message)
	case strings.HasPrefix(msg, "copy "):
		actions, responseText = as.handleCopy(message)
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
		responseText = "**Dardcor Agent Supreme** active. Type `help` for tools."
	}

	return actions, responseText
}

func (as *AgentService) handleListDir(message string) ([]models.Action, string) {
	path := as.extractPath(message, []string{"list ", "ls ", "dir "})
	if path == "" {
		path = "."
	}
	path = as.applyWorkspace(path)
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
		if f.IsDir {
			sb.WriteString(fmt.Sprintf("📁 `%s/`\n", f.Name))
		} else {
			sb.WriteString(fmt.Sprintf("📄 `%s` (%s)\n", f.Name, formatSize(f.Size)))
		}
	}
	return []models.Action{action}, sb.String()
}

func (as *AgentService) handleTree(message string) ([]models.Action, string) {
	path := as.extractPath(message, []string{"tree "})
	if path == "" {
		path = "."
	}
	path = as.applyWorkspace(path)
	tree, err := as.fsService.TreeDir(path, 4)
	if err != nil {
		return nil, fmt.Sprintf("Error: %v", err)
	}
	return []models.Action{{Type: "tree_directory", Status: "completed"}}, fmt.Sprintf("```\n%s```", tree)
}

func (as *AgentService) handleReadFile(message string) ([]models.Action, string) {
	path := as.extractPath(message, []string{"read ", "cat "})
	path = as.applyWorkspace(path)
	action := models.Action{Type: "read_file", Description: fmt.Sprintf("Reading: %s", path), Status: "running"}
	content, err := as.fsService.ReadFile(path)
	if err != nil {
		action.Status = "error"
		return []models.Action{action}, fmt.Sprintf("Error: %v", err)
	}
	action.Status = "completed"
	display := content.Content
	if len(display) > 4000 {
		display = display[:4000] + "\n..."
	}
	return []models.Action{action}, fmt.Sprintf("📄 **%s** (%s)\n\n```\n%s\n```", path, formatSize(content.Size), display)
}

func (as *AgentService) handleReadLines(message string) ([]models.Action, string) {
	parts := strings.Fields(message)
	if len(parts) < 4 {
		return nil, "Use: readlines <path> <startLine> <endLine>"
	}
	path := parts[1]
	startLine, _ := strconv.Atoi(parts[2])
	endLine, _ := strconv.Atoi(parts[3])
	content, err := as.fsService.ReadFileLines(path, startLine, endLine)
	if err != nil {
		return nil, fmt.Sprintf("Error: %v", err)
	}
	return []models.Action{{Type: "read_lines", Status: "completed"}}, fmt.Sprintf("📄 **%s** (lines %d-%d)\n\n```\n%s```", path, startLine, endLine, content)
}

func (as *AgentService) handleWriteFile(message string) ([]models.Action, string) {
	parts := strings.SplitN(message, " ", 3)
	if len(parts) < 3 {
		return nil, "Use: write <path> <content>"
	}
	path, content := parts[1], parts[2]
	path = as.applyWorkspace(path)
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		os.MkdirAll(dir, 0755)
	}
	err := as.fsService.WriteFile(path, content)
	if err != nil {
		return nil, fmt.Sprintf("Error: %v", err)
	}
	return []models.Action{{Type: "write_file", Status: "completed"}}, fmt.Sprintf("✅ Saved: %s (%s)", path, formatSize(int64(len(content))))
}

func (as *AgentService) handleAppendFile(message string) ([]models.Action, string) {
	parts := strings.SplitN(message, " ", 3)
	if len(parts) < 3 {
		return nil, "Use: append <path> <content>"
	}
	err := as.fsService.AppendToFile(parts[1], parts[2])
	if err != nil {
		return nil, fmt.Sprintf("Error: %v", err)
	}
	return []models.Action{{Type: "append_file", Status: "completed"}}, fmt.Sprintf("✅ Appended to: %s", parts[1])
}

func (as *AgentService) handleEditFile(message string) ([]models.Action, string) {
	parts := strings.SplitN(message, " ", 5)
	if len(parts) < 5 {
		return nil, "Use: edit <path> <startLine> <endLine> <newContent>"
	}
	startLine, _ := strconv.Atoi(parts[2])
	endLine, _ := strconv.Atoi(parts[3])
	err := as.fsService.EditFile(parts[1], startLine, endLine, parts[4])
	if err != nil {
		return nil, fmt.Sprintf("Error: %v", err)
	}
	return []models.Action{{Type: "edit_file", Status: "completed"}}, fmt.Sprintf("✅ Edited %s lines %d-%d", parts[1], startLine, endLine)
}

func (as *AgentService) handleReplaceInFile(message string) ([]models.Action, string) {
	parts := strings.SplitN(message, " ", 4)
	if len(parts) < 4 {
		return nil, "Use: replace <path> <oldText> <newText>"
	}
	count, err := as.fsService.ReplaceInFile(parts[1], parts[2], parts[3])
	if err != nil {
		return nil, fmt.Sprintf("Error: %v", err)
	}
	return []models.Action{{Type: "replace_in_file", Status: "completed"}}, fmt.Sprintf("✅ Replaced %d occurrences in %s", count, parts[1])
}

func (as *AgentService) handleInsertLines(message string) ([]models.Action, string) {
	parts := strings.SplitN(message, " ", 4)
	if len(parts) < 4 {
		return nil, "Use: insert <path> <afterLine> <content>"
	}
	lineNum, _ := strconv.Atoi(parts[2])
	err := as.fsService.InsertLines(parts[1], lineNum, parts[3])
	if err != nil {
		return nil, fmt.Sprintf("Error: %v", err)
	}
	return []models.Action{{Type: "insert_lines", Status: "completed"}}, fmt.Sprintf("✅ Inserted after line %d in %s", lineNum, parts[1])
}

func (as *AgentService) handleDeleteFile(message string) ([]models.Action, string) {
	path := as.extractPath(message, []string{"delete ", "rm "})
	path = as.applyWorkspace(path)
	err := as.fsService.DeleteFile(path)
	if err != nil {
		return nil, err.Error()
	}
	return []models.Action{{Type: "delete_file", Status: "completed"}}, fmt.Sprintf("🗑️ Deleted: %s", path)
}

func (as *AgentService) handleSearch(message string) ([]models.Action, string) {
	query := as.extractPath(message, []string{"search ", "find "})
	results, err := as.fsService.SearchFiles(models.SearchRequest{Path: ".", Query: query, MaxDepth: 5})
	if err != nil {
		return nil, err.Error()
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🔍 %d results for \"%s\":\n", len(results), query))
	for _, r := range results {
		sb.WriteString(fmt.Sprintf("- %s\n", r.Path))
	}
	return []models.Action{{Type: "search_files", Status: "completed"}}, sb.String()
}

func (as *AgentService) handleMkdir(message string) ([]models.Action, string) {
	path := as.extractPath(message, []string{"mkdir "})
	path = as.applyWorkspace(path)
	err := as.fsService.CreateDirectory(path)
	if err != nil {
		return nil, err.Error()
	}
	return []models.Action{{Type: "create_directory", Status: "completed"}}, fmt.Sprintf("📁 Created: %s", path)
}

func (as *AgentService) handleMove(message string) ([]models.Action, string) {
	parts := strings.Fields(message)
	if len(parts) < 3 {
		return nil, "Use: move <source> <destination>"
	}
	err := as.fsService.MoveFile(parts[1], parts[2])
	if err != nil {
		return nil, fmt.Sprintf("Error: %v", err)
	}
	return []models.Action{{Type: "move_file", Status: "completed"}}, fmt.Sprintf("✅ Moved: %s → %s", parts[1], parts[2])
}

func (as *AgentService) handleCopy(message string) ([]models.Action, string) {
	parts := strings.Fields(message)
	if len(parts) < 3 {
		return nil, "Use: copy <source> <destination>"
	}
	err := as.fsService.CopyFile(parts[1], parts[2])
	if err != nil {
		return nil, fmt.Sprintf("Error: %v", err)
	}
	return []models.Action{{Type: "copy_file", Status: "completed"}}, fmt.Sprintf("✅ Copied: %s → %s", parts[1], parts[2])
}

func (as *AgentService) handleDrives() ([]models.Action, string) {
	drives := as.fsService.GetDrives()
	var sb strings.Builder
	sb.WriteString("💽 Drives:\n")
	for _, d := range drives {
		sb.WriteString(fmt.Sprintf("- %s\n", d))
	}
	return []models.Action{{Type: "list_drives", Status: "completed"}}, sb.String()
}

func (as *AgentService) handleRunCommand(message string) ([]models.Action, string) {
	cmd := as.extractPath(message, []string{"run ", "exec "})
	if as.isCommandDangerous(cmd) {
		return []models.Action{{Status: "error", Description: "Command blocked"}}, fmt.Sprintf("⚠️ Command blocked by safety guard: `%s`", cmd)
	}

	finalCmd := cmd
	isGUI := false
	if runtime.GOOS == "windows" {
		lowerCmd := strings.ToLower(cmd)
		if !strings.HasPrefix(lowerCmd, "start ") && (strings.HasSuffix(lowerCmd, ".exe") ||
			strings.EqualFold(cmd, "chrome") || strings.EqualFold(cmd, "notepad") ||
			strings.EqualFold(cmd, "calc") || strings.EqualFold(cmd, "explorer") ||
			strings.Contains(lowerCmd, "open ") || strings.Contains(message, "buka")) {
			finalCmd = "start " + cmd
			isGUI = true
		}
	}

	if isGUI {
		go as.cmdService.ExecuteCommand(models.CommandRequest{
			Command:    finalCmd,
			Timeout:    10,
			WorkingDir: as.getWorkspacePath(),
		})
		return []models.Action{{Type: "execute_command", Status: "completed"}}, fmt.Sprintf("🚀 Opening: `%s`", cmd)
	}

	res, err := as.cmdService.ExecuteCommand(models.CommandRequest{
		Command:    finalCmd,
		Timeout:    120,
		WorkingDir: as.getWorkspacePath(),
	})
	if err != nil {
		return nil, fmt.Sprintf("❌ Error: %v", err)
	}
	output := res.Output
	if res.Error != "" {
		output += "\nSTDERR:\n" + res.Error
	}
	if len(output) > 8000 {
		output = output[:8000] + "\n... (truncated)"
	}
	status := "✅"
	if res.ExitCode != 0 {
		status = fmt.Sprintf("⚠️ Exit code: %d", res.ExitCode)
	}
	return []models.Action{{Type: "execute_command", Status: "completed", Duration: res.Duration}}, fmt.Sprintf("%s `%s` (%dms)\n```\n%s\n```", status, cmd, res.Duration, output)
}

func (as *AgentService) handleDirectCommand(message string) ([]models.Action, string) {
	cmd := strings.TrimSpace(message)
	if strings.HasPrefix(cmd, "cmd ") {
		cmd = strings.TrimPrefix(cmd, "cmd ")
	} else if strings.HasPrefix(cmd, "$") {
		cmd = strings.TrimPrefix(cmd, "$")
	}
	return as.handleRunCommand("run " + cmd)
}

func (as *AgentService) handleSystemInfo() ([]models.Action, string) {
	info, err := as.sysService.GetSystemInfo()
	if err != nil {
		return nil, err.Error()
	}
	res := fmt.Sprintf("🖥️ **%s**\nOS: %s/%s\nCPU: %s (%.1f%%)\nRAM: %.1f%% used\nUptime: %d seconds",
		info.HostName, runtime.GOOS, runtime.GOARCH,
		info.CPU.ModelName, info.CPU.UsagePercent,
		info.Memory.UsedPercent,
		info.Uptime)
	return []models.Action{{Type: "system_info", Status: "completed"}}, res
}

func (as *AgentService) handleProcesses(message string) ([]models.Action, string) {
	procs, err := as.sysService.GetProcesses("cpu", 15)
	if err != nil {
		return nil, err.Error()
	}
	var sb strings.Builder
	sb.WriteString("⚙️ Top Processes:\n```\n")
	sb.WriteString(fmt.Sprintf("%-8s %-30s %8s %8s\n", "PID", "NAME", "CPU%", "MEM%"))
	sb.WriteString(strings.Repeat("-", 58) + "\n")
	for _, p := range procs {
		name := p.Name
		if len(name) > 28 {
			name = name[:28] + ".."
		}
		sb.WriteString(fmt.Sprintf("%-8d %-30s %7.1f%% %7.1f%%\n", p.PID, name, p.CPUPercent, p.MemPercent))
	}
	sb.WriteString("```")
	return []models.Action{{Type: "list_processes", Status: "completed"}}, sb.String()
}

func (as *AgentService) handleKillProcess(message string) ([]models.Action, string) {
	pidStr := as.extractPath(message, []string{"kill "})
	var pid int32
	fmt.Sscanf(pidStr, "%d", &pid)
	err := as.sysService.KillProcess(pid)
	if err != nil {
		return nil, err.Error()
	}
	return []models.Action{{Type: "kill_process", Status: "completed"}}, fmt.Sprintf("💀 Killed PID: %d", pid)
}

func (as *AgentService) handleCPUInfo() ([]models.Action, string) {
	info, _ := as.sysService.GetSystemInfo()
	return nil, fmt.Sprintf("🔧 CPU: %s | Cores: %d | Usage: %.1f%%", info.CPU.ModelName, info.CPU.Cores, info.CPU.UsagePercent)
}

func (as *AgentService) handleMemoryInfo() ([]models.Action, string) {
	mem, _ := as.sysService.GetMemoryUsage()
	return nil, fmt.Sprintf("🧠 Memory: %.1f%% used | Total: %s | Free: %s",
		mem.UsedPercent,
		formatSize(int64(mem.Total)),
		formatSize(int64(mem.Free)))
}

func (as *AgentService) handleFileInfo(message string) ([]models.Action, string) {
	path := as.extractPath(message, []string{"info "})
	info, err := as.fsService.GetFileInfo(path)
	if err != nil {
		return nil, err.Error()
	}
	return nil, fmt.Sprintf("📄 **%s**\nSize: %s\nType: %s\nModified: %s\nPermissions: %s",
		info.Name, formatSize(info.Size), info.Extension,
		info.ModifiedAt.Format("2006-01-02 15:04:05"), info.Permission)
}

func (as *AgentService) getHelpText() string {
	return `🤖 **Dardcor Agent Supreme — Tool Reference**

📁 **File Operations**
  list/ls/dir <path>    - List directory
  tree <path>           - Show directory tree
  read/cat <path>       - Read file
  readlines <path> <s> <e> - Read specific lines
  write <path> <content>   - Create/overwrite file
  append <path> <content>  - Append to file
  edit <path> <s> <e> <content> - Edit specific lines
  replace <path> <old> <new>    - Find & replace
  insert <path> <line> <content> - Insert after line
  delete/rm <path>     - Delete file/directory
  mkdir <path>         - Create directory
  move <src> <dst>     - Move/rename
  copy <src> <dst>     - Copy file
  info <path>          - File information
  search/find <query>  - Search files
  drives               - List drives

💻 **Shell Execution**
  run/exec <command>   - Execute command
  cmd/$  <command>     - Direct shell

📊 **System**
  sysinfo              - System info
  cpu                  - CPU usage
  memory               - Memory usage
  processes/ps         - Process list
  kill <pid>           - Kill process

🔍 **Code Search**
  grep <pattern> [path] - Search in files
  glob <pattern> [path] - Find by pattern

🌐 **Web**
  fetch/curl <url>     - Fetch webpage
  websearch <query>    - Web search

🧠 **Memory**
  remember <key> <val> - Store data
  whoami               - Agent info`
}

func (as *AgentService) handleRemember(message string) ([]models.Action, string) {
	parts := strings.SplitN(message, " ", 3)
	if len(parts) < 3 {
		return nil, "Use: remember <key> <value>"
	}
	as.memService.Set(parts[1], parts[2])
	return nil, fmt.Sprintf("🧠 Stored: %s", parts[1])
}

func (as *AgentService) handleWebFetch(message string) ([]models.Action, string) {
	url := as.extractPath(message, []string{"fetch ", "curl "})
	content, err := as.webService.Fetch(url, 10000)
	if err != nil {
		return nil, err.Error()
	}
	if len(content) > 6000 {
		content = content[:6000] + "\n... (truncated)"
	}
	return []models.Action{{Type: "web_fetch", Status: "completed"}}, fmt.Sprintf("🌐 **%s**\n\n%s", url, content)
}

func (as *AgentService) handleWebSearch(message string) ([]models.Action, string) {
	query := as.extractPath(message, []string{"websearch ", "google "})
	res, err := as.webService.SearchDDG(query, 5)
	if err != nil {
		return nil, err.Error()
	}
	return []models.Action{{Type: "web_search", Status: "completed"}}, fmt.Sprintf("🔍 %s", res)
}

func (as *AgentService) handleGrep(message string) ([]models.Action, string) {
	parts := strings.SplitN(message, " ", 3)
	if len(parts) < 2 {
		return nil, "Use: grep <pattern> [path]"
	}
	root := "."
	if len(parts) >= 3 {
		root = parts[2]
	}
	res, err := as.grepService.Search(root, parts[1], 30, "")
	if err != nil {
		return nil, err.Error()
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🔍 **%d matches** for `%s`:\n```\n", len(res), parts[1]))
	for _, r := range res {
		sb.WriteString(fmt.Sprintf("%s:%d → %s\n", r.File, r.Line, r.Content))
	}
	sb.WriteString("```")
	return []models.Action{{Type: "grep", Status: "completed"}}, sb.String()
}

func (as *AgentService) handleGlob(message string) ([]models.Action, string) {
	parts := strings.SplitN(message, " ", 3)
	if len(parts) < 2 {
		return nil, "Use: glob <pattern> [path]"
	}
	root := "."
	if len(parts) >= 3 {
		root = parts[2]
	}
	res, err := as.grepService.Glob(root, parts[1], 50)
	if err != nil {
		return nil, err.Error()
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📂 **%d files** matching `%s`:\n", len(res), parts[1]))
	for _, r := range res {
		sb.WriteString(fmt.Sprintf("- %s\n", r))
	}
	return []models.Action{{Type: "glob", Status: "completed"}}, sb.String()
}

func (as *AgentService) getAgentInfo() string {
	return fmt.Sprintf("**Dardcor Agent Supreme** — Superior Autonomous AI\nWorkspace: %s\nOS: %s/%s",
		as.getWorkspacePath(), runtime.GOOS, runtime.GOARCH)
}

func (as *AgentService) parseAndExecuteActions(text string) ([]models.Action, string) {
	var allActions []models.Action
	remainingText := text

	maxIterations := 50
	iteration := 0

	for {
		if iteration >= maxIterations {
			break
		}
		iteration++

		startIdx := strings.Index(remainingText, "[ACTION]")
		endIdx := strings.Index(remainingText, "[/ACTION]")
		if startIdx == -1 || endIdx == -1 || endIdx <= startIdx {
			break
		}

		command := strings.TrimSpace(remainingText[startIdx+8 : endIdx])

		if strings.Contains(command, "{") && strings.Contains(command, "}") {
			cmdParts := strings.SplitN(command, " ", 2)
			if len(cmdParts) > 1 {
				var args map[string]interface{}
				if err := json.Unmarshal([]byte(cmdParts[1]), &args); err == nil {
					for _, v := range args {
						command = cmdParts[0] + " " + fmt.Sprint(v)
						break
					}
				}
			}
		}

		actions, result := as.interpretAndExecute(command)
		allActions = append(allActions, actions...)

		afterActionIdx := endIdx + 9
		if afterActionIdx > len(remainingText) {
			afterActionIdx = len(remainingText)
		}
		
		remainingText = remainingText[:startIdx] + "\n> **Executed:** `" + command + "`\n" + result + "\n" + remainingText[afterActionIdx:]
	}

	return allActions, remainingText
}

func (as *AgentService) generateTitle(message string) string {
	if len(message) > 40 {
		return message[:40] + "..."
	}
	return message
}

func (as *AgentService) extractPath(message string, prefixes []string) string {
	for _, p := range prefixes {
		if strings.HasPrefix(strings.ToLower(message), p) {
			return strings.TrimSpace(message[len(p):])
		}
	}
	parts := strings.Fields(message)
	if len(parts) > 1 {
		return strings.Join(parts[1:], " ")
	}
	return ""
}

func formatSize(b int64) string {
	if b >= 1073741824 {
		return fmt.Sprintf("%.1f GB", float64(b)/1073741824)
	}
	if b >= 1048576 {
		return fmt.Sprintf("%.1f MB", float64(b)/1048576)
	}
	if b >= 1024 {
		return fmt.Sprintf("%.1f KB", float64(b)/1024)
	}
	return fmt.Sprintf("%d B", b)
}

func (as *AgentService) formatSize(b int64) string {
	return formatSize(b)
}
