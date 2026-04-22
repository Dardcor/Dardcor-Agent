package models

import "time"

type MemSession struct {
	ID        int64      `json:"id"`
	SessionID string     `json:"session_id"`
	Project   string     `json:"project"`
	Summary   string     `json:"summary"`
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at"`
	CreatedAt time.Time  `json:"created_at"`
}

type MemObservation struct {
	ID        int64     `json:"id"`
	SessionID string    `json:"session_id"`
	Type      string    `json:"type"`
	ObsType   string    `json:"obs_type"`
	Content   string    `json:"content"`
	Summary   string    `json:"summary"`
	Project   string    `json:"project"`
	Files     []string  `json:"files"`
	Concepts  []string  `json:"concepts"`
	CreatedAt time.Time `json:"created_at"`
}

type MemSummary struct {
	ID        int64     `json:"id"`
	SessionID string    `json:"session_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type MemSearchResult struct {
	Observation MemObservation `json:"observation"`
	Score       float64        `json:"score"`
	Snippet     string         `json:"snippet"`
}

type MemTimeline struct {
	Anchor MemObservation   `json:"anchor"`
	Before []MemObservation `json:"before"`
	After  []MemObservation `json:"after"`
}
