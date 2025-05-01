package player

import "time"

// Clip defines an interface for audio playback sources.
type Clip interface {
	// NextChunk retrieves the next audio chunk for playback.
	// Returns the chunk and a boolean indicating if more chunks are available.
	NextChunk() (*AudioChunk, bool)

	// Stop aborts playback and stops any associated processes.
	Stop()

	// A string representation of the clip, for audio clips something like: [Artist] - [Title]
	Name() string

	// Duration of the clip.
	Duration() time.Duration
}
