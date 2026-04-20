package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"time"

	"dardcor-agent/config"
	"dardcor-agent/models"

	"github.com/google/uuid"
)

type JSONStore struct {
	mu         sync.RWMutex
	cache      map[string]*models.Conversation // ID -> Conversation
	cacheSrc   map[string]string               // ID -> Source
	indexCache map[string]*ConversationIndex   // Source -> Index
	indexMu    sync.RWMutex
}

type ConversationIndex struct {
	Entries []IndexEntry `json:"entries"`
}

type IndexEntry struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	UpdatedAt time.Time `json:"updated_at"`
	MsgCount  int       `json:"msg_count"`
}

var Store *JSONStore

func Init() {
	Store = &JSONStore{
		cache:      make(map[string]*models.Conversation),
		cacheSrc:   make(map[string]string),
		indexCache: make(map[string]*ConversationIndex),
	}
}

func (s *JSONStore) getIndexPath(source string) string {
	return filepath.Join(config.AppConfig.GetConversationsDir(source), "conversations.index.json")
}

func (s *JSONStore) loadIndexLocked(source string) *ConversationIndex {
	if idx, ok := s.indexCache[source]; ok {
		return idx
	}

	path := s.getIndexPath(source)
	data, err := os.ReadFile(path)
	if err != nil {
		idx := &ConversationIndex{Entries: []IndexEntry{}}
		s.indexCache[source] = idx
		return idx
	}

	var idx ConversationIndex
	if err := json.Unmarshal(data, &idx); err != nil {
		idx := &ConversationIndex{Entries: []IndexEntry{}}
		s.indexCache[source] = idx
		return idx
	}

	s.indexCache[source] = &idx
	return &idx
}

func (s *JSONStore) saveIndexLocked(source string, idx *ConversationIndex) error {
	path := s.getIndexPath(source)
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (s *JSONStore) updateIndexEntryLocked(source string, conv *models.Conversation) {
	idx := s.loadIndexLocked(source)
	found := false
	for i, entry := range idx.Entries {
		if entry.ID == conv.ID {
			idx.Entries[i] = IndexEntry{
				ID:        conv.ID,
				Title:     conv.Title,
				UpdatedAt: conv.UpdatedAt,
				MsgCount:  len(conv.Messages),
			}
			found = true
			break
		}
	}
	if !found {
		idx.Entries = append(idx.Entries, IndexEntry{
			ID:        conv.ID,
			Title:     conv.Title,
			UpdatedAt: conv.UpdatedAt,
			MsgCount:  len(conv.Messages),
		})
	}
	s.saveIndexLocked(source, idx)
}

func (s *JSONStore) removeIndexEntryLocked(source string, id string) {
	idx := s.loadIndexLocked(source)
	for i, entry := range idx.Entries {
		if entry.ID == id {
			idx.Entries = append(idx.Entries[:i], idx.Entries[i+1:]...)
			break
		}
	}
	s.saveIndexLocked(source, idx)
}

func (s *JSONStore) loadConversationLocked(id string, source string) (*models.Conversation, error) {
	if conv, ok := s.cache[id]; ok {
		return conv, nil
	}

	if filepath.Base(id) != id {
		return nil, fmt.Errorf("invalid conversation id")
	}
	dir := config.AppConfig.GetConversationsDir(source)
	path := filepath.Join(dir, id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("conversation not found: %w", err)
	}
	var conv models.Conversation
	if err := json.Unmarshal(data, &conv); err != nil {
		return nil, fmt.Errorf("failed to parse conversation: %w", err)
	}

	// Update cache
	s.cache[id] = &conv
	s.cacheSrc[id] = source

	return &conv, nil
}

func (s *JSONStore) saveConversationLocked(conv *models.Conversation, source string) error {
	if filepath.Base(conv.ID) != conv.ID {
		return fmt.Errorf("invalid conversation id")
	}
	conv.UpdatedAt = time.Now()
	data, err := json.MarshalIndent(conv, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal conversation: %w", err)
	}
	dir := config.AppConfig.GetConversationsDir(source)
	path := filepath.Join(dir, conv.ID+".json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}

	// Update cache and index
	s.cache[conv.ID] = conv
	s.cacheSrc[conv.ID] = source

	s.indexMu.Lock()
	s.updateIndexEntryLocked(source, conv)
	s.indexMu.Unlock()

	return nil
}

func (s *JSONStore) SaveConversation(conv *models.Conversation, source string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveConversationLocked(conv, source)
}


func (s *JSONStore) LoadConversation(id string, source string) (*models.Conversation, error) {
	s.mu.RLock()
	// Check cache first without full lock if possible or just use a single lock for simplicity in this optimization
	s.mu.RUnlock()
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadConversationLocked(id, source)
}


func (s *JSONStore) ListConversations(source string) ([]models.Conversation, error) {
	s.indexMu.Lock()
	defer s.indexMu.Unlock()

	idx := s.loadIndexLocked(source)
	var conversations []models.Conversation
	for _, entry := range idx.Entries {
		title := entry.Title
		if entry.MsgCount > 0 {
			title = fmt.Sprintf("%s (%d messages)", entry.Title, entry.MsgCount)
		}
		conversations = append(conversations, models.Conversation{
			ID:        entry.ID,
			Title:     title,
			UpdatedAt: entry.UpdatedAt,
		})
	}

	return conversations, nil
}


func (s *JSONStore) DeleteConversation(id string, source string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if filepath.Base(id) != id {
		return fmt.Errorf("invalid conversation id")
	}

	delete(s.cache, id)
	delete(s.cacheSrc, id)

	s.indexMu.Lock()
	s.removeIndexEntryLocked(source, id)
	s.indexMu.Unlock()

	dir := config.AppConfig.GetConversationsDir(source)
	path := filepath.Join(dir, id+".json")
	return os.Remove(path)
}


func (s *JSONStore) RenameConversation(id string, newTitle string, source string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	conv, err := s.loadConversationLocked(id, source)
	if err != nil {
		return err
	}
	conv.Title = newTitle
	return s.saveConversationLocked(conv, source)
}


func (s *JSONStore) CreateConversation(title string, source string) (*models.Conversation, error) {
	conv := &models.Conversation{
		ID:        uuid.New().String(),
		Title:     title,
		Messages:  []models.Message{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.SaveConversation(conv, source); err != nil {
		return nil, err
	}

	return conv, nil
}

func (s *JSONStore) AddMessage(convID string, msg models.Message, source string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	conv, err := s.loadConversationLocked(convID, source)
	if err != nil {
		return err
	}
	msg.ID = uuid.New().String()
	msg.Timestamp = time.Now()
	conv.Messages = append(conv.Messages, msg)
	return s.saveConversationLocked(conv, source)
}

func (s *JSONStore) SaveCommandHistory(cmd models.CommandResponse) error {

	s.mu.Lock()
	defer s.mu.Unlock()

	historyPath := filepath.Join(config.AppConfig.GetCommandsDir(), "history.ndjson")
	data, err := json.Marshal(cmd)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(historyPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write(append(data, '\n')); err != nil {
		return err
	}
	return nil
}


func (s *JSONStore) GetCommandHistory(limit int) ([]models.CommandResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	historyPath := filepath.Join(config.AppConfig.GetCommandsDir(), "history.ndjson")
	data, err := os.ReadFile(historyPath)
	if err != nil {
		// Fallback to old history.json if exists
		oldPath := filepath.Join(config.AppConfig.GetCommandsDir(), "history.json")
		if oldData, err := os.ReadFile(oldPath); err == nil {
			var history models.CommandHistory
			json.Unmarshal(oldData, &history)
			return history.Commands, nil
		}
		return nil, nil
	}

	lines := strings.Split(string(data), "\n")
	var commands []models.CommandResponse
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var cmd models.CommandResponse
		if err := json.Unmarshal([]byte(line), &cmd); err == nil {
			commands = append(commands, cmd)
		}
	}

	if limit > 0 && len(commands) > limit {
		return commands[len(commands)-limit:], nil
	}

	return commands, nil
}

