# auD-limit (Go rewrite)

This branch contains a Go rewrite of the original auD-limiter Python TUI.
It uses Bubble Tea for the TUI and a small SQLite cache for the library.

Prerequisites
- mpv (for playback)
- ffprobe / ffmpeg (for metadata and conversions)
- Go 1.20+

Build & run (Linux / macOS / Windows with Go installed):

# build
make build

# run (scan current directory and open TUI)
./bin/auD-limit -scan .