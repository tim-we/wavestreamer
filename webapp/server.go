package webapp

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
)

//go:embed dist/*
var content embed.FS

func StartServer() {
	// Strip the "dist" prefix so files are served at root (/)
	staticFiles, err := fs.Sub(content, "dist")
	if err != nil {
		log.Fatal(err)
	}

	// Serve embedded files
	http.Handle("/", http.FileServer(http.FS(staticFiles)))

	// Start server
	go func() {
		log.Println("Serving on http://localhost:8080")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()
}
