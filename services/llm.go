package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"dardcor-agent/config"
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

type anthropicReqBody struct {
	Model     string       `json:"model"`
	MaxTokens int          `json:"max_tokens"`
	Messages  []LLMMessage `json:"messages"`
	System    string       `json:"system,omitempty"`
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
		return p.callAnthropic(systemPrompt, messages, start)
	case "antigravity":
		return p.callAntigravity(systemPrompt, messages, start)
	case "local", "":
		return p.callLocal(systemPrompt, messages, start)
	default:
		return p.callOpenAICompat(systemPrompt, messages, start)
	}
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

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	baseURL := p.cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	req, err := http.NewRequest("POST", baseURL+"/chat/completions", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.cfg.APIKey)
	if p.cfg.Provider == "openrouter" {
		req.Header.Set("HTTP-Referer", "https://dardcor.ai")
		req.Header.Set("X-Title", "Dardcor Agent")
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBytes))
	}

	var result openAIResponseBody
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("empty choices in API response")
	}

	return &LLMResponse{
		Content:  result.Choices[0].Message.Content,
		Model:    result.Model,
		Provider: p.cfg.Provider,
		Tokens:   result.Usage.TotalTokens,
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

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Anthropic request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create Anthropic request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.cfg.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Anthropic HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Anthropic response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Anthropic API error %d: %s", resp.StatusCode, string(respBytes))
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
		return "**ULTRAWORK MODE** — AI provider not configured. Configure a provider to continue."
	}

	return fmt.Sprintf("**Dardcor Agent** (Local Mode)\n\nCommand: \"%s\"\n\nAI provider not configured. Configure a provider via CLI or Dashboard.", message)
}

func (p *LLMProvider) callAntigravity(systemPrompt string, messages []LLMMessage, start time.Time) (*LLMResponse, error) {
	if p.agSvc == nil {
		return nil, fmt.Errorf("antigravity service is not initialized")
	}

	acc, err := p.agSvc.GetActiveAccount()
	if err != nil {
		return nil, fmt.Errorf("no active Antigravity agent. Please activate one in Model > Antigravity dashboard: %v", err)
	}

	// Auto-refresh token if expired, missing, or expiring soon
	needsRefresh := acc.AccessToken == ""
	if !needsRefresh && !acc.Expiry.IsZero() && time.Now().After(acc.Expiry.Add(-2*time.Minute)) {
		needsRefresh = true
	}
	if needsRefresh {
		refreshed, refreshErr := p.agSvc.RefreshToken(acc.Email)
		if refreshErr != nil {
			return nil, fmt.Errorf("token refresh failed: %v. Please re-authenticate in the Antigravity dashboard", refreshErr)
		}
		acc = refreshed
	}

	if acc.AccessToken == "" {
		return nil, fmt.Errorf("no valid access token for active Antigravity account. Please re-authenticate in the dashboard")
	}
	if acc.ProjectID == "" {
		// Try to fetch project ID one more time
		if err := p.agSvc.FetchProjectAndQuotas(acc); err != nil {
			return nil, fmt.Errorf("active account has no project ID. Please click Refresh on the account in the dashboard")
		}
		// Re-fetch the latest account state
		if updated, err := p.agSvc.GetActiveAccount(); err == nil {
			acc = updated
		}
	}

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
			"role": role,
			"parts": []map[string]interface{}{
				{"text": m.Content},
			},
		})
	}

	modelName := p.cfg.Model

	// Load Antigravity config for user-selected model
	agCfg := p.agSvc.LoadConfig()
	if agCfg.SelectedModel != "" {
		modelName = agCfg.SelectedModel
	} else {
		// Auto-pick: find first available Gemini model key from account's quotas
		validModelFound := false
		for _, q := range acc.Quotas {
			if q.Key != "" && q.Available && q.Percentage > 0 {
				modelName = q.Key
				validModelFound = true
				// Prefer gemini-3-flash or similar free/low-tier models first
				if strings.Contains(strings.ToLower(q.Name), "flash") &&
					!strings.Contains(strings.ToLower(q.Name), "image") {
					break // flash models are best default
				}
			}
		}
		if !validModelFound {
			modelName = "gemini-3-flash-agent" // safest fallback if quota is empty
		}
	}

	temp := 0.7
	maxTok := 8192
	if agCfg.Temperature > 0 {
		temp = agCfg.Temperature
	}
	if agCfg.MaxTokens > 0 {
		maxTok = agCfg.MaxTokens
	}

	reqMap := map[string]interface{}{
		"contents": contents,
		"model":    modelName,
	}

	if systemPrompt != "" {
		reqMap["systemInstruction"] = map[string]interface{}{
			"role": "user",
			"parts": []map[string]interface{}{
				{"text": systemPrompt},
			},
		}
	}

	reqMap["generationConfig"] = map[string]interface{}{
		"temperature":     temp,
		"topK":            40,
		"topP":            1.0,
		"maxOutputTokens": maxTok,
	}

	finalPayload := map[string]interface{}{
		"project":     acc.ProjectID,
		"requestId":   fmt.Sprintf("agent/antigravity/dardcor/%d", time.Now().Unix()),
		"request":     reqMap,
		"model":       modelName,
		"userAgent":   "antigravity",
		"requestType": "agent",
	}

	bodyBytes, err := json.Marshal(finalPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal antigravity payload: %v", err)
	}

	// Try endpoints in priority order: sandbox → daily → prod
	endpoints := []string{
		"https://daily-cloudcode-pa.sandbox.googleapis.com/v1internal:generateContent",
		"https://daily-cloudcode-pa.googleapis.com/v1internal:generateContent",
		"https://cloudcode-pa.googleapis.com/v1internal:generateContent",
	}

	var lastErr error
	var respBytes []byte

	for _, endpoint := range endpoints {
		req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(bodyBytes))
		if err != nil {
			lastErr = err
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+acc.AccessToken)
		req.Header.Set("x-client-name", "antigravity")
		req.Header.Set("User-Agent", "antigravity")

		resp, err := p.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("endpoint %s failed: %v", endpoint, err)
			continue
		}
		respBytes, err = io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode == http.StatusUnauthorized {
			// Token expired - try to refresh and retry once
			refreshed, refreshErr := p.agSvc.RefreshToken(acc.Email)
			if refreshErr != nil {
				return nil, fmt.Errorf("token expired (401) and auto-refresh failed: %v. Please re-authenticate in the dashboard", refreshErr)
			}
			acc = refreshed
			// Rebuild request with new token - continue to next endpoint attempt
			lastErr = fmt.Errorf("refreshed token, retrying")
			continue
		}
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("endpoint %s returned %d, trying next", endpoint, resp.StatusCode)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("Antigravity API error (Code %d): %s", resp.StatusCode, string(respBytes))
		}

		// Success - parse response
		// v1internal wraps response in 'response' field
		var googleResp struct {
			Response struct {
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
			} `json:"response"`
		}
		// Also try direct format (some endpoints return candidates directly)
		var directResp struct {
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

		var generatedText string
		var totalTokens int

		if err := json.Unmarshal(respBytes, &googleResp); err == nil && len(googleResp.Response.Candidates) > 0 {
			for _, part := range googleResp.Response.Candidates[0].Content.Parts {
				generatedText += part.Text
			}
			totalTokens = googleResp.Response.UsageMetadata.TotalTokenCount
		} else if err := json.Unmarshal(respBytes, &directResp); err == nil && len(directResp.Candidates) > 0 {
			for _, part := range directResp.Candidates[0].Content.Parts {
				generatedText += part.Text
			}
			totalTokens = directResp.UsageMetadata.TotalTokenCount
		}

		if generatedText == "" {
			return nil, fmt.Errorf("Antigravity returned empty response from %s. Raw: %s", endpoint, string(respBytes[:intMin(200, len(respBytes))]))
		}

		return &LLMResponse{
			Content:  generatedText,
			Model:    modelName,
			Provider: "antigravity",
			Tokens:   totalTokens,
			Duration: time.Since(start).Milliseconds(),
		}, nil
	}

	return nil, fmt.Errorf("all Antigravity endpoints failed. Last error: %v", lastErr)
}

func intMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}
