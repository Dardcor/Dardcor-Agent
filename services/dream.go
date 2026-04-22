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

	egoState := s.egoSvc.GetState()
	insight := s.generateAIInsight(egoState)

	egoContext := fmt.Sprintf(" [Ego: %s | Conf: %.2f | Energy: %.2f]", egoState.Status, egoState.Confidence, egoState.Energy)

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	dreamEntry := fmt.Sprintf("[%s]%s DREAM: %s\n", timestamp, egoContext, insight)
	f.WriteString(dreamEntry)

	if s.egoSvc != nil {
		s.egoSvc.RecoverEnergy(0.05)
	}
}

func (s *DreamService) generateAIInsight(state models.EgoState) string {
	var insight string

	successRate := float64(0)
	total := state.TasksComplete + state.TasksFailed
	if total > 0 {
		successRate = float64(state.TasksComplete) / float64(total) * 100
	}

	switch {
	case state.StreakFailed > 3:
		insight = fmt.Sprintf("Critical analysis: %d consecutive failures with %.0f%% overall success rate. Root cause pattern suggests tool selection mismatch. Recommending: 1) Read target files before modification, 2) Use grep to verify assumptions, 3) Test with smaller scope first.", state.StreakFailed, successRate)
	case state.Confidence < 0.3:
		insight = fmt.Sprintf("Recovery analysis: Confidence at %.0f%%. Historical data shows recovery after targeted successes. Recommending focus on high-confidence simple tasks to rebuild momentum. Avoid complex multi-step operations until confidence > 0.5.", state.Confidence*100)
	case state.StreakSuccess > 7:
		insight = fmt.Sprintf("Peak performance detected: %d consecutive successes, %.0f%% success rate. Current strategy is optimal. Cataloging successful patterns for future reference. Energy at %.0f%% — schedule regeneration before depletion.", state.StreakSuccess, successRate, state.Energy*100)
	case state.Energy < 0.3:
		insight = fmt.Sprintf("Energy conservation advisory: %.0f%% remaining. Completed %d total actions. Recommending: consolidate pending work, defer non-critical tasks, prioritize high-impact actions only.", state.Energy*100, state.TotalActions)
	case total > 20:
		insight = fmt.Sprintf("Session analytics: %d actions completed (%.0f%% success). Current streak: %dW/%dL. Workspace patterns suggest optimizing file read caching and reducing redundant searches.", total, successRate, state.StreakSuccess, state.StreakFailed)
	default:
		base := dreamInsights[rand.Intn(len(dreamInsights))]
		insight = fmt.Sprintf("%s [Session stats: %d actions, %.0f%% success, energy %.0f%%]", base, total, successRate, state.Energy*100)
	}

	return insight
}

func (s *DreamService) AnalyzeWorkspace(workspace string) string {
	if s.fsSvc == nil {
		return ""
	}

	files, err := s.fsSvc.ListDirectory(workspace)
	if err != nil {
		return ""
	}

	fileCount := 0
	dirCount := 0
	for _, f := range files {
		if f.IsDir {
			dirCount++
		} else {
			fileCount++
		}
	}

	return fmt.Sprintf("Workspace: %s — %d files, %d directories at root level.", workspace, fileCount, dirCount)
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
