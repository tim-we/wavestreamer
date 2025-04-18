package main

import (
	"flag"
	"fmt"

	"github.com/tim-we/wavestreamer/library"
	"github.com/tim-we/wavestreamer/player"
	"github.com/tim-we/wavestreamer/player/clips"
	"github.com/tim-we/wavestreamer/scheduler"
)

func main() {
	musicDir := flag.String("d", "./music", "Path to directory containing music files")
	flag.Parse()

	fmt.Println("Using music directory:", *musicDir)
	library.ScanRootDir(*musicDir) // TODO check for existence

	player.QueueClip(clips.NewFakeTelephoneClip())
	player.QueueClip(library.PickRandomClip().CreateClip())

	scheduler.Start()
	player.Start(scheduler.GetNextClip)
}
