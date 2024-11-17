package player2

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strconv"
	"time"

	"github.com/gordonklaus/portaudio"
)

const (
	sampleRate      = 44100
	channels        = 2
	framesPerBuffer = 1024
)

func Start() {
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

			metaData, _ := GetFileMetadata(file)
			log.Printf("Reading file %s. Duration: %vs", file, metaData.Duration)
			if metaData.Title != nil {
				log.Printf("File meta data: %s", *metaData.Title)
			}

			lastAudioChunk = readEntireFile(file, lastAudioChunk, audioChunks)

			if lastAudioChunk.Length >= framesPerBuffer {
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
	stream, err := portaudio.OpenDefaultStream(0, channels, sampleRate, framesPerBuffer, playCallback)
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

// readFrame reads two 16-bit samples from the PCM stream
func readFrame(reader *bufio.Reader) (error, float32, float32) {
	var left, right int16
	if err := binary.Read(reader, binary.LittleEndian, &left); err != nil {
		return err, 0, 0
	}
	if err := binary.Read(reader, binary.LittleEndian, &right); err != nil {
		return err, 0, 0
	}
	return nil, float32(left) / 32768.0, float32(right) / 32768.0
}

func readEntireFile(file string, firstChunk *AudioChunk, chunkChan chan AudioChunk) *AudioChunk {
	cmd := exec.Command("ffmpeg", "-i", file, "-f", "s16le", "-ac", strconv.Itoa(channels), "-ar", strconv.Itoa(sampleRate), "pipe:1")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	defer stdout.Close()

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	reader := bufio.NewReader(stdout)

	nextChunk := firstChunk
	var lastChunk *AudioChunk

	for {
		var chunk AudioChunk

		if nextChunk == nil {
			chunk = AudioChunk{
				Left:  make([]float32, framesPerBuffer),
				Right: make([]float32, framesPerBuffer),
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
		for i := chunk.Length; i < framesPerBuffer; i++ {
			err, left, right = readFrame(reader)

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
			if waitErr := cmd.Wait(); waitErr != nil {
				log.Fatal(waitErr)
			}
			break
		}

		if err != nil {
			log.Fatal(err)
		}
	}

	return lastChunk
}

// GetFileMetadata fetches the duration of an audio file in seconds using ffprobe and, if available, the tracks title, artist and album.
func GetFileMetadata(filePath string) (*AudioFileMetaData, error) {
	// Run ffprobe with JSON output
	cmd := exec.Command(
		"ffprobe",
		"-v", "quiet", // Set logging level to avoid ffprobe printing non JSON data
		"-output_format", "json", // Output format
		"-show_entries", "format_tags=artist,title,album", // File meta data (ID3)
		"-show_format", // File format information, includes duration and size
		filePath,
	)
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to run ffprobe: %w", err)
	}

	// Parse the JSON output
	var probeResult FFProbeOutput
	if err := json.Unmarshal(out.Bytes(), &probeResult); err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	// Convert duration to a float64
	duration, err := strconv.ParseFloat(probeResult.Format.Duration, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid duration format: %w", err)
	}

	var title, artist, album *string

	if probeResult.Format.Tags != nil {
		title = probeResult.Format.Tags.Title
		artist = probeResult.Format.Tags.Artist
		album = probeResult.Format.Tags.Album
	}

	metaData := AudioFileMetaData{
		Duration: int(duration),
		Title:    title,
		Artist:   artist,
		Album:    album,
	}

	return &metaData, nil
}

type AudioChunk = struct {
	Left   []float32
	Right  []float32
	Length int
}

type FFProbeOutput struct {
	Format struct {
		Duration string `json:"duration"`
		Tags     *struct {
			Artist *string `json:"artist"`
			Title  *string `json:"title"`
			Album  *string `json:"album"`
		} `json:"tags"`
	} `json:"format"`
}

type AudioFileMetaData struct {
	// Duration in seconds.
	Duration int

	// Meta tags are optional.
	Title  *string
	Artist *string
	Album  *string
}
