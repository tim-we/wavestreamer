package main

import (
	"github.com/tim-we/wavestreamer/library"
	"github.com/tim-we/wavestreamer/player"
	"github.com/tim-we/wavestreamer/player/clips"
)

func main() {
	library.ScanRootDir("../../tmp/pi-music-backup/wc-music")

	player.QueueClip(clips.NewFakeTelephoneClip())
	player.QueueClip(library.PickRandomClip().CreateClip())
	player.QueueClip(library.PickRandomSong().CreateClip())
	player.QueueClip(library.PickRandomSong().CreateClip())

	player.Start()
}
