package main

import "sync"

type concurrentMap[K comparable, V any] struct {
	mu      sync.RWMutex
	storage map[K]V
}

func newConcurrentMap[K comparable, V any]() *concurrentMap[K, V] {
	return &concurrentMap[K, V]{
		storage: make(map[K]V),
	}
}

func (cm *concurrentMap[K, V]) Get(key K) (V, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	value, ok := cm.storage[key]
	return value, ok
}

func (cm *concurrentMap[K, V]) Set(key K, value V) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.storage[key] = value
}

func (cm *concurrentMap[K, V]) Delete(key K) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.storage, key)
}
