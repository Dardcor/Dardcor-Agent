package services

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const MINIMUM_CONTEXT_LENGTH = 64_000

type ModelInfo struct {
	ContextLength       int
	MaxCompletionTokens int
	Name                string
	Pricing             *ModelPricing
}

type ModelPricing struct {
	PromptPerMToken     float64
	CompletionPerMToken float64
	CacheReadPerMToken  float64
	CacheWritePerMToken float64
}

var defaultContextLengths = map[string]int{
	"claude-sonnet-4-7": 1_000_000,
	"claude-opus-4-7":   1_000_000,
	"claude-haiku-4-7":  1_000_000,
	"claude-sonnet-4-6": 1_000_000,
	"claude-opus-4-6":   1_000_000,
	"claude-haiku-4-6":  1_000_000,

	"claude": 200_000,

	"gpt-5.4": 1_050_000,

	"gpt-5.1": 400_000,
	"gpt-5.2": 400_000,
	"gpt-5.3": 400_000,
	"gpt-5":   400_000,

	"gpt-4.1": 1_047_576,

	"gpt-4": 128_000,

	"gemini": 1_048_576,

	"deepseek": 128_000,

	"llama": 131_072,

	"qwen3.5-coder-plus": 1_000_000,
	"qwen3-coder-plus":   1_000_000,
	"qwen3.5-coder":      262_144,
	"qwen3-coder":        262_144,
	"qwen":               131_072,

	"minimax": 204_800,

	"glm": 202_752,

	"grok-4-1-fast": 2_000_000,
	"grok-4-fast":   2_000_000,
	"grok-4":        256_000,
	"grok-3":        131_072,
	"grok":          131_072,

	"kimi": 262_144,

	"nemotron": 131_072,
}

var knownProviderPrefixes = []string{
	"openrouter",
	"nous",
	"openai-codex",
	"copilot",
	"gemini",
	"anthropic",
	"deepseek",
	"ollama",
	"custom",
	"local",
	"google",
	"azure",
	"bedrock",
	"vertex",
	"cohere",
	"mistral",
	"groq",
	"together",
	"fireworks",
	"perplexity",
	"huggingface",
	"replicate",
}

var contextProbeTiers = []int{128_000, 64_000, 32_000, 16_000, 8_000}

var contextLimitPattern = regexp.MustCompile(
	`(?i)(?:context(?:\s+length|\s+window)?|maximum\s+(?:context\s+)?(?:length|tokens?)|token\s+limit)[^\d]*(\d{4,7})`,
)

type ModelMetadataService struct {
	mu        sync.RWMutex
	cache     map[string]*ModelInfo
	cacheTime time.Time
	cacheTTL  time.Duration
}

func NewModelMetadataService() *ModelMetadataService {
	return &ModelMetadataService{
		cache:    make(map[string]*ModelInfo),
		cacheTTL: 1 * time.Hour,
	}
}

func StripProviderPrefix(model string) string {
	idx := strings.Index(model, ":")
	if idx < 0 {
		return model
	}

	prefix := strings.ToLower(model[:idx])
	for _, p := range knownProviderPrefixes {
		if prefix == p {
			return model[idx+1:]
		}
	}

	return model
}

func (s *ModelMetadataService) GetModelContextLength(model string) int {
	clean := strings.ToLower(StripProviderPrefix(model))

	if length, ok := defaultContextLengths[clean]; ok {
		return length
	}

	bestLen := 0
	bestMatch := 0
	for key, length := range defaultContextLengths {
		if strings.Contains(clean, key) && len(key) > bestLen {
			bestLen = len(key)
			bestMatch = length
		}
	}
	if bestMatch > 0 {
		return bestMatch
	}

	return 128_000
}

func (s *ModelMetadataService) GetModelInfo(model string) *ModelInfo {
	s.mu.RLock()
	info, hit := s.cache[model]
	expired := time.Since(s.cacheTime) > s.cacheTTL
	s.mu.RUnlock()

	if hit && !expired {
		return info
	}

	info = &ModelInfo{
		Name:                model,
		ContextLength:       s.GetModelContextLength(model),
		MaxCompletionTokens: 0,
	}

	s.mu.Lock()
	s.cache[model] = info
	s.cacheTime = time.Now()
	s.mu.Unlock()

	return info
}

func EstimateTokensRough(text string) int {
	if len(text) == 0 {
		return 0
	}
	return (len(text) + 3) / 4
}

func EstimateMessagesTokensRough(messages []map[string]interface{}) int {
	total := 0
	for _, msg := range messages {
		switch v := msg["content"].(type) {
		case string:
			total += EstimateTokensRough(v)
		case []interface{}:
			for _, block := range v {
				if bMap, ok := block.(map[string]interface{}); ok {
					if text, ok := bMap["text"].(string); ok {
						total += EstimateTokensRough(text)
					}
				}
			}
		}

		if tc, ok := msg["tool_calls"]; ok {
			switch calls := tc.(type) {
			case []interface{}:
				for _, call := range calls {
					if callMap, ok := call.(map[string]interface{}); ok {
						if fn, ok := callMap["function"].(map[string]interface{}); ok {
							if args, ok := fn["arguments"].(string); ok {
								total += EstimateTokensRough(args)
							} else if fn["arguments"] != nil {
								if b, err := json.Marshal(fn["arguments"]); err == nil {
									total += EstimateTokensRough(string(b))
								}
							}
						}
					}
				}
			}
		}

		total += 4
	}
	return total
}

func GetContextProbeTiers() []int {
	result := make([]int, len(contextProbeTiers))
	copy(result, contextProbeTiers)
	return result
}

func GetNextProbeTier(currentLength int) int {
	for _, tier := range contextProbeTiers {
		if tier < currentLength {
			return tier
		}
	}
	return 0
}

func ParseContextLimitFromError(errMsg string) int {
	matches := contextLimitPattern.FindStringSubmatch(errMsg)
	if len(matches) < 2 {
		return 0
	}
	n, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0
	}
	return n
}
