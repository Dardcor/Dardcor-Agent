package services

import (
	"dardcor-agent/models"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	ObsTypeTool      = "tool_use"
	ObsTypeOutput    = "output"
	ObsTypeDecision  = "decision"
	ObsTypeBugfix    = "bugfix"
	ObsTypeFeature   = "feature"
	ObsTypeRefactor  = "refactor"
	ObsTypeDiscovery = "discovery"
	ObsTypeChange    = "change"
)

var observationStopwords = map[string]bool{
	"the": true, "and": true, "for": true, "this": true, "that": true,
	"with": true, "from": true, "have": true, "been": true, "will": true,
	"are": true, "was": true, "not": true, "can": true, "its": true,
}

var filePathRe = regexp.MustCompile(`[a-zA-Z0-9_\-./\\]+\.[a-zA-Z]{2,6}`)

type ObservationService struct {
	store         *MemoryStore
	privacyFilter *PrivacyFilter
	mu            sync.Mutex
	sessionID     string
	project       string
}

func NewObservationService(store *MemoryStore, privacyFilter *PrivacyFilter) *ObservationService {
	return &ObservationService{
		store:         store,
		privacyFilter: privacyFilter,
	}
}

func (o *ObservationService) StartSession(sessionID, project string) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.sessionID = sessionID
	o.project = project
	if o.store == nil {
		return nil
	}
	return o.store.CreateSession(sessionID, project)
}

func (o *ObservationService) EndSession(summary string) error {
	o.mu.Lock()
	sessionID := o.sessionID
	o.sessionID = ""
	o.project = ""
	o.mu.Unlock()

	if o.store == nil || sessionID == "" {
		return nil
	}
	filtered := o.privacyFilter.Filter(summary)
	return o.store.CloseSession(sessionID, filtered)
}

func (o *ObservationService) CaptureToolUse(toolName, input, output string) error {
	content := "[" + toolName + "] input: " + input + " | output: " + output
	return o.CaptureObservation(ObsTypeTool, content, nil, nil)
}

func (o *ObservationService) CaptureOutput(content string) error {
	return o.CaptureObservation(ObsTypeOutput, content, nil, nil)
}

func (o *ObservationService) CaptureObservation(obsType, content string, files, concepts []string) error {
	filtered := o.privacyFilter.Filter(content)
	if filtered == "" {
		return nil
	}

	if files == nil {
		files = o.ExtractFilePaths(filtered)
	}
	if concepts == nil {
		concepts = o.ExtractConcepts(filtered)
	}

	o.mu.Lock()
	sessionID := o.sessionID
	project := o.project
	o.mu.Unlock()

	if o.store == nil {
		return nil
	}

	obs := &models.MemObservation{
		SessionID: sessionID,
		Type:      obsType,
		ObsType:   obsType,
		Content:   filtered,
		Summary:   "",
		Project:   project,
		Files:     files,
		Concepts:  concepts,
		CreatedAt: time.Now().UTC(),
	}
	_, err := o.store.AddObservation(obs)
	return err
}

func (o *ObservationService) ExtractConcepts(content string) []string {
	words := strings.Fields(strings.ToLower(content))
	freq := make(map[string]int)
	for _, w := range words {
		clean := strings.TrimFunc(w, func(r rune) bool {
			return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'))
		})
		if len(clean) > 4 && !observationStopwords[clean] {
			freq[clean]++
		}
	}

	type wf struct {
		word  string
		count int
	}
	var ranked []wf
	for w, c := range freq {
		ranked = append(ranked, wf{w, c})
	}
	for i := 0; i < len(ranked); i++ {
		for j := i + 1; j < len(ranked); j++ {
			if ranked[j].count > ranked[i].count {
				ranked[i], ranked[j] = ranked[j], ranked[i]
			}
		}
	}

	limit := 10
	if len(ranked) < limit {
		limit = len(ranked)
	}
	result := make([]string, limit)
	for i := 0; i < limit; i++ {
		result[i] = ranked[i].word
	}
	return result
}

func (o *ObservationService) ExtractFilePaths(content string) []string {
	matches := filePathRe.FindAllString(content, -1)
	var result []string
	seen := make(map[string]bool)
	for _, m := range matches {
		if (strings.ContainsRune(m, '/') || strings.ContainsRune(m, '\\')) && !seen[m] {
			seen[m] = true
			result = append(result, m)
		}
	}
	return result
}

func (o *ObservationService) GetCurrentSession() string {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.sessionID
}

func (o *ObservationService) GetCurrentProject() string {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.project
}
