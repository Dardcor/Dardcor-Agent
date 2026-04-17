package services

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"dardcor-agent/config"
	"dardcor-agent/models"
)

type LLMMessage struct {
	Role    string `json:"role" xml:"role"`
	Content string `json:"content" xml:"content"`
}

type LLMResponse struct {
	Content  string `json:"content"`
	Model    string `json:"model"`
	Provider string `json:"provider"`
	Tokens   int    `json:"tokens"`
	Duration int64  `json:"duration_ms"`
}

type StreamCallback func(chunk string)

type LLMErrorKind string

const (
	LLMErrRateLimit   LLMErrorKind = "rate_limit"
	LLMErrAuth        LLMErrorKind = "auth"
	LLMErrTimeout     LLMErrorKind = "timeout"
	LLMErrBadRequest  LLMErrorKind = "bad_request"
	LLMErrServerError LLMErrorKind = "server_error"
	LLMErrUnknown     LLMErrorKind = "unknown"
)

type LLMError struct {
	Kind    LLMErrorKind
	Message string
	Status  int
}

func (e *LLMError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Kind, e.Message)
}

func classifyHTTPError(status int, body string) *LLMError {
	switch {
	case status == 429:
		return &LLMError{Kind: LLMErrRateLimit, Message: "rate limit exceeded", Status: status}
	case status == 401 || status == 403:
		return &LLMError{Kind: LLMErrAuth, Message: "authentication failed: " + body, Status: status}
	case status == 400:
		return &LLMError{Kind: LLMErrBadRequest, Message: "bad request: " + body, Status: status}
	case status >= 500:
		return &LLMError{Kind: LLMErrServerError, Message: fmt.Sprintf("server error %d: %s", status, body), Status: status}
	default:
		return &LLMError{Kind: LLMErrUnknown, Message: fmt.Sprintf("API error %d: %s", status, body), Status: status}
	}
}

func isRetryable(err error) bool {
	if llmErr, ok := err.(*LLMError); ok {
		return llmErr.Kind == LLMErrRateLimit || llmErr.Kind == LLMErrServerError || llmErr.Kind == LLMErrTimeout
	}
	return true
}

type LLMProvider struct {
	cfg    config.AIConfig
	client *http.Client
	agSvc  *AntigravityService
}

type openAIResponseBody struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
	Model string `json:"model"`
}

type openAIStreamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

type anthropicReqBody struct {
	Model     string       `json:"model"`
	MaxTokens int          `json:"max_tokens"`
	Messages  []LLMMessage `json:"messages"`
	System    string       `json:"system,omitempty"`
	Stream    bool         `json:"stream,omitempty"`
}

type anthropicRespBody struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Model string `json:"model"`
}

type geminiNativeRequest struct {
	Contents          []map[string]interface{} `json:"contents"`
	SystemInstruction *map[string]interface{}  `json:"systemInstruction,omitempty"`
	GenerationConfig  map[string]interface{}   `json:"generationConfig,omitempty"`
}

type geminiNativeResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	UsageMetadata struct {
		TotalTokenCount int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
}

func NewLLMProvider(cfg config.AIConfig, agSvc *AntigravityService) *LLMProvider {
	return &LLMProvider{
		cfg: cfg,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
		agSvc: agSvc,
	}
}

func (p *LLMProvider) Complete(systemPrompt string, messages []LLMMessage) (*LLMResponse, error) {
	start := time.Now()
	provider := strings.ToLower(p.cfg.Provider)
	switch provider {
	case "anthropic":
		return p.callWithRetry(func() (*LLMResponse, error) {
			return p.callAnthropic(systemPrompt, messages, start)
		})
	case "antigravity":
		return p.callAntigravity(systemPrompt, messages, start)
	case "gemini":
		return p.callWithRetry(func() (*LLMResponse, error) {
			return p.callGeminiNative(systemPrompt, messages, start)
		})
	case "local", "":
		return p.callLocal(systemPrompt, messages, start)
	default:
		return p.callWithRetry(func() (*LLMResponse, error) {
			return p.callOpenAICompat(systemPrompt, messages, start)
		})
	}
}

func (p *LLMProvider) CompleteStream(systemPrompt string, messages []LLMMessage, cb StreamCallback) (*LLMResponse, error) {
	start := time.Now()
	provider := strings.ToLower(p.cfg.Provider)
	switch provider {
	case "anthropic":
		return p.streamAnthropic(systemPrompt, messages, start, cb)
	case "antigravity":
		resp, err := p.callAntigravity(systemPrompt, messages, start)
		if err != nil {
			return nil, err
		}
		if cb != nil {
			cb(resp.Content)
		}
		return resp, nil
	case "gemini":
		return p.streamGeminiNative(systemPrompt, messages, start, cb)
	case "local", "":
		resp, err := p.callLocal(systemPrompt, messages, start)
		if err != nil {
			return nil, err
		}
		if cb != nil {
			cb(resp.Content)
		}
		return resp, nil
	default:
		return p.streamOpenAICompat(systemPrompt, messages, start, cb)
	}
}

// Summarize uses the LLM to compress a set of messages into a concise summary.
func (p *LLMProvider) Summarize(messages []LLMMessage) (string, error) {
	if len(messages) == 0 {
		return "", nil
	}

	var sb strings.Builder
	for _, m := range messages {
		sb.WriteString(fmt.Sprintf("%s: %s\n", m.Role, m.Content))
	}

	prompt := "Summarize the following conversation concisely, preserving critical technical details and state changes:"
	resp, err := p.Complete(prompt, []LLMMessage{{Role: "user", Content: sb.String()}})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

func (p *LLMProvider) callWithRetry(fn func() (*LLMResponse, error)) (*LLMResponse, error) {
	backoffs := []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second}
	var lastErr error
	for attempt := 0; attempt <= len(backoffs); attempt++ {
		resp, err := fn()
		if err == nil {
			return resp, nil
		}
		lastErr = err
		if !isRetryable(err) {
			return nil, err
		}
		if attempt < len(backoffs) {
			time.Sleep(backoffs[attempt])
		}
	}
	return nil, lastErr
}

func (p *LLMProvider) callOpenAICompat(systemPrompt string, messages []LLMMessage, start time.Time) (*LLMResponse, error) {
	allMsgs := make([]map[string]string, 0, len(messages)+1)
	if systemPrompt != "" {
		allMsgs = append(allMsgs, map[string]string{
			"role":    "system",
			"content": systemPrompt,
		})
	}
	for _, m := range messages {
		allMsgs = append(allMsgs, map[string]string{
			"role":    m.Role,
			"content": m.Content,
		})
	}

	payload := map[string]interface{}{
		"model":       p.cfg.Model,
		"messages":    allMsgs,
		"max_tokens":  p.cfg.MaxTokens,
		"temperature": p.cfg.Temperature,
	}

	bodyBytes, _ := json.Marshal(payload)
	baseURL := p.cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	req, err := http.NewRequest("POST", baseURL+"/chat/completions", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.cfg.APIKey)
	if p.cfg.Provider == "openrouter" {
		req.Header.Set("HTTP-Referer", "https://dardcor.ai")
		req.Header.Set("X-Title", "Dardcor Agent")
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, &LLMError{Kind: LLMErrTimeout, Message: err.Error()}
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, classifyHTTPError(resp.StatusCode, string(respBytes))
	}

	var result openAIResponseBody
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("empty choices from %s", p.cfg.Provider)
	}

	return &LLMResponse{
		Content:  result.Choices[0].Message.Content,
		Model:    result.Model,
		Provider: p.cfg.Provider,
		Tokens:   result.Usage.TotalTokens,
		Duration: time.Since(start).Milliseconds(),
	}, nil
}

func (p *LLMProvider) streamOpenAICompat(systemPrompt string, messages []LLMMessage, start time.Time, cb StreamCallback) (*LLMResponse, error) {
	allMsgs := make([]map[string]string, 0, len(messages)+1)
	if systemPrompt != "" {
		allMsgs = append(allMsgs, map[string]string{
			"role":    "system",
			"content": systemPrompt,
		})
	}
	for _, m := range messages {
		allMsgs = append(allMsgs, map[string]string{
			"role":    m.Role,
			"content": m.Content,
		})
	}

	payload := map[string]interface{}{
		"model":       p.cfg.Model,
		"messages":    allMsgs,
		"max_tokens":  p.cfg.MaxTokens,
		"temperature": p.cfg.Temperature,
		"stream":      true,
	}

	bodyBytes, _ := json.Marshal(payload)
	baseURL := p.cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	req, err := http.NewRequest("POST", baseURL+"/chat/completions", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.cfg.APIKey)
	if p.cfg.Provider == "openrouter" {
		req.Header.Set("HTTP-Referer", "https://dardcor.ai")
		req.Header.Set("X-Title", "Dardcor Agent")
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, &LLMError{Kind: LLMErrTimeout, Message: err.Error()}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, classifyHTTPError(resp.StatusCode, string(body))
	}

	var fullContent strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}
		var chunk openAIStreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if len(chunk.Choices) > 0 {
			delta := chunk.Choices[0].Delta.Content
			if delta != "" {
				fullContent.WriteString(delta)
				if cb != nil {
					cb(delta)
				}
			}
		}
	}

	return &LLMResponse{
		Content:  fullContent.String(),
		Model:    p.cfg.Model,
		Provider: p.cfg.Provider,
		Duration: time.Since(start).Milliseconds(),
	}, nil
}

func (p *LLMProvider) callAnthropic(systemPrompt string, messages []LLMMessage, start time.Time) (*LLMResponse, error) {
	maxTokens := p.cfg.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	payload := anthropicReqBody{
		Model:     p.cfg.Model,
		MaxTokens: maxTokens,
		Messages:  messages,
		System:    systemPrompt,
	}

	bodyBytes, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create Anthropic request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.cfg.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, &LLMError{Kind: LLMErrTimeout, Message: err.Error()}
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, classifyHTTPError(resp.StatusCode, string(respBytes))
	}

	var result anthropicRespBody
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to parse Anthropic response: %w", err)
	}
	content := ""
	for _, c := range result.Content {
		if c.Type == "text" {
			content += c.Text
		}
	}

	return &LLMResponse{
		Content:  content,
		Model:    result.Model,
		Provider: "anthropic",
		Tokens:   result.Usage.InputTokens + result.Usage.OutputTokens,
		Duration: time.Since(start).Milliseconds(),
	}, nil
}

func (p *LLMProvider) streamAnthropic(systemPrompt string, messages []LLMMessage, start time.Time, cb StreamCallback) (*LLMResponse, error) {
	maxTokens := p.cfg.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	payload := anthropicReqBody{
		Model:     p.cfg.Model,
		MaxTokens: maxTokens,
		Messages:  messages,
		System:    systemPrompt,
		Stream:    true,
	}

	bodyBytes, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create Anthropic stream request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.cfg.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, &LLMError{Kind: LLMErrTimeout, Message: err.Error()}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, classifyHTTPError(resp.StatusCode, string(body))
	}

	var fullContent strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		var event map[string]interface{}
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}
		if evType, _ := event["type"].(string); evType == "content_block_delta" {
			if delta, ok := event["delta"].(map[string]interface{}); ok {
				if text, _ := delta["text"].(string); text != "" {
					fullContent.WriteString(text)
					if cb != nil {
						cb(text)
					}
				}
			}
		}
	}

	return &LLMResponse{
		Content:  fullContent.String(),
		Model:    p.cfg.Model,
		Provider: "anthropic",
		Duration: time.Since(start).Milliseconds(),
	}, nil
}

func (p *LLMProvider) callLocal(_ string, messages []LLMMessage, start time.Time) (*LLMResponse, error) {
	lastMsg := ""
	for _, m := range messages {
		if m.Role == "user" {
			lastMsg = m.Content
		}
	}
	response := buildLocalResponse(lastMsg)
	return &LLMResponse{
		Content:  response,
		Model:    "dardcor-local",
		Provider: "local",
		Tokens:   len(strings.Fields(response)),
		Duration: time.Since(start).Milliseconds(),
	}, nil
}

func buildLocalResponse(message string) string {
	msg := strings.ToLower(strings.TrimSpace(message))
	if strings.Contains(msg, "ultrawork") || strings.Contains(msg, "ulw") {
		return "**ULTRAWORK MODE** — Configure a provider to continue."
	}
	return "**Dardcor Agent** (Local Mode)\n\nConfigure a provider via CLI or Dashboard."
}

func (p *LLMProvider) callGeminiNative(systemPrompt string, messages []LLMMessage, start time.Time) (*LLMResponse, error) {
	if p.cfg.APIKey == "" {
		return nil, &LLMError{Kind: LLMErrAuth, Message: "gemini API key not configured"}
	}

	modelName := p.cfg.Model
	if modelName == "" {
		modelName = "gemini-1.5-flash"
	}

	var contents []map[string]interface{}
	for _, m := range messages {
		role := m.Role
		if role == "assistant" {
			role = "model"
		}
		if role == "system" {
			role = "user"
		}
		contents = append(contents, map[string]interface{}{
			"role":  role,
			"parts": []map[string]interface{}{{"text": m.Content}},
		})
	}

	reqBody := geminiNativeRequest{
		Contents: contents,
		GenerationConfig: map[string]interface{}{
			"temperature":     0.7,
			"maxOutputTokens": 8192,
		},
	}

	if p.cfg.Temperature > 0 {
		reqBody.GenerationConfig["temperature"] = p.cfg.Temperature
	}
	if p.cfg.MaxTokens > 0 {
		reqBody.GenerationConfig["maxOutputTokens"] = p.cfg.MaxTokens
	}

	if systemPrompt != "" {
		si := map[string]interface{}{
			"role":  "user",
			"parts": []map[string]interface{}{{"text": systemPrompt}},
		}
		reqBody.SystemInstruction = &si
	}

	bodyBytes, _ := json.Marshal(reqBody)
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", modelName, p.cfg.APIKey)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, &LLMError{Kind: LLMErrTimeout, Message: err.Error()}
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, classifyHTTPError(resp.StatusCode, string(respBytes))
	}

	var result geminiNativeResponse
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to parse Gemini response: %w", err)
	}

	if len(result.Candidates) == 0 {
		return nil, fmt.Errorf("empty candidates from Gemini")
	}

	var textParts []string
	for _, part := range result.Candidates[0].Content.Parts {
		if part.Text != "" {
			textParts = append(textParts, part.Text)
		}
	}

	return &LLMResponse{
		Content:  strings.Join(textParts, ""),
		Model:    modelName,
		Provider: "gemini",
		Tokens:   result.UsageMetadata.TotalTokenCount,
		Duration: time.Since(start).Milliseconds(),
	}, nil
}

func (p *LLMProvider) streamGeminiNative(systemPrompt string, messages []LLMMessage, start time.Time, cb StreamCallback) (*LLMResponse, error) {
	if p.cfg.APIKey == "" {
		return nil, &LLMError{Kind: LLMErrAuth, Message: "gemini API key not configured"}
	}

	modelName := p.cfg.Model
	if modelName == "" {
		modelName = "gemini-1.5-flash"
	}

	var contents []map[string]interface{}
	for _, m := range messages {
		role := m.Role
		if role == "assistant" {
			role = "model"
		}
		if role == "system" {
			role = "user"
		}
		contents = append(contents, map[string]interface{}{
			"role":  role,
			"parts": []map[string]interface{}{{"text": m.Content}},
		})
	}

	reqBody := geminiNativeRequest{
		Contents: contents,
		GenerationConfig: map[string]interface{}{
			"temperature":     0.7,
			"maxOutputTokens": 8192,
		},
	}

	if p.cfg.Temperature > 0 {
		reqBody.GenerationConfig["temperature"] = p.cfg.Temperature
	}
	if p.cfg.MaxTokens > 0 {
		reqBody.GenerationConfig["maxOutputTokens"] = p.cfg.MaxTokens
	}

	if systemPrompt != "" {
		si := map[string]interface{}{
			"role":  "user",
			"parts": []map[string]interface{}{{"text": systemPrompt}},
		}
		reqBody.SystemInstruction = &si
	}

	bodyBytes, _ := json.Marshal(reqBody)
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:streamGenerateContent?alt=sse&key=%s", modelName, p.cfg.APIKey)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini stream request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, &LLMError{Kind: LLMErrTimeout, Message: err.Error()}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, classifyHTTPError(resp.StatusCode, string(body))
	}

	var fullContent strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		var chunk geminiNativeResponse
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if len(chunk.Candidates) > 0 {
			for _, part := range chunk.Candidates[0].Content.Parts {
				if part.Text != "" {
					fullContent.WriteString(part.Text)
					if cb != nil {
						cb(part.Text)
					}
				}
			}
		}
	}

	return &LLMResponse{
		Content:  fullContent.String(),
		Model:    modelName,
		Provider: "gemini",
		Duration: time.Since(start).Milliseconds(),
	}, nil
}

func (p *LLMProvider) callAntigravity(systemPrompt string, messages []LLMMessage, start time.Time) (*LLMResponse, error) {
	if p.agSvc == nil {
		return nil, fmt.Errorf("antigravity service not initialized")
	}

	allAccounts := p.agSvc.GetAccounts()
	if len(allAccounts) == 0 {
		return nil, fmt.Errorf("no antigravity accounts found. please add and activate an account in the dashboard")
	}

	agCfg := p.agSvc.LoadConfig()
	isAutoRotation := agCfg.SelectedModel == ""

	// Order accounts so the active one is first, but include all others for rotation
	activeAcc, _ := p.agSvc.GetActiveAccount()
	orderedAccounts := make([]models.AntigravityAccount, 0, len(allAccounts))
	if activeAcc != nil {
		orderedAccounts = append(orderedAccounts, *activeAcc)
	}
	for _, acc := range allAccounts {
		if activeAcc == nil || acc.Email != activeAcc.Email {
			orderedAccounts = append(orderedAccounts, acc)
		}
	}

	var lastErr error
	for _, acc := range orderedAccounts {
		currAcc := &acc

		// Skip forbidden accounts
		if currAcc.Status == "FORBIDDEN" {
			continue
		}

		// Refresh token if expired
		if currAcc.AccessToken == "" || (!currAcc.Expiry.IsZero() && time.Now().After(currAcc.Expiry.Add(-2*time.Minute))) {
			refreshed, err := p.agSvc.RefreshToken(currAcc.Email)
			if err != nil {
				fmt.Printf("[Antigravity] Failed to refresh token for %s: %v\n", currAcc.Email, err)
				lastErr = err
				continue
			}
			currAcc = refreshed
		}

		if currAcc.ProjectID == "" {
			p.agSvc.FetchProjectAndQuotas(currAcc)
		}

		// Prepare contents
		var contents []map[string]interface{}
		for _, m := range messages {
			role := m.Role
			if role == "system" {
				role = "user"
			}
			if role == "assistant" {
				role = "model"
			}
			contents = append(contents, map[string]interface{}{
				"role":  role,
				"parts": []map[string]interface{}{{"text": m.Content}},
			})
		}

		// Model selection logic
		modelName := p.cfg.Model
		if agCfg.SelectedModel != "" {
			modelName = agCfg.SelectedModel
		}

		// Image Gen Auto-Selection
		isImageGen := false
		for _, m := range messages {
			if strings.Contains(m.Content, "[IMAGE_GEN_MODE]") {
				isImageGen = true
				break
			}
		}

		if isImageGen {
			foundImageModel := false
			for _, q := range currAcc.Quotas {
				nameLower := strings.ToLower(q.Name)
				if strings.Contains(nameLower, "image") || strings.Contains(nameLower, "imagen") {
					modelName = q.Key
					fmt.Printf("[Antigravity] Rotation: Auto-selected image model for %s: %s\n", currAcc.Email, q.Name)
					foundImageModel = true
					break
				}
			}
			if !foundImageModel {
				fmt.Printf("[Antigravity] Rotation: No image model found for %s, trying next account\n", currAcc.Email)
				continue
			}
		} else if isAutoRotation {
			// Find best available model in this account
			bestModel := ""
			bestQuota := -1
			// Prefer high-tier models if available
			for _, q := range currAcc.Quotas {
				if q.Percentage > 0 {
					if q.Percentage > bestQuota {
						bestQuota = q.Percentage
						bestModel = q.Key
					}
				}
			}
			if bestModel != "" {
				modelName = bestModel
				fmt.Printf("[Antigravity] Auto Rotation: Selected model %s for account %s\n", modelName, currAcc.Email)
			} else {
				fmt.Printf("[Antigravity] Rotation: No quota left for any model in %s, trying next account\n", currAcc.Email)
				continue
			}
		}

		temp, maxTok, thinkBudget := 0.7, 8192, 0
		if agCfg.Temperature > 0 {
			temp = agCfg.Temperature
		}
		if agCfg.MaxTokens > 0 {
			maxTok = agCfg.MaxTokens
		}
		if agCfg.ThinkingBudget > 0 {
			thinkBudget = agCfg.ThinkingBudget
		}
		if maxTok <= thinkBudget {
			maxTok = thinkBudget + 8192
		}

		reqMap := map[string]interface{}{
			"contents": contents,
			"model":    modelName,
		}
		if systemPrompt != "" {
			reqMap["systemInstruction"] = map[string]interface{}{
				"role":  "user",
				"parts": []map[string]interface{}{{"text": systemPrompt}},
			}
		}
		reqMap["generationConfig"] = map[string]interface{}{
			"temperature":     temp,
			"maxOutputTokens": maxTok,
		}
		if thinkBudget > 0 {
			reqMap["generationConfig"].(map[string]interface{})["thinkingConfig"] = map[string]interface{}{
				"includeThoughts": true,
				"thinkingBudget":  thinkBudget,
			}
		}

		finalPayload := map[string]interface{}{
			"project":     currAcc.ProjectID,
			"request":     reqMap,
			"model":       modelName,
			"requestId":   fmt.Sprintf("agent/antigravity/%s/%d", currAcc.ID[:8], len(messages)),
			"userAgent":   "antigravity",
			"requestType": "agent",
		}

		bodyBytes, _ := json.Marshal(finalPayload)
		endpoints := []string{
			"https://daily-cloudcode-pa.googleapis.com/v1internal:generateContent",
			"https://cloudcode-pa.googleapis.com/v1internal:generateContent",
		}

		var accountErr error

		for _, endpoint := range endpoints {
			fmt.Printf("[Antigravity] Sending chat request to %s (Account: %s, Model: %s)\n", endpoint, currAcc.Email, modelName)
			req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(bodyBytes))
			if err != nil {
				continue
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+currAcc.AccessToken)
			req.Header.Set("User-Agent", "antigravity")
			req.Header.Set("x-client-name", "antigravity")
			req.Header.Set("x-client-version", "3.3.18")
			req.Header.Set("x-machine-id", "dardcor-agent-local")

			resp, err := p.client.Do(req)
			if err != nil {
				accountErr = err
				continue
			}
			respBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				fmt.Printf("[Antigravity] Account %s error %d: %s\n", currAcc.Email, resp.StatusCode, string(respBytes))
				if resp.StatusCode == 403 || resp.StatusCode == 429 {
					// These are rotation triggers
					accountErr = fmt.Errorf("quota exceeded or forbidden (HTTP %d)", resp.StatusCode)
					break // Move to next account
				}
				accountErr = fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBytes))
				continue // Try next endpoint for THIS account
			}

			var googleResp struct {
				Response struct {
					Candidates []struct {
						Content struct {
							Parts []struct {
								Text         string `json:"text,omitempty"`
								Thought      string `json:"thought,omitempty"`
								FunctionCall *struct {
									Name string                 `json:"name"`
									Args map[string]interface{} `json:"args"`
								} `json:"functionCall,omitempty"`
							} `json:"parts"`
						} `json:"content"`
					} `json:"candidates"`
				} `json:"response"`
			}
			json.Unmarshal(respBytes, &googleResp)
			if len(googleResp.Response.Candidates) > 0 {
				text := ""
				for _, p := range googleResp.Response.Candidates[0].Content.Parts {
					if p.Text != "" {
						text += p.Text
					}
					if p.Thought != "" {
						text = "> [Thinking]\n" + p.Thought + "\n\n" + text
					}
					if p.FunctionCall != nil {
						jsonArgs, _ := json.Marshal(p.FunctionCall.Args)
						text += fmt.Sprintf("\n[ACTION] %s %s [/ACTION]", p.FunctionCall.Name, string(jsonArgs))
					}
				}
				return &LLMResponse{
					Content:  text,
					Model:    modelName,
					Provider: "antigravity",
					Duration: time.Since(start).Milliseconds(),
				}, nil
			}
		}

		lastErr = accountErr
		fmt.Printf("[Antigravity] Account %s failed, trying rotation...\n", currAcc.Email)
	}

	if lastErr != nil {
		return nil, fmt.Errorf("antigravity rotation failed: %w", lastErr)
	}
	return nil, fmt.Errorf("antigravity request failed: all accounts exhausted")
}
