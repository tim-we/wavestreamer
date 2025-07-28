package clips

import (
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tim-we/wavestreamer/config"
	"github.com/tim-we/wavestreamer/player"
	d "github.com/tim-we/wavestreamer/player/decoder"
)

type AudioClip struct {
	filepath string
	decoder  *d.DecodingProcess
	meta     *d.AudioFileMetaData
	buffer   chan *player.AudioChunk
	started  bool
	stopped  bool
	OnStart  func(meta *d.AudioFileMetaData)
	OnStop   func()
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

	if err := decoder.StartDecoding(); err != nil {
		return nil, fmt.Errorf("failed to start the decoding process of '%s'", filepath)
	}

	buffer := make(chan *player.AudioChunk, 16)

	clip := AudioClip{
		filepath: filepath,
		decoder:  &decoder,
		meta:     meta,
		buffer:   buffer,
		started:  false,
	}

	go func() {
		defer close(buffer)

		for {
			// Create empty chunk.
			chunk := player.AudioChunk{
				Left:  make([]float32, config.FRAMES_PER_BUFFER),
				Right: make([]float32, config.FRAMES_PER_BUFFER),
			}

			eofReached := false
			var peak float32 = 0.0
			var rmsAcc float64 = 0.0

			// Fill chunk and analyze data.
			for i := range config.FRAMES_PER_BUFFER {
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

				peak = max(peak, max(absf32(left), absf32(right)))
				rmsAcc += float64(left*left + right*right)

				chunk.Left[i] = left
				chunk.Right[i] = right
				chunk.Length++
			}

			chunk.Peak = peak
			chunk.RMS = float32(math.Sqrt(rmsAcc / float64(config.CHANNELS*config.FRAMES_PER_BUFFER)))

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
	if !clip.started {
		clip.started = true
		if clip.OnStart != nil {
			clip.OnStart(clip.meta)
		}
	}
	chunk, hasMore := <-clip.buffer
	if !hasMore && !clip.stopped && clip.OnStop != nil {
		clip.stopped = true
		clip.OnStop()

	}
	return chunk, hasMore
}

func (clip *AudioClip) Stop() {
	clip.decoder.Close()
	if !clip.stopped && clip.OnStop != nil {
		clip.OnStop()
	}
	clip.stopped = true
}

func (clip *AudioClip) Name() string {
	if clip == nil {
		panic("clip is nil")
	}
	if clip.meta != nil {
		if clip.meta.Artist != "" && clip.meta.Title != "" {
			return fmt.Sprintf("%s - %s", clip.meta.Artist, clip.meta.Title)
		}
	}

	filename := removeAudioExtension(filepath.Base(clip.filepath))

	return filename
}

func (clip *AudioClip) Duration() time.Duration {
	return clip.meta.Duration
}

func (clip *AudioClip) SetMetaData(title, artist, album string) {
	if title != "" {
		clip.meta.Title = title
	}
	if artist != "" {
		clip.meta.Artist = artist
	}
	if album != "" {
		clip.meta.Album = album
	}
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !errors.Is(err, os.ErrNotExist)
}

func absf32(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}

func removeAudioExtension(filename string) string {
	ext := filepath.Ext(filename)
	switch strings.ToLower(ext) {
	case ".mp3", ".ogg", ".flac", ".wav", ".aac", ".m4a", ".opus":
		// Remove known audio extensions.
		return strings.TrimSuffix(filename, ext)
	default:
		// Unknown extension, return unchanged.
		return filename
	}
}
