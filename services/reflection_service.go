package services

import (
	"fmt"
	"time"
)

type ReflectionService struct {
	egoSvc  *EgoService
	orchSvc *OrchestratorService
	history []string
}

func NewReflectionService(ego *EgoService, orch *OrchestratorService) *ReflectionService {
	return &ReflectionService{
		egoSvc:  ego,
		orchSvc: orch,
		history: make([]string, 0),
	}
}

// Reflect evaluates the current state of the agent and returns a strategic nudge.
func (rs *ReflectionService) Reflect() string {
	ego := rs.egoSvc.GetState()
	plan := rs.orchSvc.GetCurrentPlan()

	if plan == nil {
		return "Standing by for new objectives."
	}

	completed, total := rs.orchSvc.GetProgress()
	progress := float64(completed) / float64(total)

	var strategy string
	switch {
	case ego.Confidence < 0.3:
		strategy = "Confidence critical. Re-evaluating environment constraints before next action."
	case ego.Energy < 0.2:
		strategy = "Low energy. Consolidating work and recommending a rest cycle (Dreaming)."
	case progress > 0.8 && !plan.IsComplete:
		strategy = "Task nearly complete. Transitioning focus to Verification and clean-up."
	case ego.StreakFailed > 2:
		strategy = "Detected failure loop. Recommending alternative tool selection strategy."
	default:
		strategy = "Optimal execution path maintained."
	}

	rs.history = append(rs.history, fmt.Sprintf("[%s] %s", time.Now().Format("15:04"), strategy))
	if len(rs.history) > 10 {
		rs.history = rs.history[1:]
	}

	return strategy
}

func (rs *ReflectionService) GetLastInsight() string {
	if len(rs.history) == 0 {
		return "No insights yet."
	}
	return rs.history[len(rs.history)-1]
}
