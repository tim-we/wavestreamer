package clips

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/tim-we/wavestreamer/player"
	d "github.com/tim-we/wavestreamer/player/decoder"
)

type AudioClip struct {
	filepath string
	decoder  *d.DecodingProcess
	meta     *d.AudioFileMetaData
	buffer   chan *player.AudioChunk
}

func NewAudioClip(filepath string) (*AudioClip, error) {
	if !fileExists(filepath) {
		return nil, fmt.Errorf("file '%s' not found", filepath)
	}

	decoder := d.NewDecodingProcess(filepath)
	meta, metaErr := d.GetFileMetadata(filepath)

	if metaErr != nil {
		decoder.Close()
		return nil, fmt.Errorf("failed to get meta data of '%s'", filepath)
	}

	// TODO: consider checking for errors instead of panicing
	decoder.StartDecoding()

	buffer := make(chan *player.AudioChunk, 16)

	clip := AudioClip{
		filepath: filepath,
		decoder:  &decoder,
		meta:     meta,
		buffer:   buffer,
	}

	go func() {
		defer close(buffer)

		for {
			// Create empty chunk.
			chunk := player.AudioChunk{
				Left:  make([]float32, player.FRAMES_PER_BUFFER),
				Right: make([]float32, player.FRAMES_PER_BUFFER),
			}

			eofReached := false

			// Fill chunk.
			for i := range player.FRAMES_PER_BUFFER {
				left, right, err := decoder.ReadFrame()

				if err != nil {
					if err == io.EOF {
						eofReached = true
						// TODO: Do we need this?
						decoder.WaitForExit()
					} else {
						fmt.Printf("Unexpected decoding error:\n%v\n", err)
						return
					}
					break
				}

				chunk.Left[i] = left
				chunk.Right[i] = right
				chunk.Length++
			}

			// Send chunk to buffer.
			buffer <- &chunk

			if eofReached {
				break
			}
		}
	}()

	return &clip, nil
}

func (clip *AudioClip) NextChunk() (*player.AudioChunk, bool) {
	chunk, hasMore := <-clip.buffer
	return chunk, hasMore
}

func (clip *AudioClip) Stop() {
	clip.decoder.Close()
}

func (clip *AudioClip) Name() string {
	if clip.meta.Artist == "" || clip.meta.Title == "" {
		return filepath.Base(clip.filepath)
	}

	// TODO: Guess title and artist based on filename (if it includes -)

	return fmt.Sprintf("%s - %s", clip.meta.Artist, clip.meta.Title)
}

func (clip *AudioClip) Duration() int {
	return clip.meta.Duration
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !errors.Is(err, os.ErrNotExist)
}
