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

func NewLLMProvider(cfg config.AIConfig) *LLMProvider {
	return &LLMProvider{
		cfg: cfg,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func (p *LLMProvider) Complete(systemPrompt string, messages []LLMMessage) (*LLMResponse, error) {
	start := time.Now()
	switch p.cfg.Provider {
	case "anthropic":
		return p.callAnthropic(systemPrompt, messages, start)
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
