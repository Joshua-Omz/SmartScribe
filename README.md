SmartScribe 🎙️⚕️

A zero-friction, voice-to-text clinical data entry pipeline designed for HelpMum's SmartMRS ecosystem.

📌 Overview

Healthcare workers operating in HelpMum's network face the administrative burden of manual data entry, taking time away from patient care. SmartScribe solves this by capturing spoken clinical notes, transcribing them via advanced NLP, and structuring the data directly into the SmartMRS platform (e.g., SOAP notes, vitals).

Built for speed, security, and low-bandwidth environments, SmartScribe serves as a seamless bridge between a clinician's voice and the patient's electronic health record.

🚀 Features

Cross-Platform Audio Capture: Browser-based dictation using standard web APIs, requiring no heavy app installations.

Concurrent Audio Processing: High-speed backend ingestion designed to handle multiple streams simultaneously without bottlenecking.

Low-Bandwidth Optimization: OS-level audio compression before API transmission to accommodate rural or low-connectivity clinic settings.

Zero-Storage Compliance: Processes audio strictly in-memory. No voice data is saved on the server, ensuring compliance with health data privacy regulations (NDPR/HIPAA).

🛠️ Architecture & Tech Stack

Our architecture leverages a highly specialized, polyglot stack to maximize performance, security, and hardware efficiency:

Frontend (Vanilla JS/HTML5): Utilizes the MDN standard MediaRecorder API for lightweight, cross-browser microphone access without heavy frontend frameworks.

Backend Gateway (Go/Golang): Built with Go's standard net/http library. The Go community's best practices for Goroutines are used to handle concurrent audio stream ingestion and external AI API calls securely and efficiently.

Systems & Optimization (C/C++): Low-level audio compression (referencing FFmpeg community standards) reduces payload sizes, saving crucial bandwidth for clinics before data hits the cloud.

Security & Integration (Java): A secure, mock SmartMRS REST API built referencing Spring Boot documentation, ensuring our payload structures map perfectly to HelpMum's database while adhering to OWASP Healthcare standards.

⚙️ Quick Start (Development)

**Prerequisites:** [Go 1.21+](https://go.dev/dl/) installed and available in your `PATH`.

Clone the repository:

```bash
git clone https://github.com/Joshua-Omz/SmartScribe.git
cd SmartScribe
```

Run the Go Backend:

```bash
cd backend
go run main.go
```

The server starts on **http://localhost:8080** and automatically serves the frontend from the `frontend/` directory.

Test the Client:

1. Open **http://localhost:8080** in your browser.
2. Grant microphone permissions when prompted.
3. Click **▶ Start Recording** and speak a clinical note.
4. Click **⏹ Stop Recording** — the audio is sent to `POST /api/transcribe`.
5. The transcription result appears on screen.

Repository Structure:

```
SmartScribe/
├── backend/
│   ├── go.mod      # Go module definition
│   └── main.go     # HTTP server + /api/transcribe endpoint
├── frontend/
│   ├── index.html  # Responsive UI with record controls
│   └── app.js      # MediaRecorder API + fetch to backend
└── README.md
```

## Frontend Integration Guide

For frontend setup, behavior, and integration details, see frontend/docs/readme.md.
The frontend is served by the Go backend at http://localhost:8080.

API Reference:

**POST /api/transcribe**

Accepts a `multipart/form-data` request with the following field:

| Field | Type        | Description                              |
|-------|-------------|------------------------------------------|
| audio | binary file | Audio recording (WebM/Opus from browser) |

Returns JSON:

```json
{
  "status": "success",
  "text": "Patient has a mild fever."
}
```

Built with ❤️ for the HelpMum Hackathon.
