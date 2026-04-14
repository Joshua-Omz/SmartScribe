# SmartScribe Cross-Domain Integration Guide

This guide describes how the Go Backend, Web Frontend, and Mock SmartMRS (Java) interact within the Docker Compose environment.

## Architecture Overview

The SmartScribe platform consists of three main containerized services:
1. **Frontend (Nginx)**: Serves the static web application.
2. **Backend (Go)**: Orchestrates audio transcription and AI structuring.
3. **Mock SmartMRS (Java)**: Simulates an Electronic Medical Record (EMR) system.

## Communication Flow

1. **Frontend -> Backend**:
   - The user records audio in the browser.
   - The frontend sends a `POST` request to `http://localhost:8080/api/transcribe` with the audio file (field: `audio`, type: `multipart/form-data`).

2. **Backend Processing**:
   - The Go backend receives the audio.
   - It sends the audio to Google Speech-to-Text for transcription.
   - It sends the transcription to VaxLlama for SOAP structuring.

3. **Backend -> Mock SmartMRS (Java)**:
   - *Note: This integration is planned per the requirements but must be implemented in the Go handlers.*
   - The Go backend should send a `POST` request to `http://mock-mrs:8081/api/patient/notes` (internal Docker network URL).
   - Payload format: JSON containing `patient_id`, `provider_id`, `timestamp`, and the structured `clinical_data` (SOAP).

## Docker Network Configuration

All services run on a shared Docker bridge network (`smartscribe_network`).

- **Frontend URL**: `http://localhost:80` (or whichever port is mapped).
- **Backend URL (from Host)**: `http://localhost:8080`
- **Backend URL (from other containers)**: `http://backend:8080`
- **Mock SmartMRS URL (from Host)**: `http://localhost:8081`
- **Mock SmartMRS URL (from Backend container)**: `http://mock-mrs:8081`

When testing locally without Docker, use `localhost` for all cross-service communication. When running via Docker Compose, use the service names (`backend`, `mock-mrs`) for inter-container communication.
