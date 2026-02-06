package webapp

import (
	"embed"
	"encoding/json"
	"errors"
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

	addJsonEndpoint("/api/skip", func(r *http.Request) (any, error) {
		player.SkipCurrent(false)
		return ApiOkResponse{"ok"}, nil
	})

	addJsonEndpoint("/api/pause", func(r *http.Request) (any, error) {
		current := player.GetCurrentlyPlaying()

		// If the current clip is a Pause we don't schedule another one,
		// we skip the current one (see below)
		if _, isPause := current.(*clips.PauseClip); current == nil || !isPause {
			player.QueueClip(clips.NewPause(10 * time.Minute))
		}

		player.SkipCurrent(true)
		return ApiOkResponse{"ok"}, nil
	})

	addJsonEndpoint("/api/repeat", func(r *http.Request) (any, error) {
		current := player.GetCurrentlyPlaying()
		if current == nil {
			return nil, errors.New("nothing to repeat")
		}

		nextClip := current.Duplicate()
		player.QueueClipNext(nextClip)

		return ApiOkResponse{"ok"}, nil
	})

	addJsonEndpoint("/api/library/search", func(r *http.Request) (any, error) {
		// Parse query parameters and get the value of `query`
		query := r.URL.Query().Get("query")
		// Collect results
		results := searchResultsAsDTOs(library.Search(query, 100))
		return ApiSearchResponse{"ok", results}, nil
	})

	http.HandleFunc("/api/library/download", func(w http.ResponseWriter, r *http.Request) {
		rawFile := r.URL.Query().Get("file")
		fileId, err := uuid.Parse(rawFile)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid or missing 'file' query parameter")
			return
		}

		libFile := library.GetFileById(fileId)
		if libFile == nil {
			respondWithError(w, http.StatusNotFound, "file not found")
			return
		}
		// Set Content-Disposition header to force download and specify filename
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filepath.Base(libFile.Path())))
		http.ServeFile(w, r, libFile.Path())
	})

	addJsonEndpoint("/api/schedule", func(r *http.Request) (any, error) {
		if err := r.ParseForm(); err != nil {
			return nil, err
		}
		if !r.Form.Has("file") {
			return nil, errors.New("File field not set.")
		}
		rawClipId := r.Form.Get("file")
		fileId, parseErr := uuid.Parse(rawClipId)
		if parseErr != nil {
			return nil, errors.New("Invalid id value.")
		}
		file := library.GetFileById(fileId)
		if file == nil {
			// TODO: 404 code
			return nil, errors.New("File not found.")
		}
		player.QueueClip(file.CreateClip())
		return ApiOkResponse{"ok"}, nil
	})

	addJsonEndpoint("/api/schedule/news", func(r *http.Request) (any, error) {
		// TODO: avoid double scheduling
		scheduler.ScheduleTagesschauNow()
		return ApiOkResponse{"ok"}, nil
	})

	addJsonEndpoint("/api/config", func(r *http.Request) (any, error) {
		return ApiConfigResponse{"ok", news}, nil
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

func addJsonEndpoint(path string, handler func(r *http.Request) (any, error)) {
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		encoder := json.NewEncoder(w)
		result, err := handler(r)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			encoder.Encode(ApiErrorResponse{"error", err.Error()})
			return
		}

		w.WriteHeader(http.StatusOK)

		encoder.Encode(result)
	})
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ApiErrorResponse{"error", message})
}
