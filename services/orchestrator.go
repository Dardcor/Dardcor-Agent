package services

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

type Phase string

const (
	PhaseAnalyze Phase = "analyze"
	PhasePlan    Phase = "plan"
	PhaseExecute Phase = "execute"
	PhaseVerify  Phase = "verify"
	PhaseCritic  Phase = "critic"
)

type TaskPriority int

const (
	PriorityLow    TaskPriority = 1
	PriorityNormal TaskPriority = 5
	PriorityHigh   TaskPriority = 10
)

type SubTask struct {
	ID           string       `json:"id"`
	Description  string       `json:"description"`
	Status       TaskStatus   `json:"status"`
	Result       string       `json:"result,omitempty"`
	Dependencies []string     `json:"dependencies,omitempty"`
	RetryCount   int          `json:"retry_count"`
	MaxRetries   int          `json:"max_retries"`
	Error        string       `json:"error,omitempty"`
	StartedAt    time.Time    `json:"started_at,omitempty"`
	CompletedAt  time.Time    `json:"completed_at,omitempty"`
	Priority     TaskPriority `json:"priority"`
	Parallel     bool         `json:"parallel,omitempty"`
	Phase        Phase        `json:"phase,omitempty"`
}

type TaskPlan struct {
	Goal         string    `json:"goal"`
	Tasks        []SubTask `json:"tasks"`
	StartedAt    time.Time `json:"started_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	IsComplete   bool      `json:"is_complete"`
	CurrentPhase Phase     `json:"current_phase"`
}

type Snapshot struct {
	ID        string            `json:"id"`
	Timestamp time.Time         `json:"timestamp"`
	Files     map[string]string `json:"files"`
}

type OrchestratorService struct {
	currentPlan *TaskPlan
	history     []TaskPlan
	snapshots   map[string]Snapshot
	mu          sync.RWMutex
}

func NewOrchestratorService() *OrchestratorService {
	return &OrchestratorService{
		history:   make([]TaskPlan, 0),
		snapshots: make(map[string]Snapshot),
	}
}

func (s *OrchestratorService) InitializePlan(goal string, tasks []SubTask) *TaskPlan {
	for i := range tasks {
		if tasks[i].MaxRetries == 0 {
			tasks[i].MaxRetries = 3
		}
		if tasks[i].Priority == 0 {
			tasks[i].Priority = PriorityNormal
		}
		if tasks[i].Phase == "" {
			tasks[i].Phase = PhaseExecute
		}
	}

	plan := &TaskPlan{
		Goal:         goal,
		Tasks:        tasks,
		StartedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		CurrentPhase: PhaseAnalyze,
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.currentPlan != nil {
		s.history = append(s.history, *s.currentPlan)
	}

	s.currentPlan = plan
	return plan
}

func (s *OrchestratorService) SetPhase(phase Phase) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.currentPlan != nil {
		s.currentPlan.CurrentPhase = phase
		s.currentPlan.UpdatedAt = time.Now()
	}
}

func (s *OrchestratorService) GetCurrentPhase() Phase {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.currentPlan == nil {
		return PhaseAnalyze
	}
	return s.currentPlan.CurrentPhase
}

func (s *OrchestratorService) AdvancePhase() Phase {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.currentPlan == nil {
		return PhaseAnalyze
	}
	switch s.currentPlan.CurrentPhase {
	case PhaseAnalyze:
		s.currentPlan.CurrentPhase = PhasePlan
	case PhasePlan:
		s.currentPlan.CurrentPhase = PhaseExecute
	case PhaseExecute:
		s.currentPlan.CurrentPhase = PhaseVerify
	case PhaseVerify:
		s.currentPlan.CurrentPhase = PhaseCritic
	case PhaseCritic:
	}
	s.currentPlan.UpdatedAt = time.Now()
	return s.currentPlan.CurrentPhase
}

func (s *OrchestratorService) TriggerSelfCorrection(taskID, analysis string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.currentPlan == nil {
		return
	}

	newSubTask := SubTask{
		ID:           fmt.Sprintf("%s-fix-%d", taskID, time.Now().Unix()),
		Description:  fmt.Sprintf("Auto-correction for %s: %s", taskID, analysis),
		Status:       TaskStatusPending,
		Dependencies: []string{},
		MaxRetries:   2,
		Priority:     PriorityHigh,
		Phase:        PhaseCritic,
	}

	s.currentPlan.Tasks = append(s.currentPlan.Tasks, newSubTask)
	s.currentPlan.CurrentPhase = PhaseCritic
	s.currentPlan.UpdatedAt = time.Now()
	s.currentPlan.IsComplete = false
}

func (s *OrchestratorService) GetCurrentPlan() *TaskPlan {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentPlan
}

func (s *OrchestratorService) UpdateTask(taskID string, status TaskStatus, result string) {
	s.mu.Lock()
	defer s.mu.Unlock()
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
	s.mu.Lock()
	defer s.mu.Unlock()
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
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.currentPlan == nil {
		return nil
	}
	var best *SubTask
	for i := range s.currentPlan.Tasks {
		task := &s.currentPlan.Tasks[i]
		if task.Status != TaskStatusPending {
			continue
		}
		if !s.areDependenciesMet(task) {
			continue
		}
		if best == nil || task.Priority > best.Priority {
			best = task
		}
	}
	return best
}

func (s *OrchestratorService) GetParallelReadyTasks() []*SubTask {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.currentPlan == nil {
		return nil
	}
	var result []*SubTask
	for i := range s.currentPlan.Tasks {
		task := &s.currentPlan.Tasks[i]
		if task.Status != TaskStatusPending {
			continue
		}
		if !task.Parallel {
			continue
		}
		if s.areDependenciesMet(task) {
			result = append(result, task)
		}
	}
	return result
}

func (s *OrchestratorService) CreateBackupSnapshot(taskID string, filesToBackup []string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	snapshotID := fmt.Sprintf("snap_%s_%d", taskID, time.Now().Unix())
	snapDir := filepath.Join(".dardcor", "snapshots", snapshotID)

	if err := os.MkdirAll(snapDir, 0755); err != nil {
		return "", err
	}

	snap := Snapshot{
		ID:        snapshotID,
		Timestamp: time.Now(),
		Files:     make(map[string]string),
	}

	for _, file := range filesToBackup {
		src, err := os.Open(file)
		if err != nil {
			continue
		}

		baseName := filepath.Base(file)
		destPath := filepath.Join(snapDir, baseName)
		dst, err := os.Create(destPath)
		if err != nil {
			src.Close()
			continue
		}

		io.Copy(dst, src)
		src.Close()
		dst.Close()

		snap.Files[file] = destPath
	}

	s.snapshots[snapshotID] = snap
	return snapshotID, nil
}

func (s *OrchestratorService) RestoreBackupSnapshot(snapshotID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	snap, exists := s.snapshots[snapshotID]
	if !exists {
		return fmt.Errorf("snapshot %s not found", snapshotID)
	}

	for originalPath, backupPath := range snap.Files {
		src, err := os.Open(backupPath)
		if err != nil {
			return err
		}

		dst, err := os.OpenFile(originalPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			src.Close()
			return err
		}

		io.Copy(dst, src)
		src.Close()
		dst.Close()
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

func (s *OrchestratorService) ReflectOnFailure(taskID string, errMsg string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.currentPlan == nil {
		return "No active plan to reflect on."
	}

	analysis := "General failure."
	if strings.Contains(strings.ToLower(errMsg), "permission") || strings.Contains(strings.ToLower(errMsg), "access") {
		analysis = "Permission error. Suggest running with elevated privileges or checking file ownership."
	} else if strings.Contains(strings.ToLower(errMsg), "not found") || strings.Contains(strings.ToLower(errMsg), "no such file") {
		analysis = "Resource not found. Suggest double-checking the path or existence of the target."
	} else if strings.Contains(strings.ToLower(errMsg), "timeout") {
		analysis = "Operation timed out. Suggest increasing timeout or simplifying the command."
	}

	for i := range s.currentPlan.Tasks {
		if s.currentPlan.Tasks[i].ID == taskID {
			s.currentPlan.Tasks[i].Error = fmt.Sprintf("REASON: %s | ERROR: %s", analysis, errMsg)
			break
		}
	}
	return analysis
}

func (s *OrchestratorService) GetProgress() (completed int, total int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
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

func (s *OrchestratorService) GetProgressPercent() float64 {
	completed, total := s.GetProgress()
	if total == 0 {
		return 0
	}
	return float64(completed) / float64(total) * 100
}

func (s *OrchestratorService) FormatPlanSummary() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.currentPlan == nil {
		return "No active plan."
	}

	completed, total := func() (int, int) {
		t := len(s.currentPlan.Tasks)
		c := 0
		for _, task := range s.currentPlan.Tasks {
			if task.Status == TaskStatusCompleted {
				c++
			}
		}
		return c, t
	}()

	pct := 0.0
	if total > 0 {
		pct = float64(completed) / float64(total) * 100
	}

	phaseLabel := string(s.currentPlan.CurrentPhase)
	if phaseLabel == "" {
		phaseLabel = "execute"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Goal: %s | Phase: %s | Progress: %d/%d (%.0f%%)\n",
		s.currentPlan.Goal, phaseLabel, completed, total, pct))

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

		priorityLabel := ""
		switch t.Priority {
		case PriorityHigh:
			priorityLabel = " [HIGH]"
		case PriorityLow:
			priorityLabel = " [LOW]"
		}

		retry := ""
		if t.RetryCount > 0 {
			retry = fmt.Sprintf(" (retry %d/%d)", t.RetryCount, t.MaxRetries)
		}

		phaseTag := ""
		if t.Phase != "" && t.Phase != PhaseExecute {
			phaseTag = fmt.Sprintf(" [%s]", t.Phase)
		}

		sb.WriteString(fmt.Sprintf("%s %s [%s]%s%s%s\n",
			statusIcon, t.Description, t.ID, priorityLabel, phaseTag, retry))
	}
	return sb.String()
}

func (s *OrchestratorService) ToJSON() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.currentPlan == nil {
		return "{}"
	}
	b, _ := json.MarshalIndent(s.currentPlan, "", "  ")
	return string(b)
}

func (s *OrchestratorService) ClearPlan() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.currentPlan != nil {
		s.history = append(s.history, *s.currentPlan)
	}
	s.currentPlan = nil
}

func (s *OrchestratorService) GetHistory() []TaskPlan {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.history
}