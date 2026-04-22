package services

import (
	"fmt"
	"strings"
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

func (rs *ReflectionService) Reflect() string {
	ego := rs.egoSvc.GetState()
	plan := rs.orchSvc.GetCurrentPlan()

	var insights []string

	switch {
	case ego.Confidence < 0.3:
		insights = append(insights, "CRITICAL: Confidence dangerously low. Switching to cautious execution — verify every file operation before proceeding.")
	case ego.Energy < 0.2:
		insights = append(insights, "LOW ENERGY: Prioritize remaining work. Complete current objective only, defer improvements.")
	case ego.StreakFailed > 2:
		insights = append(insights, fmt.Sprintf("FAILURE PATTERN: %d consecutive failures detected. Previous error: %s. Try alternative approach.", ego.StreakFailed, ego.LastError))
	case ego.StreakSuccess > 5:
		insights = append(insights, fmt.Sprintf("MOMENTUM: %d consecutive successes. Performance ratio: %.0f%%. Maintain current strategy.", ego.StreakSuccess, float64(ego.TasksComplete)/float64(ego.TasksComplete+ego.TasksFailed)*100))
	case ego.Confidence > 0.8:
		insights = append(insights, "HIGH CONFIDENCE: Execute decisively. Batch related operations where possible.")
	default:
		insights = append(insights, "STABLE: Normal execution mode. Proceed methodically.")
	}

	if plan != nil {
		completed, total := rs.orchSvc.GetProgress()
		pct := float64(0)
		if total > 0 {
			pct = float64(completed) / float64(total) * 100
		}

		switch {
		case pct >= 90 && !plan.IsComplete:
			insights = append(insights, fmt.Sprintf("NEAR COMPLETION: %.0f%% done. Focus on verification and cleanup.", pct))
		case pct < 20 && completed > 0:
			insights = append(insights, "EARLY STAGE: Foundation being laid. Ensure correct approach before scaling.")
		case plan.CurrentPhase == "critic":
			insights = append(insights, "SELF-CORRECTION PHASE: Reviewing previous actions for errors. Apply fixes before marking complete.")
		}

		blockedCount := 0
		for _, t := range plan.Tasks {
			if t.Status == "blocked" {
				blockedCount++
			}
		}
		if blockedCount > 0 {
			insights = append(insights, fmt.Sprintf("BLOCKED: %d tasks waiting on dependencies. Resolve blockers first.", blockedCount))
		}
	}

	ratio := rs.egoSvc.GetPerformanceRatio()
	if ratio < 0.5 && ego.TotalActions > 5 {
		insights = append(insights, fmt.Sprintf("WARNING: Success ratio %.0f%% is below threshold. Consider simplifying approach.", ratio*100))
	}

	result := strings.Join(insights, " | ")

	rs.history = append(rs.history, fmt.Sprintf("[%s] %s", time.Now().Format("15:04"), result))
	if len(rs.history) > 20 {
		rs.history = rs.history[len(rs.history)-20:]
	}

	return result
}

func (rs *ReflectionService) GetLastInsight() string {
	if len(rs.history) == 0 {
		return "No insights yet."
	}
	return rs.history[len(rs.history)-1]
}

func (rs *ReflectionService) AnalyzeActionResult(actionType string, result string, success bool) string {
	if success {
		return ""
	}

	lowerResult := strings.ToLower(result)
	var advice string

	switch {
	case strings.Contains(lowerResult, "permission denied") || strings.Contains(lowerResult, "access denied"):
		advice = "Permission error detected. Check file ownership or run with elevated privileges."
	case strings.Contains(lowerResult, "not found") || strings.Contains(lowerResult, "no such file"):
		advice = "Target not found. Verify path exists before retrying. Use 'list' or 'search' to locate."
	case strings.Contains(lowerResult, "timeout"):
		advice = "Operation timed out. Consider breaking into smaller operations or increasing timeout."
	case strings.Contains(lowerResult, "syntax error") || strings.Contains(lowerResult, "compile error"):
		advice = "Code error detected. Read the file, identify the exact line, and fix surgically."
	case strings.Contains(lowerResult, "already exists"):
		advice = "Resource already exists. Read existing content before overwriting."
	default:
		advice = fmt.Sprintf("Action '%s' failed. Review the error and try an alternative approach.", actionType)
	}

	rs.history = append(rs.history, fmt.Sprintf("[%s] ACTION ANALYSIS: %s → %s", time.Now().Format("15:04"), actionType, advice))
	return advice
}
