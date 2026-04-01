# record-anywhere

A macOS CLI tool that records system audio using [BlackHole](https://github.com/ExistentialAudio/BlackHole) (virtual audio driver) + PortAudio for capture and ffmpeg for MP3 encoding.

## Prerequisites

- macOS
- [Go](https://go.dev/dl/) 1.25+
- [Homebrew](https://brew.sh/)

## Installation

```bash
git clone https://github.com/swhitake81/record-anywhere.git
cd record-anywhere
sudo make setup
```

This installs the required system dependencies (BlackHole 2ch, PortAudio, ffmpeg) via Homebrew, builds the binary, and installs it to `/usr/local/bin/`.

To uninstall:

```bash
sudo make uninstall
```

## Initial Configuration

Before your first recording, set up your output directory:

```bash
record-anywhere config init
```

This will prompt you to choose where recordings are saved.

## Usage

### Start a recording

```bash
record-anywhere start
```

This starts recording system audio in the background. The audio source is BlackHole 2ch, so make sure your system audio is routed through it (e.g., via a Multi-Output Device in Audio MIDI Setup).

**Options:**

| Flag         | Description                              | Default              |
|--------------|------------------------------------------|----------------------|
| `--name`     | Recording file name (without extension)  | Timestamp            |
| `--format`   | Output format: `mp3` or `wav`            | From config          |
| `--duration`  | Recording duration (e.g., `30m`, `1h`)  | From config          |

**Examples:**

```bash
record-anywhere start --name meeting --format mp3
record-anywhere start --duration 1h
```

### Stop a recording

```bash
record-anywhere stop
```

Stops the current recording. If the format is MP3, the WAV file is automatically converted.

### Check recording status

```bash
record-anywhere status
```

### Configuration

View or change settings stored in `~/.config/record-anywhere/config.json`:

```bash
record-anywhere config get output_dir
record-anywhere config set output_dir ~/Music/recordings
record-anywhere config set default_format mp3
record-anywhere config set default_duration 1h
```

**Config keys:**

| Key                | Description                          |
|--------------------|--------------------------------------|
| `output_dir`       | Directory where recordings are saved |
| `default_format`   | Default output format (`mp3`/`wav`)  |
| `default_duration` | Default recording duration           |

### Reinstall dependencies

If you need to reinstall the system dependencies separately:

```bash
make deps
```
