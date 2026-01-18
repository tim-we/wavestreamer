package player

import (
	"github.com/tim-we/wavestreamer/config"
	"github.com/tim-we/wavestreamer/utils"
)

func computeTargetGain(chunk *AudioChunk, inputLoudness float32) float32 {
	// We don't want to boost already loud signals or signals which are very quiet.
	if inputLoudness >= config.TARGET_MIN_RMS || inputLoudness < 0.001 {
		return 1
	}

	maxGain := utils.Clamp[float32](
		1, // no gain
		config.MAX_AMPLIFICATION,
		2, // implementation limit
	)

	// The gain is basically the ratio between current loudness and target loudness.
	gain := utils.Clamp(
		1.0, // minimum
		config.TARGET_MIN_RMS/max(0.01, inputLoudness), // ratio but protected against division by 0
		maxGain,
	)

	if chunk.RMS > config.TARGET_MIN_RMS {
		// Measure how much we are currently overshooting the target value
		over := utils.Clamp(0, (gain*chunk.RMS-config.TARGET_MIN_RMS)/config.TARGET_MIN_RMS, 1)
		// Lower the gain
		gain = utils.Lerp(gain, 1, over)
	}

	if chunk.Peak*gain > 1 {
		// Lower target gain to avoid over amplification
		gain = min(1, 1/chunk.Peak)
	}

	return gain
}

func computeCurrentLoudness(previousLoudness float32, chunk *AudioChunk) float32 {
	// 0.35 can be quite loud already
	const maxInfluenceLevel = max(config.TARGET_MIN_RMS, 0.35)

	// Louder chunks should have a faster impact, for quiet chunks the loudness should decay slower.
	factor := utils.Lerp(0.001, 0.2, min(chunk.RMS, maxInfluenceLevel)/maxInfluenceLevel)

	// Interpolate previous loudness value with current chunks loudness (RMS)
	return utils.Lerp(previousLoudness, chunk.RMS, factor)
}
