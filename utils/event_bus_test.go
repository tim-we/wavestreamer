package utils

import (
	"context"
	"testing"
	"time"
)

type testEvent struct {
	Data string
}

func TestEventBus_PublishAndReceive(t *testing.T) {
	bus := NewEventBus[testEvent](10, 2)

	ctx := t.Context()

	ch := bus.SubscribeContext(ctx)

	// Publish an event
	event := testEvent{Data: "hello"}
	bus.Publish(event)

	select {
	case ev := <-ch:
		if ev != event {
			t.Errorf("expected %v, got %v", event, ev)
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for event")
	}
}

func TestEventBus_ContextCancellation(t *testing.T) {
	bus := NewEventBus[testEvent](10, 2)

	ctx, cancel := context.WithCancel(context.Background())
	ch := bus.SubscribeContext(ctx)

	// Cancel the context
	cancel()

	// Wait a bit for unsubscribe to take effect
	time.Sleep(50 * time.Millisecond)

	// Publish something, channel should be closed
	bus.Publish(testEvent{Data: "test"})

	select {
	case _, ok := <-ch:
		if ok {
			t.Error("expected channel to be closed after context cancel")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout waiting for channel to close")
	}
}

func TestEventBus_MultipleSubscribers(t *testing.T) {
	bus := NewEventBus[testEvent](10, 2)

	ctx1 := t.Context()
	ctx2 := t.Context()

	ch1 := bus.SubscribeContext(ctx1)
	ch2 := bus.SubscribeContext(ctx2)

	event := testEvent{Data: "broadcast"}
	bus.Publish(event)

	for _, ch := range []<-chan testEvent{ch1, ch2} {
		select {
		case ev := <-ch:
			if ev != event {
				t.Errorf("expected %v, got %v", event, ev)
			}
		case <-time.After(time.Second):
			t.Error("timeout waiting for subscriber")
		}
	}
}
