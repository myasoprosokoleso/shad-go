//go:build !solution

package keylock

import (
	"slices"
	"sync"
)

type KeyLock struct {
	mu         sync.Mutex
	lockedKeys map[string]chan struct{}
}

func New() *KeyLock {
	return &KeyLock{
		lockedKeys: make(map[string]chan struct{}),
	}
}

func (l *KeyLock) LockKeys(keys []string, cancel <-chan struct{}) (canceled bool, unlock func()) {
	select {
	case <-cancel:
		return true, func() {}
	default:
	}

	// ORDER BY keys to avoid deadlocks
	sortedKeys := makeUnique(keys)
	slices.Sort(sortedKeys)

	for i, key := range sortedKeys {
		l.mu.Lock()
		ch := l.lockedKeys[key]

		if ch == nil {
			ch = make(chan struct{}, 1)
			ch <- struct{}{}
			l.lockedKeys[key] = ch
			l.mu.Unlock()
		} else {
			l.mu.Unlock()

			select {
			case <-cancel:
				l.mu.Lock()
				defer l.mu.Unlock()
				for _, key := range sortedKeys[:i] {
					<-l.lockedKeys[key]
				}

				return true, func() {}
			case ch <- struct{}{}:
			}
		}
	}

	return false, func() {
		l.mu.Lock()
		defer l.mu.Unlock()
		for _, key := range sortedKeys {
			<-l.lockedKeys[key]
		}
	}
}

func makeUnique[T comparable](s []T) []T {
	seen := make(map[T]struct{})
	res := make([]T, 0, len(s))
	for _, v := range s {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			res = append(res, v)
		}
	}
	return res
}
