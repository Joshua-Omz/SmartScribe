package main

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "os"
    "strings"
    "testing"
)

// TestNewClient verifies that the client requires an API key.
func TestNewClient(t *testing.T) {
    _, err := NewClient("")
    if err == nil {
        t.Error("Expected an error for empty API key, got nil")
    }

    client, err := NewClient("valid-key")
    if err != nil {
        t.Errorf("Unexpected error for valid API key: %v", err)
    }
    if client.APIKey != "valid-key" {
        t.Errorf("Expected APIKey to be 'valid-key', got '%s'", client.APIKey)
    }
}

// TestProcessAudio tests the multipart construction and HTTP request 
// without hitting the real external API.
func TestProcessAudio(t *testing.T) {
    // 1. Create a mock HTTP server using httptest
    mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Verify it's a POST request
        if r.Method != http.MethodPost {
            t.Errorf("Expected POST request, got %s", r.Method)
        }

        // Verify Authorization header
        authHeader := r.Header.Get("Authorization")
        if !strings.HasPrefix(authHeader, "Bearer ") {
            t.Errorf("Expected Bearer token, got %s", authHeader)
        }

        // Mock a successful JSON response (matching TranscriptionResponse)
        w.WriteHeader(http.StatusOK)
        mockResponse := TranscriptionResponse{
            // Add placeholder fields from your TranscriptionResponse struct here
            // For example: Text: "Patient has no recorded fever.",
        }
        json.NewEncoder(w).Encode(mockResponse)
    }))
    defer mockServer.Close() // Make sure to shut down the server when the test finishes

    // 2. Set the environment variable to point to our mock server instead of the real one
    // t.Setenv automatically cleans up after the test completes (Go 1.17+)
    t.Setenv("AI_PROXY_URL", mockServer.URL)

    // 3. Create a temporary dummy audio file to upload
    tmpFile, err := os.CreateTemp("", "dummy_audio_*.wav")
    if err != nil {
        t.Fatalf("Failed to create temp file: %v", err)
    }
    defer os.Remove(tmpFile.Name()) // Clean up the temp file
    tmpFile.Write([]byte("fake audio content"))
    tmpFile.Close() // Close so ProcessAudio can open it

    // 4. Initialize our client and call the function
    client, _ := NewClient("fake-testing-key")
    response, err := client.ProcessAudio(tmpFile.Name())

    // 5. Assert the results
    if err != nil {
        t.Fatalf("ProcessAudio failed unexpectedly: %v", err)
    }
    if response == nil {
        t.Fatal("Expected a TranscriptionResponse, got nil")
    }
}