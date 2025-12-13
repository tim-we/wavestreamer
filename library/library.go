package library

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/fsnotify/fsnotify"
	"github.com/google/uuid"
)

var hostClips = NewLibrarySet(128)
var songFiles = NewLibrarySet(512)
var clipFiles = NewLibrarySet(256)

func WatchRootDir(root string) {
	if !folderExists(root) {
		log.Fatalf("Folder '%s' does not exist.", root)
	}

	fmt.Printf("Searching for files in %s...\n", root)
	unknownFiles := 0

	folders := make([]string, 0, 8)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err1 error) error {
		if err1 != nil {
			return err1
		}

		// Ignore hidden files and directories
		if strings.HasPrefix(entry.Name(), ".") {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if entry.IsDir() {
			folders = append(folders, path)
			return nil
		}

		if !entry.Type().IsRegular() {
			return nil
		}

		var err2 error
		librarySet := getLibrarySetForFile(path)

		if librarySet == nil {
			unknownFiles++
			return nil
		}

		if err2 = librarySet.AddOrUpdate(path); err2 != nil {
			return err2
		}

		return nil
	})

	if err != nil {
		// If the initial scan fails we panic.
		panic(fmt.Errorf("error scanning the directory '%v' for files: %v", root, err))
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

	go watchFoldersForChanges(folders)

	go func() {
		songFiles.loadMissingMetaData()
		clipFiles.loadMissingMetaData()
		hostClips.loadMissingMetaData()

		log.Println("Finished loading meta data.")
	}()
}

func watchFoldersForChanges(folders []string) {
	fmt.Printf("Watching %d folders for changes...\n", len(folders))

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Printf("failed to create watcher: %v\n", err)
		return
	}

	for _, folder := range folders {
		if err := watcher.Add(folder); err != nil {
			log.Printf("Failed to watch folder %s: %v", folder, err)
		}
	}

	changeEvents := make(chan fsnotify.Event, 16)

	go func() {
		for {
			event := <-changeEvents
			path := event.Name

			switch {
			case event.Op&fsnotify.Create != 0:
				info, err := os.Stat(path)
				if err != nil {
					continue
				}

				if info.IsDir() {
					// A new folder has been added. We should watch it.
					_ = watcher.Add(path)
					continue
				}

				librarySet := getLibrarySetForFile(path)
				if librarySet == nil {
					continue
				}
				if err := librarySet.AddOrUpdate(path); err != nil {
					log.Printf("Warning: %v\n", err)
				} else {
					log.Printf("Added %s\n", path)
				}
			case event.Op&fsnotify.Write != 0:
				librarySet := getLibrarySetForFile(path)
				if librarySet == nil {
					continue
				}
				if err := librarySet.AddOrUpdate(path); err != nil {
					log.Printf("Warning: %v\n", err)
				} else {
					log.Printf("Updated %s\n", path)
				}
			case event.Op&fsnotify.Rename != 0:
				// Treat as remove for now (can't get the new name)
				// FIXME #9: Get the new name or trigger rescan
				fallthrough
			case event.Op&fsnotify.Remove != 0:
				removed := songFiles.Remove(path)
				removed = removed || clipFiles.Remove(path)
				removed = removed || hostClips.Remove(path)
				if removed {
					log.Printf("Removed %s\n", path)
				}
			}

			time.Sleep(10 * time.Millisecond)
		}
	}()

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			path := event.Name
			if strings.HasPrefix(filepath.Base(path), ".") {
				continue // Skip hidden files
			}

			changeEvents <- event

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("Watcher error:", err)
		}
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

// Search the library for clips matching the query. The number of results will be limited by the given limit.
func Search(query string, limit int) []*LibraryFile {
	modifiedQuery := strings.Trim(strings.ToLower(query), " ")

	if len(modifiedQuery) == 0 {
		return []*LibraryFile{}
	}

	// Prepare search
	parts := strings.Split(modifiedQuery, " ")

	// Prepare parallel search
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	results := make(chan *LibraryFile, 32)
	var wg sync.WaitGroup
	wg.Add(3)

	// Perform parallel search:
	go func() {
		defer wg.Done()
		songFiles.search(parts, ctx, results)
	}()

	go func() {
		defer wg.Done()
		clipFiles.search(parts, ctx, results)
	}()

	go func() {
		defer wg.Done()
		hostClips.search(parts, ctx, results)
	}()

	// Close channel when all searches complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// TODO: consider additional filtering

	collected := make([]*LibraryFile, 0, limit)
	for result := range results {
		collected = append(collected, result)
		if len(collected) >= limit {
			// Stop searches
			cancel()
			break
		}
	}

	return collected
}

func GetFileById(clipId uuid.UUID) *LibraryFile {
	if clip := songFiles.GetById(clipId); clip != nil {
		return clip
	}
	if clip := clipFiles.GetById(clipId); clip != nil {
		return clip
	}
	return hostClips.GetById(clipId)
}

func getLibrarySetForFile(file string) *LibrarySet {
	if matches, _ := doublestar.Match("**/music/**/*", file); matches {
		return songFiles
	}
	if matches, _ := doublestar.Match("**/night/**/*", file); matches {
		return songFiles
	}
	if matches, _ := doublestar.Match("**/hosts/**/*", file); matches {
		return hostClips
	}
	if matches, _ := doublestar.Match("**/clips/**/*", file); matches {
		return clipFiles
	}
	return nil
}

func pickRandomClipWhichHasNotBeenPlayedInAWhile(clips *LibrarySet) *LibraryFile {
	if clips == nil || clips.Size() == 0 {
		log.Println("Tried to pick a random clip from an empty library set.")
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
