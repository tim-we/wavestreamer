package main

import (
	"github.com/tim-we/wavestreamer/player"
)

func main() {
	player.QueueAudio("test-audio.ogg")
	player.QueueAudio("test-audio.ogg")
	player.QueueAudio("test-audio.ogg")

	player.Start()
}
