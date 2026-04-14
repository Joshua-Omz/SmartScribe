Final Backend Low-Level Specification (LLD)

Project: SmartMRS Voice-to-Text (Go API Gateway)

1. Architectural Role

The Go backend acts exclusively as a secure, high-concurrency API Gateway and AI Orchestrator. It does not perform heavy computation itself; it routes data between the frontend, local OS scripts, external AI APIs, and the Java storage server.

2. The Complete Execution Pipeline (POST /api/transcribe)

When a request hits the main HTTP handler, the execution must follow this exact synchronous flow:

Ingestion & Security Check:

Enforce http.MethodPost.

Enforce 10MB payload limit via http.MaxBytesReader().

Parse multipart/form-data and extract the "audio" file.

OS-Level Compression (Zero-Storage Compliance):

Save raw audio to /tmp/raw_audio_[id].webm.

Execute os/exec.Command("./compressor", "/tmp/raw_audio_[id].webm", "/tmp/compressed_[id].opus").

Wait for exit code 0.

defer os.Remove() on both temporary files.

AI Step 1: Medical Speech-to-Text (Google Cloud):

Read .opus file into a bytes.Buffer multipart payload.

Send to Google Medical Speech-to-Text API.

Extract the raw string: "Patient presents with..."

AI Step 2: Clinical Structuring (Vax-Llama Cloud API):

Wrap the raw string in a JSON payload ({"clinical_text": "...", "system_prompt": "..."}).

POST to https://api.helpmum.org/v1/ai/structure.

Decode the returned SOAP JSON.

System of Record Sync (Java Mock Server):

Construct the final clinical payload (Patient ID, Timestamp, SOAP Data).

POST to http://localhost:8081/api/patient/notes (Java Spring Boot Server) with Bearer token.

Ensure HTTP 201 Created is returned by Java.

Client Response:

Return HTTP 200 OK to the Frontend with the raw text and structured SOAP data.

3. Core Data Structures (Go Structs)

These define the exact JSON shapes expected across the system.

// Response to Frontend
type TranscriptionResponse struct {
	Status      string `json:"status"`
	RawText     string `json:"raw_text"`
	Structured  SOAP   `json:"structured_data"`
	Error       string `json:"error,omitempty"`
}

// SOAP Standard Format
type SOAP struct {
	Subjective string `json:"subjective"`
	Objective  string `json:"objective"`
	Assessment string `json:"assessment"`
	Plan       string `json:"plan"`
}

// Request to Java SmartMRS Mock
type SmartMRSSyncPayload struct {
	PatientID    string `json:"patient_id"`
	ProviderID   string `json:"provider_id"`
	Timestamp    string `json:"timestamp"`
	ClinicalData SOAP   `json:"clinical_data"`
}


4. Required Environment Variables

Do not hardcode these in the final binary. Use os.Getenv().

GOOGLE_STT_API_KEY: Credentials for Medical transcription.

LLAMA_API_KEY: Credentials for HelpMum Vax-Llama structuring.

JAVA_MOCK_TOKEN: Static bearer token to authenticate with Timi's server.

PORT: Server port (default 8080).

5. Final Directory Structure

/smartscribe-backend
├── main.go               # Router setup, MaxBytesReader, and server init
├── handlers.go           # The pipeline execution logic (handleTranscription)
├── compression.go        # Wrapper for os/exec calling the OS programmer's binary
├── ai_client.go          # The two-step logic for Google STT and Vax-Llama NLP
├── java_sync.go          # HTTP client logic to push data to the Spring Boot server
└── go.mod
