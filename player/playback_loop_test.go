package player

import (
	"testing"
	"time"
)

func TestAddToQueue(t *testing.T) {
	if QueueSize() != 0 {
		t.Errorf("Queue should have been empty. Size: %d\n", QueueSize())
	}

	QueueClip(&testClip{})

	if QueueSize() != 1 {
		t.Errorf("Unexpected queue size %d. Should have been 0.", QueueSize())
	}
}

type testClip struct{}

func (clip *testClip) NextChunk() (*AudioChunk, bool) { return nil, false }

func (clip *testClip) Stop() {}

func (clip *testClip) Name() string { return "Test Clip" }

func (clip *testClip) Duration() time.Duration { return 0 }
