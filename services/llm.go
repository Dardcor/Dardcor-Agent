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

	bodyBytes, _ := json.Marshal(payload)
	baseURL := p.cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	req, _ := http.NewRequest("POST", baseURL+"/chat/completions", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.cfg.APIKey)
	if p.cfg.Provider == "openrouter" {
		req.Header.Set("HTTP-Referer", "https://dardcor.ai")
		req.Header.Set("X-Title", "Dardcor Agent")
	}

	resp, err := p.client.Do(req)
	if err != nil { return nil, err }
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBytes))
	}

	var result openAIResponseBody
	json.Unmarshal(respBytes, &result)
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("empty choices")
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
	if maxTokens == 0 { maxTokens = 4096 }

	payload := anthropicReqBody{
		Model:     p.cfg.Model,
		MaxTokens: maxTokens,
		Messages:  messages,
		System:    systemPrompt,
	}

	bodyBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.cfg.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(req)
	if err != nil { return nil, err }
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBytes))
	}

	var result anthropicRespBody
	json.Unmarshal(respBytes, &result)
	content := ""
	for _, c := range result.Content {
		if c.Type == "text" { content += c.Text }
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
		if m.Role == "user" { lastMsg = m.Content }
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
	return fmt.Sprintf("**Dardcor Agent** (Local Mode)\n\nConfigure a provider via CLI or Dashboard.")
}

func (p *LLMProvider) callAntigravity(systemPrompt string, messages []LLMMessage, start time.Time) (*LLMResponse, error) {
	if p.agSvc == nil { return nil, fmt.Errorf("antigravity service not initialized") }
	acc, err := p.agSvc.GetActiveAccount()
	if err != nil { return nil, err }

	if acc.AccessToken == "" || (!acc.Expiry.IsZero() && time.Now().After(acc.Expiry.Add(-2*time.Minute))) {
		refreshed, err := p.agSvc.RefreshToken(acc.Email)
		if err != nil { return nil, err }
		acc = refreshed
	}

	if acc.ProjectID == "" {
		p.agSvc.FetchProjectAndQuotas(acc)
		if updated, err := p.agSvc.GetActiveAccount(); err == nil { acc = updated }
	}

	var contents []map[string]interface{}
	for _, m := range messages {
		role := m.Role
		if role == "system" { role = "user" }
		if role == "assistant" { role = "model" }
		contents = append(contents, map[string]interface{}{
			"role": role,
			"parts": []map[string]interface{}{{"text": m.Content}},
		})
	}

	modelName := p.cfg.Model
	agCfg := p.agSvc.LoadConfig()
	if agCfg.SelectedModel != "" { modelName = agCfg.SelectedModel }

	temp, maxTok, thinkBudget := 0.7, 8192, 0
	if agCfg.Temperature > 0 { temp = agCfg.Temperature }
	if agCfg.MaxTokens > 0 { maxTok = agCfg.MaxTokens }
	if agCfg.ThinkingBudget > 0 { thinkBudget = agCfg.ThinkingBudget }

	if maxTok <= thinkBudget {
		maxTok = thinkBudget + 8192
	}

	reqMap := map[string]interface{}{
		"contents": contents,
		"model":    modelName,
	}
	if systemPrompt != "" {
		reqMap["systemInstruction"] = map[string]interface{}{
			"role": "user",
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
		"project": acc.ProjectID,
		"request": reqMap,
		"model":   modelName,
		"requestId": fmt.Sprintf("agent/antigravity/%s/%d", acc.ID[:8], len(messages)),
		"userAgent": "antigravity",
		"requestType": "agent",
	}

	bodyBytes, _ := json.Marshal(finalPayload)
	endpoints := []string{
		"https://daily-cloudcode-pa.googleapis.com/v1internal:generateContent",
		"https://cloudcode-pa.googleapis.com/v1internal:generateContent",
	}

	for _, endpoint := range endpoints {
		fmt.Printf("[Antigravity] Sending chat request to %s (Project: %s, Model: %s)\n", endpoint, acc.ProjectID, modelName)
		req, _ := http.NewRequest("POST", endpoint, bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+acc.AccessToken)
		req.Header.Set("User-Agent", "antigravity")
		
		req.Header.Set("x-client-name", "antigravity")
		req.Header.Set("x-client-version", "3.3.18")
		req.Header.Set("x-machine-id", "dardcor-agent-local")
		
		resp, err := p.client.Do(req)
		if err != nil { 
			fmt.Printf("[Antigravity] Connection error: %v\n", err)
			continue 
		}
		respBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK { 
			fmt.Printf("[Antigravity] API Error %d: %s\n", resp.StatusCode, string(respBytes))
			continue 
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
	return nil, fmt.Errorf("antigravity request failed")
}
