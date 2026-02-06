package player

type PlayerEvent interface {
	Type() string
}

// NowPlayingEvent represents the current playback state of the audio player.
type NowPlayingEvent struct {
	CurrentClip Clip
}

func (event NowPlayingEvent) Type() string {
	return "now-playing"
}
