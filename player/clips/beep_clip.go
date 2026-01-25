package clips

import (
	"fmt"
	"time"

	"github.com/tim-we/wavestreamer/config"
	"github.com/tim-we/wavestreamer/player"
)

const NUMBER_OF_CHUNKS = 10

type BeepClip struct {
	buffer chan *player.AudioChunk
}

// All audio chunks of the beep will be the same.
// Therefore we can compute it once and cache it.
var beepChunk *player.AudioChunk

func NewBeep() *BeepClip {
	buffer := make(chan *player.AudioChunk, NUMBER_OF_CHUNKS)
	defer close(buffer)

	if beepChunk == nil {
		beepChunk = &player.AudioChunk{
			Left:  make([]float32, config.FRAMES_PER_BUFFER),
			Right: make([]float32, config.FRAMES_PER_BUFFER),
		}
		generateWave(beepChunk)
	}

	for range NUMBER_OF_CHUNKS {
		buffer <- beepChunk
	}

	return &BeepClip{buffer: buffer}
}

func (clip *BeepClip) NextChunk() (*player.AudioChunk, bool) {
	chunk, hasMore := <-clip.buffer
	return chunk, hasMore
}

func (clip *BeepClip) Stop() {
	// This clip is so short, no point in providing custom stop logic.
}

func (clip *BeepClip) Duration() time.Duration {
	durationInSeconds := (NUMBER_OF_CHUNKS * config.FRAMES_PER_BUFFER) / config.SAMPLE_RATE
	return time.Duration(durationInSeconds) * time.Second
}

func (clip *BeepClip) Hidden() bool {
	return true
}

func (clip *BeepClip) Duplicate() player.Clip {
	return NewBeep()
}

const wavelengthInSamples = 64
const bitmask = wavelengthInSamples - 1
const slope = float32(2.0 / 32.0)
const beepVolume = 0.2

func (clip *BeepClip) Name() string {
	frequency := config.SAMPLE_RATE / wavelengthInSamples
	return fmt.Sprintf("Beep %v Hz", frequency)
}

func generateWave(chunk *player.AudioChunk) {
	for i := range config.FRAMES_PER_BUFFER {
		x := i & bitmask

		var v float32

		if x < 32 {
			v = slope*float32(x) - 1
		} else {
			v = 1 - slope*float32(x)
		}

		v = beepVolume * v

		chunk.Left[i] = v
		chunk.Right[i] = v
	}
}
