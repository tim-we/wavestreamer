package player

import (
	"log"

	"github.com/gordonklaus/portaudio"
)

var queue = make([]Clip, 0, 12)

func Start() {
	err := portaudio.Initialize()
	if err != nil {
		log.Fatal(err)
	}
	defer portaudio.Terminate()

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
		0,                 // not reading any inputs (microphones)
		CHANNELS,          // output channels
		SAMPLE_RATE,       // output sample rate
		FRAMES_PER_BUFFER, // output buffer size
		playCallback,      // output buffer filling callback
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

		for {
			chunk, hasMore := clip.NextChunk()

			if chunk != nil {
				nextAudioChunk <- chunk
			}

			if !hasMore {
				break
			}
		}

	}
}

func QueueClip(clip Clip) {
	queue = append(queue, clip)
}

func QueueClipNext(clip Clip) {
	queue = append([]Clip{clip}, queue...)
}

func QueueSize() int {
	return len(queue)
}

func nextClip() Clip {
	if len(queue) == 0 {
		return nil
	}

	clip := queue[0]
	queue = queue[1:]
	return clip
}
