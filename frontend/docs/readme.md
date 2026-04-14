# SmartScribe Frontend Integration Guide

## Purpose
This is the frontend integration source of truth for SmartScribe.
It documents the current behavior in `frontend/app.js` and the current backend contract in `backend/handlers.go`.

## Run Locally
1. Start backend from `backend/`:

```bash
go run main.go
```

2. Open `http://localhost:8080` in your browser.
3. Grant microphone permission when prompted.

The frontend is served directly by the Go backend.

## API Contract
Endpoint:
- `POST /api/transcribe`

Request:
- `multipart/form-data`
- Field name must be `audio`
- Backend multipart parse limit is `32 MB` (`ParseMultipartForm(32 << 20)`)

Success response (current):

```json
{
  "status": "success",
  "text": "Patient has a mild fever."
}
```

Error responses (current):
- Non-200 errors are plain text from `http.Error` (for example: `method not allowed`, `failed to parse multipart form`, `audio field missing from request`).
- Frontend tries JSON parsing first and then falls back to generic failure messaging when no JSON error payload is available.

Status-code behavior matrix:

| Status code | Typical backend message | Frontend behavior |
| --- | --- | --- |
| `400` | `failed to parse multipart form` or `audio field missing from request` | Treated as transcription failure, then retry flow applies |
| `405` | `method not allowed` | Treated as transcription failure, then retry flow applies |
| `413` | plain-text payload too large response (if returned by middleware/server path) | Treated as transcription failure; user sees generic failure unless JSON payload is available |
| `500` | plain-text internal server error response | Treated as transcription failure, then retry flow applies |

## State Machine
Current app states:
- `READY`
- `RECORDING`
- `REVIEWING`

Main transitions:
- `READY -> RECORDING`: Start recording
- `RECORDING -> REVIEWING`: Stop recording or 5-minute auto-stop
- `REVIEWING -> READY`: Re-record or post-submit reset

Recording safety cutoff:
- `MAX_RECORDING_MS = 5 * 60 * 1000`

## Mock Mode
Current defaults in `frontend/app.js`:
- `MOCK_MODE = true`
- `TRANSCRIBE_ENDPOINT = "/api/transcribe"`
- `TRANSCRIBE_RETRIES = 2`
- `RETRY_DELAY_MS = 1500`
- `MOCK_DELAY_MS = 1800`
- `SUBMIT_DELAY_MS = 1100`
- `PENDING_KEY = "smartscribe_pending_note"`

When mock mode is on:
- No backend transcription request is sent.
- Fixed transcript is returned after 1800 ms.
- Console logs are prefixed with `[SmartScribe MOCK]`.

## Error Handling
Recording start errors:
- No `getUserMedia` or no `MediaRecorder` support shows an inline error banner.
- Permission denial shows a specific permission message.

Transcription errors:
- Automatic retries: up to 2 retries with 1500 ms delay.
- If retries fail, UI shows retry or re-record actions.

Submit errors:
- Empty note validation blocks submit (`Note is empty...`).
- Offline or submit failure stores the note in local storage and shows submit error banner.

## Offline Pending Notes
Pending key:
- `smartscribe_pending_note`

Stored payload:

```json
{
  "text": "...",
  "savedAt": "2026-04-14T10:30:00.000Z"
}
```

Behavior:
- On load, pending key triggers banner with View/Discard actions.
- View loads note into `REVIEWING`.
- Discard removes pending key.
- Successful submit removes pending key.
- Re-record/full reset also removes pending key.

## Quick Test Checklist
- Start app at `http://localhost:8080`.
- Verify `READY -> RECORDING -> REVIEWING` flow.
- Verify timer increments and 5-minute safety auto-stop.
- Verify mock transcription returns after ~1800 ms when `MOCK_MODE = true`.
- Verify retries trigger on transcription failure (2 retries, 1500 ms delay).
- Verify empty submit is blocked.
- Verify offline submit stores `smartscribe_pending_note`.
- Verify pending banner appears after reload and View/Discard works.
- Verify re-record clears pending key.

## Troubleshooting
- Microphone error: allow microphone permission in browser settings.
- No transcription in real mode: confirm backend is running and `MOCK_MODE = false`.
- Generic transcription failure: expected when backend returns plain-text non-200 errors.
- Pending note not clearing: complete successful submit, discard pending note, or re-record to trigger full reset.

## Known Gaps
- Backend non-200 errors are currently plain text, which can limit detailed user-facing error specificity.
- Frontend expects `status/text` on success but relies on fallback handling when non-JSON errors are returned.

## Maintenance
- Owner: Frontend team (SmartScribe)
- Last updated: 2026-04-14