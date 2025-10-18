package player

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/tim-we/wavestreamer/player/decoder"
)

// GetDisplayName returns a name of the form [Artist] - [Song] when possible.
func GetDisplayName(path string, meta *decoder.AudioFileMetaData) string {
	if meta != nil {
		if meta.Artist != "" && meta.Title != "" {
			return fmt.Sprintf("%s - %s", meta.Artist, meta.Title)
		}
	}

	filename := removeAudioExtension(filepath.Base(path))

	return filename
}

func removeAudioExtension(filename string) string {
	ext := filepath.Ext(filename)
	switch strings.ToLower(ext) {
	case ".mp3", ".ogg", ".flac", ".wav", ".aac", ".m4a", ".opus":
		// Remove known audio extensions.
		return strings.TrimSuffix(filename, ext)
	default:
		// Unknown extension, return unchanged.
		return filename
	}
}
