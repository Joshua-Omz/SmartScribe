/**
 * SmartScribe — app.js
 *
 * Uses the HTML5 MediaRecorder API (https://developer.mozilla.org/en-US/docs/Web/API/MediaRecorder)
 * to capture microphone audio, then POSTs the recorded blob to the Go backend
 * at POST /api/transcribe and displays the structured JSON response.
 */

const startBtn          = document.getElementById("startBtn");
const stopBtn           = document.getElementById("stopBtn");
const statusEl          = document.getElementById("status");
const transcriptionText = document.getElementById("transcriptionText");

let mediaRecorder = null;
let audioChunks   = [];

/**
 * Request microphone access and initialise a MediaRecorder session.
 */
async function startRecording() {
  try {
    const stream = await navigator.mediaDevices.getUserMedia({ audio: true });

    mediaRecorder  = new MediaRecorder(stream);
    audioChunks    = [];

    mediaRecorder.addEventListener("dataavailable", (event) => {
      if (event.data.size > 0) {
        audioChunks.push(event.data);
      }
    });

    mediaRecorder.addEventListener("stop", handleRecordingStop);

    mediaRecorder.start();

    startBtn.disabled = true;
    stopBtn.disabled  = false;
    statusEl.textContent = "🔴 Recording…";
    statusEl.className   = "status recording";
  } catch (err) {
    statusEl.textContent = `Microphone access denied: ${err.message}`;
    statusEl.className   = "status";
    console.error("getUserMedia error:", err);
  }
}

/**
 * Stop the active MediaRecorder. The "stop" event triggers handleRecordingStop.
 */
function stopRecording() {
  if (!mediaRecorder || mediaRecorder.state === "inactive") return;

  mediaRecorder.stop();

  // Stop all tracks so the browser releases the microphone indicator.
  mediaRecorder.stream.getTracks().forEach((track) => track.stop());

  startBtn.disabled = false;
  stopBtn.disabled  = true;
  statusEl.textContent = "Processing…";
  statusEl.className   = "status";
}

/**
 * Called when the MediaRecorder fires its "stop" event.
 * Packages the collected audio chunks as a Blob and sends them to the backend.
 */
async function handleRecordingStop() {
  const audioBlob = new Blob(audioChunks, { type: "audio/webm" });

  const formData = new FormData();
  formData.append("audio", audioBlob, "recording.webm");

  try {
    const response = await fetch("/api/transcribe", {
      method: "POST",
      body: formData,
    });

    if (!response.ok) {
      throw new Error(`Server returned ${response.status} ${response.statusText}`);
    }

    const data = await response.json();

    transcriptionText.textContent = data.text ?? "(no text field in response)";
    statusEl.textContent = `✅ Transcription received (status: ${data.status})`;
  } catch (err) {
    statusEl.textContent = `Error: ${err.message}`;
    transcriptionText.textContent = "Transcription failed. See console for details.";
    console.error("Transcription request failed:", err);
  }
}

startBtn.addEventListener("click", startRecording);
stopBtn.addEventListener("click", stopRecording);
