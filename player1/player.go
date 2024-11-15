package player1

import (
	"bufio"
	"encoding/binary"
	"io"
	"log"
	"os/exec"
	"strconv"
	"time"

	"github.com/gordonklaus/portaudio"
)

const sampleRate = 44100
const channels = 2
const framesPerBuffer = 1024 // Setting a specific buffer size

func main() {
	err := portaudio.Initialize()
	if err != nil {
		log.Fatal(err)
	}
	defer portaudio.Terminate()

	// FFmpeg command to output PCM data
	cmd := exec.Command("ffmpeg", "-i", "test-audio.ogg", "-f", "s16le", "-ac", strconv.Itoa(channels), "-ar", strconv.Itoa(sampleRate), "pipe:1")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	defer stdout.Close()

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	// Set up the PortAudio stream with a fixed buffer size
	stream, err := portaudio.OpenDefaultStream(0, channels, sampleRate, framesPerBuffer, playCallback(stdout))
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	if err := stream.Start(); err != nil {
		log.Fatal(err)
	}
	defer stream.Stop()

	if err := cmd.Wait(); err != nil {
		log.Print("Oh no, a fatal error")
		log.Fatal(err)
	}

	log.Print("Nice, no fatal error")

	time.Sleep(time.Second)
}

// playCallback reads PCM data and plays it through the audio stream
func playCallback(stdout io.Reader) func(out [][]float32) {
	reader := bufio.NewReader(stdout)
	return func(out [][]float32) {
		log.Printf("len(out[ch]) = %d", len(out[0]))
		for i := 0; i < framesPerBuffer; i++ {
			for ch := 0; ch < channels; ch++ {
				out[ch][i] = readSample(reader)
			}
		}
	}
}

// readSample reads a 16-bit sample from the PCM stream
func readSample(reader *bufio.Reader) float32 {
	var sample int16
	err := binary.Read(reader, binary.LittleEndian, &sample)
	if err != nil {
		if err == io.EOF {
			return 0
		}
		log.Printf("Argh, a fatal error")
		log.Fatal(err)
	}
	return float32(sample) / 32768.0
}
