package webapp

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/tim-we/wavestreamer/library"
	"github.com/tim-we/wavestreamer/player"
	"github.com/tim-we/wavestreamer/player/clips"
	"github.com/tim-we/wavestreamer/scheduler"
	"github.com/tim-we/wavestreamer/utils"
)

//go:embed dist/*
var content embed.FS

var startTime = time.Now()

func StartServer(port int, news bool) {
	// Strip the "dist" prefix so files are served at root (/)
	staticFiles, err := fs.Sub(content, "dist")
	if err != nil {
		log.Fatal(err)
	}

	// Serve static (embedded) files
	http.Handle("/", http.FileServer(http.FS(staticFiles)))

	// API: /now endpoint
	http.HandleFunc("/api/now", func(w http.ResponseWriter, r *http.Request) {
		response := ApiNowResponse{
			Status:      "ok",
			Current:     player.GetCurrentlyPlaying(),
			History:     player.GetHistory(),
			LibraryInfo: ApiNowLibraryInfo{},
			Uptime:      utils.PrettyDuration(time.Since(startTime), ""),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// API: /skip endpoint
	http.HandleFunc("/api/skip", func(w http.ResponseWriter, r *http.Request) {
		player.SkipCurrent()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ApiOkResponse{"ok"})
	})

	// API: /pause endpoint
	http.HandleFunc("/api/pause", func(w http.ResponseWriter, r *http.Request) {
		player.QueueClip(clips.NewPause(10 * time.Minute))
		player.SkipCurrent()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ApiOkResponse{"ok"})
	})

	http.HandleFunc("/api/library/search", func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters and get the value of `query`
		query := r.URL.Query().Get("query")
		// Collect results
		results := searchResultsAsDTOs(library.Search(query))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ApiSearchResponse{"ok", results})
	})

	http.HandleFunc("/api/schedule", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		if err := r.ParseForm(); err != nil {
			encoder.Encode(ApiErrorResponse{"error", fmt.Sprintf("%v", err)})
			return
		}
		if !r.Form.Has("file") {
			encoder.Encode(ApiErrorResponse{"error", "File field not set."})
			return
		}
		rawClipId := r.Form.Get("file")
		fileId, parseErr := uuid.Parse(rawClipId)
		if parseErr != nil {
			encoder.Encode(ApiErrorResponse{"error", "Invalid id value."})
			return
		}
		file := library.GetFileById(fileId)
		if file == nil {
			encoder.Encode(ApiErrorResponse{"error", "File not found."})
			return
		}
		player.QueueClip(file.CreateClip())
		encoder.Encode(ApiOkResponse{"ok"})
	})

	http.HandleFunc("/api/schedule/news", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		scheduler.ScheduleTagesschauNow()
		json.NewEncoder(w).Encode(ApiOkResponse{"ok"})
	})

	http.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ApiConfigResponse{"ok", news})
	})

	// Start server
	go func() {
		log.Printf("Serving on http://localhost:%d\n", port)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
	}()
}

func searchResultsAsDTOs(results []*library.LibraryFile) []SearchResultEntry {
	stringResults := make([]SearchResultEntry, len(results))
	for i, file := range results {
		stringResults[i] = SearchResultEntry{
			Id:   file.Id.String(),
			Name: file.Name(),
		}
	}
	return stringResults
}

type ApiNowResponse struct {
	Status      string                `json:"status"`
	Current     string                `json:"current"`
	History     []player.HistoryEntry `json:"history"`
	LibraryInfo ApiNowLibraryInfo     `json:"library"`
	Uptime      string                `json:"uptime"`
}

type ApiNowLibraryInfo struct {
	Music int `json:"music"`
	Hosts int `json:"hosts"`
	Other int `json:"other"`
	Night int `json:"night"`
}

type ApiOkResponse struct {
	Status string `json:"status"`
}

type ApiErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type ApiSearchResponse struct {
	Status  string              `json:"status"`
	Results []SearchResultEntry `json:"results"`
}

type SearchResultEntry struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type ApiConfigResponse struct {
	Status string `json:"status"`
	News   bool   `json:"news"`
}
