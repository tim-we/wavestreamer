package webapp

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
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

	// API endpoints:

	http.HandleFunc("/api/now", func(w http.ResponseWriter, r *http.Request) {
		current := player.GetCurrentlyPlaying()
		currentClipName := "-"

		if current != nil {
			currentClipName = current.Name()
		}

		_, isPause := current.(*clips.PauseClip)

		response := ApiNowResponse{
			Status:      "ok",
			Current:     currentClipName,
			IsPause:     isPause,
			History:     player.GetHistory(),
			LibraryInfo: ApiNowLibraryInfo{},
			Uptime:      utils.PrettyDuration(time.Since(startTime), ""),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	http.HandleFunc("/api/skip", func(w http.ResponseWriter, r *http.Request) {
		player.SkipCurrent()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ApiOkResponse{"ok"})
	})

	http.HandleFunc("/api/pause", func(w http.ResponseWriter, r *http.Request) {
		current := player.GetCurrentlyPlaying()

		// If the current clip is a Pause we don't schedule another one,
		// we skip the current one (see below)
		if _, isPause := current.(*clips.PauseClip); current == nil || !isPause {
			player.QueueClip(clips.NewPause(10 * time.Minute))
		}

		player.SkipCurrent()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ApiOkResponse{"ok"})
	})

	http.HandleFunc("/api/repeat", func(w http.ResponseWriter, r *http.Request) {
		current := player.GetCurrentlyPlaying()

		if current == nil {
			respondWithError(w, "nothing to repeat")
			return
		}

		nextClip := current.Duplicate()
		player.QueueClipNext(nextClip)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ApiOkResponse{"ok"})
	})

	http.HandleFunc("/api/library/search", func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters and get the value of `query`
		query := r.URL.Query().Get("query")
		// Collect results
		results := searchResultsAsDTOs(library.Search(query, 100))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ApiSearchResponse{"ok", results})
	})

	http.HandleFunc("/api/library/download", func(w http.ResponseWriter, r *http.Request) {
		rawFile := r.URL.Query().Get("file")
		fileId, err := uuid.Parse(rawFile)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			respondWithError(w, "invalid or missing 'file' query parameter")
			return
		}

		libFile := library.GetFileById(fileId)
		if libFile == nil {
			w.WriteHeader(http.StatusNotFound)
			respondWithError(w, "file not found")
			return
		}
		// Set Content-Disposition header to force download and specify filename
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filepath.Base(libFile.Path())))
		http.ServeFile(w, r, libFile.Path())
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

func respondWithError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ApiErrorResponse{"error", message})
}
