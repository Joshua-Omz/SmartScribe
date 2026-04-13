// Config
const MOCK_MODE = true;
const TRANSCRIBE_ENDPOINT = "/api/transcribe";
const MAX_RECORDING_MS = 5 * 60 * 1000;
const TRANSCRIBE_RETRIES = 2;
const RETRY_DELAY_MS = 1500;
const MOCK_DELAY_MS = 1800;
const SUBMIT_DELAY_MS = 1100;
const PENDING_KEY = "smartscribe_pending_note";

// State
const AppState = {
  READY: "READY",
  RECORDING: "RECORDING",
  REVIEWING: "REVIEWING",
};

let currentState = AppState.READY;
let mediaRecorder = null;
let mediaStream = null;
let audioChunks = [];
let audioBlob = null;
let recordingTimer = null;
let recordingSeconds = 0;
let safetyTimer = null;
let currentTranscribedText = "";
let submitInFlight = false;

// Elements
const stateReady = document.getElementById("state-ready");
const stateRecording = document.getElementById("state-recording");
const stateReview = document.getElementById("state-review");

const startBtn = document.getElementById("start-btn");
const stopBtn = document.getElementById("stop-btn");
const recordingTimerEl = document.getElementById("recording-timer");

const transcribeLoading = document.getElementById("transcribe-loading");
const transcribeError = document.getElementById("transcribe-error");
const transcribeErrorText = document.getElementById("transcribe-error-text");
const retryTranscribeBtn = document.getElementById("retry-transcribe-btn");
const rerecordFromErrorBtn = document.getElementById("rerecord-from-error-btn");

const editorWrap = document.getElementById("editor-wrap");
const noteTextarea = document.getElementById("note-textarea");
const wordCountEl = document.getElementById("word-count");
const charCountEl = document.getElementById("char-count");

const rerecordBtn = document.getElementById("rerecord-btn");
const submitBtn = document.getElementById("submit-btn");
const submitBtnContent = document.getElementById("submit-btn-content");

const submitError = document.getElementById("submit-error");
const submitSuccess = document.getElementById("submit-success");
const successTime = document.getElementById("success-time");

const pendingBanner = document.getElementById("pending-note-banner");
const viewPendingBtn = document.getElementById("view-pending-btn");
const discardPendingBtn = document.getElementById("discard-pending-btn");

const globalBanner = document.getElementById("global-banner");
const liveRegion = document.getElementById("live-region");
const statusText = document.getElementById("status-text");
const statusDot = document.querySelector(".status-dot");

function mockLog(message) {
  if (MOCK_MODE) {
    console.log("[SmartScribe MOCK] " + message);
  }
}

function transition(newState, payload = {}) {
  currentState = newState;
  render(newState, payload);
}

function render(state, payload = {}) {
  stateReady.classList.toggle("is-hidden", state !== AppState.READY);
  stateRecording.classList.toggle("is-hidden", state !== AppState.RECORDING);
  stateReview.classList.toggle("is-hidden", state !== AppState.REVIEWING);

  if (state === AppState.READY) {
    announce("Ready. Tap start to record a clinical note.");
    setStatusBarDefault();
  }

  if (state === AppState.RECORDING) {
    announce("Recording in progress.");
    setFooterStatus("Speak your clinical note clearly", MOCK_MODE);
  }

  if (state === AppState.REVIEWING) {
    announce(payload.reviewAnnouncement || "Review your transcribed note.");
  }
}

async function startRecording() {
  dismissBanner();
  hideSubmitMessages();
  hideTranscribeError();

  if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
    showBanner("Audio recording is not supported in this browser. Use Chrome or Firefox.", "error");
    announce("Recording unavailable. Browser does not support microphone access.");
    return;
  }

  if (!window.MediaRecorder) {
    showBanner("MediaRecorder is not available in this browser. Please use Chrome or Firefox.", "error");
    announce("Recording unavailable. MediaRecorder is not supported.");
    return;
  }

  try {
    mediaStream = await navigator.mediaDevices.getUserMedia({ audio: true });

    const mimeType = pickMimeType();
    mediaRecorder = mimeType ? new MediaRecorder(mediaStream, { mimeType }) : new MediaRecorder(mediaStream);

    audioChunks = [];
    audioBlob = null;
    recordingSeconds = 0;
    recordingTimerEl.textContent = formatTime(recordingSeconds);

    mediaRecorder.addEventListener("dataavailable", (event) => {
      if (event.data && event.data.size > 0) {
        audioChunks.push(event.data);
      }
    });

    mediaRecorder.addEventListener("stop", handleRecordingStopped);

    mediaRecorder.start();
    startTimer();
    startSafetyTimer();
    transition(AppState.RECORDING);
    mockLog("recording started");
  } catch (error) {
    const denied = error && (error.name === "NotAllowedError" || error.name === "PermissionDeniedError");
    const message = denied
      ? "Microphone permission was denied. Please allow microphone access and try again."
      : "Could not start recording. Check microphone availability and try again.";

    showBanner(message, "error");
    announce("Could not start recording.");
  }
}

function stopRecording() {
  if (!mediaRecorder || mediaRecorder.state === "inactive") {
    return;
  }

  mediaRecorder.stop();
  stopTimer();
  clearSafetyTimer();

  if (mediaStream) {
    mediaStream.getTracks().forEach((track) => track.stop());
    mediaStream = null;
  }

  mockLog("recording stopped");
}

async function handleRecordingStopped() {
  assembleBlob();
  transition(AppState.REVIEWING, { reviewAnnouncement: "Transcribing your note." });

  showLoading();
  hideTranscribeError();
  hideEditor();

  try {
    const transcript = await transcribeWithRetry(audioBlob);
    currentTranscribedText = transcript;
    noteTextarea.value = transcript;
    updateCounts();
    hideLoading();
    showEditor();
    submitBtn.disabled = false;
    submitBtnContent.textContent = "Submit to SmartMRS";
    setStatusBarDefault();
    announce("Transcription ready for review.");
  } catch (error) {
    hideLoading();
    showTranscribeError(error.message || "Transcription failed.");
    announce("Transcription failed. Retry or re-record.");
  }
}

function assembleBlob() {
  const inferredType = audioChunks[0] ? audioChunks[0].type : "audio/webm;codecs=opus";
  audioBlob = new Blob(audioChunks, { type: inferredType });
}

async function transcribeWithRetry(blob) {
  let attempt = 0;
  let lastError = null;

  while (attempt <= TRANSCRIBE_RETRIES) {
    try {
      return await transcribeAudio(blob);
    } catch (error) {
      lastError = error;
      if (attempt >= TRANSCRIBE_RETRIES) {
        break;
      }
      await delay(RETRY_DELAY_MS);
    }
    attempt += 1;
  }

  throw lastError || new Error("Transcription failed.");
}

async function transcribeAudio(blob) {
  if (MOCK_MODE) {
    mockLog("skipping backend fetch and using mock transcript");
    return mockTranscribe();
  }

  const formData = new FormData();
  formData.append("audio", blob, "clinical-note.webm");

  const response = await fetch(TRANSCRIBE_ENDPOINT, {
    method: "POST",
    body: formData,
  });

  const data = await response.json().catch(() => ({}));

  if (!response.ok || data.status === "error") {
    throw new Error(data.message || "Transcription failed.");
  }

  if (!data.text || typeof data.text !== "string") {
    throw new Error("Transcription response did not include text.");
  }

  return data.text;
}

async function mockTranscribe() {
  await delay(MOCK_DELAY_MS);
  return "Patient presents with fever of 38.5 degrees Celsius, mild cough persisting for 3 days, no chest pain or difficulty breathing. Blood pressure 120 over 80 mmHg. No known drug allergies. Recommending paracetamol 500mg twice daily, increased fluid intake, and rest. Follow-up in 5 days if symptoms persist.";
}

function startTimer() {
  stopTimer();
  recordingTimer = window.setInterval(() => {
    recordingSeconds += 1;
    recordingTimerEl.textContent = formatTime(recordingSeconds);
  }, 1000);
}

function stopTimer() {
  if (recordingTimer) {
    window.clearInterval(recordingTimer);
    recordingTimer = null;
  }
}

function startSafetyTimer() {
  clearSafetyTimer();
  safetyTimer = window.setTimeout(() => {
    announce("Maximum recording length reached. Recording stopped.");
    stopRecording();
  }, MAX_RECORDING_MS);
}

function clearSafetyTimer() {
  if (safetyTimer) {
    window.clearTimeout(safetyTimer);
    safetyTimer = null;
  }
}

function formatTime(seconds) {
  const mins = Math.floor(seconds / 60);
  const secs = seconds % 60;
  return String(mins).padStart(2, "0") + ":" + String(secs).padStart(2, "0");
}

function wordCount(text) {
  const trimmed = text.trim();
  if (!trimmed) {
    return 0;
  }
  return trimmed.split(/\s+/).length;
}

function updateCounts() {
  const text = noteTextarea.value || "";
  const words = wordCount(text);
  const chars = text.length;

  wordCountEl.textContent = words + (words === 1 ? " word" : " words");
  charCountEl.textContent = chars + (chars === 1 ? " character" : " characters");
}

async function submitNote() {
  if (submitInFlight) {
    return;
  }

  const text = noteTextarea.value.trim();
  if (!text) {
    showBanner("Note is empty. Please enter clinical notes before submitting.", "error");
    return;
  }

  submitInFlight = true;
  hideSubmitMessages();
  dismissBanner();

  submitBtn.disabled = true;
  submitBtnContent.innerHTML = "<span class=\"btn-spinner\" aria-hidden=\"true\"></span> Submitting...";

  try {
    await delay(SUBMIT_DELAY_MS);

    if (!navigator.onLine) {
      throw new Error("offline");
    }

    localStorage.removeItem(PENDING_KEY);
    submitBtnContent.innerHTML = "<span aria-hidden=\"true\">✓</span> Saved to SmartMRS";
    successTime.textContent = "Saved at " + formatTimestamp(new Date());
    submitSuccess.classList.remove("is-hidden");
    announce("Note saved to SmartMRS.");

    setTimeout(() => {
      fullReset();
    }, 1200);
  } catch (_error) {
    const payload = {
      text,
      savedAt: new Date().toISOString(),
    };
    localStorage.setItem(PENDING_KEY, JSON.stringify(payload));
    submitError.classList.remove("is-hidden");
    submitBtn.disabled = false;
    submitBtnContent.textContent = "Submit to SmartMRS";
    announce("Could not reach SmartMRS. Note saved locally.");
  } finally {
    submitInFlight = false;
  }
}

function showBanner(message, type) {
  globalBanner.textContent = message;
  globalBanner.className = "inline-banner";
  if (type === "error") {
    globalBanner.classList.add("error");
  } else if (type === "success") {
    globalBanner.classList.add("success");
  } else {
    globalBanner.classList.add("info");
  }
  globalBanner.classList.remove("is-hidden");
}

function dismissBanner() {
  globalBanner.classList.add("is-hidden");
  globalBanner.textContent = "";
}

function showLoading() {
  transcribeLoading.classList.remove("is-hidden");
}

function hideLoading() {
  transcribeLoading.classList.add("is-hidden");
}

function showEditor() {
  editorWrap.classList.remove("is-hidden");
}

function hideEditor() {
  editorWrap.classList.add("is-hidden");
}

function showTranscribeError(message) {
  transcribeErrorText.textContent = "Transcription failed. " + message;
  transcribeError.classList.remove("is-hidden");
}

function hideTranscribeError() {
  transcribeError.classList.add("is-hidden");
}

function hideSubmitMessages() {
  submitError.classList.add("is-hidden");
  submitSuccess.classList.add("is-hidden");
}

function setFooterStatus(text, amber) {
  statusText.textContent = text;
  statusDot.classList.toggle("is-amber", Boolean(amber));
}

function setStatusBarDefault() {
  if (MOCK_MODE) {
    setFooterStatus("Mock mode - backend not required", true);
  } else {
    setFooterStatus("Connected to SmartMRS", false);
  }
}

function announce(message) {
  liveRegion.textContent = message;
}

function formatTimestamp(date) {
  const options = {
    year: "numeric",
    month: "short",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  };
  return date.toLocaleString(undefined, options);
}

function pickMimeType() {
  const options = [
    "audio/webm;codecs=opus",
    "audio/webm",
    "audio/ogg;codecs=opus",
  ];

  for (const value of options) {
    if (MediaRecorder.isTypeSupported(value)) {
      return value;
    }
  }

  return "";
}

function delay(ms) {
  return new Promise((resolve) => {
    window.setTimeout(resolve, ms);
  });
}

function fullReset() {
  stopTimer();
  clearSafetyTimer();

  if (mediaRecorder && mediaRecorder.state !== "inactive") {
    mediaRecorder.stop();
  }

  if (mediaStream) {
    mediaStream.getTracks().forEach((track) => track.stop());
    mediaStream = null;
  }

  recordingSeconds = 0;
  recordingTimerEl.textContent = formatTime(0);
  audioChunks = [];
  audioBlob = null;
  currentTranscribedText = "";

  hideLoading();
  hideTranscribeError();
  hideEditor();
  hideSubmitMessages();
  dismissBanner();

  noteTextarea.value = "";
  updateCounts();
  submitBtn.disabled = false;
  submitBtnContent.textContent = "Submit to SmartMRS";

  localStorage.removeItem(PENDING_KEY);

  transition(AppState.READY);
  setStatusBarDefault();
  mockLog("state reset to READY");
}

function checkPendingNote() {
  const raw = localStorage.getItem(PENDING_KEY);
  if (!raw) {
    pendingBanner.classList.add("is-hidden");
    return;
  }

  pendingBanner.classList.remove("is-hidden");
}

function viewPendingNote() {
  const raw = localStorage.getItem(PENDING_KEY);
  if (!raw) {
    pendingBanner.classList.add("is-hidden");
    return;
  }

  let parsed;
  try {
    parsed = JSON.parse(raw);
  } catch (_error) {
    localStorage.removeItem(PENDING_KEY);
    pendingBanner.classList.add("is-hidden");
    return;
  }

  transition(AppState.REVIEWING, { reviewAnnouncement: "Loaded pending note for review." });
  hideLoading();
  hideTranscribeError();
  showEditor();
  noteTextarea.value = parsed.text || "";
  updateCounts();
  submitBtn.disabled = false;
  submitBtnContent.textContent = "Submit to SmartMRS";
  pendingBanner.classList.add("is-hidden");
}

function discardPendingNote() {
  localStorage.removeItem(PENDING_KEY);
  pendingBanner.classList.add("is-hidden");
  showBanner("Unsaved note discarded.", "info");
}

startBtn.addEventListener("click", startRecording);
stopBtn.addEventListener("click", stopRecording);
retryTranscribeBtn.addEventListener("click", async () => {
  if (!audioBlob) {
    showTranscribeError("No recording available. Please re-record.");
    return;
  }

  showLoading();
  hideTranscribeError();

  try {
    const transcript = await transcribeWithRetry(audioBlob);
    currentTranscribedText = transcript;
    noteTextarea.value = transcript;
    updateCounts();
    hideLoading();
    showEditor();
    announce("Transcription ready for review.");
  } catch (error) {
    hideLoading();
    showTranscribeError(error.message || "Transcription failed.");
  }
});

rerecordFromErrorBtn.addEventListener("click", fullReset);
rerecordBtn.addEventListener("click", fullReset);
submitBtn.addEventListener("click", submitNote);
noteTextarea.addEventListener("input", updateCounts);
viewPendingBtn.addEventListener("click", viewPendingNote);
discardPendingBtn.addEventListener("click", discardPendingNote);

document.addEventListener("DOMContentLoaded", () => {
  transition(AppState.READY);
  setStatusBarDefault();
  checkPendingNote();
  mockLog("application initialized");
});
