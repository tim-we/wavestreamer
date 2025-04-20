package webapp

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"

	"github.com/tim-we/wavestreamer/player"
)

//go:embed dist/*
var content embed.FS

func StartServer() {
	// Strip the "dist" prefix so files are served at root (/)
	staticFiles, err := fs.Sub(content, "dist")
	if err != nil {
		log.Fatal(err)
	}

	// Serve static (embedded) files
	http.Handle("/", http.FileServer(http.FS(staticFiles)))

	// API /now endpoint
	http.HandleFunc("/api/v1.0/now", func(w http.ResponseWriter, r *http.Request) {
		response := ApiNowResponse{
			Status:      "ok",
			Current:     player.GetCurrentlyPlaying(),
			History:     []string{},
			LibraryInfo: ApiNowLibraryInfo{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// API /skip endpoint
	http.HandleFunc("/api/v1.0/skip", func(w http.ResponseWriter, r *http.Request) {
		player.SkipCurrent()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ApiOkResponse{"ok"})
	})

	// Start server
	go func() {
		log.Println("Serving on http://localhost:8080")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()
}

type ApiNowResponse struct {
	Status      string            `json:"status"`
	Current     string            `json:"current"`
	History     []string          `json:"history"`
	LibraryInfo ApiNowLibraryInfo `json:"library"`
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
