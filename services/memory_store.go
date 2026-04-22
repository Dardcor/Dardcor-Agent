package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"dardcor-agent/models"

	_ "modernc.org/sqlite"
)

type MemoryStore struct {
	db *sql.DB
	mu sync.Mutex
}

func NewMemoryStore() (*MemoryStore, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("memory_store: could not determine home directory: %w", err)
	}

	dbDir := filepath.Join(homeDir, ".dardcor")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("memory_store: could not create database directory: %w", err)
	}

	dbPath := filepath.Join(dbDir, "dardcor_memory.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("memory_store: could not open database: %w", err)
	}

	db.SetMaxOpenConns(1)

	ms := &MemoryStore{db: db}
	if err := ms.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("memory_store: schema init failed: %w", err)
	}

	return ms, nil
}

func (ms *MemoryStore) initSchema() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS mem_sessions (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT    UNIQUE NOT NULL,
			project    TEXT    NOT NULL DEFAULT '',
			summary    TEXT    NOT NULL DEFAULT '',
			started_at DATETIME NOT NULL,
			ended_at   DATETIME,
			created_at DATETIME NOT NULL
		)`,

		`CREATE TABLE IF NOT EXISTS mem_observations (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT    NOT NULL,
			type       TEXT    NOT NULL DEFAULT '',
			obs_type   TEXT    NOT NULL DEFAULT '',
			content    TEXT    NOT NULL DEFAULT '',
			summary    TEXT    NOT NULL DEFAULT '',
			project    TEXT    NOT NULL DEFAULT '',
			files      TEXT    NOT NULL DEFAULT '[]',
			concepts   TEXT    NOT NULL DEFAULT '[]',
			created_at DATETIME NOT NULL
		)`,

		`CREATE TABLE IF NOT EXISTS mem_summaries (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT    NOT NULL,
			content    TEXT    NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL
		)`,

		`CREATE VIRTUAL TABLE IF NOT EXISTS mem_obs_fts
			USING fts5(
				content,
				summary,
				project,
				session_id,
				content=mem_observations,
				content_rowid=id
			)`,

		`CREATE TRIGGER IF NOT EXISTS mem_obs_fts_ai
			AFTER INSERT ON mem_observations BEGIN
				INSERT INTO mem_obs_fts(rowid, content, summary, project, session_id)
				VALUES (new.id, new.content, new.summary, new.project, new.session_id);
			END`,
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()

	for _, stmt := range stmts {
		if _, err := ms.db.Exec(stmt); err != nil {
			return fmt.Errorf("exec schema statement: %w\nSQL: %s", err, stmt)
		}
	}
	return nil
}

func (ms *MemoryStore) Close() error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return ms.db.Close()
}

func (ms *MemoryStore) CreateSession(sessionID, project string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	now := time.Now().UTC()
	_, err := ms.db.Exec(
		`INSERT INTO mem_sessions (session_id, project, summary, started_at, ended_at, created_at)
		 VALUES (?, ?, '', ?, NULL, ?)`,
		sessionID, project, now, now,
	)
	if err != nil {
		return fmt.Errorf("memory_store.CreateSession: %w", err)
	}
	return nil
}

func (ms *MemoryStore) CloseSession(sessionID, summary string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	now := time.Now().UTC()
	res, err := ms.db.Exec(
		`UPDATE mem_sessions SET ended_at = ?, summary = ? WHERE session_id = ?`,
		now, summary, sessionID,
	)
	if err != nil {
		return fmt.Errorf("memory_store.CloseSession: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("memory_store.CloseSession: session %q not found", sessionID)
	}
	return nil
}

func (ms *MemoryStore) GetSession(sessionID string) (*models.MemSession, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	row := ms.db.QueryRow(
		`SELECT id, session_id, project, summary, started_at, ended_at, created_at
		 FROM mem_sessions WHERE session_id = ?`,
		sessionID,
	)
	return scanSession(row)
}

func (ms *MemoryStore) GetRecentSessions(limit int) ([]*models.MemSession, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	rows, err := ms.db.Query(
		`SELECT id, session_id, project, summary, started_at, ended_at, created_at
		 FROM mem_sessions
		 ORDER BY created_at DESC
		 LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("memory_store.GetRecentSessions: %w", err)
	}
	defer rows.Close()

	return scanSessions(rows)
}

func (ms *MemoryStore) AddObservation(obs *models.MemObservation) (int64, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	filesJSON, err := json.Marshal(obs.Files)
	if err != nil {
		return 0, fmt.Errorf("memory_store.AddObservation: marshal files: %w", err)
	}
	conceptsJSON, err := json.Marshal(obs.Concepts)
	if err != nil {
		return 0, fmt.Errorf("memory_store.AddObservation: marshal concepts: %w", err)
	}

	now := time.Now().UTC()
	if obs.CreatedAt.IsZero() {
		obs.CreatedAt = now
	}

	res, err := ms.db.Exec(
		`INSERT INTO mem_observations
			(session_id, type, obs_type, content, summary, project, files, concepts, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		obs.SessionID,
		obs.Type,
		obs.ObsType,
		obs.Content,
		obs.Summary,
		obs.Project,
		string(filesJSON),
		string(conceptsJSON),
		obs.CreatedAt.UTC(),
	)
	if err != nil {
		return 0, fmt.Errorf("memory_store.AddObservation: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("memory_store.AddObservation: last insert id: %w", err)
	}
	return id, nil
}

func (ms *MemoryStore) GetObservationsBySession(sessionID string) ([]*models.MemObservation, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	rows, err := ms.db.Query(
		`SELECT id, session_id, type, obs_type, content, summary, project, files, concepts, created_at
		 FROM mem_observations
		 WHERE session_id = ?
		 ORDER BY created_at ASC`,
		sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("memory_store.GetObservationsBySession: %w", err)
	}
	defer rows.Close()

	return scanObservations(rows)
}

func (ms *MemoryStore) GetObservationsByProject(project string, limit int) ([]*models.MemObservation, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	rows, err := ms.db.Query(
		`SELECT id, session_id, type, obs_type, content, summary, project, files, concepts, created_at
		 FROM mem_observations
		 WHERE project = ?
		 ORDER BY created_at DESC
		 LIMIT ?`,
		project, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("memory_store.GetObservationsByProject: %w", err)
	}
	defer rows.Close()

	return scanObservations(rows)
}

func (ms *MemoryStore) GetObservationsByIDs(ids []int64) ([]*models.MemObservation, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(
		`SELECT id, session_id, type, obs_type, content, summary, project, files, concepts, created_at
		 FROM mem_observations
		 WHERE id IN (%s)
		 ORDER BY created_at ASC`,
		strings.Join(placeholders, ","),
	)

	rows, err := ms.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("memory_store.GetObservationsByIDs: %w", err)
	}
	defer rows.Close()

	return scanObservations(rows)
}

func (ms *MemoryStore) SearchFTS(query string, limit int, project string) ([]*models.MemObservation, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	var (
		rows *sql.Rows
		err  error
	)

	if project != "" {
		rows, err = ms.db.Query(
			`SELECT o.id, o.session_id, o.type, o.obs_type, o.content, o.summary, o.project,
			        o.files, o.concepts, o.created_at
			 FROM mem_obs_fts f
			 JOIN mem_observations o ON o.id = f.rowid
			 WHERE mem_obs_fts MATCH ?
			   AND o.project = ?
			 ORDER BY rank
			 LIMIT ?`,
			query, project, limit,
		)
	} else {
		rows, err = ms.db.Query(
			`SELECT o.id, o.session_id, o.type, o.obs_type, o.content, o.summary, o.project,
			        o.files, o.concepts, o.created_at
			 FROM mem_obs_fts f
			 JOIN mem_observations o ON o.id = f.rowid
			 WHERE mem_obs_fts MATCH ?
			 ORDER BY rank
			 LIMIT ?`,
			query, limit,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("memory_store.SearchFTS: %w", err)
	}
	defer rows.Close()

	return scanObservations(rows)
}

func (ms *MemoryStore) GetTimeline(anchorID int64, depthBefore, depthAfter int, project string) (*models.MemTimeline, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	anchorRow := ms.db.QueryRow(
		`SELECT id, session_id, type, obs_type, content, summary, project, files, concepts, created_at
		 FROM mem_observations WHERE id = ?`,
		anchorID,
	)
	anchor, err := scanObservation(anchorRow)
	if err != nil {
		return nil, fmt.Errorf("memory_store.GetTimeline anchor: %w", err)
	}

	projectFilter := anchor.Project
	if project != "" {
		projectFilter = project
	}

	beforeRows, err := ms.db.Query(
		`SELECT id, session_id, type, obs_type, content, summary, project, files, concepts, created_at
		 FROM mem_observations
		 WHERE session_id = ?
		   AND project    = ?
		   AND id         < ?
		 ORDER BY id DESC
		 LIMIT ?`,
		anchor.SessionID, projectFilter, anchorID, depthBefore,
	)
	if err != nil {
		return nil, fmt.Errorf("memory_store.GetTimeline before: %w", err)
	}
	defer beforeRows.Close()

	before, err := scanObservations(beforeRows)
	if err != nil {
		return nil, fmt.Errorf("memory_store.GetTimeline before scan: %w", err)
	}
	for i, j := 0, len(before)-1; i < j; i, j = i+1, j-1 {
		before[i], before[j] = before[j], before[i]
	}

	afterRows, err := ms.db.Query(
		`SELECT id, session_id, type, obs_type, content, summary, project, files, concepts, created_at
		 FROM mem_observations
		 WHERE session_id = ?
		   AND project    = ?
		   AND id         > ?
		 ORDER BY id ASC
		 LIMIT ?`,
		anchor.SessionID, projectFilter, anchorID, depthAfter,
	)
	if err != nil {
		return nil, fmt.Errorf("memory_store.GetTimeline after: %w", err)
	}
	defer afterRows.Close()

	after, err := scanObservations(afterRows)
	if err != nil {
		return nil, fmt.Errorf("memory_store.GetTimeline after scan: %w", err)
	}

	beforeVals := dereferenceSlice(before)
	afterVals := dereferenceSlice(after)

	return &models.MemTimeline{
		Anchor: *anchor,
		Before: beforeVals,
		After:  afterVals,
	}, nil
}

func (ms *MemoryStore) SaveSessionSummary(sessionID, summary string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	now := time.Now().UTC()

	_, err := ms.db.Exec(
		`INSERT INTO mem_summaries (session_id, content, created_at) VALUES (?, ?, ?)`,
		sessionID, summary, now,
	)
	if err != nil {
		return fmt.Errorf("memory_store.SaveSessionSummary insert summary: %w", err)
	}

	_, err = ms.db.Exec(
		`UPDATE mem_sessions SET summary = ? WHERE session_id = ?`,
		summary, sessionID,
	)
	if err != nil {
		return fmt.Errorf("memory_store.SaveSessionSummary update session: %w", err)
	}
	return nil
}

func (ms *MemoryStore) SaveObservation(sessionID, project, obsType, content string, files, concepts []string) error {
	if files == nil {
		files = []string{}
	}
	if concepts == nil {
		concepts = []string{}
	}
	obs := &models.MemObservation{
		SessionID: sessionID,
		Type:      obsType,
		ObsType:   obsType,
		Content:   content,
		Project:   project,
		Files:     files,
		Concepts:  concepts,
		CreatedAt: time.Now().UTC(),
	}
	_, err := ms.AddObservation(obs)
	return err
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanSession(row rowScanner) (*models.MemSession, error) {
	var s models.MemSession
	var endedAt sql.NullTime
	err := row.Scan(
		&s.ID,
		&s.SessionID,
		&s.Project,
		&s.Summary,
		&s.StartedAt,
		&endedAt,
		&s.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found")
		}
		return nil, err
	}
	if endedAt.Valid {
		t := endedAt.Time
		s.EndedAt = &t
	}
	return &s, nil
}

func scanSessions(rows *sql.Rows) ([]*models.MemSession, error) {
	var sessions []*models.MemSession
	for rows.Next() {
		s, err := scanSession(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

func scanObservation(row rowScanner) (*models.MemObservation, error) {
	var obs models.MemObservation
	var filesJSON, conceptsJSON string

	err := row.Scan(
		&obs.ID,
		&obs.SessionID,
		&obs.Type,
		&obs.ObsType,
		&obs.Content,
		&obs.Summary,
		&obs.Project,
		&filesJSON,
		&conceptsJSON,
		&obs.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("observation not found")
		}
		return nil, err
	}

	if err := json.Unmarshal([]byte(filesJSON), &obs.Files); err != nil {
		obs.Files = []string{}
	}
	if err := json.Unmarshal([]byte(conceptsJSON), &obs.Concepts); err != nil {
		obs.Concepts = []string{}
	}

	return &obs, nil
}

func scanObservations(rows *sql.Rows) ([]*models.MemObservation, error) {
	var observations []*models.MemObservation
	for rows.Next() {
		obs, err := scanObservation(rows)
		if err != nil {
			return nil, err
		}
		observations = append(observations, obs)
	}
	return observations, rows.Err()
}

func dereferenceSlice(ptrs []*models.MemObservation) []models.MemObservation {
	out := make([]models.MemObservation, len(ptrs))
	for i, p := range ptrs {
		out[i] = *p
	}
	return out
}
