package main

import (
	"log"
	"net/http"
	"os"
)

type Server struct {
	aiClient *AIClient
}

func main() {
	transcribeModel_Key := os.Getenv("GOOGLE_STT_API_KEY")
	llmKey := os.Getenv("LLM_API_KEY")

	client, err := NewClient(transcribeModel_Key, llmKey)

	if err != nil {
		log.Fatalf("Failed to initialize AI client: %v", err)
	}

	server := &Server{
		aiClient: client,
	}

	// Register the handler function to the specific URL path
	http.HandleFunc("/api/transcribe", server.handleTranscription)

	log.Println("SmartScribe Gateway starting on port 8080...")

	// Start the server. If it crashes, log.Fatal will print the error.
	log.Fatal(http.ListenAndServe(":8080", nil))
}
