# Video Stream Recorder (VSR) Documentation

A comprehensive HLS video stream recorder with pre-processing, recording, and post-processing capabilities.

## 📚 Documentation Index

- [**Architecture Overview**](architecture.md) - System design and component interactions
- [**Implementation Trade-offs**](tradeoffs.md) - Current implementation decisions and Go language considerations
- [**Deployment Guide**](deployment.md) - Server requirements for production deployment
- [**API Reference**](api.md) - HTTP endpoints and usage examples
- [**Configuration**](configuration.md) - Command-line options and settings

## 🚀 Quick Start

```bash
# Basic usage
./vsr --url "https://example.com/stream.m3u8" --tail 24

# With preprocessing
./vsr --url "https://youtube.com/watch?v=example" --preprocessor ytdlp --tail 48

# Access streams
curl http://localhost:8080/live/stream.m3u8
```

## 🏗️ Project Structure

```
video-stream-recorder/
├── go/                     # Core application
│   ├── main.go            # Main application logic
│   ├── database.go        # BoltDB storage layer
│   ├── fifocache.go       # FIFO caching mechanism
│   ├── go.mod             # Go module dependencies
│   └── go.sum             # Dependency checksums
├── docs/                  # Documentation
│   ├── README.md          # This file
│   ├── architecture.md    # System architecture
│   ├── tradeoffs.md       # Implementation decisions
│   └── deployment.md      # Production deployment
├── build.sh               # Cross-platform build script
└── README.md              # Project overview
```

## 🎯 Core Features

- **Multi-source Support**: Direct HLS, YouTube, Twitch, and 1000+ sites via yt-dlp
- **Real-time Recording**: Continuous HLS segment downloading and storage
- **Time-shifted Playback**: VOD access to any recorded timeframe
- **Automatic Cleanup**: Configurable retention policies
- **HTTP API**: RESTful endpoints for live and VOD access
- **Cross-platform**: Linux, macOS, Windows, and ARM support

## 📈 Performance Characteristics

- **Memory Usage**: ~50MB base + ~10MB per concurrent stream
- **Storage**: ~1GB per hour per stream (varies by bitrate)
- **CPU Usage**: ~5% per stream on modern hardware
- **Network**: Dependent on stream bitrate (typically 1-10 Mbps per stream)

## 🔗 Related Documentation

For detailed information about specific aspects of the system, please refer to the individual documentation files linked above.
