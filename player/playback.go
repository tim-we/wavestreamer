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

var priorityQueue = make(chan Clip, 1)

var currentlyPlaying Clip = nil

var skipSignal = make(chan struct{}, 1)

func Start(clipProvider func() Clip, normalize bool) {
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

	nextClipProvider := func() Clip {
		if len(userQueue) > 0 {
			clip := userQueue[0]
			userQueue = userQueue[1:]
			return clip
		}
		if clip := clipProvider(); clip != nil {
			return clip
		}
		return nil
	}

	// Default & priority playback loops
	priorityLoop := NewPlaybackLoop("Priority Loop", normalize, func() Clip { return <-priorityQueue })
	mainLoop := NewPlaybackLoop("Main Loop", normalize, nextClipProvider)

	playCallback := func(out [][]float32) {
		select {
		case chunk := <-priorityLoop.NextAudioChunk:
			copy(out[0], chunk.Left)
			copy(out[1], chunk.Right)
			// Priority chunks should replace normal ones.
			// Otherwise you would hear the remaining chunks after a pause beep.
			utils.DropOne(mainLoop.NextAudioChunk)
		case chunk := <-mainLoop.NextAudioChunk:
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

	mainLoop.clipEndCallback = func(clip Clip, skipped bool) {
		addClipToHistory(clip, skipped)
		currentlyPlaying = nil
	}

	go priorityLoop.Run()
	mainLoop.Run()
}

func QueueClip(clip Clip) {
	if clip == nil {
		return
	}
	userQueue = append(userQueue, clip)
}

func QueueClipNext(clip Clip) {
	if clip == nil {
		return
	}
	userQueue = append([]Clip{clip}, userQueue...)
}

func QueueSize() int {
	return len(userQueue)
}

func GetCurrentlyPlaying() Clip {
	return currentlyPlaying
}

func SkipCurrent() {
	skipSignal <- struct{}{}
}

func PlayPriorityClip(clip Clip) {
	priorityQueue <- clip
}
