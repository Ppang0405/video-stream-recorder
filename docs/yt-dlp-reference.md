# yt-dlp Command Reference for Stream Extraction

## 🎯 Overview

This document provides comprehensive yt-dlp commands for extracting video stream information without downloading the actual video content, specifically focusing on HLS stream URLs for the VSR project.

## 🔍 Basic Information Extraction

### **Extract All Available Formats**
```bash
# Get all available formats with detailed info
yt-dlp -F "https://youtube.com/watch?v=VIDEO_ID"

# Output example:
# [youtube] VIDEO_ID: Downloading webpage
# [info] Available formats for VIDEO_ID:
# format code  extension  resolution note
# 139          m4a        audio only tiny   49k , m4a_dash container, mp4a.40.5@ 49k (22050Hz), 1.09MiB
# 140          m4a        audio only tiny  129k , m4a_dash container, mp4a.40.2@129k (44100Hz), 2.87MiB
# 298          mp4        256x144    144p   62k , mp4_dash container, avc1.4d400c@ 62k, 30fps, video only, 1.38MiB
# 396          mp4        256x144    144p   92k , mp4_dash container, av01.0.00M.08@ 92k, 30fps, video only, 2.05MiB
```

### **Extract Only HLS Streams**
```bash
# Filter for HLS/m3u8 streams only
yt-dlp -F "https://youtube.com/watch?v=VIDEO_ID" | grep "m3u8\|hls"

# More specific HLS extraction
yt-dlp --list-formats --format-sort "proto:m3u8" "https://youtube.com/watch?v=VIDEO_ID"
```

## 📊 JSON Metadata Extraction

### **Complete Metadata in JSON Format**
```bash
# Extract all metadata without downloading
yt-dlp -J "https://youtube.com/watch?v=VIDEO_ID"

# Save to file for processing
yt-dlp -J "https://youtube.com/watch?v=VIDEO_ID" > video_info.json
```

### **Extract Specific HLS URL**
```bash
# Get the best HLS stream URL
yt-dlp -f "best[protocol=m3u8]" --get-url "https://youtube.com/watch?v=VIDEO_ID"

# Get all HLS URLs
yt-dlp -f "best[protocol=m3u8]/bestvideo[protocol=m3u8]" --get-url "https://youtube.com/watch?v=VIDEO_ID"

# Alternative method
yt-dlp --format "bestvideo[ext=m3u8]+bestaudio[ext=m3u8]" --get-url "https://youtube.com/watch?v=VIDEO_ID"
```

## 🔧 Advanced Extraction Commands

### **Extract with Quality Preferences**
```bash
# Best quality HLS stream
yt-dlp -f "best[protocol=m3u8][height<=1080]" --get-url "https://youtube.com/watch?v=VIDEO_ID"

# Specific resolution HLS
yt-dlp -f "best[protocol=m3u8][height=720]" --get-url "https://youtube.com/watch?v=VIDEO_ID"

# Mobile-friendly HLS
yt-dlp -f "worst[protocol=m3u8][height>=480]" --get-url "https://youtube.com/watch?v=VIDEO_ID"
```

### **Platform-Specific Extraction**

#### **YouTube**
```bash
# YouTube HLS streams
yt-dlp -f "best[protocol=m3u8_native]" --get-url "https://youtube.com/watch?v=VIDEO_ID"

# YouTube with fallback
yt-dlp -f "best[protocol=m3u8_native]/best[protocol=m3u8]/best" --get-url "https://youtube.com/watch?v=VIDEO_ID"
```

#### **Twitch**
```bash
# Twitch live streams
yt-dlp -f "best[protocol=m3u8]" --get-url "https://twitch.tv/CHANNEL_NAME"

# Twitch VODs
yt-dlp -f "best[protocol=m3u8]" --get-url "https://twitch.tv/videos/VIDEO_ID"
```

#### **Generic Streams**
```bash
# Any platform HLS extraction
yt-dlp -f "best[protocol=m3u8]" --get-url "https://example.com/stream"
```

## 🛠️ Go Integration Examples

### **Simple Command Execution**
```go
package main

import (
    "encoding/json"
    "os/exec"
    "strings"
)

type VideoInfo struct {
    ID          string `json:"id"`
    Title       string `json:"title"`
    URL         string `json:"url"`
    Duration    int    `json:"duration"`
    Formats     []Format `json:"formats"`
}

type Format struct {
    FormatID   string `json:"format_id"`
    URL        string `json:"url"`
    Protocol   string `json:"protocol"`
    Extension  string `json:"ext"`
    Resolution string `json:"resolution"`
    Filesize   int64  `json:"filesize"`
}

func extractVideoInfo(videoURL string) (*VideoInfo, error) {
    cmd := exec.Command("yt-dlp", "-J", videoURL)
    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }
    
    var info VideoInfo
    err = json.Unmarshal(output, &info)
    return &info, err
}

func getHLSURL(videoURL string) (string, error) {
    cmd := exec.Command("yt-dlp", 
        "-f", "best[protocol=m3u8]", 
        "--get-url", 
        videoURL)
    
    output, err := cmd.Output()
    if err != nil {
        return "", err
    }
    
    return strings.TrimSpace(string(output)), nil
}
```

### **Advanced Processing Function**
```go
func extractHLSWithQuality(videoURL string, maxHeight int) (string, error) {
    // First, try to get HLS stream with quality limit
    formatSelector := fmt.Sprintf("best[protocol=m3u8][height<=%d]", maxHeight)
    
    cmd := exec.Command("yt-dlp", 
        "-f", formatSelector,
        "--get-url", 
        videoURL)
    
    output, err := cmd.Output()
    if err != nil {
        // Fallback to any HLS stream
        cmd = exec.Command("yt-dlp", 
            "-f", "best[protocol=m3u8]",
            "--get-url", 
            videoURL)
        
        output, err = cmd.Output()
        if err != nil {
            return "", fmt.Errorf("no HLS stream found: %v", err)
        }
    }
    
    return strings.TrimSpace(string(output)), nil
}
```

## 🚀 Integration with VSR

### **Pre-processor Implementation**
```go
type YTDLPPreprocessor struct {
    maxRetries int
    timeout    time.Duration
}

func (p *YTDLPPreprocessor) ExtractHLS(inputURL string) (*StreamInfo, error) {
    // Check if already HLS
    if strings.Contains(inputURL, ".m3u8") {
        return &StreamInfo{HLSURL: inputURL}, nil
    }
    
    // Extract using yt-dlp
    cmd := exec.Command("yt-dlp", 
        "-J",
        "--no-warnings",
        inputURL)
    
    var stdout bytes.Buffer
    cmd.Stdout = &stdout
    
    ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
    defer cancel()
    
    err := cmd.Run()
    if err != nil {
        return nil, fmt.Errorf("yt-dlp failed: %v", err)
    }
    
    var info VideoInfo
    err = json.Unmarshal(stdout.Bytes(), &info)
    if err != nil {
        return nil, fmt.Errorf("failed to parse yt-dlp output: %v", err)
    }
    
    // Find best HLS format
    for _, format := range info.Formats {
        if format.Protocol == "m3u8" || format.Protocol == "m3u8_native" {
            return &StreamInfo{
                HLSURL:     format.URL,
                Title:      info.Title,
                Duration:   info.Duration,
                Resolution: format.Resolution,
            }, nil
        }
    }
    
    return nil, fmt.Errorf("no HLS stream found")
}
```

## 📋 Command Reference Quick Guide

| Purpose | Command |
|---------|---------|
| **List all formats** | `yt-dlp -F URL` |
| **Get HLS URL only** | `yt-dlp -f "best[protocol=m3u8]" --get-url URL` |
| **JSON metadata** | `yt-dlp -J URL` |
| **Best quality HLS** | `yt-dlp -f "best[protocol=m3u8][height<=1080]" --get-url URL` |
| **Check if live** | `yt-dlp -J URL \| jq '.is_live'` |
| **Get thumbnail** | `yt-dlp --get-thumbnail URL` |
| **Silent extraction** | `yt-dlp -J --no-warnings --quiet URL` |

## ⚠️ Important Notes

1. **Rate Limiting**: Some platforms have rate limits; implement backoff strategies
2. **Authentication**: Some streams require cookies or authentication
3. **Geoblocking**: Use proxies if needed for region-restricted content
4. **Updates**: Keep yt-dlp updated regularly (`yt-dlp -U`)
5. **Error Handling**: Always handle cases where no HLS stream is available
6. **Timeout**: Set reasonable timeouts for extraction commands
7. **Caching**: Cache extraction results to avoid repeated API calls

## 🔧 Installation

```bash
# Install yt-dlp
pip install yt-dlp

# Or via package manager
brew install yt-dlp  # macOS
apt install yt-dlp    # Ubuntu/Debian

# Update regularly
yt-dlp -U
```
