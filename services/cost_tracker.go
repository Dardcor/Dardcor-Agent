package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"dardcor-agent/config"
)

// CostTrackerService tracks token usage and estimated cost — inspired by MIAW-CLI cost tracking.
type CostTrackerService struct {
	mu      sync.Mutex
	data    CostData
	filePath string
}

type CostData struct {
	TotalInputTokens  int                    `json:"total_input_tokens"`
	TotalOutputTokens int                    `json:"total_output_tokens"`
	TotalCost         float64                `json:"total_cost"`
	TotalRequests     int                    `json:"total_requests"`
	ByProvider        map[string]*ProviderStats `json:"by_provider"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

type ProviderStats struct {
	InputTokens  int     `json:"input_tokens"`
	OutputTokens int     `json:"output_tokens"`
	Requests     int     `json:"requests"`
	Cost         float64 `json:"cost"`
}

// Pricing per 1M tokens (input/output) in USD
var modelPricing = map[string][2]float64{
	"gpt-4o":              {5.0, 15.0},
	"gpt-4o-mini":         {0.15, 0.6},
	"gpt-4-turbo":         {10.0, 30.0},
	"gpt-3.5-turbo":       {0.5, 1.5},
	"o1":                  {15.0, 60.0},
	"o1-mini":             {3.0, 12.0},
	"o3-mini":             {1.1, 4.4},
	"claude-opus-4-5":     {15.0, 75.0},
	"claude-sonnet-4-5":   {3.0, 15.0},
	"claude-haiku-4-5":    {0.25, 1.25},
	"gemini-2.5-pro":      {1.25, 5.0},
	"gemini-2.5-flash":    {0.075, 0.3},
	"gemini-2.0-flash":    {0.075, 0.3},
	"deepseek-chat":       {0.27, 1.1},
	"deepseek-reasoner":   {0.55, 2.19},
	"llama-3.3-70b-versatile": {0.59, 0.79},
}

func NewCostTrackerService() *CostTrackerService {
	filePath := "database/stats.json"
	if config.AppConfig != nil {
		filePath = filepath.Join(config.AppConfig.DataDir, "stats.json")
	}

	svc := &CostTrackerService{
		filePath: filePath,
		data: CostData{
			ByProvider: make(map[string]*ProviderStats),
		},
	}
	svc.load()
	return svc
}

func (c *CostTrackerService) load() {
	if data, err := os.ReadFile(c.filePath); err == nil {
		json.Unmarshal(data, &c.data)
		if c.data.ByProvider == nil {
			c.data.ByProvider = make(map[string]*ProviderStats)
		}
	}
}

func (c *CostTrackerService) save() {
	os.MkdirAll(filepath.Dir(c.filePath), 0755)
	data, _ := json.MarshalIndent(c.data, "", "  ")
	os.WriteFile(c.filePath, data, 0644)
}

// Track records token usage for a request.
func (c *CostTrackerService) Track(provider, model string, inputTokens, outputTokens int) float64 {
	cost := c.EstimateCost(model, inputTokens, outputTokens)

	c.mu.Lock()
	defer c.mu.Unlock()

	c.data.TotalInputTokens += inputTokens
	c.data.TotalOutputTokens += outputTokens
	c.data.TotalCost += cost
	c.data.TotalRequests++
	c.data.UpdatedAt = time.Now()

	if _, ok := c.data.ByProvider[provider]; !ok {
		c.data.ByProvider[provider] = &ProviderStats{}
	}
	ps := c.data.ByProvider[provider]
	ps.InputTokens += inputTokens
	ps.OutputTokens += outputTokens
	ps.Requests++
	ps.Cost += cost

	c.save()
	return cost
}

// EstimateCost calculates estimated cost in USD.
func (c *CostTrackerService) EstimateCost(model string, inputTokens, outputTokens int) float64 {
	pricing, ok := modelPricing[model]
	if !ok {
		// Default fallback: $1/$3 per 1M tokens
		pricing = [2]float64{1.0, 3.0}
	}
	return float64(inputTokens)/1_000_000*pricing[0] + float64(outputTokens)/1_000_000*pricing[1]
}

func (c *CostTrackerService) GetStats() CostData {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.data
}

func (c *CostTrackerService) FormatStats() string {
	c.mu.Lock()
	d := c.data
	c.mu.Unlock()

	return fmt.Sprintf("Requests: %d | Tokens: %d in / %d out | Est. Cost: $%.4f",
		d.TotalRequests,
		d.TotalInputTokens,
		d.TotalOutputTokens,
		d.TotalCost,
	)
}

func (c *CostTrackerService) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = CostData{ByProvider: make(map[string]*ProviderStats), UpdatedAt: time.Now()}
	c.save()
}
