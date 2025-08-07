# API Reference

## 🔌 HTTP Endpoints

### **Live Streaming Endpoints**

#### `GET /live/stream.m3u8`
Returns an HLS playlist with the most recent segments for live streaming.

**Response:**
```http
HTTP/1.1 200 OK
Content-Type: application/vnd.apple.mpegurl
Content-Length: 285

#EXTM3U
#EXT-X-TARGETDURATION:2
#EXT-X-VERSION:4
#EXT-X-MEDIA-SEQUENCE:12345
#EXTINF:2.000000,
1642678920000000000.ts
#EXTINF:2.000000,
1642678922000000000.ts
#EXTINF:2.000000,
1642678924000000000.ts
```

**Query Parameters:**
- `utc` (optional): Unix timestamp for specific time point

**Example:**
```bash
curl http://localhost:8080/live/stream.m3u8
curl http://localhost:8080/live/stream.m3u8?utc=1642678920
```

#### `GET /live/{segment}`
Serves individual TS segment files.

**Parameters:**
- `segment`: Segment filename (e.g., `1642678920000000000.ts`)

**Response:**
```http
HTTP/1.1 200 OK
Content-Type: text/vnd.trolltech.linguist
Content-Length: 2048576

[Binary TS data]
```

### **Video On Demand (VOD) Endpoints**

#### `GET /start/{timestamp}/{duration}/vod.m3u8`
Returns an HLS playlist for a specific time range.

**Parameters:**
- `timestamp`: Start time in format `YYYYMMDDHHMMSS` (e.g., `20240115143000`)
- `duration`: Duration in minutes (e.g., `300` for 5 hours)

**Response:**
```http
HTTP/1.1 200 OK
Content-Type: application/vnd.apple.mpegurl

#EXTM3U
#EXT-X-PLAYLIST-TYPE:VOD
#EXT-X-TARGETDURATION:20
#EXT-X-VERSION:4
#EXT-X-MEDIA-SEQUENCE:0
#EXTINF:2.000000,
1642678920000000000.ts
#EXTINF:2.000000,
1642678922000000000.ts
#EXT-X-ENDLIST
```

**Example:**
```bash
# Get 2 hours starting from Jan 15, 2024 at 14:30
curl http://localhost:8080/start/20240115143000/120/vod.m3u8
```

#### `GET /start/{timestamp}/{duration}/stream.m3u8`
Alternative VOD endpoint with streaming format.

**Parameters:**
- Same as above

#### `GET /start/{unix_timestamp}/{duration}/vod.m3u8`
VOD endpoint using Unix timestamp.

**Parameters:**
- `unix_timestamp`: Unix timestamp (e.g., `1705327800`)
- `duration`: Duration in minutes

**Example:**
```bash
curl http://localhost:8080/start/1705327800/120/vod.m3u8
```

#### `GET /start/{timestamp}/{duration}/{segment}`
Serves segment files for VOD playback.

### **Health and Monitoring**

#### `GET /health`
Returns service health status (if implemented).

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T14:30:00Z",
  "streams": {
    "active": 5,
    "total": 100
  },
  "storage": {
    "used_gb": 1250,
    "available_gb": 2750
  }
}
```

## 📝 Configuration API (Future Enhancement)

### **Stream Management**

#### `POST /api/streams`
Add a new stream to record.

**Request:**
```json
{
  "url": "https://example.com/stream.m3u8",
  "name": "stream1",
  "quality": "1080p",
  "retention_hours": 24
}
```

**Response:**
```json
{
  "id": "stream_123",
  "status": "starting",
  "created_at": "2024-01-15T14:30:00Z"
}
```

#### `GET /api/streams`
List all active streams.

**Response:**
```json
{
  "streams": [
    {
      "id": "stream_123",
      "name": "stream1",
      "url": "https://example.com/stream.m3u8",
      "status": "active",
      "started_at": "2024-01-15T14:30:00Z",
      "segments_count": 1205,
      "last_segment_at": "2024-01-15T16:30:00Z"
    }
  ],
  "total": 1,
  "active": 1
}
```

#### `DELETE /api/streams/{id}`
Stop recording a stream.

**Response:**
```json
{
  "status": "stopped",
  "message": "Stream recording stopped successfully"
}
```

### **Analytics API**

#### `GET /api/analytics/streams/{id}`
Get stream analytics data.

**Response:**
```json
{
  "stream_id": "stream_123",
  "period": "24h",
  "stats": {
    "total_segments": 1205,
    "total_duration_seconds": 2410,
    "average_bitrate": 5000000,
    "storage_used_bytes": 1073741824,
    "uptime_percentage": 99.5
  }
}
```

## 🔧 Client SDK Examples

### **JavaScript/Node.js**

```javascript
class VSRClient {
  constructor(baseUrl) {
    this.baseUrl = baseUrl;
  }
  
  async getLiveStream() {
    const response = await fetch(`${this.baseUrl}/live/stream.m3u8`);
    return response.text();
  }
  
  async getVODStream(timestamp, duration) {
    const response = await fetch(
      `${this.baseUrl}/start/${timestamp}/${duration}/vod.m3u8`
    );
    return response.text();
  }
  
  async getHealth() {
    const response = await fetch(`${this.baseUrl}/health`);
    return response.json();
  }
}

// Usage
const client = new VSRClient('http://localhost:8080');
const playlist = await client.getLiveStream();
```

### **Python**

```python
import requests
from datetime import datetime, timedelta

class VSRClient:
    def __init__(self, base_url):
        self.base_url = base_url.rstrip('/')
    
    def get_live_stream(self):
        response = requests.get(f'{self.base_url}/live/stream.m3u8')
        response.raise_for_status()
        return response.text
    
    def get_vod_stream(self, start_time, duration_minutes):
        timestamp = start_time.strftime('%Y%m%d%H%M%S')
        url = f'{self.base_url}/start/{timestamp}/{duration_minutes}/vod.m3u8'
        response = requests.get(url)
        response.raise_for_status()
        return response.text
    
    def get_health(self):
        response = requests.get(f'{self.base_url}/health')
        response.raise_for_status()
        return response.json()

# Usage
client = VSRClient('http://localhost:8080')
playlist = client.get_live_stream()

# Get VOD for last 2 hours
start_time = datetime.now() - timedelta(hours=2)
vod_playlist = client.get_vod_stream(start_time, 120)
```

### **Go**

```go
package main

import (
    "fmt"
    "io"
    "net/http"
    "time"
)

type VSRClient struct {
    baseURL string
    client  *http.Client
}

func NewVSRClient(baseURL string) *VSRClient {
    return &VSRClient{
        baseURL: baseURL,
        client:  &http.Client{Timeout: 30 * time.Second},
    }
}

func (c *VSRClient) GetLiveStream() (string, error) {
    resp, err := c.client.Get(c.baseURL + "/live/stream.m3u8")
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }
    
    return string(body), nil
}

func (c *VSRClient) GetVODStream(timestamp string, duration int) (string, error) {
    url := fmt.Sprintf("%s/start/%s/%d/vod.m3u8", c.baseURL, timestamp, duration)
    resp, err := c.client.Get(url)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }
    
    return string(body), nil
}

// Usage
func main() {
    client := NewVSRClient("http://localhost:8080")
    playlist, err := client.GetLiveStream()
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    fmt.Println(playlist)
}
```

## 🎬 Media Player Integration

### **Video.js**

```html
<!DOCTYPE html>
<html>
<head>
    <link href="https://vjs.zencdn.net/8.6.1/video-js.css" rel="stylesheet">
    <script src="https://vjs.zencdn.net/8.6.1/video.min.js"></script>
</head>
<body>
    <video-js
        id="vsr-player"
        class="vjs-default-skin"
        controls
        preload="auto"
        data-setup="{}">
    </video-js>

    <script>
        const player = videojs('vsr-player');
        
        // Live stream
        player.src({
            src: 'http://localhost:8080/live/stream.m3u8',
            type: 'application/x-mpegURL'
        });
        
        // VOD stream
        function playVOD(timestamp, duration) {
            player.src({
                src: `http://localhost:8080/start/${timestamp}/${duration}/vod.m3u8`,
                type: 'application/x-mpegURL'
            });
        }
    </script>
</body>
</html>
```

### **HLS.js**

```html
<script src="https://cdn.jsdelivr.net/npm/hls.js@latest"></script>
<video id="video" controls></video>

<script>
    const video = document.getElementById('video');
    const videoSrc = 'http://localhost:8080/live/stream.m3u8';
    
    if (Hls.isSupported()) {
        const hls = new Hls();
        hls.loadSource(videoSrc);
        hls.attachMedia(video);
        
        hls.on(Hls.Events.MANIFEST_PARSED, function() {
            video.play();
        });
    } else if (video.canPlayType('application/vnd.apple.mpegurl')) {
        video.src = videoSrc;
    }
</script>
```

## 🔍 Error Responses

### **Common Error Codes**

| Status Code | Description | Example |
|-------------|-------------|---------|
| 200 | Success | Playlist returned successfully |
| 400 | Bad Request | Invalid timestamp format |
| 404 | Not Found | No segments found for time range |
| 500 | Internal Error | Database connection failed |
| 503 | Service Unavailable | Too many concurrent requests |

### **Error Response Format**

```json
{
  "error": {
    "code": "INVALID_TIMESTAMP",
    "message": "Timestamp format must be YYYYMMDDHHMMSS",
    "details": "Received: 2024011514300"
  },
  "request_id": "req_123456789",
  "timestamp": "2024-01-15T14:30:00Z"
}
```

## 📊 Rate Limiting

### **Default Limits**
- **Live Stream Requests**: 100 requests/minute per IP
- **VOD Requests**: 50 requests/minute per IP
- **Segment Downloads**: 1000 requests/minute per IP

### **Rate Limit Headers**
```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1705327860
```

## 🔐 Authentication (Future Enhancement)

### **API Key Authentication**

```http
GET /live/stream.m3u8
Authorization: Bearer your-api-key-here
```

### **JWT Authentication**

```http
GET /api/streams
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

## 📈 Metrics Endpoint

### **Prometheus Metrics**

```http
GET /metrics

# HELP vsr_active_streams Number of currently active streams
# TYPE vsr_active_streams gauge
vsr_active_streams 5

# HELP vsr_segments_processed_total Total number of segments processed
# TYPE vsr_segments_processed_total counter
vsr_segments_processed_total{stream_id="stream_1"} 1205

# HELP vsr_errors_total Total number of errors by type
# TYPE vsr_errors_total counter
vsr_errors_total{type="network_error"} 3
vsr_errors_total{type="parse_error"} 1
```

This API reference provides comprehensive documentation for all current and planned endpoints, making it easy for developers to integrate with the VSR system.
