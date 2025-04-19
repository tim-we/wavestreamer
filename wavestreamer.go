package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/tim-we/wavestreamer/library"
	"github.com/tim-we/wavestreamer/player"
	"github.com/tim-we/wavestreamer/player/clips"
	"github.com/tim-we/wavestreamer/scheduler"
)

func main() {
	musicDir := flag.String("d", "./music", "Path to directory containing music files")
	flag.Parse()

	fmt.Println("Using music directory:", *musicDir)
	library.ScanRootDir(*musicDir)

	player.QueueClip(clips.NewPause(2 * time.Second))
	player.QueueClip(clips.NewFakeTelephoneClip())
	player.QueueClip(library.PickRandomClip().CreateClip())

	fmt.Println("Start playback loop...")
	player.Start(scheduler.GetNextClip)
	// Give PortAudio/ALSA/The audio system some time to start.
	// Otherwise we get stutters in the beginning.
	time.Sleep(1 * time.Second)

	fmt.Println("Starting scheduler...")
	scheduler.Start()
}
