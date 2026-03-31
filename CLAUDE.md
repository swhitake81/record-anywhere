# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

record-anywhere is a macOS CLI tool that records system audio using BlackHole (virtual audio driver) + PortAudio for capture and ffmpeg for MP3 encoding. Built with Go and Cobra.

## Build & Run

```bash
go build -o record-anywhere .        # Build the binary
go run . <command>                    # Run without building
```

Requires system dependencies installed via `record-anywhere setup`: BlackHole 2ch (cask), PortAudio, ffmpeg. All installed via Homebrew.

PortAudio must be available at build time (uses cgo via `github.com/gordonklaus/portaudio`).

## Architecture

**Daemon-style recording**: `start` spawns a detached child process that runs the hidden `_record` command. The parent exits immediately. Communication between `start`/`stop`/`status` and the recorder process happens via PID and status files stored in `~/.config/record-anywhere/`.

- `cmd/` — Cobra commands. `record_internal.go` contains the hidden `_record` command that does the actual recording in the background process.
- `internal/audio/` — Audio capture (PortAudio → float32 → 16-bit PCM → WAV) and MP3 conversion (shells out to ffmpeg).
- `internal/process/` — Process lifecycle: PID file management, status file (JSON) for inter-process state, signal handling (SIGTERM/SIGINT), and daemon spawning via `os/exec` with `Setsid`.
- `internal/config/` — JSON config at `~/.config/record-anywhere/config.json`. Keys: `output_dir`, `default_format` (mp3/wav), `default_duration`.
- `internal/setup/` — Dependency checking and Homebrew-based installation.

**Recording flow**: `start` → spawns `_record` → writes PID file → opens PortAudio stream on BlackHole 2ch → captures PCM to WAV → on SIGTERM: stops stream, optionally converts to MP3 via ffmpeg, writes final status, removes PID file.

**WAV writer** flushes the header every 10 seconds for crash safety, then finalizes on close.

## No Tests

There are currently no tests in this project.
