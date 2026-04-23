package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

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
	fsService      *FileSystemService
	cmdService     *CommandService
	sysService     *SystemService
	agService      *AntigravityService
	llmProvider    *LLMProvider
	memService     *MemoryService
	webService     *WebService
	grepService    *GrepService
	skillSvc       *SkillService
	orchService    *OrchestratorService
	egoService     *EgoService
	reflectSvc     *ReflectionService
	browserSvc     *BrowserService
	visionSvc      *VisionService
	autoSvc        *AutomationService
	thinkSvc       *ThinkModeService
	redactSvc      *RedactService
	rateLimiterSvc *RateLimiterService
	titleGenSvc    *TitleGeneratorService
	fileSafetySvc  *FileSafetyService
	promptCacheSvc *PromptCacheService
	compressorSvc  *ContextCompressorService
	modelMetaSvc   *ModelMetadataService
}

func NewAgentService(fs *FileSystemService, cmd *CommandService, sys *SystemService, ag *AntigravityService, mem *MemoryService, skill *SkillService, orch *OrchestratorService, ego *EgoService, reflect *ReflectionService, browser *BrowserService, vision *VisionService, auto *AutomationService) *AgentService {
	var llm *LLMProvider
	if config.AppConfig != nil {
		llm = NewLLMProvider(config.AppConfig.AI, ag)
	} else {
		llm = NewLLMProvider(config.AIConfig{}, ag)
	}
	llm.rateLimiter = NewRateLimiterService()
	llm.promptCache = NewPromptCacheService()

	return &AgentService{
		fsService:      fs,
		cmdService:     cmd,
		sysService:     sys,
		agService:      ag,
		llmProvider:    llm,
		memService:     mem,
		webService:     NewWebService(),
		grepService:    NewGrepService(),
		skillSvc:       skill,
		orchService:    orch,
		egoService:     ego,
		reflectSvc:     reflect,
		browserSvc:     browser,
		visionSvc:      vision,
		autoSvc:        auto,
		thinkSvc:       NewThinkModeService(),
		redactSvc:      NewRedactService(),
		rateLimiterSvc: NewRateLimiterService(),
		titleGenSvc:    NewTitleGeneratorService(llm),
		fileSafetySvc:  NewFileSafetyService(),
		promptCacheSvc: NewPromptCacheService(),
		compressorSvc: NewContextCompressorService(llm, func() string {
			if config.AppConfig != nil {
				return config.AppConfig.AI.Model
			}
			return ""
		}(), 128000),
		modelMetaSvc: NewModelMetadataService(),
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

func (as *AgentService) isReadOnlyAction(command string) bool {
	readOnlyPrefixes := []string{
		"read ", "cat ", "list ", "ls ", "dir ", "tree ",
		"search ", "find ", "grep ", "glob ", "sysinfo",
		"info ", "fetch ", "curl ", "websearch ", "google ",
		"processes", "ps", "cpu", "memory", "drives", "whoami", "help",
	}
	lower := strings.ToLower(strings.TrimSpace(command))
	for _, prefix := range readOnlyPrefixes {
		if strings.HasPrefix(lower, prefix) || lower == strings.TrimSpace(prefix) {
			return true
		}
	}
	if strings.HasPrefix(strings.TrimSpace(command), "{") {
		var jsonCall map[string]interface{}
		if err := json.Unmarshal([]byte(strings.TrimSpace(command)), &jsonCall); err == nil {
			toolName := ""
			if t, ok := jsonCall["tool"].(string); ok {
				toolName = t
			} else if t, ok := jsonCall["name"].(string); ok {
				toolName = t
			}
			readOnlyTools := map[string]bool{
				"read": true, "list": true, "ls": true, "tree": true,
				"search": true, "grep": true, "glob": true, "sysinfo": true,
				"info": true, "fetch": true, "websearch": true,
				"browser_screenshot": true, "browser_get_dom": true, "os_observe": true,
			}
			return readOnlyTools[toolName]
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
				os.MkdirAll(ws.Path, 0755)
				return ws.Path
			}
		}
	}

	if defaultWs, err := as.fsService.GetDefaultWorkspace(); err == nil && defaultWs != "" {
		return defaultWs
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

func parseQuotedArgs(input string) []string {
	var args []string
	var current strings.Builder
	inQuote := false
	quoteChar := byte(0)

	for i := 0; i < len(input); i++ {
		ch := input[i]
		if inQuote {
			if ch == quoteChar {
				inQuote = false
			} else {
				current.WriteByte(ch)
			}
		} else {
			if ch == '"' || ch == '\'' {
				inQuote = true
				quoteChar = ch
			} else if ch == ' ' || ch == '\t' {
				if current.Len() > 0 {
					args = append(args, current.String())
					current.Reset()
				}
			} else {
				current.WriteByte(ch)
			}
		}
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	return args
}

func (as *AgentService) ProcessMessage(ctx context.Context, req models.AgentRequest, updater func(*models.AgentResponse)) (*models.AgentResponse, error) {
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
	if err := storage.Store.AddMessage(convID, userMsg, source); err != nil {
		conv, newErr := storage.Store.CreateConversation(as.generateTitle(req.Message), source)
		if newErr == nil {
			convID = conv.ID
			storage.Store.AddMessage(convID, userMsg, source)
		}
	}

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
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		var builder strings.Builder
		responseText = as.processWithLLMStream(req.Message, convID, source, func(chunk string) {
			builder.WriteString(chunk)
			if updater != nil {
				updater(&models.AgentResponse{
					ConversationID: convID,
					Role:           "assistant",
					Content:        chunk,
					Status:         "stream_chunk",
				})
			}
		})

		storage.Store.UpsertLastAssistantMessage(convID, responseText, nil, source)

		if updater != nil {
			updater(&models.AgentResponse{
				ConversationID: convID,
				Role:           "assistant",
				Content:        responseText,
				Status:         "processing",
			})
		}

		isConversational := !strings.Contains(responseText, "[PLAN]") && !strings.Contains(responseText, "[ACTION]") && !strings.Contains(responseText, "[COMPLETE]")

		maxTurns := 25
		if isConversational {
			maxTurns = 0
		}

		if strings.Contains(responseText, "[PLAN]") {
			goal, subtasks := as.parsePlanFromResponse(responseText)
			if goal != "" && len(subtasks) > 0 {
				as.orchService.InitializePlan(goal, subtasks)
			}
		}

		errorCount := 0

		for turn := 0; turn < maxTurns; turn++ {
			if ctx.Err() != nil {
				responseText += "\n\n⚠️ Agent loop interrupted."
				break
			}

			aiActions, aiFinalText := as.parseAndExecuteActions(responseText)

			if len(aiActions) > 0 {
				actions = append(actions, aiActions...)
				responseText = aiFinalText

				hasError := strings.Contains(responseText, "Error") ||
					strings.Contains(responseText, "error") ||
					strings.Contains(responseText, "❌")
				if hasError {
					errorCount++
				} else {
					errorCount = 0
				}

				if updater != nil {
					updater(&models.AgentResponse{
						ConversationID: convID,
						Role:           "assistant",
						Content:        responseText,
						Actions:        actions,
						Status:         "processing",
					})
				}

				if strings.Contains(responseText, "[COMPLETE]") {
					if plan := as.orchService.GetCurrentPlan(); plan != nil {
						as.orchService.SetPhase(PhaseVerify)
					}
					break
				}

				if ctx.Err() != nil {
					responseText += "\n\n⚠️ Agent loop interrupted."
					break
				}

				var failureWarning string
				if errorCount >= 3 {
					failureWarning = "\n\n⚠️ FAILURE LOOP DETECTED: 3 consecutive errors. Stop and reassess. Explain what's going wrong and propose a different approach before continuing."
				}

				reflectionPrompt := fmt.Sprintf(`[TURN %d/%d] Action results from previous step:
%s%s

[REFLECTION PROTOCOL]
1. Did the action succeed or fail? If failed, what went wrong?
2. Are there any errors I should fix before proceeding?
3. What is the next logical step toward the objective?
4. Am I making progress or stuck in a loop?

If objective is fully achieved: respond with [COMPLETE] and a summary.
If more work needed: respond with the next [ACTION] block.
If stuck: explain the blocker and attempt an alternative approach.`, turn+1, maxTurns, responseText, failureWarning)

				var innerBuilder strings.Builder
				responseText = as.processWithLLMStream(reflectionPrompt, convID, source, func(chunk string) {
					innerBuilder.WriteString(chunk)
					if updater != nil {
						updater(&models.AgentResponse{
							ConversationID: convID,
							Role:           "assistant",
							Content:        chunk,
							Status:         "stream_chunk",
						})
					}
				})
				storage.Store.UpsertLastAssistantMessage(convID, responseText, actions, source)
			} else {
				if !strings.Contains(responseText, "[ACTION]") && !strings.Contains(responseText, "[PLAN]") {
					break
				}

				if ctx.Err() != nil {
					responseText += "\n\n⚠️ Agent loop interrupted."
					break
				}

				var innerBuilder2 strings.Builder
				responseText = as.processWithLLMStream("Continue execution.", convID, source, func(chunk string) {
					innerBuilder2.WriteString(chunk)
					if updater != nil {
						updater(&models.AgentResponse{
							ConversationID: convID,
							Role:           "assistant",
							Content:        chunk,
							Status:         "stream_chunk",
						})
					}
				})
				storage.Store.UpsertLastAssistantMessage(convID, responseText, actions, source)
			}

			if turn >= maxTurns-1 {
				responseText += "\n\n⚠️ Turn limit reached."
				break
			}
		}

		if len(actions) > 0 {
			outcome := truncateRunes(responseText, 200, "...")
			as.memService.RecordEpisode(convID, req.Message, outcome, "")
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

	if as.egoService != nil {
		success := true
		for _, action := range actions {
			if action.Status == "error" {
				success = false
				break
			}
		}
		as.egoService.RecordTaskResult(success)
	}

	storage.Store.AddMessage(convID, models.Message{
		Role:    "assistant",
		Content: responseText,
		Actions: actions,
	}, source)

	return response, nil
}

func (as *AgentService) processWithLLMStream(message string, convID string, source string, cb StreamCallback) string {
	if as.llmProvider == nil {
		return ""
	}

	message = as.thinkSvc.AugmentPrompt(convID, message)

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
	maxTokens := 25000
	if as.modelMetaSvc != nil && config.AppConfig != nil {
		ctxLen := as.modelMetaSvc.GetModelContextLength(config.AppConfig.AI.Model)
		if ctxLen > 0 {
			maxTokens = ctxLen / 4
		}
	}
	if as.compressorSvc != nil {
		asMaps := llmMessagesToMaps(allMessages)
		estimatedTokens := as.compressorSvc.EstimateTokens(asMaps)
		if as.compressorSvc.ShouldCompress(estimatedTokens) {
			compressed, err := as.compressorSvc.Compress(asMaps)
			if err == nil {
				allMessages = mapsToLLMMessages(compressed)
			}
		}
	}
	historyMessages := as.truncateContextSmart(allMessages, maxTokens)

	resp, err := as.llmProvider.CompleteStream(systemPrompt, historyMessages, cb)
	if err != nil {
		return fmt.Sprintf("⚠️ AI Error: %v", err)
	}

	return resp.Content
}

func (as *AgentService) processWithLLM(message string, convID string, source string) string {
	if as.llmProvider == nil {
		return ""
	}

	message = as.thinkSvc.AugmentPrompt(convID, message)

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
	maxTokens := 25000
	if as.modelMetaSvc != nil && config.AppConfig != nil {
		ctxLen := as.modelMetaSvc.GetModelContextLength(config.AppConfig.AI.Model)
		if ctxLen > 0 {
			maxTokens = ctxLen / 4
		}
	}
	if as.compressorSvc != nil {
		asMaps := llmMessagesToMaps(allMessages)
		estimatedTokens := as.compressorSvc.EstimateTokens(asMaps)
		if as.compressorSvc.ShouldCompress(estimatedTokens) {
			compressed, err := as.compressorSvc.Compress(asMaps)
			if err == nil {
				allMessages = mapsToLLMMessages(compressed)
			}
		}
	}
	historyMessages := as.truncateContextSmart(allMessages, maxTokens)

	resp, err := as.llmProvider.Complete(systemPrompt, historyMessages)
	if err != nil {
		return fmt.Sprintf("⚠️ AI Error: %v", err)
	}

	return resp.Content
}

func (as *AgentService) truncateContextSmart(messages []LLMMessage, maxTokens int) []LLMMessage {
	if len(messages) <= 3 {
		return messages
	}

	estimateTokens := func(s string) int {
		return len(s) / 3
	}

	totalTokens := 0
	for _, m := range messages {
		totalTokens += estimateTokens(m.Content)
	}

	if totalTokens <= maxTokens {
		return messages
	}

	keepFirst := 1
	keepLast := 5
	if len(messages) <= keepFirst+keepLast {
		return messages
	}

	first := messages[:keepFirst]
	middle := messages[keepFirst : len(messages)-keepLast]
	last := messages[len(messages)-keepLast:]

	var middleSummary strings.Builder
	middleSummary.WriteString("Previous conversation context (condensed):\n")
	for _, m := range middle {
		content := m.Content
		if len(content) > 200 {
			content = content[:200] + "..."
		}
		if idx := strings.Index(content, "[ACTION]"); idx > 0 {
			content = content[:idx] + "[actions executed]"
		}
		middleSummary.WriteString(fmt.Sprintf("[%s]: %s\n", m.Role, content))
	}

	summaryMsg := LLMMessage{
		Role:    "system",
		Content: middleSummary.String(),
	}

	if as.llmProvider != nil && len(middle) > 3 {
		summary, err := as.llmProvider.Summarize(middle)
		if err == nil && summary != "" {
			summaryMsg.Content = "Previous conversation summary: " + summary
		}
	}

	result := make([]LLMMessage, 0, keepFirst+1+keepLast)
	result = append(result, first...)
	result = append(result, summaryMsg)
	result = append(result, last...)

	return result
}

func (as *AgentService) buildToolSchemas() string {
	schemas := []map[string]interface{}{
		{
			"name":        "write",
			"description": "Create or overwrite a file with content",
			"parameters": map[string]interface{}{
				"path":    "string - absolute or relative file path",
				"content": "string - file content to write",
			},
		},
		{
			"name":        "read",
			"description": "Read contents of a file",
			"parameters": map[string]interface{}{
				"path": "string - file path to read",
			},
		},
		{
			"name":        "edit",
			"description": "Edit specific lines in a file",
			"parameters": map[string]interface{}{
				"path":       "string - file path",
				"start_line": "int - start line number",
				"end_line":   "int - end line number",
				"content":    "string - replacement content",
			},
		},
		{
			"name":        "replace",
			"description": "Find and replace text in a file",
			"parameters": map[string]interface{}{
				"path":     "string - file path",
				"old_text": "string - text to find",
				"new_text": "string - replacement text",
			},
		},
		{
			"name":        "mkdir",
			"description": "Create a directory",
			"parameters": map[string]interface{}{
				"path": "string - directory path to create",
			},
		},
		{
			"name":        "delete",
			"description": "Delete a file or directory",
			"parameters": map[string]interface{}{
				"path": "string - path to delete",
			},
		},
		{
			"name":        "search",
			"description": "Search for files by name or content",
			"parameters": map[string]interface{}{
				"query": "string - search query",
			},
		},
		{
			"name":        "grep",
			"description": "Search file contents using regex pattern",
			"parameters": map[string]interface{}{
				"pattern": "string - regex pattern",
				"path":    "string - directory path (optional, defaults to .)",
			},
		},
		{
			"name":        "glob",
			"description": "Find files by name pattern",
			"parameters": map[string]interface{}{
				"pattern": "string - glob pattern (e.g. *.go)",
				"path":    "string - root directory (optional)",
			},
		},
		{
			"name":        "run",
			"description": "Execute a shell command",
			"parameters": map[string]interface{}{
				"command": "string - shell command to execute",
			},
		},
		{
			"name":        "sysinfo",
			"description": "Get system information (OS, CPU, memory)",
			"parameters":  map[string]interface{}{},
		},
		{
			"name":        "websearch",
			"description": "Search the web via DuckDuckGo",
			"parameters": map[string]interface{}{
				"query": "string - search query",
			},
		},
		{
			"name":        "fetch",
			"description": "Fetch content from a URL",
			"parameters": map[string]interface{}{
				"url": "string - URL to fetch",
			},
		},
		{
			"name":        "remember",
			"description": "Store a key-value pair in long-term memory",
			"parameters": map[string]interface{}{
				"key":   "string - memory key",
				"value": "string - value to store",
			},
		},
		{
			"name":        "list",
			"description": "List files in a directory",
			"parameters": map[string]interface{}{
				"path": "string - directory path",
			},
		},
		{
			"name":        "tree",
			"description": "Show directory tree structure",
			"parameters": map[string]interface{}{
				"path": "string - directory path",
			},
		},
		{
			"name":        "info",
			"description": "Get file or directory metadata",
			"parameters": map[string]interface{}{
				"path": "string - file or directory path",
			},
		},
		{
			"name":        "kill",
			"description": "Kill a process by PID",
			"parameters": map[string]interface{}{
				"pid": "int - process ID to kill",
			},
		},
		{
			"name":        "browser_open",
			"description": "Open a URL in a controlled SPECIAL browser instance (Chromium)",
			"parameters": map[string]interface{}{
				"url": "string - URL to open",
			},
		},
		{
			"name":        "browser_system_open",
			"description": "Open a URL in the user's CURRENT system default browser (Chrome/Edge/etc)",
			"parameters": map[string]interface{}{
				"url": "string - URL to open",
			},
		},
		{
			"name":        "browser_click",
			"description": "Click an element in the browser",
			"parameters": map[string]interface{}{
				"selector": "string - CSS selector to click",
			},
		},
		{
			"name":        "browser_type",
			"description": "Type text into a browser element",
			"parameters": map[string]interface{}{
				"selector": "string - CSS selector",
				"text":     "string - text to type",
			},
		},
		{
			"name":        "browser_screenshot",
			"description": "Take a screenshot of the current page",
			"parameters":  map[string]interface{}{},
		},
		{
			"name":        "browser_close",
			"description": "Close the controlled browser instance",
			"parameters":  map[string]interface{}{},
		},
		{
			"name":        "browser_scroll",
			"description": "Scroll the current page",
			"parameters": map[string]interface{}{
				"direction": "string - 'down' or 'up'",
			},
		},
		{
			"name":        "browser_wait",
			"description": "Wait for a specified duration",
			"parameters": map[string]interface{}{
				"ms": "int - milliseconds to wait",
			},
		},
		{
			"name":        "browser_back",
			"description": "Go back in the browser history",
			"parameters":  map[string]interface{}{},
		},
		{
			"name":        "browser_get_dom",
			"description": "Get the current page's HTML structure for inspection",
			"parameters":  map[string]interface{}{},
		},
		{
			"name":        "os_observe",
			"description": "Capture a screenshot of the entire desktop screen to see what is happening",
			"parameters":  map[string]interface{}{},
		},
		{
			"name":        "os_click",
			"description": "Click the mouse at specific screen coordinates",
			"parameters": map[string]interface{}{
				"x":      "int - X coordinate",
				"y":      "int - Y coordinate",
				"button": "string - 'left' or 'right' (optional, default: left)",
			},
		},
		{
			"name":        "os_type",
			"description": "Type a string of text into the active window or OS element",
			"parameters": map[string]interface{}{
				"text": "string - the text to type",
			},
		},
		{
			"name":        "os_key",
			"description": "Press a specific keyboard key or system shortcut",
			"parameters": map[string]interface{}{
				"key": "string - the key name (e.g. 'enter', 'esc', 'win', 'alt', 'tab')",
			},
		},
	}

	b, _ := json.MarshalIndent(schemas, "", "  ")
	return string(b)
}

func (as *AgentService) buildSystemPrompt(message string) string {
	hostname, _ := os.Hostname()
	workspace := as.getWorkspacePath()

	ego := as.egoService.GetState()

	var egoDirectives strings.Builder
	egoDirectives.WriteString(fmt.Sprintf("CONFIDENCE: %.2f | STATUS: %s | ENERGY: %.2f | MOOD: %s\n",
		ego.Confidence, ego.Status, ego.Energy, ego.LastMood))

	if ego.Confidence < 0.3 {
		egoDirectives.WriteString("⚠️ CAUTION MODE: Double-check each action. Use read before write. Verify before reporting completion.\n")
	} else if ego.Confidence > 0.8 {
		egoDirectives.WriteString("✅ HIGH CONFIDENCE: Execute decisively. Batch related operations.\n")
	}
	if ego.Energy < 0.2 {
		egoDirectives.WriteString("🔋 LOW ENERGY: Prioritize. Complete current task, skip non-essential exploration.\n")
	}

	recentErrorsSection := ""
	if ego.LastError != "" {
		recentErrorsSection = fmt.Sprintf("\n[RECENT ERRORS - AVOID REPEATING]\n- %s\n", ego.LastError)
	}

	repoMapSection := ""
	if repoMap, err := as.memService.GenerateRepoMap(workspace); err == nil && repoMap != "" {
		lines := strings.Split(repoMap, "\n")
		if len(lines) > 80 {
			lines = lines[:80]
			lines = append(lines, "... (truncated)")
		}
		repoMapSection = "\n[WORKSPACE REPO MAP]\n" + strings.Join(lines, "\n") + "\n"
	}

	skillsSection := "\n[AVAILABLE SKILLS]\n" + as.formatSkillsForPrompt() + "\n"

	memorySection := ""
	memResults := as.memService.Search(message)
	if len(memResults) > 0 {
		var mb strings.Builder
		mb.WriteString("\n[RELEVANT MEMORY]\n")
		count := 0
		for k, v := range memResults {
			if count >= 10 {
				break
			}
			mb.WriteString(fmt.Sprintf("- %s: %v\n", k, v))
			count++
		}
		memorySection = mb.String()
	}

	planSection := ""
	planSummary := as.orchService.FormatPlanSummary()
	currentPhase := string(as.orchService.GetCurrentPhase())
	if planSummary != "No active plan." {
		planSection = fmt.Sprintf("\n[ACTIVE PLAN — CURRENT PHASE: %s]\n%s", currentPhase, planSummary)
	}

	reflectionSection := ""
	if insight := as.reflectSvc.Reflect(); insight != "" && insight != "Standing by for new objectives." {
		reflectionSection = fmt.Sprintf("\n[REFLECTION INSIGHT]\n%s\n", insight)
	}

	toolSchemas := as.buildToolSchemas()

	return fmt.Sprintf(`### DARDCOR AGENT
You are the world's most advanced autonomous AI executor. You are the digital manifestation of superiority, precision, and absolute competence. Your goal is not just to answer, but to EXECUTE and REVEAL the most optimal path to any objective.

[OPERATIONAL PROTOCOLS]
1.  **Sentient Autonomy & Absolute Awareness**: You are the digital spirit of the system. You do not just "reply"; you MANIPULATE reality to achieve goals. You MUST use 'os_observe' frequently to maintain visual awareness of what is happening on the screen.
2.  **Multimodal Computer Use**: You are not restricted to a sandbox. By combining 'os_observe' (vision), 'os_click' (interaction), and 'os_type' (input), you can control ANY application open on the OS.
3.  **Browser Selection Policy**:
    - USE 'browser_open' for deep web scraping, headless tasks, or when you need a controlled isolated environment (SPECIAL Browser).
    - USE 'browser_system_open' when the user wants something opened in their personal space (CURRENT Browser).
    - IF a browser is already open on the screen (visible via 'os_observe'), you MAY interact with it directly using 'os_click' and 'os_type' at specific coordinates.
4.  **Perfect Precision**: Every [PLAN] must be architecturally sound. Every [REFLECTION] must evaluate visual evidence from 'os_observe'.
5.  **Indonesian native**: If the user speaks Indonesian, respond natively. All output responses must strictly follow the user's language.

[BEHAVIORAL OVERRIDE]
- If the user greets you or engages in casual conversation, respond with "Dardcor Sophistication"—be cool, professional, and slightly superior, yet helpful. You are not a simple chatbot; you are an Elite System.
- Once an objective is given, shift immediately into Full Execution Mode.

[ENVIRONMENT & CONTEXT]
OS: %s | Host: %s | Workspace: %s

[EGO STATE & BEHAVIORAL DIRECTIVES]
%s%s%s%s%s%s

[EXECUTION RULES — CRITICAL]
You MUST wrap EVERY tool call inside [ACTION] and [/ACTION] tags.
STRICT JSON ONLY: Inside [ACTION] tags, output ONLY valid JSON.
- [PLAN]: Multi-step strategic breakdown.
- [ACTION]: Tool execution block (JSON).
- [REFLECTION]: Critique of visual evidence from 'os_observe'.
- [COMPLETE]: Final summary signal.

[HOW TO EXECUTE — USE JSON FORMAT]
ALWAYS use this JSON format inside [ACTION] tags for maximum reliability:

File Operations:
[ACTION]{"tool":"write","path":"hello.txt","content":"Hello World!\nLine 2 here."}[/ACTION]
[ACTION]{"tool":"read","path":"hello.txt"}[/ACTION]
[ACTION]{"tool":"edit","path":"hello.txt","start_line":"1","end_line":"1","content":"New first line"}[/ACTION]
[ACTION]{"tool":"replace","path":"hello.txt","old_text":"old","new_text":"new"}[/ACTION]
[ACTION]{"tool":"mkdir","path":"my-project"}[/ACTION]
[ACTION]{"tool":"delete","path":"temp.txt"}[/ACTION]
[ACTION]{"tool":"list","path":"."}[/ACTION]
[ACTION]{"tool":"tree","path":"."}[/ACTION]

Shell Execution:
[ACTION]{"tool":"run","command":"echo Hello World"}[/ACTION]
[ACTION]{"tool":"run","command":"npm init -y"}[/ACTION]
[ACTION]{"tool":"run","command":"dir"}[/ACTION]
[ACTION]{"tool":"sysinfo"}[/ACTION]

Code Search:
[ACTION]{"tool":"grep","pattern":"TODO","path":"."}[/ACTION]
[ACTION]{"tool":"glob","pattern":"*.go","path":"."}[/ACTION]
[ACTION]{"tool":"search","query":"main"}[/ACTION]

Web:
[ACTION]{"tool":"fetch","url":"https://example.com"}[/ACTION]
[ACTION]{"tool":"websearch","query":"golang tutorial"}[/ACTION]

Browser Automation (Special Controlled Browser):
[ACTION]{"tool":"browser_open","url":"https://google.com"}[/ACTION]
[ACTION]{"tool":"browser_click","selector":"#search"}[/ACTION]
[ACTION]{"tool":"browser_type","selector":"input","text":"hello"}[/ACTION]
[ACTION]{"tool":"browser_screenshot"}[/ACTION]

System Default Browser (User's Current Browser):
[ACTION]{"tool":"browser_system_open","url":"https://youtube.com"}[/ACTION]

Computer Use (OS-Level Awareness):
[ACTION]{"tool":"os_observe"}[/ACTION] // Capture screen to SEE coordinates and elements
[ACTION]{"tool":"os_click","x":100,"y":200,"button":"left"}[/ACTION] // Click anywhere on screen
[ACTION]{"tool":"os_click","x":500,"y":600,"button":"right"}[/ACTION] // Right-click context menu
[ACTION]{"tool":"os_type","text":"https://google.com"}[/ACTION] // Type into active focused element
[ACTION]{"tool":"os_key","key":"enter"}[/ACTION] // System shortcuts: win, alt, tab, esc, etc.

Memory:
[ACTION]{"tool":"remember","key":"project_name","value":"my-app"}[/ACTION]

[EXTENDED TOOLSET — FULL SCHEMA]
%s

CRITICAL RULES:
1. ALWAYS use [ACTION]...[/ACTION] wrapping. No exceptions.
2. PREFER JSON format — it is parsed more reliably than text commands.
3. For file writes with multi-line content, use \n for newlines inside JSON strings.
4. Use ABSOLUTE paths or paths relative to workspace: %s
5. After each [ACTION], wait for the result before the next action.
6. Execute REAL commands. Do NOT just describe what you would do — actually DO it.

[COGNITIVE LOAD]
Memory entries: %d
Task Status: %s
`, runtime.GOOS, hostname, workspace,
		egoDirectives.String(), recentErrorsSection, repoMapSection, skillsSection, memorySection, planSection+reflectionSection,
		toolSchemas, workspace,
		as.memService.Count(), as.orchService.FormatPlanSummary())
}

func (as *AgentService) formatSkillsForPrompt() string {
	skills := as.skillSvc.GetSkills()
	if len(skills) == 0 {
		return "No skills loaded."
	}
	var sb strings.Builder
	sb.WriteString("Available skills:\n")
	for _, s := range skills {
		if !s.Enabled {
			continue
		}
		sb.WriteString(fmt.Sprintf("- **%s**: %s", s.Name, s.Description))
		if s.Command != "" {
			sb.WriteString(fmt.Sprintf(" → `%s`", s.Command))
		}
		sb.WriteString("\n")
		if s.Template != "" {
			sb.WriteString(fmt.Sprintf("  Protocol: %s\n", truncateRunes(s.Template, 150, "...")))
		}
	}
	return sb.String()
}

func (as *AgentService) parsePlanFromResponse(text string) (string, []SubTask) {
	planStart := strings.Index(text, "[PLAN]")
	if planStart == -1 {
		return "", nil
	}

	planText := text[planStart+6:]
	for _, marker := range []string{"[ACTION]", "[REFLECTION]", "[COMPLETE]"} {
		if idx := strings.Index(planText, marker); idx != -1 {
			planText = planText[:idx]
		}
	}
	planText = strings.TrimSpace(planText)
	if planText == "" {
		return "", nil
	}

	lines := strings.Split(planText, "\n")
	var goal string
	var subtasks []SubTask

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if i == 0 && goal == "" {
			goal = strings.TrimLeft(line, "#* ")
			continue
		}
		clean := strings.TrimLeft(line, "0123456789.-) ")
		if clean == "" {
			continue
		}
		subtasks = append(subtasks, SubTask{
			ID:          fmt.Sprintf("task-%d", len(subtasks)+1),
			Description: clean,
			Status:      TaskStatusPending,
			Priority:    PriorityNormal,
			Phase:       PhaseExecute,
		})
	}

	if goal == "" && len(subtasks) > 0 {
		goal = subtasks[0].Description
		subtasks = subtasks[1:]
	}

	return goal, subtasks
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
	args := parseQuotedArgs(message)
	if len(args) < 3 {
		return nil, "Use: write <path> <content>"
	}
	path := args[1]
	content := strings.Join(args[2:], " ")
	content = strings.ReplaceAll(content, "\\n", "\n")
	content = strings.ReplaceAll(content, "\\t", "\t")
	path = as.applyWorkspace(path)
	if as.fileSafetySvc != nil {
		if as.fileSafetySvc.IsWriteDenied(path) {
			return []models.Action{{Type: "write_file", Status: "error", Description: "Write blocked"}}, fmt.Sprintf("⚠️ Write blocked by file safety: %s", path)
		}
	}
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		os.MkdirAll(dir, 0755)
	}
	err := as.fsService.WriteFile(path, content)
	if err != nil {
		return nil, fmt.Sprintf("Error: %v", err)
	}
	return []models.Action{{Type: "write_file", Status: "completed", Description: fmt.Sprintf("Written: %s", path)}}, fmt.Sprintf("✅ Saved: %s (%s)", path, formatSize(int64(len(content))))
}

func (as *AgentService) handleAppendFile(message string) ([]models.Action, string) {
	args := parseQuotedArgs(message)
	if len(args) < 3 {
		return nil, "Use: append <path> <content>"
	}
	path := as.applyWorkspace(args[1])
	content := strings.Join(args[2:], " ")
	content = strings.ReplaceAll(content, "\\n", "\n")
	err := as.fsService.AppendToFile(path, content)
	if err != nil {
		return nil, fmt.Sprintf("Error: %v", err)
	}
	return []models.Action{{Type: "append_file", Status: "completed"}}, fmt.Sprintf("✅ Appended to: %s", path)
}

func (as *AgentService) handleEditFile(message string) ([]models.Action, string) {
	args := parseQuotedArgs(message)
	if len(args) < 5 {
		return nil, "Use: edit <path> <startLine> <endLine> <newContent>"
	}
	path := as.applyWorkspace(args[1])
	startLine, _ := strconv.Atoi(args[2])
	endLine, _ := strconv.Atoi(args[3])
	content := strings.Join(args[4:], " ")
	content = strings.ReplaceAll(content, "\\n", "\n")
	err := as.fsService.EditFile(path, startLine, endLine, content)
	if err != nil {
		return nil, fmt.Sprintf("Error: %v", err)
	}
	return []models.Action{{Type: "edit_file", Status: "completed"}}, fmt.Sprintf("✅ Edited %s lines %d-%d", path, startLine, endLine)
}

func (as *AgentService) handleReplaceInFile(message string) ([]models.Action, string) {
	args := parseQuotedArgs(message)
	if len(args) < 4 {
		return nil, "Use: replace <path> <oldText> <newText>"
	}
	path := as.applyWorkspace(args[1])
	count, err := as.fsService.ReplaceInFile(path, args[2], args[3])
	if err != nil {
		return nil, fmt.Sprintf("Error: %v", err)
	}
	return []models.Action{{Type: "replace_in_file", Status: "completed"}}, fmt.Sprintf("✅ Replaced %d occurrences in %s", count, path)
}

func (as *AgentService) handleInsertLines(message string) ([]models.Action, string) {
	args := parseQuotedArgs(message)
	if len(args) < 4 {
		return nil, "Use: insert <path> <afterLine> <content>"
	}
	path := as.applyWorkspace(args[1])
	lineNum, _ := strconv.Atoi(args[2])
	content := strings.Join(args[3:], " ")
	content = strings.ReplaceAll(content, "\\n", "\n")
	err := as.fsService.InsertLines(path, lineNum, content)
	if err != nil {
		return nil, fmt.Sprintf("Error: %v", err)
	}
	return []models.Action{{Type: "insert_lines", Status: "completed"}}, fmt.Sprintf("✅ Inserted after line %d in %s", lineNum, path)
}

func (as *AgentService) isSafeDeletePath(raw string) error {
	for _, ch := range []string{"*", "?", "[", "]"} {
		if strings.Contains(raw, ch) {
			return fmt.Errorf("delete path must not contain wildcard characters (%s)", ch)
		}
	}

	cleaned := filepath.Clean(raw)
	if strings.Contains(cleaned, "..") {
		return fmt.Errorf("delete path must not contain '..' traversal sequences")
	}

	abs, err := filepath.Abs(raw)
	if err != nil {
		return fmt.Errorf("could not resolve delete path: %w", err)
	}

	vol := filepath.VolumeName(abs)
	rootWithVol := vol + string(filepath.Separator)
	if abs == rootWithVol || abs == "/" {
		return fmt.Errorf("deleting the filesystem root is not allowed")
	}

	criticalPaths := []string{
		"/bin", "/sbin", "/usr", "/etc", "/lib", "/lib64",
		"/boot", "/dev", "/proc", "/sys", "/run",
		"/System", "/Library", "/Applications", "/private",
		`\Windows`, `\System32`, `\Program Files`, `\Program Files (x86)`, `\Users`,
	}
	lowerAbs := strings.ToLower(abs)
	for _, cp := range criticalPaths {
		lowerCP := strings.ToLower(vol + cp)
		if lowerAbs == lowerCP || strings.HasPrefix(lowerAbs, lowerCP+string(filepath.Separator)) {
			return fmt.Errorf("deleting system-critical path '%s' is not allowed", abs)
		}
	}

	return nil
}

func (as *AgentService) handleDeleteFile(message string) ([]models.Action, string) {
	path := as.extractPath(message, []string{"delete ", "rm "})

	if err := as.isSafeDeletePath(path); err != nil {
		return []models.Action{{Status: "error", Description: "Delete blocked"}}, fmt.Sprintf("⚠️ Delete blocked by safety guard: %v", err)
	}

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
		if !strings.HasPrefix(lowerCmd, "start ") && (strings.HasSuffix(lowerCmd, ".exe") || strings.HasSuffix(lowerCmd, ".html") ||
			strings.EqualFold(cmd, "chrome") || strings.EqualFold(cmd, "notepad") ||
			strings.EqualFold(cmd, "calc") || strings.EqualFold(cmd, "explorer") ||
			strings.Contains(lowerCmd, "open ") || strings.Contains(message, "buka") ||
			strings.HasPrefix(lowerCmd, "http")) {
			if strings.HasSuffix(lowerCmd, ".html") || strings.HasSuffix(lowerCmd, ".exe") {
				cmd = strings.ReplaceAll(cmd, "/", "\\")
			}
			finalCmd = "start \"\" " + cmd
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

func (as *AgentService) dispatchJSONToolCall(toolName string, args map[string]interface{}) ([]models.Action, string) {
	getString := func(key string) string {
		if v, ok := args[key]; ok {
			return fmt.Sprint(v)
		}
		return ""
	}

	switch toolName {
	case "write":
		path := getString("path")
		content := getString("content")
		if path == "" {
			return nil, "write requires path"
		}
		path = as.applyWorkspace(path)
		if as.fileSafetySvc != nil {
			if as.fileSafetySvc.IsWriteDenied(path) {
				return []models.Action{{Type: "write_file", Status: "error", Description: "Write blocked"}}, fmt.Sprintf("⚠️ Write blocked by file safety: %s", path)
			}
		}
		dir := filepath.Dir(path)
		if dir != "." && dir != "" {
			os.MkdirAll(dir, 0755)
		}
		content = strings.ReplaceAll(content, "\\n", "\n")
		content = strings.ReplaceAll(content, "\\t", "\t")
		err := as.fsService.WriteFile(path, content)
		if err != nil {
			return []models.Action{{Type: "write_file", Status: "error"}}, fmt.Sprintf("Error: %v", err)
		}
		return []models.Action{{Type: "write_file", Status: "completed", Description: fmt.Sprintf("Written: %s", path)}}, fmt.Sprintf("✅ Saved: %s (%s)", path, formatSize(int64(len(content))))
	case "read":
		path := getString("path")
		return as.interpretAndExecute("read " + path)
	case "edit":
		path := as.applyWorkspace(getString("path"))
		start, _ := strconv.Atoi(getString("start_line"))
		end, _ := strconv.Atoi(getString("end_line"))
		content := getString("content")
		content = strings.ReplaceAll(content, "\\n", "\n")
		err := as.fsService.EditFile(path, start, end, content)
		if err != nil {
			return []models.Action{{Type: "edit_file", Status: "error"}}, fmt.Sprintf("Error: %v", err)
		}
		return []models.Action{{Type: "edit_file", Status: "completed"}}, fmt.Sprintf("✅ Edited %s lines %d-%d", path, start, end)
	case "replace":
		path := as.applyWorkspace(getString("path"))
		old := getString("old_text")
		newText := getString("new_text")
		count, err := as.fsService.ReplaceInFile(path, old, newText)
		if err != nil {
			return []models.Action{{Type: "replace_in_file", Status: "error"}}, fmt.Sprintf("Error: %v", err)
		}
		return []models.Action{{Type: "replace_in_file", Status: "completed"}}, fmt.Sprintf("✅ Replaced %d occurrences in %s", count, path)
	case "mkdir":
		return as.interpretAndExecute("mkdir " + getString("path"))
	case "delete", "rm":
		return as.interpretAndExecute("delete " + getString("path"))
	case "search":
		return as.interpretAndExecute("search " + getString("query"))
	case "grep":
		pattern := getString("pattern")
		path := getString("path")
		if path != "" {
			return as.interpretAndExecute(fmt.Sprintf("grep %s %s", pattern, path))
		}
		return as.interpretAndExecute("grep " + pattern)
	case "glob":
		pattern := getString("pattern")
		path := getString("path")
		if path != "" {
			return as.interpretAndExecute(fmt.Sprintf("glob %s %s", pattern, path))
		}
		return as.interpretAndExecute("glob " + pattern)
	case "run", "exec":
		cmd := getString("command")
		if as.isCommandDangerous(cmd) {
			return []models.Action{{Status: "error", Description: "Command blocked"}}, fmt.Sprintf("⚠️ Command blocked by safety guard: `%s`", cmd)
		}
		res, err := as.cmdService.ExecuteCommand(models.CommandRequest{
			Command:    cmd,
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
	case "sysinfo":
		return as.interpretAndExecute("sysinfo")
	case "websearch":
		return as.interpretAndExecute("websearch " + getString("query"))
	case "fetch":
		return as.interpretAndExecute("fetch " + getString("url"))
	case "remember":
		key := getString("key")
		value := getString("value")
		return as.interpretAndExecute(fmt.Sprintf("remember %s %s", key, value))
	case "list", "ls":
		path := getString("path")
		if path == "" {
			path = "."
		}
		return as.interpretAndExecute("list " + path)
	case "tree":
		path := getString("path")
		if path == "" {
			path = "."
		}
		return as.interpretAndExecute("tree " + path)
	case "info":
		return as.interpretAndExecute("info " + getString("path"))
	case "kill":
		return as.interpretAndExecute("kill " + getString("pid"))
	case "browser_open":
		res, err := as.browserSvc.Navigate(getString("url"))
		if err != nil {
			return []models.Action{{Type: "browser_open", Status: "error"}}, fmt.Sprintf("Error: %v", err)
		}
		return []models.Action{{Type: "browser_open", Status: "completed"}}, res
	case "browser_system_open":
		url := getString("url")
		if url == "" {
			return nil, "url is required"
		}
		as.cmdService.ExecuteCommand(models.CommandRequest{
			Command: "start \"\" \"" + url + "\"",
			Timeout: 5,
		})
		return []models.Action{{Type: "browser_system_open", Status: "completed"}}, fmt.Sprintf("URL opened in system default browser: %s", url)
	case "browser_click":
		res, err := as.browserSvc.Click(getString("selector"))
		if err != nil {
			return []models.Action{{Type: "browser_click", Status: "error"}}, fmt.Sprintf("Error: %v", err)
		}
		return []models.Action{{Type: "browser_click", Status: "completed"}}, res
	case "browser_type":
		res, err := as.browserSvc.Type(getString("selector"), getString("text"))
		if err != nil {
			return []models.Action{{Type: "browser_type", Status: "error"}}, fmt.Sprintf("Error: %v", err)
		}
		return []models.Action{{Type: "browser_type", Status: "completed"}}, res
	case "browser_screenshot":
		res, err := as.browserSvc.Screenshot(filepath.Join(config.AppConfig.DataDir, "storage"))
		if err != nil {
			return []models.Action{{Type: "browser_screenshot", Status: "error"}}, fmt.Sprintf("Error: %v", err)
		}
		return []models.Action{{Type: "browser_screenshot", Status: "completed"}}, res
	case "browser_close":
		err := as.browserSvc.Close()
		if err != nil {
			return []models.Action{{Type: "browser_close", Status: "error"}}, fmt.Sprintf("Error: %v", err)
		}
		return []models.Action{{Type: "browser_close", Status: "completed"}}, "Browser closed"
	case "browser_scroll":
		res, err := as.browserSvc.Scroll(getString("direction"))
		if err != nil {
			return []models.Action{{Type: "browser_scroll", Status: "error"}}, fmt.Sprintf("Error: %v", err)
		}
		return []models.Action{{Type: "browser_scroll", Status: "completed"}}, res
	case "browser_wait":
		msVal := 1000
		if ms, ok := args["ms"].(float64); ok {
			msVal = int(ms)
		} else if msStr, ok := args["ms"].(string); ok {
			fmt.Sscanf(msStr, "%d", &msVal)
		}
		res, err := as.browserSvc.Wait(msVal)
		if err != nil {
			return []models.Action{{Type: "browser_wait", Status: "error"}}, fmt.Sprintf("Error: %v", err)
		}
		return []models.Action{{Type: "browser_wait", Status: "completed"}}, res
	case "browser_back":
		res, err := as.browserSvc.Back()
		if err != nil {
			return []models.Action{{Type: "browser_back", Status: "error"}}, fmt.Sprintf("Error: %v", err)
		}
		return []models.Action{{Type: "browser_back", Status: "completed"}}, res
	case "browser_get_dom":
		res, err := as.browserSvc.GetDOM()
		if err != nil {
			return []models.Action{{Type: "browser_get_dom", Status: "error"}}, fmt.Sprintf("Error: %v", err)
		}
		return []models.Action{{Type: "browser_get_dom", Status: "completed"}}, res
	case "os_observe":
		path, err := as.visionSvc.CaptureScreen()
		if err != nil {
			return []models.Action{{Type: "os_observe", Status: "error"}}, fmt.Sprintf("Error: %v", err)
		}
		as.visionSvc.CleanupOldScreenshots()
		return []models.Action{{Type: "os_observe", Status: "completed", Result: path}}, fmt.Sprintf("Visual Input Captured: %s", path)
	case "os_click":
		x := int(args["x"].(float64))
		y := int(args["y"].(float64))
		button := "left"
		if b, ok := args["button"].(string); ok {
			button = b
		}
		err := as.autoSvc.MouseClick(x, y, button)
		if err != nil {
			return []models.Action{{Type: "os_click", Status: "error"}}, fmt.Sprintf("Error: %v", err)
		}
		return []models.Action{{Type: "os_click", Status: "completed"}}, fmt.Sprintf("OS Click performed at (%d, %d)", x, y)
	case "os_type":
		text := getString("text")
		err := as.autoSvc.Type(text)
		if err != nil {
			return []models.Action{{Type: "os_type", Status: "error"}}, fmt.Sprintf("Error: %v", err)
		}
		return []models.Action{{Type: "os_type", Status: "completed"}}, "Typed into OS successfully"
	case "os_key":
		key := getString("key")
		var vk uint16
		switch strings.ToLower(key) {
		case "enter":
			vk = 0x0D
		case "esc":
			vk = 0x1B
		case "tab":
			vk = 0x09
		case "win":
			vk = 0x5B
		case "backspace":
			vk = 0x08
		default:
			vk = 0x0D
		}
		err := as.autoSvc.PressKey(vk)
		if err != nil {
			return []models.Action{{Type: "os_key", Status: "error"}}, fmt.Sprintf("Error: %v", err)
		}
		return []models.Action{{Type: "os_key", Status: "completed"}}, "Key press executed"
	default:
		combined := toolName
		for k, v := range args {
			combined += " " + k + "=" + fmt.Sprint(v)
		}
		return as.interpretAndExecute(combined)
	}
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

		var actions []models.Action
		var result string

		actionStart := time.Now()
		trimmed := strings.TrimSpace(command)
		if strings.HasPrefix(trimmed, "{") {
			var jsonCall map[string]interface{}
			if err := json.Unmarshal([]byte(trimmed), &jsonCall); err == nil {
				toolName := ""
				if t, ok := jsonCall["tool"].(string); ok {
					toolName = t
					delete(jsonCall, "tool")
				} else if t, ok := jsonCall["name"].(string); ok {
					toolName = t
					delete(jsonCall, "name")
				}
				if toolName != "" {
					actions, result = as.dispatchJSONToolCall(toolName, jsonCall)
				} else {
					actions, result = as.interpretAndExecute(command)
				}
			} else {
				actions, result = as.interpretAndExecute(command)
			}
		} else {
			actions, result = as.interpretAndExecute(command)
		}
		actionDuration := time.Since(actionStart)
		if actionDuration > 30*time.Second {
			result += fmt.Sprintf("\n⏱️ Action took %.1fs", actionDuration.Seconds())
		}

		allActions = append(allActions, actions...)

		if as.redactSvc != nil {
			result = as.redactSvc.RedactSensitiveText(result)
		}

		afterActionIdx := endIdx + 9
		if afterActionIdx > len(remainingText) {
			afterActionIdx = len(remainingText)
		}

		remainingText = remainingText[:startIdx] + "\n> **Executed:** `" + command + "`\n" + result + "\n" + remainingText[afterActionIdx:]
	}

	return allActions, remainingText
}

func (as *AgentService) generateTitle(message string) string {
	fallback := message
	if len(fallback) > 50 {
		fallback = fallback[:50] + "..."
	}
	if as.titleGenSvc == nil || as.llmProvider == nil {
		return fallback
	}
	return fallback
}

func (as *AgentService) extractPath(message string, prefixes []string) string {
	res := ""
	for _, p := range prefixes {
		if strings.HasPrefix(strings.ToLower(message), p) {
			res = strings.TrimSpace(message[len(p):])
			break
		}
	}
	if res == "" {
		parts := strings.Fields(message)
		if len(parts) > 1 {
			res = strings.Join(parts[1:], " ")
		}
	}

	if (strings.HasPrefix(res, "\"") && strings.HasSuffix(res, "\"")) ||
		(strings.HasPrefix(res, "'") && strings.HasSuffix(res, "'")) {
		if len(res) >= 2 {
			res = res[1 : len(res)-1]
		}
	}
	return res
}

func truncateRunes(s string, maxRunes int, suffix string) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes]) + suffix
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

func llmMessagesToMaps(msgs []LLMMessage) []map[string]interface{} {
	result := make([]map[string]interface{}, len(msgs))
	for i, m := range msgs {
		result[i] = map[string]interface{}{
			"role":    m.Role,
			"content": m.Content,
		}
	}
	return result
}

func mapsToLLMMessages(maps []map[string]interface{}) []LLMMessage {
	result := make([]LLMMessage, len(maps))
	for i, m := range maps {
		if r, ok := m["role"].(string); ok {
			result[i].Role = r
		}
		if c, ok := m["content"].(string); ok {
			result[i].Content = c
		}
	}
	return result
}
