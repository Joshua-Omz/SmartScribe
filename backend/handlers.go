package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func (s *Server) handleTranscription(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request to %s", r.Method, r.URL.Path)

	// 1. Method Validation
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Payload Security (10MB limit)
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	// 3. Parse Multipart Form
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		log.Printf("Error parsing multipart form: %v", err)
		http.Error(w, "Payload too large or malformed", http.StatusBadRequest)
		return
	}

	// 4. Extract the incoming File
	file, _, err := r.FormFile("audio")
	if err != nil {
		log.Printf("Error extracting audio file: %v", err)
		http.Error(w, "Invalid or missing 'audio' field", http.StatusBadRequest)
		return
	}
	defer file.Close() // Release the multipart file from memory

	// 5. Create a secure temporary file on the OS
	// The first argument "" means use the default OS temp directory (e.g., /tmp on Linux)
	// The second argument dictates the naming pattern.
	tempFile, err := os.CreateTemp("", "raw_audio_*.webm")
	if err != nil {
		log.Printf("Error creating temp file: %v", err)
		http.Error(w, "Failed to allocate temporary storage", http.StatusInternalServerError)
		return
	}
	// Close the file descriptor when we are done writing to it
	defer tempFile.Close()

	// 6. Zero-Storage Compliance (CRITICAL)
	// Guarantee the file is wiped from the hard drive the moment this request finishes.
	defer os.Remove(tempFile.Name())

	// 7. Stream the data from the network directly to the hard drive
	// io.Copy does this in small, efficient chunks (usually 32KB).
	// It never loads the whole 10MB into RAM at once.
	if _, err := io.Copy(tempFile, file); err != nil {
		log.Printf("Error copying audio stream to temp file: %v", err)
		http.Error(w, "Failed to save audio stream", http.StatusInternalServerError)
		return
	}

	// --- MODULE 1 COMPLETE ---
	// At this exact line, tempFile.Name() holds the absolute path to the safely ingested audio.
	// You are now ready to pass tempFile.Name() directly to your AIClient.

	// For testing purposes right now, just return a success message:
	log.Printf("Audio securely ingested to temporary file: %s", tempFile.Name())

	// 8. Call the AI Client for Processing
	rawText, err := s.aiClient.TranscribeMedicalAudioAudio(tempFile.Name())
	if err != nil {
		log.Printf("Error during AI transcription: %v", err)
		http.Error(w, "AI transcription failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 9. Return the transcribed text as JSON
	structuredData, err := s.aiClient.StructureTextToSOAP(rawText)
	if err != nil {
		log.Printf("Error during AI structuring: %v", err)
		http.Error(w, "AI structuring failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := TranscriptionResponse2{
		Status:     "success",
		RawText:    rawText,
		Structured: *structuredData,
	}

	// 10. Set headers and stream the JSON back to the frontend
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Handle CORS for local frontend testing
	w.WriteHeader(http.StatusOK)

	// Idiomatic Go: Use json.Encoder to stream directly to the ResponseWriter without buffering in memory
	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("JSON Encode Error: %v\n", err)
	}

}
