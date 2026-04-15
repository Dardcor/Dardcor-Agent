package services

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type MemoryEntry struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

type MemoryService struct {
	path  string
	data  map[string]interface{}
	mutex sync.RWMutex
}

func NewMemoryService(dataDir string) *MemoryService {
	p := filepath.Join(dataDir, "memory.json")
	ms := &MemoryService{
		path: p,
		data: make(map[string]interface{}),
	}
	ms.load()
	return ms
}

func (ms *MemoryService) load() {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	if _, err := os.Stat(ms.path); os.IsNotExist(err) {
		return
	}

	raw, err := os.ReadFile(ms.path)
	if err != nil {
		return
	}

	json.Unmarshal(raw, &ms.data)
}

func (ms *MemoryService) save() {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	raw, _ := json.MarshalIndent(ms.data, "", "  ")
	os.MkdirAll(filepath.Dir(ms.path), 0755)
	os.WriteFile(ms.path, raw, 0644)
}

func (ms *MemoryService) Set(key string, value interface{}) {
	ms.mutex.Lock()
	ms.data[key] = value
	ms.mutex.Unlock()
	ms.save()
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
	ms.mutex.Unlock()
	ms.save()
}

func (ms *MemoryService) GetAll() map[string]interface{} {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()
	copy := make(map[string]interface{}, len(ms.data))
	for k, v := range ms.data {
		copy[k] = v
	}
	return copy
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
