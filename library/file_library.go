package library

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

var hostClips = NewLibrarySet(128)
var songFiles = NewLibrarySet(512)
var clipFiles = NewLibrarySet(256)

func ScanRootDir(root string) {
	if !folderExists(root) {
		log.Fatalf("Folder '%s' does not exist.", root)
	}

	fmt.Printf("Searching for files in %s...\n", root)
	unknownFiles := 0

	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err1 error) error {
		if err1 != nil {
			return err1
		}

		if !entry.Type().IsRegular() {
			return nil
		}

		var err2 error

		if isSong(path) {
			err2 = songFiles.AddOrUpdate(path)
		} else if isClipFile(path) {
			err2 = clipFiles.AddOrUpdate(path)
		} else if isHostClip(path) {
			err2 = hostClips.AddOrUpdate(path)
		} else {
			unknownFiles++
		}

		if err2 != nil {
			return err2
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error scanning the directory '%v' for files: %v\n", root, err)
	}

	fmt.Printf(
		"Scanning complete. Found:\n - %d songs\n - %d clips\n - %d host/dj clips\n",
		songFiles.Size(),
		clipFiles.Size(),
		hostClips.Size(),
	)

	if unknownFiles > 0 {
		fmt.Printf("%d files could not be classified.\n", unknownFiles)
	}
}

func PickRandomSong() *LibraryFile {
	return pickRandomClipWhichHasNotBeenPlayedInAWhile(songFiles)
}

func PickRandomClip() *LibraryFile {
	return pickRandomClipWhichHasNotBeenPlayedInAWhile(clipFiles)
}

func PickRandomHostClip() *LibraryFile {
	return pickRandomClipWhichHasNotBeenPlayedInAWhile(hostClips)
}

func Search(query string) []*LibraryFile {
	modifiedQuery := strings.Trim(strings.ToLower(query), " ")

	if len(modifiedQuery) == 0 {
		return []*LibraryFile{}
	}

	parts := strings.Split(modifiedQuery, " ")

	// TODO: consider additional filtering
	results := append(songFiles.search(parts), clipFiles.search(parts)...)
	results = append(results, hostClips.search(parts)...)

	return results
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

func pickRandomClipWhichHasNotBeenPlayedInAWhile(clips *LibrarySet) *LibraryFile {
	if clips.Size() == 0 {
		return nil
	}

	var candidate *LibraryFile
	for range 2 {
		newCandidate := clips.GetRandom()
		if newCandidate.lastPlayed == nil {
			return newCandidate
		}
		if candidate == nil {
			// No checks required for the first candidate
			candidate = newCandidate
			continue
		}
		// Update current candidate if the new one has not been played for longer
		if newCandidate.lastPlayed.Before(*candidate.lastPlayed) {
			candidate = newCandidate
		}
	}

	return candidate
}

func folderExists(folder string) bool {
	info, err := os.Stat(folder)

	if os.IsNotExist(err) {
		// Does not exist
		return false
	} else if err != nil {
		// Unknown error
		return false
	} else if !info.IsDir() {
		// Not a folder
		return false
	}

	return true
}
