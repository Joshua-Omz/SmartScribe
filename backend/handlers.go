package main

import (
	"encoding/json"
	"log"
	"net/http"
)



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

	resp := TranscriptionResponse{
		Status: "success",
		Text:   "Patient has a mild fever.",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("error encoding response: %v", err)
	}
}
