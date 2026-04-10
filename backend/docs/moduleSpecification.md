Backend Low-Level Specification (LLD)

Project: SmartScribe (Go API Gateway)

1. Core Architecture & Concurrency Model

Following standard Go community practices outlined in Effective Go, the server will rely entirely on the net/http standard library.

Concurrency: We will not manually manage thread pools. http.ListenAndServe will automatically spawn a new, lightweight Goroutine for every incoming audio stream from the frontend clients.

Memory Management: The server operates as a pass-through gateway. Audio data will be streamed via io.Reader/io.Writer interfaces where possible, rather than loading entire multi-megabyte files into RAM.

2. API Contract

Endpoint: POST /api/transcribe

Purpose: Receives raw audio from the Web/Mobile client, coordinates processing, and returns structured clinical data.

Content-Type: multipart/form-data

Payload Boundary: Field name strictly set to audio.

Response Content-Type: application/json

Security: Enforced 10MB limit using http.MaxBytesReader() to prevent OOM attacks.

3. Data Models (Struct Definitions)

To ensure correct JSON serialization mapping with the Java mock server and the frontend, we use explicit struct tags (referencing the encoding/json standard library).

// TranscriptionResponse represents the final JSON sent back to the frontend.

type TranscriptionResponse struct {
	Status      string `json:"status"`
	RawText     string `json:"raw_text"`
	Structured  SOAP   `json:"structured_data"`
	ErrorMsg    string `json:"error,omitempty"` // Omitted if empty
}

// SOAP represents the structured clinical format extracted by the NLP service.
type SOAP struct {
	Subjective string `json:"subjective"`
	Objective  string `json:"objective"`
	Assessment string `json:"assessment"`
	Plan       string `json:"plan"`
}


4. The Request Processing Pipeline (Step-by-Step)

Inside the handleTranscription HTTP handler, the execution flow must follow these precise steps:

Method Validation: if r.Method != http.MethodPost -> Return 405 Method Not Allowed.

Size Limiting: Wrap the request body: r.Body = http.MaxBytesReader(w, r.Body, 10<<20).

Multipart Parsing: Call err := r.ParseMultipartForm(10 << 20). If error, return 400 Bad Request.

File Extraction: Retrieve the file: file, header, err := r.FormFile("audio").

Resource Cleanup: Immediately declare defer file.Close() to ensure OS file descriptors are released.

Hand-off to OS Script (C/C++ Integration): * Save the multipart.File to a temporary directory using os.CreateTemp.

Use Go's os/exec standard package to invoke the OS Programmer's compression binary (e.g., ./compressor /tmp/audio.wav /tmp/audio.opus).

defer os.Remove() on both temp files to guarantee zero-storage compliance (NDPR/HIPAA).

External AI API Call: * Read the compressed .opus file.

Construct a new http.Request to the Whisper/Speech-to-Text API.

Use a custom http.Client{Timeout: 30 * time.Second} to ensure the request doesn't hang indefinitely.

Response Formatting: * Parse the AI's response into the TranscriptionResponse struct.

Stream the JSON to the client: json.NewEncoder(w).Encode(response).

5. Integration Points

Frontend Integration: The Go server will configure CORS (Cross-Origin Resource Sharing) headers if the frontend is hosted on a different port during local development: w.Header().Set("Access-Control-Allow-Origin", "*").

SmartMRS Mock Integration: In the background (or as a separate Goroutine), the validated SOAP data will be marshaled into JSON and sent via http.Post to the Java Security Analyst's mock endpoint (http://mock-mrs:8081/api/notes).

6. File Structure Convention

To keep the hackathon code clean, we will follow a simplified version of the Go community's standard layout:

/smartscribe
├── /backend
│   ├── main.go               # Server setup and router
│   ├── handlers.go           # HTTP handler functions (handleTranscription)
│   ├── ai_client.go          # Logic to communicate with Whisper API
│   ├── compression.go        # Wrapper for os/exec calling the C++ script
│   └── go.mod                # Module definition
