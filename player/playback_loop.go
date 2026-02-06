package player

import (
	"log"
	"time"

	"github.com/tim-we/wavestreamer/config"
	"github.com/tim-we/wavestreamer/utils"
)

type PlaybackLoop struct {
	NextAudioChunk    chan *AudioChunk
	ClipStartCallback func(Clip)
	currentClip       Clip
	name              string
	skipSignal        chan struct{}
	clipProvider      func() Clip
	normalize         bool
	clipEndCallback   func(Clip, bool)
}

func NewPlaybackLoop(name string, normalize bool, clipProvider func() Clip) *PlaybackLoop {
	return &PlaybackLoop{
		NextAudioChunk: make(chan *AudioChunk, 2),
		name:           name,
		clipProvider:   clipProvider,
		normalize:      normalize,
		skipSignal:     make(chan struct{}, 1),
	}
}

func (loop *PlaybackLoop) Run() {
	for {
		clip := loop.clipProvider()
		skipped := false

		if clip == nil {
			log.Printf("No more clips to play in %s.", loop.name)
			break
		}

		if loop.ClipStartCallback != nil {
			loop.ClipStartCallback(clip)
		}
		loop.currentClip = clip

		// Reset measured loudness for new clip
		var inputLoudness float32 = config.TARGET_MIN_RMS
		var lastGain float32 = 1.0

		// We perform this check only once per clip. That way it is a cheap operation and
		// does not cause weird audio glitches when we dynamically toggle features like normalization.
		reduceCPULoad := utils.ShouldReduceCPU()

		for {
			// Check if there is a skip signal
			if utils.TryDropOne(loop.skipSignal) {
				clip.Stop()
				skipped = true
				break
			}

			chunk, hasMore := clip.NextChunk()

			if !hasMore || chunk == nil {
				// We have reached the end of clip
				break
			}

			if !reduceCPULoad && loop.normalize {
				inputLoudness = computeCurrentLoudness(inputLoudness, chunk)
				gain := computeTargetGain(chunk, inputLoudness)
				chunk.ApplyGain(lastGain, gain)
				lastGain = gain
			}

			loop.NextAudioChunk <- chunk
		}

		if loop.clipEndCallback != nil {
			loop.clipEndCallback(clip, skipped)
		}

		if reduceCPULoad {
			time.Sleep(20 * time.Millisecond)
		}
	}
}

func (loop *PlaybackLoop) Skip() {
	select {
	case loop.skipSignal <- struct{}{}:
		// Skip signal sent
	default:
		// Skip already pending, ignore
	}

}

func (loop *PlaybackLoop) GetCurrentClip() Clip {
	return loop.currentClip
}

func (loop *PlaybackLoop) OnClipEnd(callback func(Clip, bool)) {
	if loop.clipEndCallback != nil {
		panic("OnClipEnd should only be called once")
	}

	loop.clipEndCallback = callback
}
