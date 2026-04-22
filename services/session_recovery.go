package services

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type SessionRecoveryService struct {
	mu       sync.RWMutex
	attempts map[string]*RecoveryAttempt
}

type RecoveryAttempt struct {
	SessionID   string    `json:"session_id"`
	ErrorType   string    `json:"error_type"`
	AttemptedAt time.Time `json:"attempted_at"`
	Success     bool      `json:"success"`
}

type RecoveryErrorType string

const (
	ErrToolResultMissing  RecoveryErrorType = "tool_result_missing"
	ErrContextWindowLimit RecoveryErrorType = "context_window_limit"
	ErrRateLimitExceeded  RecoveryErrorType = "rate_limit_exceeded"
	ErrTokenLimitExceeded RecoveryErrorType = "token_limit_exceeded"
	ErrEmptyResponse      RecoveryErrorType = "empty_response"
	ErrThinkingBlockOrder RecoveryErrorType = "thinking_block_order"
	ErrUnknown            RecoveryErrorType = "unknown"
)

func NewSessionRecoveryService() *SessionRecoveryService {
	return &SessionRecoveryService{
		attempts: make(map[string]*RecoveryAttempt),
	}
}

func (s *SessionRecoveryService) DetectErrorType(errMsg string) RecoveryErrorType {
	lower := strings.ToLower(errMsg)

	if strings.Contains(lower, "tool") && (strings.Contains(lower, "result") || strings.Contains(lower, "missing")) {
		return ErrToolResultMissing
	}
	if strings.Contains(lower, "context") && strings.Contains(lower, "limit") {
		return ErrContextWindowLimit
	}
	if strings.Contains(lower, "context_length") || strings.Contains(lower, "context length") {
		return ErrContextWindowLimit
	}
	if strings.Contains(lower, "rate limit") || strings.Contains(lower, "ratelimit") || strings.Contains(lower, "429") {
		return ErrRateLimitExceeded
	}
	if strings.Contains(lower, "max_tokens") || strings.Contains(lower, "token limit") {
		return ErrTokenLimitExceeded
	}
	if strings.Contains(lower, "thinking") && strings.Contains(lower, "order") {
		return ErrThinkingBlockOrder
	}
	if errMsg == "" {
		return ErrEmptyResponse
	}
	return ErrUnknown
}

func (s *SessionRecoveryService) IsRecoverable(errMsg string) bool {
	t := s.DetectErrorType(errMsg)
	switch t {
	case ErrContextWindowLimit, ErrTokenLimitExceeded, ErrEmptyResponse, ErrThinkingBlockOrder:
		return true
	}
	return false
}

func (s *SessionRecoveryService) BuildRecoveryPrompt(errType RecoveryErrorType, originalPrompt string) string {
	switch errType {
	case ErrContextWindowLimit, ErrTokenLimitExceeded:
		return fmt.Sprintf("[SESSION RECOVERY — Context limit reached]\n\nSummarize the conversation so far in 200 words, then continue with the original task:\n\n%s", originalPrompt)
	case ErrEmptyResponse:
		return fmt.Sprintf("[SESSION RECOVERY — Empty response]\n\nPlease retry: %s", originalPrompt)
	case ErrThinkingBlockOrder:
		return fmt.Sprintf("[SESSION RECOVERY — Message structure error]\n\nRetry the following without using thinking blocks:\n\n%s", originalPrompt)
	default:
		return fmt.Sprintf("[SESSION RECOVERY]\n\nAn error occurred. Retrying: %s", originalPrompt)
	}
}

func (s *SessionRecoveryService) RecordAttempt(sessionID string, errType RecoveryErrorType, success bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.attempts[sessionID] = &RecoveryAttempt{
		SessionID:   sessionID,
		ErrorType:   string(errType),
		AttemptedAt: time.Now(),
		Success:     success,
	}
}

func (s *SessionRecoveryService) GetAttempt(sessionID string) *RecoveryAttempt {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if a, ok := s.attempts[sessionID]; ok {
		copy := *a
		return &copy
	}
	return nil
}

func (s *SessionRecoveryService) TruncateHistory(messages []LLMMessage, maxTokens int) []LLMMessage {
	if len(messages) == 0 {
		return messages
	}

	var systemMsg *LLMMessage
	var convMsgs []LLMMessage
	for i, m := range messages {
		if m.Role == "system" {
			systemMsg = &messages[i]
		} else {
			convMsgs = append(convMsgs, m)
		}
	}

	targetChars := maxTokens * 4
	totalChars := 0

	var kept []LLMMessage
	for i := len(convMsgs) - 1; i >= 0; i-- {
		msgChars := len(convMsgs[i].Content)
		if totalChars+msgChars > targetChars {
			break
		}
		kept = append([]LLMMessage{convMsgs[i]}, kept...)
		totalChars += msgChars
	}

	if systemMsg != nil {
		return append([]LLMMessage{*systemMsg}, kept...)
	}
	return kept
}
