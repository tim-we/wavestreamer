package player

import "time"

type HistoryEntry struct {
	StartTime     time.Time `json:"start"`
	Title         string    `json:"title"`
	Skipped       bool      `json:"skipped"`
	UserScheduled bool      `json:"userScheduled"`
}

const historyLength = 10

var history []HistoryEntry

func addClipToHistory(clip Clip) {
	history = append(history, HistoryEntry{
		StartTime: time.Now(),
		Title:     clip.Name(),
	})
	if len(history) > historyLength {
		history = history[1:] // remove the oldest entry
	}
}

func GetHistory() []HistoryEntry {
	return history[:]
}
