package clips

import (
	"fmt"
	"time"

	"github.com/tim-we/wavestreamer/player"
)

type PauseClip struct {
	duration time.Duration
	progress time.Duration
}

var emptyChunk = player.AudioChunk{
	Left:   make([]float32, player.FRAMES_PER_BUFFER),
	Right:  make([]float32, player.FRAMES_PER_BUFFER),
	Length: player.FRAMES_PER_BUFFER,
}

const emptyChunkDuration = (player.FRAMES_PER_BUFFER * time.Second) / player.SAMPLE_RATE

func NewPause() *PauseClip {
	clip := PauseClip{
		duration: 10 * time.Second,
		progress: 0,
	}

	return &clip
}

func (clip *PauseClip) NextChunk() (*player.AudioChunk, bool) {
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
