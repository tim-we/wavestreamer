package main

import (
	"github.com/tim-we/wavestreamer/library"
	"github.com/tim-we/wavestreamer/player"
)

func main() {
	library.ScanRootDir("../../tmp/pi-music-backup/wc-music")

	player.QueueClip(player.NewFakeTelephoneClip())
	player.QueueClip(library.PickRandomClip().CreateClip())
	player.QueueClip(library.PickRandomSong().CreateClip())
	player.QueueClip(library.PickRandomSong().CreateClip())

	player.Start()
}
