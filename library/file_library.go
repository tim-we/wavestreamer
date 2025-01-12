package library

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
)

var hostClips = make([]*LibraryFile, 0, 128)
var songFiles = make([]*LibraryFile, 0, 512)
var clipFiles = make([]*LibraryFile, 0, 256)

func ScanRootDir(root string) {
	fmt.Printf("Searching for files in %s...\n", root)
	unknownFiles := 0

	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err1 error) error {
		if err1 != nil {
			return err1
		}

		if !entry.Type().IsRegular() {
			return nil
		}

		file, err2 := NewLibraryFile(path)

		if err2 != nil {
			return err2
		}

		if isSong(path) {
			songFiles = append(songFiles, file)
		} else if isClipFile(path) {
			clipFiles = append(clipFiles, file)
		} else if isHostClip(path) {
			hostClips = append(hostClips, file)
		} else {
			unknownFiles++
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error scanning the directory '%v' for files: %v\n", root, err)
	}

	fmt.Printf(
		"Scanning complete. Found:\n - %d songs\n - %d clips\n - %d host/dj clips\n",
		len(songFiles),
		len(clipFiles),
		len(hostClips),
	)

	if unknownFiles > 0 {
		fmt.Printf("%d files could not be classified.\n", unknownFiles)
	}
}

func PickRandomSong() *LibraryFile {
	if len(songFiles) == 0 {
		panic("No songs to select from.")
	}

	return songFiles[rand.Intn(len(songFiles))]
}

func PickRandomClip() *LibraryFile {
	if len(clipFiles) == 0 {
		panic("No clips to select from.")
	}

	return clipFiles[rand.Intn(len(clipFiles))]
}

func PickRandomHostClip() *LibraryFile {
	if len(hostClips) == 0 {
		panic("No host clips to select from.")
	}

	return hostClips[rand.Intn(len(hostClips))]
}

func isHostClip(file string) bool {
	matches, _ := doublestar.Match("**/hosts/**/*", file)
	return matches
}

func isClipFile(file string) bool {
	matches, _ := doublestar.Match("**/clips/**/*", file)
	return matches
}

func isSong(file string) bool {
	matches, _ := doublestar.Match("**/music/**/*", file)
	return matches
}
