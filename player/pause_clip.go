package player

import (
	"fmt"
	"time"
)

type PauseClip struct {
	duration time.Duration
	progress time.Duration
}

var emptyChunk = AudioChunk{
	Left:   make([]float32, FRAMES_PER_BUFFER),
	Right:  make([]float32, FRAMES_PER_BUFFER),
	Length: FRAMES_PER_BUFFER,
}

const emptyChunkDuration = (FRAMES_PER_BUFFER * time.Second) / SAMPLE_RATE

func NewPause() *PauseClip {
	clip := PauseClip{
		duration: 10 * time.Second,
		progress: 0,
	}

	return &clip
}

func (clip *PauseClip) NextChunk() (*AudioChunk, bool) {
	clip.progress += emptyChunkDuration
	return &emptyChunk, clip.progress < clip.duration
}

func (clip *PauseClip) Stop() {
	clip.progress = clip.duration
}

func (clip *PauseClip) Name() string {
	return fmt.Sprintf("Pause %v", clip.duration)
}

func (clip *PauseClip) Duration() int {
	return int(clip.duration.Seconds())
}
