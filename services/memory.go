package services

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type MemoryEntry struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

type EpisodicLog struct {
	TaskID  string `json:"task_id"`
	Action  string `json:"action"`
	Outcome string `json:"outcome"`
	Insight string `json:"insight"`
}

type MemoryService struct {
	path          string
	episodicPath  string
	data          map[string]interface{}
	episodicData  []EpisodicLog
	mutex         sync.RWMutex
	dirty         bool
	dirtyEpisodic bool
	flushCh       chan struct{}
	doneCh        chan struct{}
}

func NewMemoryService(dataDir string) *MemoryService {
	p := filepath.Join(dataDir, "memory.json")
	ePath := filepath.Join(dataDir, "episodic.json")
	ms := &MemoryService{
		path:         p,
		episodicPath: ePath,
		data:         make(map[string]interface{}),
		episodicData: make([]EpisodicLog, 0),
		flushCh:      make(chan struct{}, 1),
		doneCh:       make(chan struct{}),
	}
	ms.load()
	go ms.flushLoop()
	return ms
}

func (ms *MemoryService) flushLoop() {
	defer close(ms.doneCh)
	for range ms.flushCh {
		ms.mutex.Lock()
		ms.dirty = false
		ms.dirtyEpisodic = false
		raw, err1 := json.MarshalIndent(ms.data, "", "  ")
		rawEpi, err2 := json.MarshalIndent(ms.episodicData, "", "  ")
		ms.mutex.Unlock()

		if err1 == nil {
			os.MkdirAll(filepath.Dir(ms.path), 0755)
			os.WriteFile(ms.path, raw, 0644)
		}
		if err2 == nil {
			os.MkdirAll(filepath.Dir(ms.episodicPath), 0755)
			os.WriteFile(ms.episodicPath, rawEpi, 0644)
		}
	}
}

func (ms *MemoryService) scheduleSave() {
	ms.dirty = true
	select {
	case ms.flushCh <- struct{}{}:
	default:
	}
}

func (ms *MemoryService) scheduleEpisodicSave() {
	ms.dirtyEpisodic = true
	select {
	case ms.flushCh <- struct{}{}:
	default:
	}
}

func (ms *MemoryService) Flush() {
	ms.mutex.Lock()
	if !ms.dirty && !ms.dirtyEpisodic {
		ms.mutex.Unlock()
		return
	}
	ms.dirty = false
	ms.dirtyEpisodic = false
	raw, err1 := json.MarshalIndent(ms.data, "", "  ")
	rawEpi, err2 := json.MarshalIndent(ms.episodicData, "", "  ")
	ms.mutex.Unlock()

	if err1 == nil {
		os.MkdirAll(filepath.Dir(ms.path), 0755)
		os.WriteFile(ms.path, raw, 0644)
	}
	if err2 == nil {
		os.MkdirAll(filepath.Dir(ms.episodicPath), 0755)
		os.WriteFile(ms.episodicPath, rawEpi, 0644)
	}
}

func (ms *MemoryService) load() {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	if _, err := os.Stat(ms.path); err == nil {
		if raw, err := os.ReadFile(ms.path); err == nil {
			json.Unmarshal(raw, &ms.data)
		}
	}

	if _, err := os.Stat(ms.episodicPath); err == nil {
		if raw, err := os.ReadFile(ms.episodicPath); err == nil {
			json.Unmarshal(raw, &ms.episodicData)
		}
	}
}

func (ms *MemoryService) Set(key string, value interface{}) {
	ms.mutex.Lock()
	ms.data[key] = value
	ms.scheduleSave()
	ms.mutex.Unlock()
}

func (ms *MemoryService) Get(key string) interface{} {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()
	return ms.data[key]
}

func (ms *MemoryService) Recall(key string) (interface{}, bool) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()
	val, exists := ms.data[key]
	return val, exists
}

func (ms *MemoryService) Delete(key string) {
	ms.mutex.Lock()
	delete(ms.data, key)
	ms.scheduleSave()
	ms.mutex.Unlock()
}

func (ms *MemoryService) RecordEpisode(taskID, action, outcome, insight string) {
	ms.mutex.Lock()
	ms.episodicData = append(ms.episodicData, EpisodicLog{
		TaskID:  taskID,
		Action:  action,
		Outcome: outcome,
		Insight: insight,
	})
	ms.scheduleEpisodicSave()
	ms.mutex.Unlock()
}

func (ms *MemoryService) RetrieveEpisodesByOutcome(outcome string) []EpisodicLog {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()
	var results []EpisodicLog
	for _, ep := range ms.episodicData {
		if strings.Contains(strings.ToLower(ep.Outcome), strings.ToLower(outcome)) {
			results = append(results, ep)
		}
	}
	return results
}

func (ms *MemoryService) GetAll() map[string]interface{} {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()
	cp := make(map[string]interface{}, len(ms.data))
	for k, v := range ms.data {
		cp[k] = v
	}
	return cp
}

func (ms *MemoryService) Search(query string) map[string]interface{} {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	if query == "" {
		return ms.GetAll()
	}

	keywords := strings.Fields(strings.ToLower(query))
	result := make(map[string]interface{})

	for k, v := range ms.data {
		lk := strings.ToLower(k)
		for _, kw := range keywords {
			if strings.Contains(lk, kw) {
				result[k] = v
				break
			}
		}
	}
	return result
}

func (ms *MemoryService) GenerateRepoMap(directory string) (string, error) {
	var sb strings.Builder
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == ".git" || info.Name() == "node_modules" || info.Name() == "dist" {
				return filepath.SkipDir
			}
			return nil
		}
		rel, _ := filepath.Rel(directory, path)
		sb.WriteString(rel + "\n")
		return nil
	})
	return sb.String(), err
}

func (ms *MemoryService) Count() int {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()
	return len(ms.data)
}

func (ms *MemoryService) Keys() []string {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()
	keys := make([]string, 0, len(ms.data))
	for k := range ms.data {
		keys = append(keys, k)
	}
	return keys
}
