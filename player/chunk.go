package player

import (
	"github.com/tim-we/wavestreamer/utils"
)

type AudioChunk struct {
	// Left channel samples.
	Left []float32

	// Right channel samples.
	Right []float32

	// Number of samples in this chunk (up to FRAMES_PER_BUFFER).
	Length int

	// Root mean square of this chunk's audio (stereo average).
	RMS float32

	// Maximum absolute sample value across both channels.
	Peak float32
}

func (chunk *AudioChunk) ApplyGain(startGain, endGain float32) {
	if startGain == 1 && endGain == 1 {
		// Nothing to do.
		return
	}

	a := startGain
	b := (endGain - startGain) / float32(max(1, chunk.Length))

	// Linearly interpolate gain and apply gain with soft limit:
	for i := range chunk.Length {
		gain := a + float32(i)*b
		chunk.Left[i] = utils.SoftLimitGain(chunk.Left[i], gain)
		chunk.Right[i] = utils.SoftLimitGain(chunk.Right[i], gain)
	}
}
