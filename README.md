# mplar

A lightweight terminal UI (TUI) music player built in Go. Search YouTube directly from your terminal, pick a result, and control playback — all without leaving the keyboard or opening a browser.

## Features

- Search YouTube from an in-terminal search bar
- Browse results in a scrollable list
- Play audio via `mpv` (no video window, minimal resource usage)
- Control playback (pause/resume, volume, seek) via `mpv`'s IPC socket

## Prerequisites

Make sure the following are installed before building:

- **Go** 1.21 or later — check with `go version`
- **mpv** — used for audio playback
- **socat** — used to send playback commands to mpv over its IPC socket
- **yt-dlp** — used to search and resolve YouTube results

On most Linux distros:

```bash
sudo apt install mpv socat
pip install yt-dlp --break-system-packages
```

(or use your distro's package manager / `pipx` for `yt-dlp`)

## Setup

1. Clone the repository:

   ```bash
   git clone https://github.com/abdu11ahfa1zan/mplar.git
   cd mplar
   ```

2. Install Go dependencies:

   ```bash
   go mod tidy
   ```

   This reads `go.mod`/`go.sum` and pulls in everything the project needs (`bubbletea`, `bubbles`, etc.) — no manual `go get` required.

## Running

Run directly without producing a binary:

```bash
go run main.go
```

Or build a binary and run it:

```bash
go build -o mplar
./mplar
```

## How it works

- Searching runs `yt-dlp` in `ytsearch` mode and parses the titles/links into a scrollable list.
- Selecting a result launches `mpv` in the background with `--no-video` and `--input-ipc-server=/tmp/mpvsocket`, so it plays audio only and exposes a control socket.
- Playback controls (pause, volume, seek) are sent as JSON commands to that socket via `socat`, so the same running `mpv` process can be controlled without restarting it.

**Note:** only one `mpv` instance should use `/tmp/mpvsocket` at a time — starting a second track while one is already playing may conflict on the socket path.

## License

Add your license of choice here (MIT, Apache 2.0, etc.) if you plan to share this publicly.
