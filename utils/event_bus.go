package utils

import (
	"context"
	"runtime"
	"sync"
)

// EventBus is a generic, non-blocking publish/subscribe event bus.
//
// Publishing is intentionally lossy: if the bus or a subscriber is congested,
// events are dropped instead of blocking the publisher.
type EventBus[Event any] struct {
	mu          sync.RWMutex
	subscribers map[chan Event]struct{}

	incoming   chan Event
	bufferSize int
}

// NewEventBus creates a new EventBus.
//
// busBuffer controls how many events can be queued for fan-out before
// published events are dropped.
//
// subscriberBuffer controls how many events an individual subscriber can lag
// behind before events for that subscriber are dropped.
func NewEventBus[Event any](busBuffer, subscriberBuffer int) *EventBus[Event] {
	b := &EventBus[Event]{
		subscribers: make(map[chan Event]struct{}),
		incoming:    make(chan Event, busBuffer),
		bufferSize:  subscriberBuffer,
	}

	go b.run()
	return b
}

// Publish broadcasts an event to all current subscribers.
//
// Publish never blocks. If the internal bus buffer is full, the event is
// dropped silently.
func (b *EventBus[Event]) Publish(event Event) {
	select {
	case b.incoming <- event:
		// queued
	default:
		// bus congested, drop event
	}
}

// Subscribe registers a new subscriber.
//
// It returns a receive-only channel on which events are delivered and a
// function that must be called to unsubscribe. Once unsubscribed, the
// channel is closed.
func (b *EventBus[Event]) Subscribe() (<-chan Event, func()) {
	ch := make(chan Event, b.bufferSize)

	b.mu.Lock()
	b.subscribers[ch] = struct{}{}
	b.mu.Unlock()

	unsubscribe := func() {
		b.mu.Lock()
		if _, ok := b.subscribers[ch]; ok {
			delete(b.subscribers, ch)
			close(ch)
		}
		b.mu.Unlock()
	}

	return ch, unsubscribe
}

// SubscribeContext registers a new subscriber that is automatically
// unsubscribed when the provided context is done.
//
// The returned receive-only channel is closed when the context is canceled or
// the request ends. Each subscriber has its own buffer; if the buffer is full,
// events for that subscriber are dropped.
func (b *EventBus[Event]) SubscribeContext(ctx context.Context) <-chan Event {
	ch := make(chan Event, b.bufferSize)

	b.mu.Lock()
	b.subscribers[ch] = struct{}{}
	b.mu.Unlock()

	go func() {
		<-ctx.Done()
		b.mu.Lock()
		if _, ok := b.subscribers[ch]; ok {
			delete(b.subscribers, ch)
			close(ch)
		}
		b.mu.Unlock()
	}()

	return ch
}

func (b *EventBus[Event]) run() {
	for event := range b.incoming {
		b.mu.RLock()
		for ch := range b.subscribers {
			select {
			case ch <- event:
				// delivered
			default:
				// subscriber too slow, drop
			}
		}
		b.mu.RUnlock()
		runtime.Gosched()
	}
}
