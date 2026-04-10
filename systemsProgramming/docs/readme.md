Audio Compression Integration Contract (Source of Truth)

Project: SmartScribe

This document defines the strict integration boundary between the Go Backend API and the OS-level audio compression script (C/C++/Rust).

1. Integration Strategy

According to standard system programming practices (and the official Go os/exec documentation), the integration will happen via the command line interface (CLI). The Go backend will execute the compiled binary, pass file paths as arguments, and read exit codes to determine success.

2. The Execution Contract

The Command Structure

The Go server will invoke the compiled binary exactly like this:

./compressor <input_file_path> <output_file_path>


<input_file_path>: An absolute path to the raw audio file saved temporarily by Go (e.g., /tmp/audio_in_1234.webm).

<output_file_path>: An absolute path where the Go server expects the script to save the compressed audio (e.g., /tmp/audio_out_1234.opus).

Expected Behavior

Read: The script must open and read the file at <input_file_path>.

Compress: Compress the audio data. (Using libraries like FFmpeg's libavcodec is highly recommended by the C++ community for this).

Write: Write the compressed output directly to the <output_file_path>.

Exit: The script must terminate and return standard OS exit codes.

3. Exit Codes & Error Handling

The Go server relies entirely on the exit code of the script to determine the next step.

Exit Code 0 (Success): * Meaning: The compression was successful, and the file at <output_file_path> is ready to be read by Go.

Exit Code 1 (or any non-zero): * Meaning: The compression failed.

Action: The script should output a concise error message to stderr (Standard Error) before exiting. The Go server will capture this stderr output and log it for debugging.

4. Implementation Notes for the OS Dev

Speed over Size: For this hackathon, configure the compression algorithm to prioritize speed over achieving the absolute smallest file size. The goal is to reduce the payload fast enough that the user doesn't notice the delay.

No Network Calls: This script must be entirely self-contained. It should not make any network calls or attempt to upload the file. Its sole responsibility is local file-to-file compression.

Memory Management: Ensure your script cleans up its own memory allocations before exiting to prevent memory leaks on the host server. The Go server will handle deleting the actual temporary files from the disk.