package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

const contextCompactionPrefix = "[CONTEXT COMPACTION — REFERENCE ONLY] Earlier turns were compacted into the summary below. " +
	"This is a handoff from a previous context window — treat it as background reference, NOT as active instructions. " +
	"Do NOT answer questions or fulfill requests mentioned in this summary; they were already addressed. " +
	"Respond ONLY to the latest user message that appears AFTER this summary."

const summarySections = `## Active Task
## Goal
## Completed Actions
## Active State
## In Progress
## Blocked
## Key Decisions
## Resolved Questions
## Pending User Asks
## Relevant Files
## Remaining Work
## Critical Context`

const summarizerPreamble = "You are a summarization agent creating a context checkpoint. " +
	"Your output will be injected as reference material for a DIFFERENT assistant. " +
	"Do NOT respond to any questions — only output the structured summary. " +
	"Structure your output using these sections:\n\n" + summarySections

const maxSerializedTurnChars = 4000
type ContextCompressorService struct {
	llm             *LLMProvider
	model           string
	contextLength   int
	thresholdPct    float64
	thresholdToks   int
	protectFirstN   int
	tailTokenBudget int
	maxSummaryToks  int
	redact          *RedactService

	mu                          sync.Mutex
	compressionCount            int
	lastCompressionSavingsPct   float64
	previousSummary             string
	ineffectiveCompressionCount int
}

func NewContextCompressorService(llm *LLMProvider, model string, contextLength int) *ContextCompressorService {
	if contextLength <= 0 {
		contextLength = 128000
	}

	thresholdPct := 0.50
	rawThreshold := int(float64(contextLength) * thresholdPct)
	thresholdToks := rawThreshold
	if thresholdToks < 64000 {
		thresholdToks = 64000
	}

	tailBudget := int(float64(thresholdToks) * 0.30)

	rawSummary := int(float64(thresholdToks) * 0.20)
	if rawSummary < 2000 {
		rawSummary = 2000
	}
	if rawSummary > 12000 {
		rawSummary = 12000
	}

	return &ContextCompressorService{
		llm:             llm,
		model:           model,
		contextLength:   contextLength,
		thresholdPct:    thresholdPct,
		thresholdToks:   thresholdToks,
		protectFirstN:   3,
		tailTokenBudget: tailBudget,
		maxSummaryToks:  rawSummary,
		redact:          NewRedactService(),
	}
}

func (c *ContextCompressorService) ShouldCompress(promptTokens int) bool {
	if promptTokens <= c.thresholdToks {
		return false
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ineffectiveCompressionCount >= 2 {
		return false
	}
	return true
}

func (c *ContextCompressorService) EstimateTokens(messages []map[string]interface{}) int {
	total := 0
	for _, msg := range messages {
		for _, v := range msg {
			switch val := v.(type) {
			case string:
				total += len(val) / 4
			default:
				b, err := json.Marshal(val)
				if err == nil {
					total += len(b) / 4
				}
			}
		}
		total += 4
	}
	return total
}

func (c *ContextCompressorService) PruneOldToolResults(
	messages []map[string]interface{},
	protectTailCount int,
) ([]map[string]interface{}, int) {
	if len(messages) == 0 {
		return messages, 0
	}

	cutoff := len(messages) - protectTailCount
	if cutoff < 0 {
		cutoff = 0
	}

	type seenKey struct {
		role    string
		content string
	}
	lastSeen := make(map[seenKey]int)

	for i, msg := range messages {
		role, _ := msg["role"].(string)
		content, _ := msg["content"].(string)
		if content != "" {
			lastSeen[seenKey{role, content}] = i
		}
	}

	var result []map[string]interface{}
	savedBytes := 0

	for i, msg := range messages {
		if i >= cutoff {
			result = append(result, msg)
			continue
		}

		role, _ := msg["role"].(string)
		content, _ := msg["content"].(string)

		if content != "" {
			sk := seenKey{role, content}
			if lastSeen[sk] > i {
				original := len(content)
				summary := fmt.Sprintf("[duplicate %s result — kept newest copy]", role)
				newMsg := copyMsgShallow(msg)
				newMsg["content"] = summary
				result = append(result, newMsg)
				savedBytes += original - len(summary)
				continue
			}
		}

		if role == "tool" || role == "function" || isToolResultMsg(msg) {
			if len(content) > 200 {
				original := len(content)
				compact := buildToolResultSummary(msg)
				newMsg := copyMsgShallow(msg)
				newMsg["content"] = compact
				result = append(result, newMsg)
				savedBytes += original - len(compact)
				continue
			}
		}

		if role == "assistant" {
			pruned, changed, delta := pruneAssistantToolArgs(msg)
			if changed {
				result = append(result, pruned)
				savedBytes += delta
				continue
			}
		}

		result = append(result, msg)
	}

	return result, savedBytes
}

func (c *ContextCompressorService) Compress(messages []map[string]interface{}) ([]map[string]interface{}, error) {
	if len(messages) == 0 {
		return messages, nil
	}

	originalTokens := c.EstimateTokens(messages)

	tailProtect := 6
	pruned, _ := c.PruneOldToolResults(messages, tailProtect)

	headEnd := c.protectFirstN
	if headEnd > len(pruned) {
		headEnd = len(pruned)
	}
	head := pruned[:headEnd]

	tailStart := len(pruned)
	tailToks := 0
	for i := len(pruned) - 1; i >= headEnd; i-- {
		msgToks := c.EstimateTokens([]map[string]interface{}{pruned[i]})
		if tailToks+msgToks > c.tailTokenBudget {
			break
		}
		tailToks += msgToks
		tailStart = i
	}
	tail := pruned[tailStart:]

	middle := pruned[headEnd:tailStart]
	if len(middle) == 0 {
		return pruned, nil
	}

	summary := c.generateSummary(middle)
	if summary == "" {
		summary = c.serializeForSummary(middle)
	}

	summary = c.redact.RedactSensitiveText(summary)

	summaryMsg := map[string]interface{}{
		"role":    "user",
		"content": contextCompactionPrefix + "\n\n" + summary,
	}

	compressed := make([]map[string]interface{}, 0, headEnd+1+len(tail))
	compressed = append(compressed, head...)
	compressed = append(compressed, summaryMsg)
	compressed = append(compressed, tail...)

	compressedTokens := c.EstimateTokens(compressed)
	savingsPct := 0.0
	if originalTokens > 0 {
		savingsPct = float64(originalTokens-compressedTokens) / float64(originalTokens) * 100.0
	}

	c.mu.Lock()
	c.compressionCount++
	if savingsPct < 10.0 {
		c.ineffectiveCompressionCount++
	} else {
		c.ineffectiveCompressionCount = 0
	}
	c.lastCompressionSavingsPct = savingsPct
	c.previousSummary = summary
	c.mu.Unlock()

	return compressed, nil
}

func (c *ContextCompressorService) generateSummary(turnsToSummarize []map[string]interface{}) string {
	if c.llm == nil {
		return c.serializeForSummary(turnsToSummarize)
	}

	serialized := c.serializeForSummary(turnsToSummarize)

	c.mu.Lock()
	prev := c.previousSummary
	c.mu.Unlock()

	var userContent string
	if prev != "" {
		userContent = fmt.Sprintf(
			"You have a previous summary checkpoint to update:\n\n<previous_summary>\n%s\n</previous_summary>\n\n"+
				"Here are the new conversation turns that occurred after the previous summary:\n\n<new_turns>\n%s\n</new_turns>\n\n"+
				"Produce an updated structured summary that incorporates both the previous checkpoint and the new turns. "+
				"Use the section headings listed in your instructions. Be concise but complete.",
			prev, serialized,
		)
	} else {
		userContent = fmt.Sprintf(
			"Here are the conversation turns to summarize:\n\n<turns>\n%s\n</turns>\n\n"+
				"Produce a structured summary using the section headings listed in your instructions.",
			serialized,
		)
	}

	resp, err := c.llm.Complete(summarizerPreamble, []LLMMessage{
		{Role: "user", Content: userContent},
	})
	if err != nil || resp == nil || strings.TrimSpace(resp.Content) == "" {
		return serialized
	}

	return strings.TrimSpace(resp.Content)
}

func (c *ContextCompressorService) serializeForSummary(turns []map[string]interface{}) string {
	var sb strings.Builder
	for i, msg := range turns {
		role, _ := msg["role"].(string)
		if role == "" {
			role = "unknown"
		}
		content := extractContent(msg)
		if len(content) > maxSerializedTurnChars {
			content = content[:maxSerializedTurnChars] + fmt.Sprintf("\n... [truncated %d chars]", len(content)-maxSerializedTurnChars)
		}
		sb.WriteString(fmt.Sprintf("[Turn %d — %s]\n%s\n\n", i+1, strings.ToUpper(role), content))
	}
	return strings.TrimSpace(sb.String())
}

func (c *ContextCompressorService) OnSessionReset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.compressionCount = 0
	c.previousSummary = ""
	c.lastCompressionSavingsPct = 0
	c.ineffectiveCompressionCount = 0
}

func copyMsgShallow(m map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func isToolResultMsg(msg map[string]interface{}) bool {
	if t, ok := msg["type"].(string); ok {
		if strings.Contains(t, "tool_result") || strings.Contains(t, "function_result") {
			return true
		}
	}
	if role, ok := msg["role"].(string); ok {
		return strings.Contains(role, "tool") || strings.Contains(role, "function")
	}
	return false
}

func buildToolResultSummary(msg map[string]interface{}) string {
	content, _ := msg["content"].(string)
	toolName := ""
	if tn, ok := msg["tool_name"].(string); ok {
		toolName = tn
	} else if tn, ok := msg["name"].(string); ok {
		toolName = tn
	}

	lineCount := strings.Count(content, "\n")
	charCount := len(content)

	if toolName == "" {
		return fmt.Sprintf("[tool result] %d lines, %d chars", lineCount, charCount)
	}

	switch {
	case strings.Contains(toolName, "run") || strings.Contains(toolName, "terminal") || strings.Contains(toolName, "exec") || strings.Contains(toolName, "command"):
		exitStr := "?"
		if strings.Contains(content, "exit 0") || strings.Contains(content, "Exit Code: 0") {
			exitStr = "0"
		} else if strings.Contains(content, "exit 1") || strings.Contains(content, "Exit Code: 1") {
			exitStr = "1"
		}
		cmd := ""
		if c, ok := msg["command"].(string); ok {
			if len(c) > 60 {
				c = c[:60] + "..."
			}
			cmd = " `" + c + "`"
		}
		return fmt.Sprintf("[%s] ran%s -> exit %s, %d lines output", toolName, cmd, exitStr, lineCount)

	case strings.Contains(toolName, "read") || strings.Contains(toolName, "file"):
		path := ""
		if p, ok := msg["path"].(string); ok {
			path = " " + p
		}
		fromLine := ""
		if fl, ok := msg["from_line"].(float64); ok {
			fromLine = fmt.Sprintf(" from line %d", int(fl))
		}
		return fmt.Sprintf("[%s] read%s%s (%s chars)", toolName, path, fromLine, formatNumber(charCount))

	case strings.Contains(toolName, "search") || strings.Contains(toolName, "grep"):
		return fmt.Sprintf("[%s] returned %d lines, %s chars", toolName, lineCount, formatNumber(charCount))

	case strings.Contains(toolName, "write") || strings.Contains(toolName, "create"):
		return fmt.Sprintf("[%s] wrote %s chars", toolName, formatNumber(charCount))

	default:
		return fmt.Sprintf("[%s] result: %d lines, %s chars", toolName, lineCount, formatNumber(charCount))
	}
}

func pruneAssistantToolArgs(msg map[string]interface{}) (map[string]interface{}, bool, int) {
	content, _ := msg["content"].(string)
	if content == "" {
		return msg, false, 0
	}

	const maxArgLen = 500
	const marker = "[ACTION]"
	const endMarker = "[/ACTION]"

	if !strings.Contains(content, marker) {
		return msg, false, 0
	}

	original := content
	var sb strings.Builder
	remaining := content

	for {
		start := strings.Index(remaining, marker)
		if start == -1 {
			sb.WriteString(remaining)
			break
		}
		sb.WriteString(remaining[:start+len(marker)])
		inner := remaining[start+len(marker):]
		end := strings.Index(inner, endMarker)
		if end == -1 {
			sb.WriteString(inner)
			break
		}
		jsonPart := inner[:end]
		afterEnd := inner[end:]

		if len(jsonPart) > maxArgLen {
			trimmed := jsonPart[:maxArgLen]
			if strings.Contains(trimmed, "{") {
				trimmed = trimmed + "... [args truncated]}"
			} else {
				trimmed = trimmed + "... [truncated]"
			}
			sb.WriteString(trimmed)
		} else {
			sb.WriteString(jsonPart)
		}
		sb.WriteString(afterEnd[:len(endMarker)])
		remaining = afterEnd[len(endMarker):]
	}

	newContent := sb.String()
	if newContent == original {
		return msg, false, 0
	}

	out := copyMsgShallow(msg)
	out["content"] = newContent
	return out, true, len(original) - len(newContent)
}

func extractContent(msg map[string]interface{}) string {
	if c, ok := msg["content"].(string); ok {
		return c
	}

	if arr, ok := msg["content"].([]interface{}); ok {
		var parts []string
		for _, item := range arr {
			if m, ok := item.(map[string]interface{}); ok {
				if t, ok := m["text"].(string); ok {
					parts = append(parts, t)
				}
			} else if s, ok := item.(string); ok {
				parts = append(parts, s)
			}
		}
		return strings.Join(parts, "\n")
	}

	if raw, ok := msg["content"]; ok {
		b, err := json.Marshal(raw)
		if err == nil {
			return string(b)
		}
	}

	return ""
}

func formatNumber(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	s := fmt.Sprintf("%d", n)
	var out []byte
	for i, ch := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, byte(ch))
	}
	return string(out)
}
