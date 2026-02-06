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

var userQueue = utils.NewConcurrentQueue[Clip](12)

var priorityQueue = make(chan Clip, 2)

var mainLoop *PlaybackLoop

var beepClipProvider func() Clip

func Start(clipProvider func() Clip, normalize bool) {
	nextClipProvider := func() Clip {
		if !userQueue.IsEmpty() {
			clip, _ := userQueue.GetNext()
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
	mainLoop.ClipStartCallback = func(clip Clip) { log.Printf("Now playing %s", clip.Name()) }

	playCallback := func(out [][]float32) {
		// Check priority queue first:
		select {
		case chunk := <-priorityLoop.NextAudioChunk:
			copy(out[0], chunk.Left)
			copy(out[1], chunk.Right)
			// Priority chunks should replace normal ones.
			// Otherwise you would hear the remaining chunks after a pause beep.
			utils.DropOne(mainLoop.NextAudioChunk)
			return
		default:
			// No priority clips.
		}

		// There are no priority clips so we proceed with the main queue:
		select {
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
	userQueue.Add(clip)
}

func QueueClipNext(clip Clip) {
	if clip == nil {
		return
	}
	userQueue.Prepend(clip)
}

func QueueSize() int {
	return userQueue.Size()
}

func GetCurrentlyPlaying() Clip {
	return mainLoop.GetCurrentClip()
}

func SkipCurrent(silent bool) {
	if mainLoop == nil {
		// This should not happen...
		return
	}

	if !silent && beepClipProvider != nil {
		PlayPriorityClip(beepClipProvider())
	}

	mainLoop.Skip()
}

func PlayPriorityClip(clip Clip) {
	if clip == nil {
		return
	}
	priorityQueue <- clip
}

func SetBeepProvider(provider func() Clip) {
	beepClipProvider = provider
}
