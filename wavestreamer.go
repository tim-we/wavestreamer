package main

import (
	"github.com/tim-we/wavestreamer/library"
	"github.com/tim-we/wavestreamer/player"
)

func main() {
	library.ScanRootDir("../../tmp/pi-music-backup/wc-music")

	player.QueuePauseNext()
	player.QueueAudio("test-audio.ogg")
	player.QueueAudio("test-audio.ogg")
	player.QueueAudio("test-audio.ogg")

	player.Start()
}
