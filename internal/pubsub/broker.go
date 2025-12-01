package pubsub

import (
	"context"
	"sync"
)

const bufferSize = 64

type Broker[E ~string, T any] struct {
	subs      map[chan Event[E, T]]struct{}
	mu        sync.RWMutex
	done      chan struct{}
	subCount  int
	maxEvents int
}

func NewBroker[E ~string, T any]() *Broker[E, T] {
	return NewBrokerWithOptions[E, T](bufferSize, 1000)
}

func NewBrokerWithOptions[E ~string, T any](channelBufferSize, maxEvents int) *Broker[E, T] {
	b := &Broker[E, T]{
		subs:      make(map[chan Event[E, T]]struct{}),
		done:      make(chan struct{}),
		subCount:  0,
		maxEvents: maxEvents,
	}
	return b
}

func (b *Broker[E, T]) Shutdown() {
	select {
	case <-b.done: // Already closed
		return
	default:
		close(b.done)
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	for ch := range b.subs {
		delete(b.subs, ch)
		close(ch)
	}

	b.subCount = 0
}

func (b *Broker[E, T]) Subscribe(ctx context.Context) <-chan Event[E, T] {
	b.mu.Lock()
	defer b.mu.Unlock()

	select {
	case <-b.done:
		ch := make(chan Event[E, T])
		close(ch)
		return ch
	default:
	}

	sub := make(chan Event[E, T], bufferSize)
	b.subs[sub] = struct{}{}
	b.subCount++

	go func() {
		<-ctx.Done()

		b.mu.Lock()
		defer b.mu.Unlock()

		select {
		case <-b.done:
			return
		default:
		}

		delete(b.subs, sub)
		close(sub)
		b.subCount--
	}()

	return sub
}

func (b *Broker[E, T]) GetSubscriberCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.subCount
}

func (b *Broker[E, T]) Publish(t E, payload T) {
	b.mu.RLock()
	select {
	case <-b.done:
		b.mu.RUnlock()
		return
	default:
	}

	subscribers := make([]chan Event[E, T], 0, len(b.subs))
	for sub := range b.subs {
		subscribers = append(subscribers, sub)
	}
	b.mu.RUnlock()

	event := Event[E, T]{Type: t, Payload: payload}

	for _, sub := range subscribers {
		select {
		case sub <- event:
		default:
			// Channel is full, subscriber is slow - skip this event
			// This prevents blocking the publisher
		}
	}
}
