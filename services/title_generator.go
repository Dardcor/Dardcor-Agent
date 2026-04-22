package services

import (
	"log"
	"strings"
)

const (
	titleSystemPrompt = "Generate a short, descriptive title (3-7 words) for a conversation that starts with the following exchange. The title should capture the main topic or intent. Return ONLY the title text, nothing else. No quotes, no punctuation at the end, no prefixes."
	titleMaxInputLen  = 500
	titleMaxLen       = 80
	titleMaxTokens    = 30
	titleTemperature  = 0.3
)

type TitleGeneratorService struct {
	llm *LLMProvider
}

func NewTitleGeneratorService(llm *LLMProvider) *TitleGeneratorService {
	return &TitleGeneratorService{llm: llm}
}

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen])
}

func cleanTitle(raw string) string {
	title := strings.TrimSpace(raw)

	if len(title) >= 2 {
		first, last := title[0], title[len(title)-1]
		if (first == '"' && last == '"') ||
			(first == '\'' && last == '\'') ||
			(first == '`' && last == '`') {
			title = title[1 : len(title)-1]
		}
	}

	if after, found := strings.CutPrefix(title, "Title: "); found {
		title = after
	}
	if after, found := strings.CutPrefix(title, "title: "); found {
		title = after
	}

	title = strings.TrimSpace(title)

	runes := []rune(title)
	if len(runes) > titleMaxLen {
		title = string(runes[:titleMaxLen-3]) + "..."
	}

	return title
}

func (s *TitleGeneratorService) GenerateTitle(userMessage, assistantResponse string) string {
	if s.llm == nil {
		return ""
	}

	userSnippet := truncate(userMessage, titleMaxInputLen)
	assistantSnippet := truncate(assistantResponse, titleMaxInputLen)

	exchangeText := "User: " + userSnippet + "\n\nAssistant: " + assistantSnippet

	messages := []LLMMessage{
		{Role: "user", Content: exchangeText},
	}

	resp, err := s.llm.Complete(titleSystemPrompt, messages)
	if err != nil {
		log.Printf("[TitleGenerator] LLM call failed: %v", err)
		return ""
	}
	if resp == nil || strings.TrimSpace(resp.Content) == "" {
		return ""
	}

	return cleanTitle(resp.Content)
}

func (s *TitleGeneratorService) MaybeAutoTitle(
	conversationID string,
	userMessage string,
	assistantResponse string,
	messageCount int,
	onTitle func(conversationID, title string),
) {
	if messageCount > 2 {
		return
	}

	go func() {
		title := s.GenerateTitle(userMessage, assistantResponse)
		if title == "" {
			return
		}
		if onTitle != nil {
			onTitle(conversationID, title)
		}
	}()
}
