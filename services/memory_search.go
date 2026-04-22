package services

import (
	"fmt"
	"strings"
	"sync"

	"dardcor-agent/models"
)

type MemorySearchService struct {
	store *MemoryStore
	mu    sync.RWMutex
}

func NewMemorySearchService(store *MemoryStore) *MemorySearchService {
	return &MemorySearchService{store: store}
}

func (s *MemorySearchService) Search(query string, limit int, project string) ([]*models.MemSearchResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 20
	}

	obs, err := s.store.SearchFTS(query, limit, project)
	if err != nil {
		return nil, fmt.Errorf("search FTS failed: %w", err)
	}

	results := make([]*models.MemSearchResult, 0, len(obs))
	for _, o := range obs {
		snippet := o.Content
		if len(snippet) > 200 {
			snippet = snippet[:200]
		}
		results = append(results, &models.MemSearchResult{
			Observation: *o,
			Score:       1.0,
			Snippet:     snippet,
		})
	}
	return results, nil
}

func (s *MemorySearchService) Timeline(anchorID int64, depthBefore, depthAfter int, project string) (*models.MemTimeline, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	timeline, err := s.store.GetTimeline(anchorID, depthBefore, depthAfter, project)
	if err != nil {
		return nil, fmt.Errorf("get timeline failed: %w", err)
	}
	return timeline, nil
}

func (s *MemorySearchService) GetObservations(ids []int64) ([]*models.MemObservation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	obs, err := s.store.GetObservationsByIDs(ids)
	if err != nil {
		return nil, fmt.Errorf("get observations by IDs failed: %w", err)
	}
	return obs, nil
}

func (s *MemorySearchService) InjectContext(project string, query string, maxTokens int) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	maxChars := maxTokens * 4
	if maxChars <= 0 {
		maxChars = 4000
	}

	var allObs []*models.MemObservation
	if project != "" {
		obs, err := s.store.GetObservationsByProject(project, 15)
		if err != nil {
			return "", fmt.Errorf("get observations by project failed: %w", err)
		}
		allObs = obs
	} else {
		sessions, err := s.store.GetRecentSessions(5)
		if err != nil {
			return "", fmt.Errorf("get recent sessions failed: %w", err)
		}
		for _, sess := range sessions {
			obs, err := s.store.GetObservationsBySession(sess.SessionID)
			if err != nil {
				continue
			}
			if len(obs) > 3 {
				obs = obs[len(obs)-3:]
			}
			allObs = append(allObs, obs...)
		}
	}

	if len(allObs) == 0 {
		return "", nil
	}

	formatted := s.FormatObservationsForPrompt(allObs, maxChars)
	if formatted == "" {
		return "", nil
	}

	return "### Past Context\n\n" + formatted, nil
}

func (s *MemorySearchService) FormatObservationsForPrompt(obs []*models.MemObservation, maxChars int) string {
	var sb strings.Builder

	for _, o := range obs {
		text := o.Summary
		if text == "" {
			text = o.Content
			if len(text) > 300 {
				text = text[:300]
			}
		}
		line := fmt.Sprintf("[%s] %s\n", o.Type, text)
		sb.WriteString(line)
	}

	result := sb.String()
	if maxChars > 0 && len(result) > maxChars {
		suffix := "... (truncated)"
		cutAt := maxChars - len(suffix)
		if cutAt < 0 {
			cutAt = 0
		}
		result = result[:cutAt] + suffix
	}
	return result
}
