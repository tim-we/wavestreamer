package player

/*
#cgo linux,arm64 LDFLAGS: -lportaudio
*/
import "C"

import (
	"log"

	"github.com/gordonklaus/portaudio"
	"github.com/tim-we/wavestreamer/utils"
)

var userQueue = make([]Clip, 0, 12)

var priorityQueue = make(chan Clip, 2)

var mainLoop *PlaybackLoop

func Start(clipProvider func() Clip, normalize bool) {
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
	priorityLoop := NewPlaybackLoop("Priority Loop", false, func() Clip { return <-priorityQueue })
	mainLoop = NewPlaybackLoop("Main Loop", normalize, nextClipProvider)

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

	stream := initPortAudioStream(playCallback)

	defer portaudio.Terminate()
	defer stream.Close()

	if err := stream.Start(); err != nil {
		log.Fatal(err)
	}
	defer stream.Stop()

	mainLoop.clipEndCallback = func(clip Clip, skipped bool) {
		addClipToHistory(clip, skipped)
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
	return mainLoop.GetCurrentClip()
}

func SkipCurrent() {
	if mainLoop != nil {
		mainLoop.Skip()
	}
}

func PlayPriorityClip(clip Clip) {
	priorityQueue <- clip
}
