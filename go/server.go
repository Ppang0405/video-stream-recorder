package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// HTTPServer handles all HTTP endpoints and routing
type HTTPServer struct {
	router    *mux.Router
	bindAddr  string
	host      string
	processor *HLSProcessor
}

// NewHTTPServer creates a new HTTP server instance
func NewHTTPServer(bindAddr, host string, processor *HLSProcessor) *HTTPServer {
	server := &HTTPServer{
		router:    mux.NewRouter(),
		bindAddr:  bindAddr,
		host:      host,
		processor: processor,
	}
	
	server.setupRoutes()
	return server
}

// setupRoutes configures all HTTP routes
func (s *HTTPServer) setupRoutes() {
	// Live stream endpoints
	s.router.HandleFunc("/live/stream.m3u8", s.handleLiveStream)
	s.router.HandleFunc("/live/{ts:.+}", s.handleServeFile)
	
	// VOD endpoints
	s.router.HandleFunc("/start/{start}/{limit}/vod.m3u8", s.handleVODPlaylist)
	s.router.HandleFunc("/start/{start}/{limit}/stream.m3u8", s.handleVODStream)
	s.router.HandleFunc("/start/{start}/{limit}/{ts:.+}", s.handleServeFile)
	
	// Convenience endpoints
	s.router.HandleFunc("/", s.handleRoot)
	s.router.HandleFunc("/live", s.handleLiveRedirect)
	
	// API endpoints
	s.router.HandleFunc("/api/status", s.handleAPIStatus)
	s.router.HandleFunc("/api/stats", s.handleAPIStats)
}

// handleLiveStream serves the live HLS manifest
func (s *HTTPServer) handleLiveStream(w http.ResponseWriter, r *http.Request) {
	utc := r.URL.Query().Get("utc")
	if utc != "" {
		s.handleVODRequest(w, r)
		return
	}

	out := "#EXTM3U\n"
	out += "#EXT-X-TARGETDURATION:2\n"
	out += "#EXT-X-VERSION:4\n"

	items := database_last_5()
	if len(items) <= 0 {
		http.Error(w, "No segments available", http.StatusNotFound)
		return
	}

	last := items[len(items)-1]
	out += fmt.Sprintf("#EXT-X-MEDIA-SEQUENCE:%d\n", last.ID)

	for _, item := range items {
		out += fmt.Sprintf("#EXTINF:%f\n", item.Len)
		out += fmt.Sprintf("%s\n", item.Name)
	}

	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(out)))
	w.Write([]byte(out))
}

// handleVODPlaylist serves VOD playlists
func (s *HTTPServer) handleVODPlaylist(w http.ResponseWriter, r *http.Request) {
	varz := mux.Vars(r)

	out := "#EXTM3U\n"
	out += "#EXT-X-PLAYLIST-TYPE:VOD\n"
	out += "#EXT-X-TARGETDURATION:20\n"
	out += "#EXT-X-VERSION:4\n"
	out += "#EXT-X-MEDIA-SEQUENCE:0\n"

	t, err := time.Parse("20060102150405", varz["start"])
	if err != nil {
		http.Error(w, "Invalid timestamp format", http.StatusBadRequest)
		return
	}
	start := fmt.Sprintf("%d", t.Unix())

	items := database_get(start, varz["limit"])

	for _, item := range items {
		out += fmt.Sprintf("#EXTINF:%f\n", item.Len)
		out += fmt.Sprintf("%s\n", item.Name)
	}

	out += "#EXT-X-ENDLIST\n"

	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(out)))
	w.Write([]byte(out))
}

// handleVODStream serves VOD streams
func (s *HTTPServer) handleVODStream(w http.ResponseWriter, r *http.Request) {
	s.handleVODRequest(w, r)
}

// handleVODRequest handles VOD requests with parameters
func (s *HTTPServer) handleVODRequest(w http.ResponseWriter, r *http.Request) {
	varz := mux.Vars(r)

	utc := r.URL.Query().Get("utc")
	if utc != "" {
		varz["start"] = utc
		varz["limit"] = "300"
	}

	out := "#EXTM3U\n"
	out += "#EXT-X-PLAYLIST-TYPE:VOD\n"
	out += "#EXT-X-TARGETDURATION:20\n"
	out += "#EXT-X-VERSION:4\n"
	out += "#EXT-X-MEDIA-SEQUENCE:0\n"

	items := database_get(varz["start"], varz["limit"])

	for _, item := range items {
		out += fmt.Sprintf("#EXTINF:%f\n", item.Len)
		out += fmt.Sprintf("%s\n", item.Name)
	}

	out += "#EXT-X-ENDLIST\n"

	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(out)))
	w.Write([]byte(out))
}

// handleServeFile serves individual TS segment files
func (s *HTTPServer) handleServeFile(w http.ResponseWriter, r *http.Request) {
	varz := mux.Vars(r)
	
	filePath := fmt.Sprintf("./files/%s", varz["ts"])
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Printf("error reading file %s: %v", filePath, err)
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "video/mp2t")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(b)))
	w.Write(b)
}

// handleRoot serves the main dashboard
func (s *HTTPServer) handleRoot(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>VSR - Video Stream Recorder</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background: #f5f5f5; }
        .container { max-width: 800px; margin: 0 auto; background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #333; border-bottom: 2px solid #007cba; padding-bottom: 10px; }
        .endpoint { background: #f8f9fa; padding: 15px; margin: 10px 0; border-radius: 5px; border-left: 4px solid #007cba; }
        .endpoint h3 { margin-top: 0; color: #007cba; }
        code { background: #e9ecef; padding: 2px 6px; border-radius: 3px; font-family: monospace; }
        .live { border-left-color: #28a745; }
        .vod { border-left-color: #ffc107; }
        .api { border-left-color: #6f42c1; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🎥 Video Stream Recorder (VSR)</h1>
        <p>Welcome to VSR! Your HLS stream recorder is running.</p>
        
        <div class="endpoint live">
            <h3>📡 Live Stream</h3>
            <p><strong>HLS Manifest:</strong> <code>/live/stream.m3u8</code></p>
            <p><strong>Example:</strong> <code>http://localhost:8080/live/stream.m3u8</code></p>
        </div>
        
        <div class="endpoint vod">
            <h3>🎬 Video On Demand (VOD)</h3>
            <p><strong>VOD Manifest:</strong> <code>/start/{timestamp}/{duration}/vod.m3u8</code></p>
            <p><strong>Stream Manifest:</strong> <code>/start/{timestamp}/{duration}/stream.m3u8</code></p>
            <p><strong>Timestamp Format:</strong> YYYYMMDDHHMMSS (e.g., 20240107120000)</p>
        </div>
        
        <div class="endpoint api">
            <h3>🔧 API Endpoints</h3>
            <p><strong>Status:</strong> <code>/api/status</code></p>
            <p><strong>Statistics:</strong> <code>/api/stats</code></p>
        </div>
        
        <div class="endpoint">
            <h3>🔗 Quick Links</h3>
            <p><a href="/live/stream.m3u8">Live Stream (M3U8)</a></p>
            <p><a href="/api/status">System Status</a></p>
            <p><a href="/api/stats">Statistics</a></p>
        </div>
    </div>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// handleLiveRedirect redirects /live to the live stream
func (s *HTTPServer) handleLiveRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/live/stream.m3u8", http.StatusFound)
}

// handleAPIStatus provides system status information
func (s *HTTPServer) handleAPIStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"status":    "running",
		"timestamp": time.Now().Unix(),
		"version":   "v0.2",
	}
	
	if s.processor != nil {
		processorStats := s.processor.GetStats()
		status["processor"] = processorStats
	}
	
	w.Header().Set("Content-Type", "application/json")
	// Simple JSON response without external dependencies
	fmt.Fprintf(w, `{"status":"%s","timestamp":%d,"version":"%s"}`, 
		status["status"], status["timestamp"], status["version"])
}

// handleAPIStats provides detailed statistics
func (s *HTTPServer) handleAPIStats(w http.ResponseWriter, r *http.Request) {
	// Get database statistics
	segmentCount := database_count_segments()
	
	stats := map[string]interface{}{
		"segments_total":    segmentCount,
		"segments_recent":   len(database_last_5()),
		"cache_size":        globalCache.Size(),
		"uptime_seconds":    time.Since(startTime).Seconds(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	// Simple JSON response
	fmt.Fprintf(w, `{"segments_total":%d,"segments_recent":%d,"cache_size":%d,"uptime_seconds":%f}`,
		stats["segments_total"], stats["segments_recent"], stats["cache_size"], stats["uptime_seconds"])
}

// Start starts the HTTP server
func (s *HTTPServer) Start() error {
	log.Printf("Starting HTTP server on %s", s.bindAddr)
	return http.ListenAndServe(s.bindAddr, s.router)
}

// Global variable to track start time for uptime calculation
var startTime = time.Now()
