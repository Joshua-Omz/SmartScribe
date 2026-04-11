package main


import (
	"fmt"
	"log"
	"net/http"
)

// transcribeResponse is the dummy structured response returned by the API.

// transcribeHandler handles POST /api/transcribe.
// It accepts a multipart/form-data audio upload, logs receipt, and returns a
// dummy transcription result. No audio data is persisted to disk.

func main() {
	// Serve the compiled frontend from the sibling frontend/ directory.
	fs := http.FileServer(http.Dir("../frontend"))
	http.Handle("/", fs)

	http.HandleFunc("/api/transcribe", transcribeHandler)

	addr := ":8080"
	fmt.Printf("SmartScribe server listening on http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
