package services

import (
	"strings"
	"sync"
)

// ThinkModeService detects "think" keywords in prompts and augments
// them with deep reasoning instructions — inspired by oh-my-openagent.
type ThinkModeService struct {
	mu       sync.RWMutex
	sessions map[string]*ThinkModeState
}

type ThinkModeState struct {
	Requested bool   `json:"requested"`
	SessionID string `json:"session_id"`
}

var thinkKeywords = []string{
	"think", "thinking", "reason", "reasoning",
	"reflect", "contemplate", "analyze deeply",
	"step by step", "step-by-step", "chain of thought",
	"cot", "ultrawork", "supreme",
}

func NewThinkModeService() *ThinkModeService {
	return &ThinkModeService{sessions: make(map[string]*ThinkModeState)}
}

func (t *ThinkModeService) DetectThinkKeyword(prompt string) bool {
	lower := strings.ToLower(prompt)
	for _, kw := range thinkKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// AugmentPrompt injects deep-reasoning instructions when think mode is active.
func (t *ThinkModeService) AugmentPrompt(sessionID, prompt string) string {
	if !t.DetectThinkKeyword(prompt) {
		return prompt
	}

	t.mu.Lock()
	t.sessions[sessionID] = &ThinkModeState{Requested: true, SessionID: sessionID}
	t.mu.Unlock()

	prefix := `[THINK MODE ACTIVE]
Before answering, work through your reasoning step-by-step:
1. Restate the problem in your own words
2. Identify assumptions and constraints
3. Consider multiple approaches
4. Evaluate trade-offs
5. Choose the best approach and explain why

Now respond:

`
	return prefix + prompt
}

func (t *ThinkModeService) GetState(sessionID string) *ThinkModeState {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.sessions[sessionID]
}

func (t *ThinkModeService) ClearState(sessionID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.sessions, sessionID)
}
