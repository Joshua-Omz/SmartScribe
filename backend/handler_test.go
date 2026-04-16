package main

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/joho/godotenv"
)

func TestHandleTranscription(t *testing.T) {
	// 1. Create a buffer to hold the multipart form data
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// 2. Add the required "audio" field to the form
	fileWriter, err := writer.CreateFormFile("audio", "audio.webm")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	// read the real audio file in your backend directory and write its bytes to the form file
	audioBytes, err := os.ReadFile("smartscribe_test (1).webm") // Make sure this file exists in your backend directory
	if err != nil {
		t.Fatalf("Failed to read test audio file: %v", err)
	}
	fileWriter.Write(audioBytes)

	// MUST close the writer to finalize the multipart boundary data
	writer.Close()

	// 3. Create the mock HTTP request
	req, err := http.NewRequest(http.MethodPost, "/transcribe", &requestBody)
	if err != nil {
		t.Fatalf("Failed to create HTTP request: %v", err)
	}

	// Critical: Set the Content-Type header so the handler knows it's multipart
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 4. Create a ResponseRecorder to capture what the handler writes back
	recorder := httptest.NewRecorder()

	// 5. Initialize your Server struct
	// Note: You will need to attach a dummy or mock AIClient so the handler
	// doesn't panic when it tries to call s.aiClient.TranscribeMedicalAudioAudio()
	godotenv.Overload() // Force override any system cache of the old env vars
	googleKey := os.Getenv("GOOGLE_STT_API_KEY")
	llmKey := os.Getenv("LLM_API_KEY")

	// Strip literal quotes if powershell accidentally included them
	llmKey = strings.TrimSpace(strings.Trim(llmKey, "\""))
	googleKey = strings.TrimSpace(strings.Trim(googleKey, "\""))

	if googleKey == "" || llmKey == "" {
		t.Skip("Skipping test: Requires GOOGLE_API_KEY and LLM_API_KEY environment variables")
	}
	test_client, err := NewClient(googleKey, llmKey)
	if err != nil {
		t.Fatalf("Failed to create test client: %v", err)
	}

	server := &Server{
		aiClient: test_client,
	}

	// 6. Call the handler directly
	server.handleTranscription(recorder, req)

	// 7. Assert the results
	// The exact status will depend on what dummyClient returns. If you don't mock
	// the actual API call, it might return a 500 error because the AI API call fails.
	// You can check that it at least didn't return a 400 Bad Request.
	if recorder.Code == http.StatusBadRequest {
		t.Errorf("Expected request to pass validation, but got 400 Bad Request. Body: %s", recorder.Body.String())
	}
}
func TestLLMFallback(t *testing.T) {
    // 1. Initialize the client with a deliberately broken LLM API key
    // We'll provide a dummy STT key just to satisfy the constructor
    brokenLLMKey := "invalid_key_to_force_error"
    dummySTTKey := "dummy_stt_key"
    
    test_client, err := NewClient(dummySTTKey, brokenLLMKey)
    if err != nil {
        t.Fatalf("Failed to create test client: %v", err)
    }

    // 2. Define the raw text we want to process
    rawText := "Patient presents with mild headache and fatigue. History of hypertension."

    // 3. Call the StructureTextToSOAP method
    soap, err := test_client.StructureTextToSOAP(rawText)

    // 4. Assert that no error is returned (because we caught it and used the fallback)
    if err != nil {
        t.Fatalf("Expected no error due to fallback, but got: %v", err)
    }

    // 5. Assert that the fallback was properly populated
    if soap == nil {
        t.Fatal("Expected SOAP object to be returned, got nil")
    }

    // 6. Check that the predefined fallback fields are present
    if soap.Subjective != "Unable to transcribe audio. This is a fallback SOAP note." {
        t.Errorf("Expected fallback Subjective text, but got: %q", soap.Subjective)
    }

    if soap.Objective != "Audio transcription failed. No objective data available." {
        t.Errorf("Expected fallback Objective text, but got: %q", soap.Objective)
    }
}
