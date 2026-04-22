package services

import (
	"context"
	"fmt"
	"strings"
)

type MemoryCompressor struct {
	llm *LLMProvider
}

func NewMemoryCompressor(llm *LLMProvider) *MemoryCompressor {
	return &MemoryCompressor{llm: llm}
}

func (c *MemoryCompressor) CompressObservations(_ context.Context, observations []string, project string) (string, error) {
	if len(observations) == 0 {
		return "", nil
	}

	joined := strings.Join(observations, "\n")
	if len(joined) > 8000 {
		joined = joined[:8000]
	}

	if project != "" {
		joined = fmt.Sprintf("[Project: %s]\n\n%s", project, joined)
	}

	systemPrompt := "You are a memory compression system. Summarize the following observations into a concise, actionable summary. Focus on decisions made, problems solved, patterns used, and important context. Be brief but complete."

	resp, err := c.llm.Complete(systemPrompt, []LLMMessage{
		{Role: "user", Content: joined},
	})
	if err != nil || resp == nil || strings.TrimSpace(resp.Content) == "" {
		return c.fallback(joined), nil
	}
	return strings.TrimSpace(resp.Content), nil
}

func (c *MemoryCompressor) CompressSession(_ context.Context, sessionObservations []string, project string) (string, error) {
	if len(sessionObservations) == 0 {
		return "", nil
	}

	joined := strings.Join(sessionObservations, "\n")
	if len(joined) > 8000 {
		joined = joined[:8000]
	}

	if project != "" {
		joined = fmt.Sprintf("[Project: %s]\n\n%s", project, joined)
	}

	systemPrompt := "Summarize this coding session into 3-5 key points. Include: what was accomplished, key technical decisions, files changed, and any important patterns or lessons learned."

	resp, err := c.llm.Complete(systemPrompt, []LLMMessage{
		{Role: "user", Content: joined},
	})
	if err != nil || resp == nil || strings.TrimSpace(resp.Content) == "" {
		return c.fallback(joined), nil
	}
	return strings.TrimSpace(resp.Content), nil
}

func (c *MemoryCompressor) SummarizeObservation(_ context.Context, content, obsType string) (string, error) {
	if content == "" {
		return "", nil
	}

	systemPrompt := fmt.Sprintf("Summarize this %s in 1-2 sentences. Be specific and actionable.", obsType)

	resp, err := c.llm.Complete(systemPrompt, []LLMMessage{
		{Role: "user", Content: content},
	})
	if err != nil || resp == nil || strings.TrimSpace(resp.Content) == "" {
		return c.fallback(content), nil
	}
	return strings.TrimSpace(resp.Content), nil
}

func (c *MemoryCompressor) fallback(input string) string {
	if len(input) <= 500 {
		return input
	}
	return input[:500]
}
