package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Server struct {
	aiClient *AIClient
}

func main() {
	godotenv.Overload() // Loads variables from .env file and overrides any blank host vars
	transcribeModel_Key := os.Getenv("GOOGLE_STT_API_KEY")
	llmKey := os.Getenv("LLM_API_KEY")

	// Strip literal quotes if powershell accidentally included them
	llmKey = strings.TrimSpace(strings.Trim(llmKey, "\""))
	transcribeModel_Key = strings.TrimSpace(strings.Trim(transcribeModel_Key, "\""))

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
