package player

type Clip interface {
	// Get the next AudioChunk of this clip. Returns nil if there is no more data.
	NextChunk() (*AudioChunk, bool)

	// Abort playback of the current clip. Stop decoding or other connected processes.
	Stop()

	// A string representation of the clip, for audio clips something like: [Artist] - [Title]
	Name() string

	// Duration of the clip in seconds.
	Duration() int
}

type AudioChunk = struct {
	Left   []float32
	Right  []float32
	Length int
}
