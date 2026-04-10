package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type TranscribeResponse struct {
	Status string `json:"status"`
	Text   string `json:"text"`
}

func transcribeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit upload size to 32 MB
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "failed to parse multipart form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("audio")
	if err != nil {
		http.Error(w, "audio file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	log.Printf("received audio file: name=%s size=%d bytes", header.Filename, header.Size)

	resp := TranscribeResponse{
		Status: "success",
		Text:   "Patient has a mild fever.",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

func main() {
	// Serve static frontend files
	fs := http.FileServer(http.Dir("../frontend"))
	http.Handle("/", fs)

	// API endpoints
	http.HandleFunc("/api/transcribe", transcribeHandler)

	addr := ":8080"
	fmt.Printf("SmartScribe server listening on http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
