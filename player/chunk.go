package player

import "github.com/tim-we/wavestreamer/utils"

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
	a := startGain
	b := (endGain - startGain) / float32(max(1, chunk.Length))

	// Its not safe to just multiply by gain, it would cause clipping.
	// To avoid clipping we use a soft limiting function.
	// For optimization purposes we precompute some parameters:
	xThreshold, alpha := utils.SoftLimitParameters(0.5 * (startGain + endGain))

	// Apply soft limit with the computed parameters:
	for i := range chunk.Length {
		gain := a + float32(i)*b
		chunk.Left[i] = utils.SoftLimit(chunk.Left[i], xThreshold, gain, alpha)
		chunk.Right[i] = utils.SoftLimit(chunk.Right[i], xThreshold, gain, alpha)
	}
}
