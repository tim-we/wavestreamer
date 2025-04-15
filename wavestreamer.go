package main

import (
	"github.com/tim-we/wavestreamer/library"
	"github.com/tim-we/wavestreamer/player"
	"github.com/tim-we/wavestreamer/player/clips"
	"github.com/tim-we/wavestreamer/scheduler"
)

func main() {
	library.ScanRootDir("../../tmp/pi-music-backup/wc-music")

	player.QueueClip(clips.NewFakeTelephoneClip())
	player.QueueClip(library.PickRandomClip().CreateClip())

	scheduler.Start()
	player.Start(scheduler.GetNextClip)
}
