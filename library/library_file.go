package library

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/tim-we/wavestreamer/player"
)

type LibraryFile struct {
	filepath   string
	meta       *player.AudioFileMetaData
	playCount  int32
	skipCount  int32
	lastPlayed *time.Time
}

func NewLibraryFile(filepath string) (*LibraryFile, error) {
	if !fileExists(filepath) {
		return nil, fmt.Errorf("File '%s' not found.", filepath)
	}

	return &LibraryFile{
		filepath:   filepath,
		meta:       nil,
		playCount:  0,
		skipCount:  0,
		lastPlayed: nil,
	}, nil
}

func (file *LibraryFile) CreateClip() *player.AudioClip {
	file.playCount += 1
	now := time.Now()
	file.lastPlayed = &now
	clip, _ := player.NewAudioClip(file.filepath)
	// TODO: get meta data
	return clip
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !errors.Is(err, os.ErrNotExist)
}
