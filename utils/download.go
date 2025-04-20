package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func DownloadToTempFile(url string) (*os.File, error) {
	// Make HTTP GET request
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "download-*.tmp")
	if err != nil {
		return nil, err
	}

	// Copy response body to temp file
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name()) // clean up on error
		return nil, err
	}

	// Seek to beginning because we likely want to read from it again
	_, err = tmpFile.Seek(0, 0)
	if err != nil {
		tmpFile.Close()
		return nil, err
	}

	return tmpFile, nil
}

func DownloadToMemory(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}
