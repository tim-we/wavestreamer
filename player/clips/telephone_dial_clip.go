package clips

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/tim-we/wavestreamer/config"
	"github.com/tim-we/wavestreamer/player"
)

type TelephoneDialClip struct {
	buffer          chan *player.AudioChunk
	duration        int
	telephoneNumber string
}

// Dual-Tone Multi-Frequency
type DTMFFrequencies struct {
	Lower  int
	Higher int
}

var frequencyMap = map[rune]DTMFFrequencies{
	'1': {697, 1209},
	'2': {697, 1336},
	'3': {697, 1477},
	'4': {770, 1209},
	'5': {770, 1336},
	'6': {770, 1477},
	'7': {852, 1209},
	'8': {852, 1336},
	'9': {852, 1477},
	'*': {941, 1209},
	'0': {941, 1336},
	'#': {941, 1477},
}

var telephoneNumbers = []string{
	"555-2368",     // Ghostbusters
	"555-0113",     // Simpsons
	"555-0134",     // Fight Club
	"618-625-8313", // Stranger Things
	"212-555-0175", // Mr. Robot
	"555-0690",     // The Matrix
}

var dialFrequencies = DTMFFrequencies{350, 440}

// var busyFrequencies = frequencyPair{480, 620}

const VOLUME = 0.25

// Roughly a third of a second
const BEEP_DURATION_IN_CHUNKS = max(1, (config.SAMPLE_RATE/3)/config.FRAMES_PER_BUFFER-1)

// Roughly 1.5 seconds
const DIAL_DURATION_IN_CHUNKS = max(1, (3*config.SAMPLE_RATE/2)/config.FRAMES_PER_BUFFER)

// Roughly half a second. Pause between beeps and dial sound.
const PAUSE_DURATION_IN_CHUNKS = max(1, (config.SAMPLE_RATE/2)/config.FRAMES_PER_BUFFER)

func NewFakeTelephoneClip() *TelephoneDialClip {
	// Pick a random telephone number
	telNumber := telephoneNumbers[rand.Intn(len(telephoneNumbers))]

	buffer := make(chan *player.AudioChunk, 8)

	durationInChunks := len(telNumber)*BEEP_DURATION_IN_CHUNKS + PAUSE_DURATION_IN_CHUNKS + DIAL_DURATION_IN_CHUNKS
	durationInSeconds := (durationInChunks * config.FRAMES_PER_BUFFER) / config.SAMPLE_RATE

	go func() {
		defer close(buffer)

		for _, ch := range telNumber {
			if ch == ' ' || ch == '-' || ch == '/' {
				for range 3 {
					chunk := silence()
					buffer <- &chunk
				}
			}

			frequencies, found := frequencyMap[ch]
			if !found {
				continue
			}

			for i := range BEEP_DURATION_IN_CHUNKS {
				chunk := silence()
				fillChunkWithFrequencies(chunk, frequencies, i*config.FRAMES_PER_BUFFER, i == BEEP_DURATION_IN_CHUNKS-1)
				buffer <- &chunk
			}
		}

		for range PAUSE_DURATION_IN_CHUNKS {
			chunk := silence()
			buffer <- &chunk
		}

		for i := range DIAL_DURATION_IN_CHUNKS {
			chunk := silence()

			fillChunkWithFrequencies(chunk, dialFrequencies, i*config.FRAMES_PER_BUFFER, i == DIAL_DURATION_IN_CHUNKS-1)
			buffer <- &chunk
		}
	}()

	return &TelephoneDialClip{
		buffer,
		durationInSeconds,
		telNumber,
	}
}

func (clip *TelephoneDialClip) NextChunk() (*player.AudioChunk, bool) {
	chunk, hasMore := <-clip.buffer
	return chunk, hasMore
}

func (clip *TelephoneDialClip) Stop() {
}

func (clip *TelephoneDialClip) Name() string {
	return fmt.Sprintf("Telephone Dial (%s)", clip.telephoneNumber)
}

func (clip *TelephoneDialClip) Duration() int {
	return clip.duration
}

func fillChunkWithFrequencies(chunk player.AudioChunk, pair DTMFFrequencies, timeOffset int, fadeOut bool) {
	freqA := float64(pair.Lower)
	freqB := float64(pair.Higher)
	for i := range config.FRAMES_PER_BUFFER {
		t := 2.0 * math.Pi * float64(timeOffset+i) / config.SAMPLE_RATE
		value := float32(VOLUME * (math.Sin(t*freqA) + math.Sin(t*freqB)))
		if fadeOut {
			value = value * (float32(config.FRAMES_PER_BUFFER-i) / config.FRAMES_PER_BUFFER)
		}
		chunk.Left[i] = value
		chunk.Right[i] = value
	}
}

func silence() player.AudioChunk {
	return player.AudioChunk{
		Left:  make([]float32, config.FRAMES_PER_BUFFER),
		Right: make([]float32, config.FRAMES_PER_BUFFER),
	}
}
