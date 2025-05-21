package library

import (
	"errors"
	"fmt"
	"log"
	"os"
	fp "path/filepath"
	"strings"
	"time"

	"github.com/tim-we/wavestreamer/player/clips"
	"github.com/tim-we/wavestreamer/player/decoder"

	"github.com/google/uuid"
)

type LibraryFile struct {
	Id         uuid.UUID
	filepath   string
	searchData string
	meta       *decoder.AudioFileMetaData
	playCount  int32
	skipCount  int32
	lastPlayed *time.Time
}

func NewLibraryFile(filepath string) (*LibraryFile, error) {
	if !fileExists(filepath) {
		return nil, fmt.Errorf("file '%s' not found", filepath)
	}

	return &LibraryFile{
		Id:         uuid.New(),
		filepath:   filepath,
		searchData: createSearchData(filepath, nil),
		meta:       nil,
		playCount:  0,
		skipCount:  0,
		lastPlayed: nil,
	}, nil
}

func (file *LibraryFile) CreateClip() *clips.AudioClip {
	clip, err := clips.NewAudioClip(file.filepath)
	if err != nil {
		log.Println(err)
		return nil
	}
	clip.OnStart = func(meta *decoder.AudioFileMetaData) {
		now := time.Now()
		file.lastPlayed = &now
		file.playCount++
		file.meta = meta
		file.searchData = createSearchData(file.filepath, meta)
	}
	return clip
}

func (file *LibraryFile) Name() string {
	// TODO: Share implementation with AudioClip
	return fp.Base(file.filepath)
}

func (file *LibraryFile) Matches(query string) bool {
	return strings.Contains(file.searchData, query)
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !errors.Is(err, os.ErrNotExist)
}

func createSearchData(filepath string, meta *decoder.AudioFileMetaData) string {
	searchData := strings.ToLower(fp.Base(filepath))

	if meta == nil {
		return searchData
	}

	searchData += strings.ToLower(fmt.Sprintf(" %s %s %s", meta.Artist, meta.Title, meta.Album))

	return searchData
}
