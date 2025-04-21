package clips

import (
	"fmt"
	"time"

	"github.com/tim-we/wavestreamer/config"
	"github.com/tim-we/wavestreamer/player"
)

type PauseClip struct {
	duration time.Duration
	progress time.Duration
}

var emptyChunk = player.AudioChunk{
	Left:   make([]float32, config.FRAMES_PER_BUFFER),
	Right:  make([]float32, config.FRAMES_PER_BUFFER),
	Length: config.FRAMES_PER_BUFFER,
}

const emptyChunkDuration = (config.FRAMES_PER_BUFFER * time.Second) / config.SAMPLE_RATE

func NewPause(duration time.Duration) *PauseClip {
	clip := PauseClip{
		duration: duration,
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
	return fmt.Sprintf("Pause %s", formatDuration(clip.duration))
}

func (clip *PauseClip) Duration() time.Duration {
	return clip.duration
}

func formatDuration(d time.Duration) string {
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60

	if minutes == 0 {
		return fmt.Sprintf("%ds", seconds)
	}

	if seconds == 0 {
		return fmt.Sprintf("%dmin", minutes)
	}

	return fmt.Sprintf("%dmin %ds", minutes, seconds)
}
