package decoder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type AudioFileMetaData struct {
	Duration time.Duration

	// Meta tags are optional. The fallback value is "".
	Title  string
	Artist string
	Album  string
}

// GetFileMetadata fetches the duration of an audio file in seconds using ffprobe and, if available, the tracks title, artist and album.
func GetFileMetadata(filePath string) (*AudioFileMetaData, error) {
	// Run ffprobe with JSON output
	cmd := exec.Command(
		"ffprobe",
		"-v", "quiet", // Set logging level to prevent ffprobe printing non JSON data
		"-of", "json", // Output format (-output_format is not available in older versions of ffprobe like 5.1.6)
		"-show_entries", "format_tags:stream_tags", // File meta data can be either in the container (format_tags) or stream (stream_tags)
		"-show_format", // File format information, includes duration and size
		filePath,
	)
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to run ffprobe: %w", err)
	}

	// Parse the JSON output
	var probeResult FFProbeOutput
	if err := json.Unmarshal(out.Bytes(), &probeResult); err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	// Convert duration to a float64 (representing seconds)
	duration, err := strconv.ParseFloat(probeResult.Format.Duration, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid duration format: %w", err)
	}

	// Collect metadata
	metadata := make(map[string]string)

	// Add format tags if available
	if probeResult.Format.Tags != nil {
		for key, value := range probeResult.Format.Tags {
			metadata[strings.ToLower(key)] = value
		}
	}

	// Fallback to stream tags if format tags are missing
	if len(metadata) == 0 && len(probeResult.Streams) > 0 {
		for _, stream := range probeResult.Streams {
			if stream.Tags != nil {
				for key, value := range stream.Tags {
					metadata[strings.ToLower(key)] = value
				}
			}
		}
	}

	return &AudioFileMetaData{
		Duration: time.Duration(duration * float64(time.Second)),
		Title:    metadata["title"],
		Artist:   metadata["artist"],
		Album:    metadata["album"],
	}, nil
}

type FFProbeOutput struct {
	Format struct {
		Duration string            `json:"duration"`
		Tags     map[string]string `json:"tags"`
	} `json:"format"`
	Streams []struct {
		Tags map[string]string `json:"tags"`
	} `json:"streams"`
}
