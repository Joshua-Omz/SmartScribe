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

Clone the repository:

git clone [https://github.com/your-org/smartscribe.git](https://github.com/your-org/smartscribe.git)
cd smartscribe


Run the Go Backend:
Ensure you have Go installed, then start the server:

cd backend
go run main.go


The server will start on http://localhost:8080 and serve the frontend statically.

Test the Client:
Open your browser and navigate to http://localhost:8080. Grant microphone permissions, click "Start Recording," and dictate a clinical note.

Built with ❤️ for the HelpMum Hackathon.
