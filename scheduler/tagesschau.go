package scheduler

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
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

func fetchRSS() ([]byte, error) {
	resp, err := http.Get(PODCAST_FEED)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}

func extractLatestMP3URL(xmlContent []byte) (*EpisodeInfo, error) {
	var rss RSSRoot
	if err := xml.Unmarshal(xmlContent, &rss); err != nil {
		return nil, err
	}

	var latestURL string
	var latestTime time.Time

	for _, item := range rss.Channel.Items {
		if item.Enclosure == nil || item.Enclosure.Type != "audio/mpeg" {
			continue
		}

		var pubTime time.Time
		var err error

		switch {
		case item.DCDate != "":
			pubTime, err = time.Parse(time.RFC3339, strings.TrimSpace(item.DCDate))
		case item.PubDate != "":
			pubTime, err = time.Parse(time.RFC1123Z, strings.TrimSpace(item.PubDate))
		default:
			continue // no recognizable date
		}

		if err != nil {
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
