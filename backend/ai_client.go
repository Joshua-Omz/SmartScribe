// This module is to give a SOAP note response from the AI API. It handles the audio file processing and communication with the AI API, ensuring that the audio data is sent correctly and that the response is parsed into a structured format for further use in the application.
package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"time"
)

const (
	SpeechToTextURL = "https://speech.googleapis.com/v1/speech:recognize"
	LLMAPIURL       = "https://api.groq.com/openai/v1/chat/completions"
)

type AIClient struct {
	Client              *http.Client
	TranscribeModel_key string
	LLMAPIKey           string
}

// NewClient initializes and returns a new AIClient with the provided API key and a configured HTTP client to prevent goroutine leaks.
func NewClient(apikey string, llmKey string) (*AIClient, error) {
	if apikey == "" {
		return nil, errors.New("AI API key is missing. Please set the AI_API_KEY environment variable.")
	}
	return &AIClient{
		Client: &http.Client{
			Timeout: 45 * time.Second,
		},
		TranscribeModel_key: apikey,
		LLMAPIKey:           llmKey,
	}, nil
}

// Process Audio sends the audio data to the AI API and returns the transcribed text.
func (c *AIClient) TranscribeMedicalAudioAudio(audioFilePath string) (string, error) {
	log.Printf("Starting transcription for audio file: %s", audioFilePath)

	// 1.read the entire audio file into memory
	fileBytes, err := os.ReadFile(audioFilePath)

	if err != nil {
		log.Printf("Error reading audio file: %v", err)
		return "", fmt.Errorf("failed to read audio file: %w", err)
	}

	//2.Base64 encode the audio bytes
	encodedAudio := base64.StdEncoding.EncodeToString(fileBytes)
	log.Printf("Successfully read and base64 encoded audio. Base64 length: %d chars", len(encodedAudio))

	//3. Create the JSON payload for the Google Speech-to-Text API
	payload := map[string]interface{}{
		"config": map[string]interface{}{
			"model":                      "medical_conversation",
			"languageCode":               "en-US",
			"encoding":                   "WEBM_OPUS", // 1. Set specific codec
			"sampleRateHertz":            48000,       // 2. Mandatory for Opus
			"enableAutomaticPunctuation": true,        // 3. Recommended addition
		},
		"audio": map[string]interface{}{
			"content": encodedAudio,
		},
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling JSON payload: %v", err)
		return "", fmt.Errorf("failed to marshal STT request: %w", err)
	}

	//4. Create the HTTP request
	req, err := http.NewRequest(http.MethodPost, SpeechToTextURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error creating HTTP request for STT: %v", err)
		return "", fmt.Errorf("failed to create STT request: %w", err)
	}

	//5. Attach the Google API Key to authenticate the request
	req.Header.Set("X-Goog-Api-Key", c.TranscribeModel_key)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	log.Printf("Sending request to Google STT API (%s)", SpeechToTextURL)
	startTime := time.Now()

	//6. Send the request to the Google Speech-to-Text API
	resp, err := c.Client.Do(req)

	if err != nil {
		log.Printf("Error executing HTTP request: %v", err)
		return "", fmt.Errorf("failed to send STT request: %w", err)
	}
	defer resp.Body.Close()
	log.Printf("Received response from Google STT API in %v. Status Code: %d", time.Since(startTime), resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("STT API failed with status %d: %s", resp.StatusCode, string(bodyBytes))
		return "", fmt.Errorf("STT API returned non-200 status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// 8. Parse the response
	var result GoogleSTTResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("Error decoding JSON response from Google: %v", err)
		return "", fmt.Errorf("failed to decode AI API response: %w", err)
	}

	// 9. Extract the transcript
	// 9. Extract the transcript
	if len(result.Results) > 0 && len(result.Results[0].Alternatives) > 0 {
		transcript := result.Results[0].Alternatives[0].Transcript

		// If the first segment is just a single character (like "."),
		// iterate through all results to combine them instead.
		if len(transcript) <= 1 {
			log.Println("Initial transcript was 1 character. Combining all segments...")
			var fullTranscript string
			for _, res := range result.Results {
				if len(res.Alternatives) > 0 {
					// Use strings.TrimSpace if you want to cleanly format spaces
					fullTranscript += res.Alternatives[0].Transcript + " "
				}
			}
			log.Printf("Successfully transcribed audio (combined). Transcript length: %d chars. Transcript: %q", len(fullTranscript), fullTranscript)
			return fullTranscript, nil
		}

		// Original behavior: return just the first segment if it seems valid
		log.Printf("Successfully transcribed audio. Transcript length: %d chars. Transcript: %q", len(transcript), transcript)
		return transcript, nil
	}
	log.Println("Google API returned a successful 200 response, but no transcription results were found.")
	return "", fmt.Errorf("no transcription found in the API response")

}

type LLMRequest struct {
	Text         string `json:"clinical_text"`
	SystemPrompt string `json:"system_prompt"`
}

// StructureTextToSOAP takes raw clinical text and sends it to the LLM API to be structured into a SOAP note format. It returns a SOAP struct containing the Subjective, Objective, Assessment, and Plan sections extracted from the clinical text.

func (c *AIClient) StructureTextToSOAP(text string) (*SOAP, error) {
	log.Printf("Starting LLM SOAP structuring for text length: %d chars", len(text))
	prompt := fmt.Sprintf("System: You are a clinical AI. Extract the following text into a strict SOAP format JSON object.\n\nText: %s\n\nJSON:", text)
	// The Standard OpenAI/Groq Request Shape
	payload := map[string]interface{}{
		"model": "llama-3.1-8b-instant", // Groq's lightning-fast Llama 3 model
		"messages": []map[string]interface{}{
			{
				"role":    "system",
				"content": "You are a clinical AI. Extract the text into a strict SOAP JSON format. Output ONLY valid JSON containing the keys: subjective, objective, assessment, plan.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"response_format": map[string]interface{}{
			"type": "json_object", // Physically forces the AI to return clean JSON
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling LLM payload: %v", err)
		return nil, fmt.Errorf("failed to marshal LLM request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, LLMAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error creating HTTP request for LLM: %v", err)
		return nil, fmt.Errorf("failed to create LLM request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.LLMAPIKey)
	req.Header.Set("Content-Type", "application/json")

	log.Printf("Sending request to LLM API (%s)", LLMAPIURL)
	startTime := time.Now()

	resp, err := c.Client.Do(req)

	if err != nil {
		log.Printf("Error executing LLM HTTP request: %v", err)
		return nil, fmt.Errorf("failed to send LLM request: %w", err)
	}
	defer resp.Body.Close()
	log.Printf("Received response from LLM API in %v. Status Code: %d", time.Since(startTime), resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("LLM API failed with status %d: %s", resp.StatusCode, string(bodyBytes))
		return nil, fmt.Errorf("LLM API returned non-200 status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var result SOAP
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("Error decoding LLM JSON response: %v", err)
		return nil, fmt.Errorf("failed to decode LLM response: %w", err)
	}

	log.Println("Successfully formatted text into SOAP notes.")
	return &result, nil
}
