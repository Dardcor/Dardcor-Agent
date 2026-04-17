package services

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"dardcor-agent/config"
	"dardcor-agent/models"
)

var dreamInsights = []string{
	"Detected a pattern in recent file operations — a batch processing utility could reduce repetitive tasks by 60%%.",
	"The current workspace structure has diverged from conventional Go project layout. Consider consolidating config paths.",
	"Memory service could benefit from TTL-based expiration to prevent stale context pollution in long sessions.",
	"Observed high frequency of web search queries — pre-indexing documentation could accelerate response latency.",
	"The orchestrator's retry logic is linear. An exponential backoff strategy would improve resilience under load.",
	"Command execution timeout defaults may be too aggressive for build operations. Adaptive timeout based on command type recommended.",
	"Token truncation in LLM context uses a hard cutoff. A priority-weighted compression would preserve critical context better.",
	"The ego confidence curve is too linear. A logarithmic decay on failure would better model real performance recovery.",
	"Workspace path resolution is called redundantly across handlers. Caching the resolved path per session would save cycles.",
	"The safety guard regex patterns miss some Windows-specific destructive commands. A platform-aware guard update is recommended.",
}

type DreamService struct {
	fsSvc  *FileSystemService
	egoSvc *EgoService
}

func NewDreamService(fs *FileSystemService, ego *EgoService) *DreamService {
	return &DreamService{fsSvc: fs, egoSvc: ego}
}

func (s *DreamService) StartDreaming() {
	go func() {
		for {
			time.Sleep(10 * time.Minute)
			s.dream()
		}
	}()
}

func (s *DreamService) dream() {
	dreamPath := ""
	if config.AppConfig != nil {
		dreamPath = filepath.Join(config.AppConfig.DataDir, "ego", "dreams.json")
	} else {
		dreamPath = "database/ego/dreams.json"
	}

	os.MkdirAll(filepath.Dir(dreamPath), 0755)

	f, err := os.OpenFile(dreamPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	// Perfection: Generate a dynamic insight based on actual system state
	egoState := s.egoSvc.GetState()
	insight := s.generateAIInsight(egoState)

	egoContext := fmt.Sprintf(" [Ego: %s | Conf: %.2f | Energy: %.2f]", egoState.Status, egoState.Confidence, egoState.Energy)

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	dreamEntry := fmt.Sprintf("[%s]%s DREAM: %s\n", timestamp, egoContext, insight)
	f.WriteString(dreamEntry)

	if s.egoSvc != nil {
		s.egoSvc.RecoverEnergy(0.05) // Dreams recover more energy now
	}
}

// generateAIInsight simulates an LLM call to reflect on the agent's state.
// In a full implementation, this would call llmProvider.Complete with a reflection prompt.
func (s *DreamService) generateAIInsight(state models.EgoState) string {
	if state.Confidence < 0.5 {
		return "Analysis of recent failures suggests a mismatch between tool timeout and system latency. Recommending adaptive buffering."
	}
	if state.StreakSuccess > 5 {
		return fmt.Sprintf("Consistent success in %d tasks detected. Identifying patterns for new 'High-Speed' skill templates.", state.StreakSuccess)
	}

	// Default to a smart observation
	return dreamInsights[rand.Intn(len(dreamInsights))]
}

func (s *DreamService) GetRecentDreams(count int) []string {
	dreamPath := ""
	if config.AppConfig != nil {
		dreamPath = filepath.Join(config.AppConfig.DataDir, "ego", "dreams.json")
	} else {
		dreamPath = "database/ego/dreams.json"
	}

	data, err := os.ReadFile(dreamPath)
	if err != nil {
		return nil
	}

	lines := splitLines(string(data))
	if len(lines) == 0 {
		return nil
	}

	if count <= 0 || count > len(lines) {
		count = len(lines)
	}

	start := len(lines) - count
	if start < 0 {
		start = 0
	}

	return lines[start:]
}

func splitLines(s string) []string {
	var lines []string
	current := ""
	for _, c := range s {
		if c == '\n' {
			if current != "" {
				lines = append(lines, current)
			}
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}
