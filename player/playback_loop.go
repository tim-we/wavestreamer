package player

/*
#cgo linux,arm64 LDFLAGS: -lportaudio
*/
import "C"

import (
	"log"

	"github.com/gordonklaus/portaudio"
	"github.com/tim-we/wavestreamer/config"
)

var userQueue = make([]Clip, 0, 12)

var currentlyPlaying string = "?"

var skipSignal = make(chan struct{}, 1)

func Start(clipProvider func() Clip) {
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

	nextAudioChunk := make(chan *AudioChunk, 1)

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

		for {
			if shouldSkipCurrentClip() {
				clip.Stop()
				break
			}

			chunk, hasMore := clip.NextChunk()

			if chunk != nil {
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
