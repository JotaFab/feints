# Feints

_Feints_ is a Discord music bot written in Go. It enables music playback in Discord voice channels, supports streaming audio from various sources (including YouTube), slash-command interactions, and manages concurrent playback sessions across multiple Discord guilds.

---

## Features

- Streams audio from YouTube and other sources  
- Slash commands for playback control (play, pause, skip, stop, etc.)  
- Multi-guild support: each guild has independent voice/music sessions  
- Concurrent playback management  
- Configuration via config files / environment variables  
- Docker support for easy deployment  

---

## Requirements

- Go 1.25+ (or whatever Go version you have as minimum)  
- A Discord Bot Token with appropriate permissions (voice, commands)  
- FFmpeg installed / accessible in environment (if required for audio processing)
- ytdlp uses cookies.txt for yt dowloading the music.
- Internet access for fetching/streaming audio
- Docker and Docker-compose 

---

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/JotaFab/feints.git
   cd feints
2. Using Docker:
- docker-compose up --build

Usage

Once running, use slash commands in your Discord server, for example:

/play <url or search query> — Play a song or add to queue

/pause — Pause current playback

/resume — Resume playback

/skip — Skip current track

/stop — Stop playing and clear queue

Contributing

Contributions are welcome! If you want to help:

Fork the repo

Create a feature branch (git checkout -b feature/YourFeature)

Write tests for new functionality

Submit a Pull Request

