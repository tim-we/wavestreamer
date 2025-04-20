package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/tim-we/wavestreamer/library"
	"github.com/tim-we/wavestreamer/player"
	"github.com/tim-we/wavestreamer/player/clips"
	"github.com/tim-we/wavestreamer/scheduler"
	"github.com/tim-we/wavestreamer/webapp"
)

type Options struct {
	MusicDir   string `short:"d" long:"music-dir" description:"Path to directory containing music files"`
	News       bool   `short:"n" long:"news" description:"Enable hourly news (Tagesschau in 100s)"`
	WebApp     bool   `short:"w" long:"webapp" description:"Enable web app" `
	WebAppPort int    `short:"p" long:"port" description:"Web App Port" default:"6969"`
}

func main() {
	var opts Options
	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}

	fmt.Println("Using music directory:", opts.MusicDir)
	library.ScanRootDir(opts.MusicDir)

	// Give PortAudio/ALSA/The audio system some time to start.
	// Otherwise we get stutters in the beginning.
	player.QueueClip(clips.NewPause(1 * time.Second))
	player.QueueClip(clips.NewFakeTelephoneClip())
	player.QueueClip(library.PickRandomClip().CreateClip())

	fmt.Println("Starting scheduler...")
	scheduler.Start()

	if opts.News {
		fmt.Println("Starting Tagesschau loop...")
		scheduler.StartTagesschauScheduler()
	}

	if opts.WebApp {
		fmt.Println("Starting web server...")
		if opts.WebAppPort < 1024 {
			log.Println("Warning: Ports below 1024 require root access.")
		}
		webapp.StartServer(opts.WebAppPort)
	}

	fmt.Println("Starting playback loop...")
	player.Start(scheduler.GetNextClip)

	fmt.Println("Player stopped.")
}
