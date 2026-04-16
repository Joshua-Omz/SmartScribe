# SmartScribe: Bootstrap & Setup Guide

Welcome to **SmartScribe**! This guide covers the steps required to set up, build, and run both the frontend and backend services for development and testing.

## Prerequisites

Before you begin, ensure you have the following installed on your machine:
* **Docker** and **Docker Desktop/Compose**
* **Git**
* A modern web browser (Chrome, Firefox, or Edge) with microphone access enabled.

---

## Step 1: Environment Configuration

SmartScribe relies on external AI APIs for its transcription and note-structuring pipeline. You must provide these credentials for the application to function.

1. Navigate to the `backend/` directory.
2. You will see an `.env.example` file. Create a copy of it and name it `.env` (or just create a new `.env` file):

   ```env
   # backend/.env

   # Database settings (if currently used)
   DB_URL="postgres://postgres:secret@localhost:5433/smartscribe?sslmode=disable"

   # AI Credentials
   LLM_API_KEY="your_gemini_api_key_here"
   GOOGLE_STT_API_KEY="your_google_stt_api_key_here"
   ```

3. Replace the placeholder strings with your actual valid API keys.

*Note: The Docker build process will automatically copy this `.env` file into the container, so the Go server can authenticate properly.*

---

## Step 2: Build and Run with Docker

The entire application state (Frontend + Backend) is orchestrated via Docker Compose.

1. Open your terminal and navigate to the project root directory (where the `docker-compose.yml` file is located).
2. Run the following command to build the images and start the containers in detached mode:

   ```powershell
   docker-compose up --build -d
   ```

3. Wait a few moments for the build to complete. You should see `smartscribe-backend` and `smartscribe-frontend` listed as "Started".

---

## Step 3: Access the Application

Once Docker is running, the services are mapped to your local machine:

* **Frontend (User Interface):** `http://localhost` (or `http://localhost:80`)
* **Backend (API Gateway):** `http://localhost:8082`

1. Open your browser and go to **`http://localhost`**.
2. **Important:** Your browser may prompt you for Microphone permissions. You must **Allow** this for the recording feature to work.
3. Click "Start Recording", speak a medical note, and click "Stop". It will be transcribed and structured automatically!

---

## Troubleshooting & Tips

* **Frontend Not Updating?**
  If you make changes to `frontend/app.js` or `index.html`, rebuild the frontend container:
  ```powershell
  docker-compose up --build -d frontend
  ```
  Afterward, perform a **Hard Refresh** in your browser (`Ctrl + Shift + R` or `Cmd + Shift + R`) to bypass the browser cache.

* **Backend Crashing or API Keys Missing?**
  If the backend container continually stops, verify your `.env` file is properly formatted and placed in the `backend/` folder. Check the backend logs with:
  ```powershell
  docker logs smartscribe-backend
  ```

* **Port Conflicts:**
  If port `8082` or `80` is already in use by another application on your system, you can edit the `ports` bindings inside the `docker-compose.yml` file. If you change the backend port, be sure to update `TRANSCRIBE_ENDPOINT` inside `frontend/app.js` to match!

* **Mock Mode (Testing without APIs):**
  If you quickly want to test the UI without hitting the real AI APIs, open `frontend/app.js`, change `const MOCK_MODE = false;` to `true`, and rebuild the frontend container.