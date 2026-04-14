# Backend API Testing Bootstrap

This guide is a quick-start reference for anyone (frontend developers, testers, or backend contributors) who needs to start up the SmartScribe Go backend locally and send API requests to it.

## 1. Prerequisites (Environment Setup)

Before starting the server, you must have your API keys configured properly. 
In the `backend` directory, ensure you have a `.env` file (you can copy `.env.example` if available) with valid keys:

```properties
DB_URL="postgres://postgres:secret@localhost:5433/smartscribe?sslmode=disable"
LLM_API_KEY="gsk_your_real_groq_api_key_here"
GOOGLE_STT_API_KEY="AIzaSy_your_real_google_api_key_here"
```
*(Note: We recently migrated from `VAX_API_KEY` to a generic `LLM_API_KEY`).*

## 2. Start the Server Local

Open a terminal right inside the `backend` directory and run:

```bash
go run .
```

You should see log output looking like:
`2026/04/14 12:00:00 SmartScribe Gateway starting on port 8080...`

## 3. Making an API Request

The primary endpoint ingestion requires a Multipart form containing an audio file.

*   **URL:** `http://localhost:8080/api/transcribe`
*   **Method:** `POST`
*   **Enc Type:** `multipart/form-data`
*   **Field Name:** `audio` *(This must be exactly "audio")*

---

### Option A: Testing via Postman (Recommended for manual testing)

1. Open Postman and create a new **POST** request.
2. Set the URL to: `http://localhost:8080/api/transcribe`
3. Go directly to the **Body** tab underneath the URL.
4. Select the **form-data** radio button.
5. In the `Key` column, type `audio`.
6. **Crucial Step:** Hover over the `audio` cell, and a hidden dropdown showing "Text" will appear on the right side of the cell. Click it and change it to **File**.
7. In the `Value` column, click **Select Files** and choose a `.webm` or `.mp3` audio file from your computer (there is a `smartscribe_test (1).webm` in the backend folder you can use).
8. Click **Send**!

---

### Option B: Testing via cURL

If you are on Linux, macOS, or using Git Bash/WSL on Windows, you can initiate a request from the terminal. Assuming you are in the `backend` directory where `smartscribe_test (1).webm` exists:

```bash
curl -X POST http://localhost:8080/api/transcribe \
  -F "audio=@\"smartscribe_test (1).webm\""
```

---

### Option C: Frontend Interaction (JavaScript)

If you're building the frontend, the fetch request looks like this using `FormData`:

```javascript
const formData = new FormData();
// 'audioBlob' is the audio data from the user's microphone MediaRecorder
formData.append("audio", audioBlob, "dictation.webm");

const response = await fetch("http://localhost:8080/api/transcribe", {
    method: "POST",
    body: formData // Note: Let the browser set the "Content-Type" header automatically!
});
const data = await response.json();
console.log(data);
```

## 4. Expected Response

Once the server finishes querying Google Speech-to-Text and the LLM API (which can take ~5 to 10 seconds), it will return an HTTP `200 OK` with JSON matching this structure:

```json
{
  "status": "success",
  "raw_text": "So the patient is a male with a slight fever...",
  "structured_data": {
    "subjective": "Patient reports a slight fever.",
    "objective": "Fever present.",
    "assessment": "Febrile illness.",
    "plan": "Monitor temperature."
  }
}
```

## 5. Running Internal Unit Tests

If you want to verify the backend logic is working without using Postman/cURL at all, we have an internal test configured (`handler_test.go`). 

Run this command inside the `backend` folder:
```bash
go test -count=1 -v ./...
```

*Note: The test uses the `smartscribe_test (1).webm` file automatically. If it fails with `401 Invalid API Key`, make sure your `LLM_API_KEY` in `.env` is valid!*