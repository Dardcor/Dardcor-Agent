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

func (s *JSONStore) SaveConversation(conv *models.Conversation) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	conv.UpdatedAt = time.Now()
	data, err := json.MarshalIndent(conv, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal conversation: %w", err)
	}

	path := filepath.Join(config.AppConfig.GetConversationsDir(), conv.ID+".json")
	return os.WriteFile(path, data, 0644)
}

func (s *JSONStore) LoadConversation(id string) (*models.Conversation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := filepath.Join(config.AppConfig.GetConversationsDir(), id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("conversation not found: %w", err)
	}

	var conv models.Conversation
	if err := json.Unmarshal(data, &conv); err != nil {
		return nil, fmt.Errorf("failed to parse conversation: %w", err)
	}

	return &conv, nil
}

func (s *JSONStore) ListConversations() ([]models.Conversation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dir := config.AppConfig.GetConversationsDir()
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

func (s *JSONStore) DeleteConversation(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(config.AppConfig.GetConversationsDir(), id+".json")
	return os.Remove(path)
}

func (s *JSONStore) RenameConversation(id string, newTitle string) error {
	conv, err := s.LoadConversation(id)
	if err != nil {
		return err
	}

	conv.Title = newTitle
	return s.SaveConversation(conv)
}

func (s *JSONStore) CreateConversation(title string) (*models.Conversation, error) {
	conv := &models.Conversation{
		ID:        uuid.New().String(),
		Title:     title,
		Messages:  []models.Message{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.SaveConversation(conv); err != nil {
		return nil, err
	}

	return conv, nil
}

func (s *JSONStore) AddMessage(convID string, msg models.Message) error {
	conv, err := s.LoadConversation(convID)
	if err != nil {
		return err
	}

	msg.ID = uuid.New().String()
	msg.Timestamp = time.Now()
	conv.Messages = append(conv.Messages, msg)

	return s.SaveConversation(conv)
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
