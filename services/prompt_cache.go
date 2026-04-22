package services

import "encoding/json"

type PromptCacheService struct{}

func NewPromptCacheService() *PromptCacheService {
	return &PromptCacheService{}
}

func (s *PromptCacheService) ApplyAnthropicCacheControl(
	messages []map[string]interface{},
	cacheTTL string,
	nativeAnthropic bool,
) []map[string]interface{} {
	raw, err := json.Marshal(messages)
	if err != nil {
		return messages
	}
	var copied []map[string]interface{}
	if err := json.Unmarshal(raw, &copied); err != nil {
		return messages
	}

	marker := map[string]interface{}{
		"type": "ephemeral",
	}
	if cacheTTL == "1h" {
		marker["ttl"] = "1h"
	}

	systemIdx := -1
	for i, msg := range copied {
		role, _ := msg["role"].(string)
		if role == "system" {
			systemIdx = i
			break
		}
	}

	var nonSystemIndices []int
	for i, msg := range copied {
		role, _ := msg["role"].(string)
		if role != "system" {
			nonSystemIndices = append(nonSystemIndices, i)
		}
	}

	markIndices := make(map[int]bool)
	if systemIdx >= 0 {
		markIndices[systemIdx] = true
	}
	start := len(nonSystemIndices) - 3
	if start < 0 {
		start = 0
	}
	for _, idx := range nonSystemIndices[start:] {
		markIndices[idx] = true
	}

	for idx := range markIndices {
		msg := copied[idx]
		role, _ := msg["role"].(string)

		if role == "tool" && nativeAnthropic {
			msg["cache_control"] = marker
			continue
		}

		content, hasContent := msg["content"]

		if !hasContent || content == nil {
			msg["cache_control"] = marker
			continue
		}

		switch v := content.(type) {
		case string:
			msg["content"] = []interface{}{
				map[string]interface{}{
					"type":          "text",
					"text":          v,
					"cache_control": marker,
				},
			}

		case []interface{}:
			if len(v) == 0 {
				msg["cache_control"] = marker
			} else {
				last := v[len(v)-1]
				if block, ok := last.(map[string]interface{}); ok {
					block["cache_control"] = marker
				}
			}

		default:
			msg["cache_control"] = marker
		}
	}

	return copied
}
