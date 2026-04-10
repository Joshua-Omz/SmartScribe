'use strict';

const startBtn = document.getElementById('startBtn');
const stopBtn = document.getElementById('stopBtn');
const statusEl = document.getElementById('status');
const resultEl = document.getElementById('result');
const transcriptionEl = document.getElementById('transcriptionText');

let mediaRecorder = null;
let audioChunks = [];

function setStatus(message, isRecording = false) {
  statusEl.textContent = message;
  statusEl.className = isRecording ? 'recording' : '';
}

async function startRecording() {
  try {
    const stream = await navigator.mediaDevices.getUserMedia({ audio: true });

    audioChunks = [];
    mediaRecorder = new MediaRecorder(stream);

    mediaRecorder.addEventListener('dataavailable', (event) => {
      if (event.data.size > 0) {
        audioChunks.push(event.data);
      }
    });

    mediaRecorder.addEventListener('stop', () => {
      // Stop all tracks so the microphone indicator clears
      stream.getTracks().forEach((track) => track.stop());
      sendAudio();
    });

    mediaRecorder.start();

    startBtn.disabled = true;
    stopBtn.disabled = false;
    resultEl.classList.remove('visible');
    setStatus('🔴 Recording…', true);
  } catch (err) {
    setStatus(`Microphone access denied: ${err.message}`);
    console.error('getUserMedia error:', err);
  }
}

function stopRecording() {
  if (mediaRecorder && mediaRecorder.state !== 'inactive') {
    mediaRecorder.stop();
    startBtn.disabled = false;
    stopBtn.disabled = true;
    setStatus('Processing…');
  }
}

async function sendAudio() {
  const audioBlob = new Blob(audioChunks, { type: 'audio/webm' });
  const formData = new FormData();
  formData.append('audio', audioBlob, 'recording.webm');

  try {
    setStatus('Uploading to server…');
    const response = await fetch('/api/transcribe', {
      method: 'POST',
      body: formData,
    });

    if (!response.ok) {
      throw new Error(`Server returned ${response.status}`);
    }

    const data = await response.json();
    transcriptionEl.textContent = data.text || JSON.stringify(data, null, 2);
    resultEl.classList.add('visible');
    setStatus('Transcription complete.');
  } catch (err) {
    setStatus(`Error: ${err.message}`);
    console.error('Transcription error:', err);
  }
}

startBtn.addEventListener('click', startRecording);
stopBtn.addEventListener('click', stopRecording);
