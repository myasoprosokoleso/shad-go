//go:build !solution

package batcher

import (
	"sync"

	"gitlab.com/slon/shad-go/batcher/slow"
)

type batch struct {
	result   any
	ready    chan struct{}
	isStable bool
}

type Batcher struct {
	value   *slow.Value
	mu      sync.Mutex
	current *batch
}

func NewBatcher(v *slow.Value) *Batcher {
	return &Batcher{value: v}
}

func (b *Batcher) Load() any {
	b.mu.Lock()
	// Сохраняем локальный батч, работаем с ним
	currentBatch := b.current
	if currentBatch != nil {
		b.mu.Unlock()
		<-currentBatch.ready
	} else {
		currentBatch = &batch{ready: make(chan struct{})}
		b.current = currentBatch
		b.mu.Unlock()

		lastResult := b.value.Load()
		// b.value may have changed during the Loading...
		isStable := lastResult == b.value.Load()

		currentBatch.result, currentBatch.isStable = lastResult, isStable

		b.mu.Lock()
		// Очищаем указатель на глобальный батч
		b.current = nil
		b.mu.Unlock()

		close(currentBatch.ready)
	}

	if !currentBatch.isStable {
		return b.Load()
	}
	return currentBatch.result
}
