//go:build !solution

package pubsub

import (
	"context"
	"errors"
	"iter"
	"maps"
	"runtime"
	"sync"
)

var (
	errPubSubClosed = errors.New("PubSub is closed")
	parallelJobsCnt = runtime.NumCPU()
)

var _ Subscription = (*MySubscription)(nil)

type MySubscription struct {
	msgHandler MsgHandler
	msgQueue   chan any
	publisher  *MyPubSub
	topic      string
}

func (s *MySubscription) Unsubscribe() {
	p := s.publisher
	p.mu.Lock()
	defer p.mu.Unlock()

	subs := p.subscribers[s.topic]
	_, ok := subs[s]
	if !ok {
		return
	}

	delete(subs, s)
	if len(subs) == 0 {
		delete(p.subscribers, s.topic)
	}
	close(s.msgQueue)
}

var _ PubSub = (*MyPubSub)(nil)

type MyPubSub struct {
	mu          sync.RWMutex
	isClosed    bool
	subscribers map[string]map[*MySubscription]struct{}
	listeners   sync.WaitGroup
}

func NewPubSub() PubSub {
	return &MyPubSub{subscribers: make(map[string]map[*MySubscription]struct{})}
}

func (p *MyPubSub) Subscribe(subj string, cb MsgHandler) (Subscription, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isClosed {
		return nil, errPubSubClosed
	}

	sub := &MySubscription{
		msgHandler: cb,
		msgQueue:   make(chan any, 100),
		publisher:  p,
		topic:      subj,
	}

	p.listeners.Go(func() {
		for msg := range sub.msgQueue {
			sub.msgHandler(msg)
		}
	})

	subs, ok := p.subscribers[subj]
	if !ok {
		subs = make(map[*MySubscription]struct{})
	}
	subs[sub] = struct{}{}
	p.subscribers[subj] = subs

	return sub, nil
}

func (p *MyPubSub) Publish(subj string, msg interface{}) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.isClosed {
		return errPubSubClosed
	}

	subs := p.subscribers[subj]
	subsCh := seqToChan(maps.Keys(subs))

	var wg sync.WaitGroup
	for range parallelJobsCnt {
		wg.Go(func() {
			for sub := range subsCh {
				sub.msgQueue <- msg
			}
		})
	}

	// Дождаться всех отправок перед разблокировкой, чтобы не писать в закрытый канал
	wg.Wait()
	return nil
}

func (p *MyPubSub) Close(ctx context.Context) error {
	p.mu.Lock()
	if p.isClosed {
		p.mu.Unlock()
		return nil
	}
	p.isClosed = true

	allSubs := make([]*MySubscription, 0, len(p.subscribers))
	for _, subs := range p.subscribers {
		for sub := range subs {
			allSubs = append(allSubs, sub)
		}
	}
	p.mu.Unlock()

	subsCh := make(chan *MySubscription, parallelJobsCnt)
	go func() {
		for _, sub := range allSubs {
			subsCh <- sub
		}
		close(subsCh)
	}()

	var wg sync.WaitGroup
	for range parallelJobsCnt {
		wg.Go(func() {
			for sub := range subsCh {
				sub.Unsubscribe()
			}
		})
	}
	wg.Wait() // дождаться закытия всех каналов

	done := make(chan struct{})
	go func() {
		p.listeners.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}

func seqToChan[T any](seq iter.Seq[T]) <-chan T {
	res := make(chan T, parallelJobsCnt)
	go func() {
		for v := range seq {
			res <- v
		}
		close(res)
	}()
	return res
}
