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

type AppOptions struct {
	MusicDir    string `short:"d" long:"music-dir" description:"Path to directory containing music files"`
	News        bool   `short:"n" long:"news" description:"Enable hourly news (Tagesschau in 100s)"`
	WebApp      bool   `short:"w" long:"webapp" description:"Enable web app" `
	WebAppPort  int    `short:"p" long:"port" description:"Web App Port" default:"6969"`
	GPIO        bool   `short:"i" long:"gpio" description:"Enable GPIO controls"`
	NoNormalize bool   `long:"no-normalize" description:"Disable automatic loudness normalization"`
	Version     bool   `short:"v" long:"version" description:"Display version & build information"`
}

// These will be replaced in the GitHub Actions workflow
var GitCommit string = "dev"
var BuildTime string = "unknown"

func main() {
	var opts AppOptions
	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}

	if opts.Version {
		fmt.Println("wavestreamer (github.com/tim-we/wavestreamer)")
		fmt.Printf("Commit: %s\n", GitCommit)
		fmt.Printf("Build time: %s\n", BuildTime)
		return
	}

	fmt.Println("Using music directory:", opts.MusicDir)
	library.WatchRootDir(opts.MusicDir)

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
		webapp.StartServer(opts.WebAppPort, opts.News)
	}

	if !opts.NoNormalize {
		fmt.Println("Loudness normalization is enabled (default).")
	}
	fmt.Println("Starting playback loop...")
	player.Start(scheduler.GetNextClip, !opts.NoNormalize)

	fmt.Println("Player stopped.")
}
