//go:build !solution

package dupcall

import (
	"context"
	"sync"
)

type callContext struct {
	cancel  context.CancelFunc
	done    chan struct{}
	waiters int
	result  interface{}
	err     error
}

type Call struct {
	mu  sync.Mutex
	ctx *callContext
}

func (o *Call) Do(
	ctx context.Context,
	cb func(context.Context) (interface{}, error),
) (result interface{}, err error) {
	o.mu.Lock()
	currentCtx := o.ctx // работаем с копией o.ctx

	if currentCtx != nil {
		currentCtx.waiters++
		o.mu.Unlock()
	} else {
		// Новый контекст, чтобы не отменять все cb при отмене одного Do
		cbContext, cancel := context.WithCancel(context.Background())
		currentCtx = &callContext{
			cancel:  cancel,
			done:    make(chan struct{}),
			waiters: 1,
		}
		o.ctx = currentCtx
		o.mu.Unlock()

		// Не ждем завершения cb, чтобы следить за отменой контекста
		go func() {
			currentCtx.result, currentCtx.err = cb(cbContext)

			o.mu.Lock()
			o.ctx = nil
			close(currentCtx.done)
			o.mu.Unlock()
		}()
	}

	select {
	case <-currentCtx.done:
		result, err = currentCtx.result, currentCtx.err
	case <-ctx.Done():
		o.mu.Lock()
		currentCtx.waiters--
		if currentCtx.waiters == 0 {
			currentCtx.cancel()
		}
		o.mu.Unlock()

		result, err = nil, ctx.Err()
	}

	return
}
