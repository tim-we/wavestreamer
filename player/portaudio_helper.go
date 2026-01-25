package player

import (
	"fmt"
	"log"

	"github.com/gordonklaus/portaudio"
	"github.com/tim-we/wavestreamer/config"
)

var portaudioInitialized = false

// initPortAudioStream creates and configures a low-latency PortAudio output stream.
// Initializes PortAudio on first call.
//
// playCallback is invoked repeatedly by PortAudio to fill output buffers with stereo audio samples.
// Panics on error.
func initPortAudioStream(playCallback func(out [][]float32)) *portaudio.Stream {
	if !portaudioInitialized {
		if init_err := portaudio.Initialize(); init_err != nil {
			log.Fatal(init_err)
		}
		portaudioInitialized = true
	}

	outputDevice, devErr := portaudio.DefaultOutputDevice()
	if devErr != nil {
		log.Fatal(devErr)
	}

	// Set up a low-latency PortAudio stream with a fixed buffer size
	streamParams := portaudio.StreamParameters{
		Output: portaudio.StreamDeviceParameters{
			Device:   outputDevice,
			Channels: config.CHANNELS,                      // output channels
			Latency:  outputDevice.DefaultLowOutputLatency, // use device's low latency setting
		},
		SampleRate:      float64(config.SAMPLE_RATE), // output sample rate
		FramesPerBuffer: config.FRAMES_PER_BUFFER,    // output buffer size
	}

	stream, streamErr := portaudio.OpenStream(streamParams, playCallback)
	if streamErr != nil {
		log.Fatal(streamErr)
	}

	info := stream.Info()
	fmt.Printf("Output latency: %d ms\n", info.OutputLatency.Milliseconds())

	return stream
}
