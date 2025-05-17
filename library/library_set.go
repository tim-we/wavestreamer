package library

import (
	"fmt"
	"math/rand"
	"sync"
)

type LibrarySet struct {
	files map[string]*LibraryFile // holds the data (ground truth)
	list  []*LibraryFile          // for random access (derived from `files`)
	dirty bool                    // dirty=true implies the list is outdated and must be regenerated
	mu    sync.RWMutex            // different threads/goroutines access this struct
}

func NewLibrarySet(initialCapacity int) *LibrarySet {
	return &LibrarySet{
		files: make(map[string]*LibraryFile),
		list:  make([]*LibraryFile, initialCapacity),
		dirty: false,
	}
}

// AddOrUpdate adds a new file or updates an existing one.
func (ls *LibrarySet) AddOrUpdate(path string) error {
	file, err := NewLibraryFile(path)

	if err != nil {
		return fmt.Errorf("failed to load new library file %s. Error: %v\n", path, err)
	}

	ls.mu.Lock()
	defer ls.mu.Unlock()

	ls.files[path] = file
	ls.dirty = true

	return nil
}

// Remove deletes the file entry.
func (ls *LibrarySet) Remove(path string) {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	if _, ok := ls.files[path]; ok {
		delete(ls.files, path)
		ls.dirty = true
	}
}

// Rename changes the key and preserves the file object and metadata.
func (ls *LibrarySet) Rename(oldPath, newPath string) {
	if oldPath == newPath {
		return
	}

	ls.mu.Lock()
	defer ls.mu.Unlock()

	if file, ok := ls.files[oldPath]; ok {
		delete(ls.files, oldPath)
		file.filepath = newPath
		ls.files[newPath] = file
		ls.dirty = true
	}
}

// Regenerate internal list WITHOUT locking.
func (ls *LibrarySet) regenerateList() {
	ls.list = make([]*LibraryFile, 0, len(ls.files))
	for _, f := range ls.files {
		ls.list = append(ls.list, f)
	}
	ls.dirty = false
}

// UpdateInternals should be called after all modifications have been performed.
// It refreshes the slice used for random access.
func (ls *LibrarySet) UpdateInternals() {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	ls.regenerateList()
}

// GetRandom returns a random file.
func (ls *LibrarySet) GetRandom() *LibraryFile {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	if ls.dirty {
		// List is outdated so we need to regenerate it.
		ls.regenerateList()
	}

	if len(ls.list) == 0 {
		panic(fmt.Errorf("no files available"))
	}

	return ls.list[rand.Intn(len(ls.list))]
}

// Size returns the number of files in this set.
func (ls *LibrarySet) Size() int {
	return len(ls.files)
}

func (ls *LibrarySet) search(queryParts []string) []*LibraryFile {
	results := make([]*LibraryFile, 0, 16)

clipLoop:
	for _, clip := range ls.list {
		// All parts must match.
		for _, part := range queryParts {
			if !clip.Matches(part) {
				continue clipLoop
			}
		}
		results = append(results, clip)
	}

	return results
}
