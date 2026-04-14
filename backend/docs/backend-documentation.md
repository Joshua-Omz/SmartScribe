# SmartScribe Go Backend Documentation

## Overview
The SmartScribe backend is built in Go and serves as the central orchestration layer. It receives audio dictations from the frontend, processes them through AI services (Google Speech-to-Text and VaxLlama), and forwards the structured SOAP notes to the mock SmartMRS system (Java Spring Boot).

## Architecture
- **Language**: Go
- **Port**: 8080
- **Main Endpoints**:
  - `POST /api/transcribe`: Receives audio file, processes it, and returns structured SOAP notes.

## Run Locally
1. Ensure you have Go installed.
2. Create a `.env` file based on `.env.example`.
3. Run the server: `go run .`

## Docker
The backend is containerized using a multi-stage Dockerfile and orchestrated via Docker Compose along with the rest of the stack.

## Environment Variables
- `GOOGLE_STT_API_KEY`: API key for Google Speech-to-Text.
- `VAX_API_KEY`: API key for VaxLlama.

## Integrations
- **Frontend**: Expects `multipart/form-data` with an `audio` field.
- **Mock SmartMRS (Java)**: Sends structured JSON data to `http://mock-mrs:8081/api/patient/notes` (or `http://localhost:8081` locally).

See the [Cross-Domain Integration Guide](./integration-guide.md) for more details.
