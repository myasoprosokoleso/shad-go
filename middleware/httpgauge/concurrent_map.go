package httpgauge

import (
	"maps"
	"sync"
)

type concurrentMap struct {
	mu      sync.RWMutex
	storage map[string]int
}

func newConcurrentMap() *concurrentMap {
	return &concurrentMap{
		storage: make(map[string]int),
	}
}

func (cm *concurrentMap) Inc(key string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.storage[key]++
}

func (cm *concurrentMap) GetCopy() map[string]int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return maps.Clone(cm.storage)
}
