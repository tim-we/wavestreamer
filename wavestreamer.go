package main

import (
	"io"
	"log"
	"time"

	"github.com/gordonklaus/portaudio"
	"github.com/tim-we/wavestreamer/player"
)

func main() {
	err := portaudio.Initialize()
	if err != nil {
		log.Fatal(err)
	}
	defer portaudio.Terminate()

	files := make(chan string, 3)
	files <- "test-audio.ogg"
	files <- "test-audio.ogg"
	files <- "test-audio.ogg"

	audioChunks := make(chan AudioChunk, 16)

	go func() {
		var lastAudioChunk *AudioChunk

		for {
			file := <-files

			metaData, _ := player.GetFileMetadata(file)
			log.Printf("Reading file %s", file)
			log.Printf("File metadata:\n%+v\n", *metaData)

			lastAudioChunk = readEntireFile(file, lastAudioChunk, audioChunks)

			if lastAudioChunk.Length >= player.FRAMES_PER_BUFFER {
				lastAudioChunk = nil
			}
		}
	}()

	// playCallback sends audio chunks to the PortAudio stream.
	playCallback := func(out [][]float32) {
		select {
		case chunk := <-audioChunks:
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
		player.CHANNELS,          // output channels
		player.SAMPLE_RATE,       // output sample rate
		player.FRAMES_PER_BUFFER, // output buffer size
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

	time.Sleep(25 * time.Second)
}

func readEntireFile(file string, firstChunk *AudioChunk, chunkChan chan AudioChunk) *AudioChunk {
	process := player.NewDecodingProcess(file)
	process.StartDecoding()

	nextChunk := firstChunk
	var lastChunk *AudioChunk

	for {
		var chunk AudioChunk

		if nextChunk == nil {
			chunk = AudioChunk{
				Left:  make([]float32, player.FRAMES_PER_BUFFER),
				Right: make([]float32, player.FRAMES_PER_BUFFER),
			}
		} else {
			chunk = *nextChunk
			nextChunk = nil
		}

		lastChunk = &chunk

		var err error
		var left float32
		var right float32

		// By starting at chunk.Length we can fill up the remaining space in the chunk.
		for i := chunk.Length; i < player.FRAMES_PER_BUFFER; i++ {
			err, left, right = process.ReadFrame()

			if err != nil {
				break
			}

			chunk.Left[i] = left
			chunk.Right[i] = right
			chunk.Length++
		}

		chunkChan <- chunk

		if err == io.EOF {
			// Ensure the process is properly waited on before returning
			process.WaitForExit()
			break
		}

		if err != nil {
			log.Fatal(err)
		}
	}

	return lastChunk
}

type AudioChunk = struct {
	Left   []float32
	Right  []float32
	Length int
}
