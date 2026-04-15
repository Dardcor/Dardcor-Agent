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
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusBlocked    TaskStatus = "blocked"
)

type SubTask struct {
	ID           string     `json:"id"`
	Description  string     `json:"description"`
	Status       TaskStatus `json:"status"`
	Result       string     `json:"result,omitempty"`
	Dependencies []string   `json:"dependencies,omitempty"`
	RetryCount   int        `json:"retry_count"`
	MaxRetries   int        `json:"max_retries"`
	Error        string     `json:"error,omitempty"`
	StartedAt    time.Time  `json:"started_at,omitempty"`
	CompletedAt  time.Time  `json:"completed_at,omitempty"`
}

type TaskPlan struct {
	Goal       string    `json:"goal"`
	Tasks      []SubTask `json:"tasks"`
	StartedAt  time.Time `json:"started_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	IsComplete bool      `json:"is_complete"`
}

type OrchestratorService struct {
	currentPlan *TaskPlan
	history     []TaskPlan
}

func NewOrchestratorService() *OrchestratorService {
	return &OrchestratorService{
		history: make([]TaskPlan, 0),
	}
}

func (s *OrchestratorService) InitializePlan(goal string, tasks []SubTask) *TaskPlan {
	for i := range tasks {
		if tasks[i].MaxRetries == 0 {
			tasks[i].MaxRetries = 3
		}
	}

	plan := &TaskPlan{
		Goal:      goal,
		Tasks:     tasks,
		StartedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if s.currentPlan != nil {
		s.history = append(s.history, *s.currentPlan)
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
			if status == TaskStatusInProgress && s.currentPlan.Tasks[i].StartedAt.IsZero() {
				s.currentPlan.Tasks[i].StartedAt = time.Now()
			}
			if status == TaskStatusCompleted || status == TaskStatusFailed {
				s.currentPlan.Tasks[i].CompletedAt = time.Now()
			}
			break
		}
	}
	s.currentPlan.UpdatedAt = time.Now()
	s.checkCompletion()
}

func (s *OrchestratorService) FailTask(taskID string, errMsg string) bool {
	if s.currentPlan == nil {
		return false
	}
	for i := range s.currentPlan.Tasks {
		if s.currentPlan.Tasks[i].ID == taskID {
			task := &s.currentPlan.Tasks[i]
			task.RetryCount++
			task.Error = errMsg
			if task.RetryCount >= task.MaxRetries {
				task.Status = TaskStatusFailed
				task.CompletedAt = time.Now()
				s.currentPlan.UpdatedAt = time.Now()
				return false
			}
			task.Status = TaskStatusPending
			s.currentPlan.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}

func (s *OrchestratorService) GetNextPendingTask() *SubTask {
	if s.currentPlan == nil {
		return nil
	}
	for i := range s.currentPlan.Tasks {
		task := &s.currentPlan.Tasks[i]
		if task.Status != TaskStatusPending {
			continue
		}
		if s.areDependenciesMet(task) {
			return task
		}
	}
	return nil
}

func (s *OrchestratorService) areDependenciesMet(task *SubTask) bool {
	if len(task.Dependencies) == 0 {
		return true
	}
	for _, depID := range task.Dependencies {
		for _, t := range s.currentPlan.Tasks {
			if t.ID == depID && t.Status != TaskStatusCompleted {
				return false
			}
		}
	}
	return true
}

func (s *OrchestratorService) checkCompletion() {
	if s.currentPlan == nil {
		return
	}
	allDone := true
	for _, t := range s.currentPlan.Tasks {
		if t.Status != TaskStatusCompleted && t.Status != TaskStatusFailed {
			allDone = false
			break
		}
	}
	s.currentPlan.IsComplete = allDone
}

func (s *OrchestratorService) GetProgress() (completed int, total int) {
	if s.currentPlan == nil {
		return 0, 0
	}
	total = len(s.currentPlan.Tasks)
	for _, t := range s.currentPlan.Tasks {
		if t.Status == TaskStatusCompleted {
			completed++
		}
	}
	return
}

func (s *OrchestratorService) FormatPlanSummary() string {
	if s.currentPlan == nil {
		return "No active plan."
	}

	completed, total := s.GetProgress()

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Goal: %s [%d/%d]\n", s.currentPlan.Goal, completed, total))
	for _, t := range s.currentPlan.Tasks {
		statusIcon := "⏳"
		switch t.Status {
		case TaskStatusInProgress:
			statusIcon = "🔄"
		case TaskStatusCompleted:
			statusIcon = "✅"
		case TaskStatusFailed:
			statusIcon = "❌"
		case TaskStatusBlocked:
			statusIcon = "🚫"
		}
		retry := ""
		if t.RetryCount > 0 {
			retry = fmt.Sprintf(" (retry %d/%d)", t.RetryCount, t.MaxRetries)
		}
		sb.WriteString(fmt.Sprintf("%s %s [%s]%s\n", statusIcon, t.Description, t.ID, retry))
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

func (s *OrchestratorService) ClearPlan() {
	if s.currentPlan != nil {
		s.history = append(s.history, *s.currentPlan)
	}
	s.currentPlan = nil
}

func (s *OrchestratorService) GetHistory() []TaskPlan {
	return s.history
}
