package library

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
)

var hostClips = make([]string, 0, 128)
var songFiles = make([]string, 0, 512)
var clipFiles = make([]string, 0, 256)

func ScanRootDir(root string) {
	fmt.Printf("Searching for files in %s...\n", root)
	unknownFiles := 0

	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.Type().IsRegular() {
			if isSong(path) {
				songFiles = append(songFiles, path)
			} else if isClipFile(path) {
				clipFiles = append(clipFiles, path)
			} else if isHostClip(path) {
				hostClips = append(hostClips, path)
			} else {
				unknownFiles++
			}
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
