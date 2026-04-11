package main

import (
	"errors"
	"bytes"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"fmt"
	"mime/multipart"
	"io"
	"encoding/json"

)

type AIClient struct {
	Client  *http.Client	
	APIKey  string
	
		
}

// NewClient initializes and returns a new AIClient with the provided API key and a configured HTTP client to prevent goroutine leaks.
func NewClient(apikey string) (*AIClient,error ) {
	if apikey == "" {
		return nil, errors.New("AI API key is missing. Please set the AI_API_KEY environment variable.")
	}
	return &AIClient{
		Client:  &http.Client{
			Timeout: 45 * time.Second	,
		},
		APIKey:  apikey,
	},nil
}

//  Process Audio sends the audio data to the AI API and returns the transcribed text.
func (c *AIClient) ProcessAudio(filePath string) (*TranscriptionResponse, error) {
  file , err := os.Open(filePath)

  if err != nil {
	return nil, fmt.Errorf("failed to open compressed audio: %w", err)
  }
  defer file.Close()

  // 2. Construct the Multipart Form strictly in-memory
  var requestBody bytes.Buffer	

  writer := multipart.NewWriter(&requestBody)

  part, err := writer.CreateFormFile("audio", filepath.Base(filePath))
  if err != nil {
	return nil, fmt.Errorf("failed to create form file: %w", err)
  }

  if _, err := io.Copy(part, file); err != nil {
	return nil, fmt.Errorf("failed to copy audio data to form: %w", err)
  }

  systemInstruction := `You are a clinical AI assistant for HelpMum. 
Listen to this audio and extract the data strictly into a SOAP note format. 
Focus specifically on maternal health and immunization terms.`

writer.WriteField("model","vax-llama-8b")
writer.WriteField("task",systemInstruction)
writer.WriteField("response_format","json")


if err := writer.Close(); err != nil {
	return nil, fmt.Errorf("failed to close multipart writer: %w", err)
}

// 3. Create the HTTP request
req, err := http.NewRequest("POST", os.Getenv("AI_PROXY_URL"), &requestBody)
if err != nil {
	return nil, fmt.Errorf("failed to create HTTP request: %w", err)
}

req.Header.Set("Content-Type", writer.FormDataContentType())
req.Header.Set("Authorization", "Bearer "+c.APIKey)

// 4. Send the request
resp, err := c.Client.Do(req)
if err != nil {
	return nil, fmt.Errorf("failed to send request to AI API: %w", err)
}
defer resp.Body.Close()

// 5. Handle the response
if resp.StatusCode != http.StatusOK {
	bytes, _ := io.ReadAll(resp.Body)
	return nil, fmt.Errorf("AI API returned non-200 status: %d, body: %s", resp.StatusCode, string(bytes))
}

var result TranscriptionResponse
if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
	return nil, fmt.Errorf("failed to decode AI API response: %w", err)
}
return &result, nil
}