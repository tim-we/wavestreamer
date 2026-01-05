package clips

import (
	"fmt"
	"time"

	"github.com/tim-we/wavestreamer/config"
	"github.com/tim-we/wavestreamer/player"
)

type PauseClip struct {
	duration        time.Duration
	progress        time.Duration
	hidden          bool
	manuallyStopped bool
}

var emptyChunk = player.AudioChunk{
	Left:   make([]float32, config.FRAMES_PER_BUFFER),
	Right:  make([]float32, config.FRAMES_PER_BUFFER),
	Length: config.FRAMES_PER_BUFFER,
}

const emptyChunkDuration = (config.FRAMES_PER_BUFFER * time.Second) / config.SAMPLE_RATE

// NewPause creates a new Pause clip with the given duration.
// Pass 0 to create an indefinite pause clip.
func NewPause(duration time.Duration) *PauseClip {
	clip := PauseClip{
		duration: duration,
		progress: 0,
		hidden:   duration < 2*time.Second,
	}

	return &clip
}

func (clip *PauseClip) NextChunk() (*player.AudioChunk, bool) {
	clip.progress += emptyChunkDuration
	hasMore := clip.progress < clip.duration
	if clip.duration == 0 {
		// Indefinite clip
		hasMore = true
	}
	if clip.manuallyStopped {
		hasMore = false
	}
	return &emptyChunk, hasMore
}

func (clip *PauseClip) Stop() {
	clip.manuallyStopped = true

	if clip.progress <= time.Second {
		// Hide sub-second pauses from the UI. Those typically occur during short GPIO button presses.
		clip.hidden = true
	}
	if clip.duration > 0 {
		clip.progress = clip.duration
	}
}

func (clip *PauseClip) Name() string {
	if clip.duration == 0 {
		return "Pause"
	}
	return fmt.Sprintf("Pause %s", formatDuration(clip.duration))
}

func (clip *PauseClip) Duration() time.Duration {
	return clip.duration
}

func (clip *PauseClip) Duplicate() player.Clip {
	return NewPause(clip.duration)
}

func (clip *PauseClip) Hidden() bool {
	return clip.hidden
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
