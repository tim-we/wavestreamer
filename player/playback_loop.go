package player

import (
	"log"

	"github.com/tim-we/wavestreamer/config"
	"github.com/tim-we/wavestreamer/utils"
)

type PlaybackLoop struct {
	NextAudioChunk  chan *AudioChunk
	currentClip     Clip
	name            string
	skipSignal      chan struct{}
	clipProvider    func() Clip
	normalize       bool
	clipEndCallback func(Clip, bool)
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

func (queue *PlaybackLoop) Run() {
	for {
		clip := queue.clipProvider()
		skipped := false

		if clip == nil {
			log.Printf("No more clips to play in queue %s.", queue.name)
			break
		}

		// Reset measured loudness for new clip
		var inputLoudness float32 = config.TARGET_MIN_RMS
		var lastGain float32 = 1.0

		for {
			// Check if there is a skip signal
			if utils.TryDropOne(queue.skipSignal) {
				clip.Stop()
				skipped = true
				break
			}

			chunk, hasMore := clip.NextChunk()

			if !hasMore || chunk == nil {
				// We have reached the end of clip
				break
			}

			if queue.normalize {
				inputLoudness = computeCurrentLoudness(inputLoudness, chunk)
				gain := computeTargetGain(chunk, inputLoudness)
				chunk.ApplyGain(lastGain, gain)
				lastGain = gain
			}

			queue.NextAudioChunk <- chunk
		}

		if queue.clipEndCallback != nil {
			queue.clipEndCallback(clip, skipped)
		}
	}
}

func (queue *PlaybackLoop) Skip() {
	queue.skipSignal <- struct{}{}
}

func (queue *PlaybackLoop) GetCurrentClip() Clip {
	return queue.currentClip
}

func (queue *PlaybackLoop) OnClipEnd(callback func(Clip, bool)) {
	if queue.clipEndCallback != nil {
		panic("OnClipEnd should only be called once")
	}

	queue.clipEndCallback = callback
}
