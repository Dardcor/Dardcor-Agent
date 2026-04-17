package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"dardcor-agent/config"
	"dardcor-agent/models"

	"github.com/google/uuid"
)

type JSONStore struct {
	mu sync.RWMutex
}

var Store *JSONStore

func Init() {
	Store = &JSONStore{}
}

func (s *JSONStore) SaveConversation(conv *models.Conversation, source string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	conv.UpdatedAt = time.Now()
	data, err := json.MarshalIndent(conv, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal conversation: %w", err)
	}

	dir := config.AppConfig.GetConversationsDir(source)
	path := filepath.Join(dir, conv.ID+".json")
	return os.WriteFile(path, data, 0644)
}

func (s *JSONStore) LoadConversation(id string, source string) (*models.Conversation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dir := config.AppConfig.GetConversationsDir(source)
	path := filepath.Join(dir, id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("conversation not found")
	}

	var conv models.Conversation
	if err := json.Unmarshal(data, &conv); err != nil {
		return nil, fmt.Errorf("failed to parse conversation: %w", err)
	}

	return &conv, nil
}

func (s *JSONStore) ListConversations(source string) ([]models.Conversation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dir := config.AppConfig.GetConversationsDir(source)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var conversations []models.Conversation
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue
		}

		var conv models.Conversation
		if err := json.Unmarshal(data, &conv); err != nil {
			continue
		}

		msgCount := len(conv.Messages)
		conv.Messages = nil
		if msgCount > 0 {
			conv.Title = fmt.Sprintf("%s (%d messages)", conv.Title, msgCount)
		}

		conversations = append(conversations, conv)
	}

	return conversations, nil
}

func (s *JSONStore) DeleteConversation(id string, source string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	dir := config.AppConfig.GetConversationsDir(source)
	path := filepath.Join(dir, id+".json")
	return os.Remove(path)
}

func (s *JSONStore) RenameConversation(id string, newTitle string, source string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	dir := config.AppConfig.GetConversationsDir(source)
	path := filepath.Join(dir, id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("conversation not found: %w", err)
	}

	var conv models.Conversation
	if err := json.Unmarshal(data, &conv); err != nil {
		return fmt.Errorf("failed to parse conversation: %w", err)
	}

	conv.Title = newTitle
	conv.UpdatedAt = time.Now()
	out, err := json.MarshalIndent(&conv, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal conversation: %w", err)
	}
	return os.WriteFile(path, out, 0644)
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

// addMessageLocked performs the read-modify-write atomically.
// Caller MUST hold s.mu (write lock) before calling.
func (s *JSONStore) addMessageLocked(convID string, msg models.Message, source string) error {
	dir := config.AppConfig.GetConversationsDir(source)
	path := filepath.Join(dir, convID+".json")

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("conversation not found")
	}

	var conv models.Conversation
	if err := json.Unmarshal(data, &conv); err != nil {
		return fmt.Errorf("failed to parse conversation: %w", err)
	}

	msg.ID = uuid.New().String()
	msg.Timestamp = time.Now()
	conv.Messages = append(conv.Messages, msg)
	conv.UpdatedAt = time.Now()

	out, err := json.MarshalIndent(&conv, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal conversation: %w", err)
	}
	return os.WriteFile(path, out, 0644)
}

func (s *JSONStore) AddMessage(convID string, msg models.Message, source string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.addMessageLocked(convID, msg, source)
}

func (s *JSONStore) SaveCommandHistory(cmd models.CommandResponse) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	historyPath := filepath.Join(config.AppConfig.GetCommandsDir(), "history.json")

	var history models.CommandHistory
	if data, err := os.ReadFile(historyPath); err == nil {
		json.Unmarshal(data, &history)
	}

	if len(history.Commands) >= 500 {
		history.Commands = history.Commands[len(history.Commands)-499:]
	}

	history.Commands = append(history.Commands, cmd)

	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(historyPath, data, 0644)
}

func (s *JSONStore) GetCommandHistory(limit int) ([]models.CommandResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	historyPath := filepath.Join(config.AppConfig.GetCommandsDir(), "history.json")

	var history models.CommandHistory
	if data, err := os.ReadFile(historyPath); err == nil {
		json.Unmarshal(data, &history)
	}

	if limit > 0 && limit < len(history.Commands) {
		return history.Commands[len(history.Commands)-limit:], nil
	}

	return history.Commands, nil
}
