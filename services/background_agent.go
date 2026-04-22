package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"dardcor-agent/models"

	"github.com/google/uuid"
)

type BackgroundAgentService struct {
	mu    sync.RWMutex
	tasks map[string]*BackgroundTask
	agent *AgentService
}

type BackgroundTask struct {
	ID          string             `json:"id"`
	Prompt      string             `json:"prompt"`
	Status      string             `json:"status"`
	Result      string             `json:"result,omitempty"`
	Error       string             `json:"error,omitempty"`
	CreatedAt   time.Time          `json:"created_at"`
	CompletedAt time.Time          `json:"completed_at,omitempty"`
	ConvID      string             `json:"conv_id,omitempty"`
	Priority    int                `json:"priority,omitempty"`
	cancel      context.CancelFunc `json:"-"`
	ctx         context.Context    `json:"-"`
}

func NewBackgroundAgentService(agent *AgentService) *BackgroundAgentService {
	return &BackgroundAgentService{
		tasks: make(map[string]*BackgroundTask),
		agent: agent,
	}
}

func (b *BackgroundAgentService) Submit(prompt string) string {
	return b.SubmitWithPriority(prompt, 0)
}

func (b *BackgroundAgentService) SubmitWithPriority(prompt string, priority int) string {
	taskID := uuid.New().String()
	ctx, cancel := context.WithCancel(context.Background())
	task := &BackgroundTask{
		ID:        taskID,
		Prompt:    prompt,
		Status:    "pending",
		CreatedAt: time.Now(),
		Priority:  priority,
		cancel:    cancel,
		ctx:       ctx,
	}

	b.mu.Lock()
	b.tasks[taskID] = task
	b.mu.Unlock()

	go b.execute(task)
	return taskID
}

func (b *BackgroundAgentService) execute(task *BackgroundTask) {
	b.setStatus(task.ID, "running", "", "")

	req := models.AgentRequest{
		Message: task.Prompt,
		Source:  "background",
	}
	resp, err := b.agent.ProcessMessage(task.ctx, req, nil)

	if task.ctx.Err() != nil {
		b.setStatus(task.ID, "cancelled", "", "Task was cancelled")
		return
	}

	if err != nil {
		b.setStatus(task.ID, "failed", "", fmt.Sprintf("Error: %v", err))
		return
	}

	convID := ""
	result := ""
	if resp != nil {
		convID = resp.ConversationID
		result = resp.Content
	}
	b.mu.Lock()
	if t, ok := b.tasks[task.ID]; ok {
		t.Status = "completed"
		t.Result = result
		t.ConvID = convID
		t.CompletedAt = time.Now()
	}
	b.mu.Unlock()
}

func (b *BackgroundAgentService) setStatus(id, status, result, errMsg string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if t, ok := b.tasks[id]; ok {
		t.Status = status
		if result != "" {
			t.Result = result
		}
		if errMsg != "" {
			t.Error = errMsg
		}
		if status == "completed" || status == "failed" {
			t.CompletedAt = time.Now()
		}
	}
}

func (b *BackgroundAgentService) GetTask(id string) *BackgroundTask {
	b.mu.RLock()
	defer b.mu.RUnlock()
	t := b.tasks[id]
	if t == nil {
		return nil
	}
	copy := *t
	return &copy
}

func (b *BackgroundAgentService) ListTasks() []*BackgroundTask {
	b.mu.RLock()
	defer b.mu.RUnlock()
	var list []*BackgroundTask
	for _, t := range b.tasks {
		copy := *t
		list = append(list, &copy)
	}
	return list
}

func (b *BackgroundAgentService) CancelTask(id string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if t, ok := b.tasks[id]; ok {
		if t.Status == "pending" || t.Status == "running" {
			if t.cancel != nil {
				t.cancel()
			}
			t.Status = "cancelled"
			t.CompletedAt = time.Now()
			return true
		}
	}
	return false
}

func (b *BackgroundAgentService) GetRunningCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	count := 0
	for _, t := range b.tasks {
		if t.Status == "running" {
			count++
		}
	}
	return count
}

func (b *BackgroundAgentService) CancelAll() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	count := 0
	for _, t := range b.tasks {
		if t.Status == "pending" || t.Status == "running" {
			if t.cancel != nil {
				t.cancel()
			}
			t.Status = "cancelled"
			t.CompletedAt = time.Now()
			count++
		}
	}
	return count
}

func (b *BackgroundAgentService) PurgeCompleted() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	count := 0
	for id, t := range b.tasks {
		if t.Status == "completed" || t.Status == "failed" || t.Status == "cancelled" {
			delete(b.tasks, id)
			count++
		}
	}
	return count
}
