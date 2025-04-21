package scheduler

import (
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/tim-we/wavestreamer/player"
	"github.com/tim-we/wavestreamer/player/clips"
	"github.com/tim-we/wavestreamer/utils"
)

const PODCAST_FEED = "https://www.tagesschau.de/multimedia/sendung/tagesschau_in_100_sekunden/podcast-ts100-audio-100~podcast.xml"

type RSSRoot struct {
	Channel RSSChannel `xml:"channel"`
}

type RSSChannel struct {
	Items []RSSItem `xml:"item"`
}

type RSSItem struct {
	Title     string        `xml:"title"`
	PubDate   string        `xml:"pubDate"` // still here as a fallback if needed
	DCDate    string        `xml:"http://purl.org/dc/elements/1.1/ date"`
	Enclosure *RSSEnclosure `xml:"enclosure"`
}

type RSSEnclosure struct {
	URL    string `xml:"url,attr"`
	Type   string `xml:"type,attr"`
	Length int64  `xml:"length,attr"`
}

type EpisodeInfo struct {
	URL     string
	PubDate time.Time
}

var userSignal = make(chan struct{})

func extractLatestMP3URL(xmlContent []byte) (*EpisodeInfo, error) {
	var rss RSSRoot
	if err := xml.Unmarshal(xmlContent, &rss); err != nil {
		return nil, err
	}

	var latestURL string
	var latestTime time.Time

	// Iterate over all items and remember the latest (valid) one
	for _, item := range rss.Channel.Items {
		if item.Enclosure == nil || item.Enclosure.Type != "audio/mpeg" {
			continue
		}

		var pubTime time.Time
		var parseError error

		switch {
		case item.DCDate != "":
			// Example: 2025-04-20T11:06:00Z
			pubTime, parseError = time.Parse(time.RFC3339, strings.TrimSpace(item.DCDate))
		case item.PubDate != "":
			// Example: Sun, 20 Apr 2025 13:06:22 +0200
			pubTime, parseError = time.Parse(time.RFC1123Z, strings.TrimSpace(item.PubDate))
		default:
			continue // no recognizable date -> skip
		}

		if parseError != nil {
			continue
		}

		if pubTime.After(latestTime) {
			latestTime = pubTime
			latestURL = item.Enclosure.URL
		}
	}

	if latestURL == "" {
		return nil, fmt.Errorf("no valid MP3 entries found")
	}

	return &EpisodeInfo{latestURL, latestTime}, nil
}

func timeUntilNextShow(last time.Time) time.Duration {
	now := time.Now()
	nextShow := now.Truncate(time.Hour).Add(time.Hour)
	timeBetweenShows := nextShow.Sub(last)
	if timeBetweenShows < (30 * time.Minute) {
		nextShow = nextShow.Add(15 * time.Minute)
	}

	return nextShow.Sub(now)
}

func StartTagesschauScheduler() {
	go func() {
		for {
			delay := timeUntilNextShow(time.Now())

			// Wait for next opportunity to play
			select {
			case <-userSignal:
				log.Println("Tagesschau scheduled by user.")
			case <-time.After(delay):
				log.Println("Tagesschau automatically scheduled.")
			}

			// Fetch latest episode
			rssDataRaw, rssDownloadError := utils.DownloadToMemory(PODCAST_FEED)
			if rssDownloadError != nil {
				log.Printf("Error downloading Tagesschau RSS:\n%v\n", rssDownloadError)
				// Lets try again later
				continue
			}

			episode, decodeError := extractLatestMP3URL(rssDataRaw)
			if decodeError != nil {
				log.Printf("Error decoding Tagesschau RSS:\n%v\n", decodeError)
				// Lets try again later
				continue
			}

			if time.Now().Sub(episode.PubDate) > (24 * time.Hour) {
				log.Printf("No recent Tagesschau episode available (%v)\n", episode.PubDate)
				// Lets try again later
				continue
			}

			tmpFile, downloadErr := utils.DownloadToTempFile(episode.URL)
			if downloadErr != nil {
				log.Printf("Error downloading Tagesschau episode:\n%v\n", downloadErr)
				// Lets try again later
				continue
			}
			tmpFile.Close()

			// Create clip with custom meta data
			clip, err := clips.NewAudioClip(tmpFile.Name())
			if err != nil {
				log.Printf("Failed to create Tagesschau clip:\n%v\n", err)
			}
			clip.SetMetaData(episode.PubDate.Format("02.01.06 - 15:04"), "Tagesschau in 100s", "")

			// Cleanup
			clip.OnStop = func() {
				if err := os.Remove(tmpFile.Name()); err != nil {
					log.Printf("Failed to remove temporary file %s.\n", err)
				}
			}

			// And finally... schedule the clip
			player.QueueClip(clip)
		}
	}()
}

func ScheduleTagesschauNow() {
	go func() {
		userSignal <- struct{}{}
	}()
}
