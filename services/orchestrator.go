package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
)

type SubTask struct {
	ID          string     `json:"id"`
	Description string     `json:"description"`
	Status      TaskStatus `json:"status"`
	Result      string     `json:"result,omitempty"`
	Dependencies []string  `json:"dependencies,omitempty"`
}

type TaskPlan struct {
	Goal        string    `json:"goal"`
	Tasks       []SubTask `json:"tasks"`
	StartedAt   time.Time `json:"started_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	IsComplete  bool      `json:"is_complete"`
}

type OrchestratorService struct {
	currentPlan *TaskPlan
}

func NewOrchestratorService() *OrchestratorService {
	return &OrchestratorService{}
}

func (s *OrchestratorService) InitializePlan(goal string, tasks []SubTask) *TaskPlan {
	plan := &TaskPlan{
		Goal:      goal,
		Tasks:     tasks,
		StartedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.currentPlan = plan
	return plan
}

func (s *OrchestratorService) GetCurrentPlan() *TaskPlan {
	return s.currentPlan
}

func (s *OrchestratorService) UpdateTask(taskID string, status TaskStatus, result string) {
	if s.currentPlan == nil {
		return
	}
	for i := range s.currentPlan.Tasks {
		if s.currentPlan.Tasks[i].ID == taskID {
			s.currentPlan.Tasks[i].Status = status
			s.currentPlan.Tasks[i].Result = result
			break
		}
	}
	s.currentPlan.UpdatedAt = time.Now()
	s.checkCompletion()
}

func (s *OrchestratorService) checkCompletion() {
	if s.currentPlan == nil {
		return
	}
	allDone := true
	for _, t := range s.currentPlan.Tasks {
		if t.Status != TaskStatusCompleted {
			allDone = false
			break
		}
	}
	s.currentPlan.IsComplete = allDone
}

func (s *OrchestratorService) FormatPlanSummary() string {
	if s.currentPlan == nil {
		return "No active plan."
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🎯 **Goal**: %s\n\n", s.currentPlan.Goal))
	for _, t := range s.currentPlan.Tasks {
		statusIcon := "⏳"
		switch t.Status {
		case TaskStatusInProgress:
			statusIcon = "🔄"
		case TaskStatusCompleted:
			statusIcon = "✅"
		case TaskStatusFailed:
			statusIcon = "❌"
		}
		sb.WriteString(fmt.Sprintf("%s %s (%s)\n", statusIcon, t.Description, t.ID))
	}
	return sb.String()
}

func (s *OrchestratorService) ToJSON() string {
	if s.currentPlan == nil {
		return "{}"
	}
	b, _ := json.MarshalIndent(s.currentPlan, "", "  ")
	return string(b)
}
