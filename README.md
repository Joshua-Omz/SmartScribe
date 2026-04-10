# SmartScribe 🩺

A hackathon healthcare voice-to-text application built with a **Go backend** and a **Vanilla JS frontend**.

## Project Structure

```
SmartScribe/
├── backend/
│   ├── main.go      # Go HTTP server
│   └── go.mod       # Go module file
└── frontend/
    ├── index.html   # Responsive UI
    └── app.js       # MediaRecorder + fetch client
```

## Running the Application

### Prerequisites

- [Go 1.21+](https://go.dev/dl/)
- A modern web browser (Chrome, Firefox, Edge)

### Start the Go Server

```bash
cd backend
go run main.go
```

The server starts on **http://localhost:8080** and serves the frontend automatically.

### Using the App

1. Open **http://localhost:8080** in your browser.
2. Click **▶ Start Recording** – grant microphone access when prompted.
3. Speak your clinical notes.
4. Click **⏹ Stop Recording** – the audio is uploaded to the backend.
5. The transcription result appears on the screen.

## API

| Method | Path              | Description                              |
|--------|-------------------|------------------------------------------|
| POST   | `/api/transcribe` | Upload audio (`multipart/form-data`, field `audio`), returns JSON transcription |

### Example Response

```json
{
  "status": "success",
  "text": "Patient has a mild fever."
}
```

## Development Notes

- The backend uses only the Go standard library (`net/http`).
- The frontend uses the [MediaRecorder API](https://developer.mozilla.org/en-US/docs/Web/API/MediaRecorder) – no external dependencies.
- Microphone access requires either `localhost` or an HTTPS origin.
