package clips

import (
	"testing"
)

func TestIndefiniteClip(t *testing.T) {
	clip := NewPause(0)

	if clip.Duration() != 0 {
		t.Errorf("Wrong duration: %s", clip.duration)
	}

	for range 42 {
		chunk, hasMore := clip.NextChunk()

		if !hasMore {
			t.Errorf("Clip is not infinite")
		}

		if chunk == nil {
			t.Errorf("Got a nil chunk")
		}
	}

	clip.Stop()

	_, hasMore := clip.NextChunk()

	if hasMore {
		t.Errorf("A stopped clip should not have any more chunks")
	}
}
