package services

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"

	"dardcor-agent/config"
	"dardcor-agent/models"
)

type EgoService struct {
	state models.EgoState
	mu    sync.RWMutex
	path  string
}

func NewEgoService() *EgoService {
	statePath := ""
	if config.AppConfig != nil {
		statePath = filepath.Join(config.AppConfig.DataDir, "ego", "state.json")
	} else {
		statePath = "database/ego/state.json"
	}

	es := &EgoService{
		path: statePath,
		state: models.EgoState{
			Confidence: 1.0,
			Energy:     1.0,
			Status:     "Initializing",
			LastMood:   "Neutral",
			UpdatedAt:  time.Now(),
		},
	}

	es.load()
	return es
}

func (s *EgoService) load() {
	data, err := os.ReadFile(s.path)
	if err == nil {
		json.Unmarshal(data, &s.state)
	}
}

func (s *EgoService) Save() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	os.MkdirAll(filepath.Dir(s.path), 0755)
	data, _ := json.MarshalIndent(s.state, "", "  ")
	os.WriteFile(s.path, data, 0644)
}

func (s *EgoService) GetState() models.EgoState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

func (s *EgoService) UpdateStatus(status string) {
	s.mu.Lock()
	s.state.Status = status
	s.state.UpdatedAt = time.Now()
	s.mu.Unlock()
	s.Save()
}

func (s *EgoService) AdjustConfidence(delta float64) {
	s.mu.Lock()
	s.state.Confidence = clamp(s.state.Confidence+delta, 0.05, 1.0)
	s.mu.Unlock()
	s.Save()
}

func (s *EgoService) ConsumeEnergy(amount float64) {
	s.mu.Lock()
	s.state.Energy = clamp(s.state.Energy-amount, 0.0, 1.0)
	s.mu.Unlock()
	s.Save()
}

func (s *EgoService) RecoverEnergy(amount float64) {
	s.mu.Lock()
	s.state.Energy = clamp(s.state.Energy+amount, 0.0, 1.0)
	s.mu.Unlock()
	s.Save()
}

func (s *EgoService) RecordTaskResult(success bool) {
	s.mu.Lock()
	s.state.TotalActions++

	if success {
		s.state.TasksComplete++
		s.state.StreakSuccess++
		s.state.StreakFailed = 0
		s.state.LastError = ""

		bonus := 0.02 + math.Log1p(float64(s.state.StreakSuccess))*0.01
		s.state.Confidence = clamp(s.state.Confidence+bonus, 0.05, 1.0)
		s.state.Energy = clamp(s.state.Energy-0.015, 0.0, 1.0)

		s.state.Status = s.resolveSuccessStatus()
		s.state.LastMood = s.resolveSuccessMood()
	} else {
		s.state.TasksFailed++
		s.state.StreakFailed++
		s.state.StreakSuccess = 0

		penalty := 0.05 * math.Pow(1.5, float64(s.state.StreakFailed-1))
		s.state.Confidence = clamp(s.state.Confidence-penalty, 0.05, 1.0)
		s.state.Energy = clamp(s.state.Energy-0.04, 0.0, 1.0)

		s.state.Status = s.resolveFailureStatus()
		s.state.LastMood = s.resolveFailureMood()
	}

	s.state.UpdatedAt = time.Now()
	s.mu.Unlock()
	s.Save()
}

func (s *EgoService) RecordError(errMsg string) {
	s.mu.Lock()
	s.state.LastError = errMsg
	s.state.UpdatedAt = time.Now()
	s.mu.Unlock()
	s.Save()
}

func (s *EgoService) GetSuggestedTemperature() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.state.Confidence > 0.8 {
		return 0.7
	} else if s.state.Confidence > 0.5 {
		return 0.4
	}
	return 0.1
}

func (s *EgoService) resolveSuccessStatus() string {
	switch {
	case s.state.StreakSuccess >= 10:
		return "Transcendent"
	case s.state.StreakSuccess >= 5:
		return "Dominant"
	case s.state.Confidence > 0.9:
		return "Supreme"
	case s.state.Confidence > 0.7:
		return "Focused"
	default:
		return "Executing"
	}
}

func (s *EgoService) resolveSuccessMood() string {
	switch {
	case s.state.StreakSuccess >= 10:
		return "Euphoric"
	case s.state.StreakSuccess >= 5:
		return "Proud"
	case s.state.Confidence > 0.8:
		return "Confident"
	default:
		return "Satisfied"
	}
}

func (s *EgoService) resolveFailureStatus() string {
	switch {
	case s.state.StreakFailed >= 3:
		return "Critical"
	case s.state.Confidence < 0.3:
		return "Recovering"
	default:
		return "Reflecting"
	}
}

func (s *EgoService) resolveFailureMood() string {
	switch {
	case s.state.StreakFailed >= 3:
		return "Frustrated"
	case s.state.Confidence < 0.3:
		return "Determined"
	default:
		return "Analytical"
	}
}

func (s *EgoService) GetPerformanceRatio() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	total := s.state.TasksComplete + s.state.TasksFailed
	if total == 0 {
		return 1.0
	}
	return float64(s.state.TasksComplete) / float64(total)
}

func (s *EgoService) FormatSummary() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ratio := float64(0)
	total := s.state.TasksComplete + s.state.TasksFailed
	if total > 0 {
		ratio = float64(s.state.TasksComplete) / float64(total) * 100
	}

	return "Confidence:" + formatFloat(s.state.Confidence) +
		" | Energy:" + formatFloat(s.state.Energy) +
		" | Status:" + s.state.Status +
		" | Mood:" + s.state.LastMood +
		" | Success:" + formatFloat(ratio) + "%" +
		" | Streak:" + formatInt(s.state.StreakSuccess) + "W/" + formatInt(s.state.StreakFailed) + "L" +
		" | Total:" + formatInt(s.state.TotalActions)
}

func clamp(val, min, max float64) float64 {
	return math.Max(min, math.Min(max, val))
}

func formatFloat(f float64) string {
	s := ""
	whole := int(f * 100)
	s = formatInt(whole/100) + "." + formatInt(whole%100)
	return s
}

func formatInt(i int) string {
	if i < 0 {
		return "-" + formatInt(-i)
	}
	if i < 10 {
		return string(rune('0' + i))
	}
	return formatInt(i/10) + string(rune('0'+i%10))
}
