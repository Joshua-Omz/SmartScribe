package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// transcribeResponse is the dummy structured response returned by the API.
type transcribeResponse struct {
	Status string `json:"status"`
	Text   string `json:"text"`
}

// transcribeHandler handles POST /api/transcribe.
// It accepts a multipart/form-data audio upload, logs receipt, and returns a
// dummy transcription result. No audio data is persisted to disk.
func transcribeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit upload size to 32 MB in memory; remainder spills to temp files.
	// 32 MB comfortably covers several minutes of compressed WebM/Opus audio.
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "failed to parse multipart form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("audio")
	if err != nil {
		http.Error(w, "audio field missing from request", http.StatusBadRequest)
		return
	}
	defer file.Close()

	log.Printf("received audio file: name=%q size=%d bytes content-type=%q",
		header.Filename, header.Size, header.Header.Get("Content-Type"))

	resp := transcribeResponse{
		Status: "success",
		Text:   "Patient has a mild fever.",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("error encoding response: %v", err)
	}
}

func main() {
	// Serve the compiled frontend from the sibling frontend/ directory.
	fs := http.FileServer(http.Dir("../frontend"))
	http.Handle("/", fs)

	http.HandleFunc("/api/transcribe", transcribeHandler)

	addr := ":8080"
	fmt.Printf("SmartScribe server listening on http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
