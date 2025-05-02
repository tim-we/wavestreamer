package player

/*
#cgo linux,arm64 LDFLAGS: -lportaudio
*/
import "C"

import (
	"log"

	"github.com/gordonklaus/portaudio"
	"github.com/tim-we/wavestreamer/config"
	"github.com/tim-we/wavestreamer/utils"
)

var userQueue = make([]Clip, 0, 12)

var currentlyPlaying string = "?"

var skipSignal = make(chan struct{}, 1)

func Start(clipProvider func() Clip, normalize bool) {
	nextClip := func() Clip {
		if len(userQueue) > 0 {
			clip := userQueue[0]
			userQueue = userQueue[1:]
			return clip
		}
		if clip := clipProvider(); clip != nil {
			return clip
		}
		// TODO: what do we do now?
		return nil
	}

	if err := portaudio.Initialize(); err != nil {
		log.Fatal(err)
	}
	defer portaudio.Terminate()

	devices, dev_err := portaudio.Devices()
	if dev_err != nil {
		log.Fatal(dev_err)
	}

	if len(devices) == 0 {
		log.Fatal("No audio devices found.")
	}

	nextAudioChunk := make(chan *AudioChunk, 2)

	playCallback := func(out [][]float32) {
		select {
		case chunk := <-nextAudioChunk:
			copy(out[0], chunk.Left)
			copy(out[1], chunk.Right)
		default:
			// Handle underflow (e.g., fill with silence)
			for i := range out[0] {
				out[0][i] = 0
				out[1][i] = 0
			}
		}
	}

	// Set up the PortAudio stream with a fixed buffer size
	stream, err := portaudio.OpenDefaultStream(
		0,                        // not reading any inputs (microphones)
		config.CHANNELS,          // output channels
		config.SAMPLE_RATE,       // output sample rate
		config.FRAMES_PER_BUFFER, // output buffer size
		playCallback,             // output buffer filling callback
	)

	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	if err := stream.Start(); err != nil {
		log.Fatal(err)
	}
	defer stream.Stop()

	for {
		clip := nextClip()

		if clip == nil {
			log.Print("No more clips to play.")
			break
		}

		log.Printf("Now playing %s", clip.Name())
		currentlyPlaying = clip.Name()
		addClipToHistory(clip)

		// Reset measured loudness for new clip
		var inputLoudness float32 = config.TARGET_MIN_RMS
		var lastGain float32 = 1.0

		for {
			if shouldSkipCurrentClip() {
				clip.Stop()
				break
			}

			chunk, hasMore := clip.NextChunk()

			if chunk != nil {
				inputLoudness = computeCurrentLoudness(inputLoudness, chunk)
				gain := computeTargetGain(chunk, inputLoudness)
				if normalize {
					chunk.ApplyGain(lastGain, gain)
				}
				lastGain = gain
				nextAudioChunk <- chunk
			}

			if !hasMore {
				break
			}
		}

		currentlyPlaying = "-"

	}
}

func QueueClip(clip Clip) {
	userQueue = append(userQueue, clip)
}

func QueueClipNext(clip Clip) {
	userQueue = append([]Clip{clip}, userQueue...)
}

func QueueSize() int {
	return len(userQueue)
}

func GetCurrentlyPlaying() string {
	return currentlyPlaying
}

func SkipCurrent() {
	skipSignal <- struct{}{}
}

func shouldSkipCurrentClip() bool {
	select {
	case <-skipSignal:
		return true
	default:
		return false
	}
}

func computeTargetGain(chunk *AudioChunk, inputLoudness float32) float32 {
	// We don't want to boost already loud signals or signals which are very quiet.
	if inputLoudness >= config.TARGET_MIN_RMS || inputLoudness < 0.001 {
		return 1
	}

	// The gain is basically the ratio between current loudness and target loudness.
	gain := utils.Clamp(
		1.0, // minimum
		config.TARGET_MIN_RMS/max(0.01, inputLoudness), // ratio but protected against division by 0
		2.0, // maximum (implementation limit)
	)

	// TODO
	if chunk.RMS > config.TARGET_MIN_RMS {
		rel := utils.Clamp(0, 2*(chunk.RMS-config.TARGET_MIN_RMS)/config.TARGET_MIN_RMS, 1)
		// bring gain closer to 1
		gain = rel*gain + (1-rel)*1
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
	factor := utils.Lerp(0.01, 0.25, min(chunk.RMS, maxInfluenceLevel)/maxInfluenceLevel)

	// Interpolate previous loudness value with current chunks loudness (RMS)
	return utils.Lerp(previousLoudness, chunk.RMS, factor)
}
