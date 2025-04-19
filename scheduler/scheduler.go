package scheduler

import (
	"math/rand"
	"time"

	"github.com/tim-we/wavestreamer/library"
	"github.com/tim-we/wavestreamer/player"
)

var schedulerQueue = make(chan player.Clip, 3)

func Start() {
	go func() {
		for {
			// First play music for ~10min
			musicTime := 0 * time.Second
			for musicTime < 10*time.Minute {
				if t := enqueueFile(library.PickRandomSong()); t > 0 {
					musicTime += t
				} else {
					break
				}
			}

			// Half of the time play a host clip...
			if rand.Intn(100) < 50 {
				if t := enqueueFile(library.PickRandomHostClip()); t > 0 {
					continue
				}
			}

			// ... the other half play some random clips.
			clipsTime := 0 * time.Second
			clipsCount := 0
			for clipsTime < time.Minute && clipsCount < 2 {
				if t := enqueueFile(library.PickRandomClip()); t > 0 {
					clipsTime += t
					clipsCount++
				}
			}
		}
	}()
}

// GetNextClip returns a Clip or nil. It does not block.
func GetNextClip() player.Clip {
	select {
	case clip := <-schedulerQueue:
		return clip
	default:
		return nil
	}
}

func enqueueFile(file *library.LibraryFile) time.Duration {
	if file == nil {
		return 0
	}

	clip := file.CreateClip()

	if clip == nil {
		return 0
	}

	schedulerQueue <- clip

	return clip.Duration()
}
