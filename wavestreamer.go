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
	news := flag.Bool("news", false, "Enable hourly news (Tagesschau in 100s)")
	flag.Parse()

	fmt.Println("Using music directory:", *musicDir)
	library.ScanRootDir(*musicDir)

	// Give PortAudio/ALSA/The audio system some time to start.
	// Otherwise we get stutters in the beginning.
	player.QueueClip(clips.NewPause(1 * time.Second))
	player.QueueClip(clips.NewFakeTelephoneClip())
	player.QueueClip(library.PickRandomClip().CreateClip())

	fmt.Println("Starting scheduler...")
	scheduler.Start()

	if *news {
		fmt.Println("Starting Tagesschau loop...")
		scheduler.StartTagesschau()
	}

	fmt.Println("Start playback loop...")
	player.Start(scheduler.GetNextClip)

	fmt.Println("Player stopped.")
}
