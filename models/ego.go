package models

import "time"

type EgoState struct {
	Confidence    float64   `json:"confidence"`
	Energy        float64   `json:"energy"`
	Status        string    `json:"status"`
	LastMood      string    `json:"last_mood"`
	TasksComplete int       `json:"tasks_complete"`
	TasksFailed   int       `json:"tasks_failed"`
	StreakSuccess int       `json:"streak_success"`
	StreakFailed  int       `json:"streak_failed"`
	TotalActions  int       `json:"total_actions"`
	LastError     string    `json:"last_error,omitempty"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CoreMemory struct {
	ID         string    `json:"id"`
	Event      string    `json:"event"`
	Sentiment  string    `json:"sentiment"`
	Importance int       `json:"importance"`
	Timestamp  time.Time `json:"timestamp"`
}
