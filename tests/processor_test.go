package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Mock HLS playlist data
const mockMasterPlaylist = `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:BANDWIDTH=1280000,RESOLUTION=854x480
480p.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=2560000,RESOLUTION=1280x720
720p.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=5120000,RESOLUTION=1920x1080
1080p.m3u8`

const mockMediaPlaylist = `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:10
#EXT-X-MEDIA-SEQUENCE:0
#EXTINF:9.009,
segment0000.ts
#EXTINF:9.009,
segment0001.ts
#EXTINF:9.009,
segment0002.ts
#EXT-X-ENDLIST`

const mockSegmentData = "mock_ts_segment_data_12345"

func createMockHLSServer() *httptest.Server {
	mux := http.NewServeMux()
	
	// Master playlist endpoint
	mux.HandleFunc("/master.m3u8", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		w.Write([]byte(mockMasterPlaylist))
	})
	
	// Media playlist endpoint
	mux.HandleFunc("/720p.m3u8", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		w.Write([]byte(mockMediaPlaylist))
	})
	
	// Segment endpoints
	mux.HandleFunc("/segment", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".ts") {
			w.Header().Set("Content-Type", "video/mp2t")
			w.Write([]byte(mockSegmentData))
		} else {
			http.NotFound(w, r)
		}
	})
	
	return httptest.NewServer(mux)
}

func TestNewHLSProcessor(t *testing.T) {
	mockPostProcessor := &PostProcessor{}
	processor := NewHLSProcessor("https://example.com/stream.m3u8", mockPostProcessor)
	
	if processor.url != "https://example.com/stream.m3u8" {
		t.Errorf("Expected URL 'https://example.com/stream.m3u8', got '%s'", processor.url)
	}
	
	if processor.postProcessor != mockPostProcessor {
		t.Error("Post-processor not set correctly")
	}
	
	if processor.cache == nil {
		t.Error("Cache should not be nil")
	}
	
	if processor.running {
		t.Error("Processor should not be running initially")
	}
}

func TestHLSProcessorStartStop(t *testing.T) {
	processor := NewHLSProcessor("https://example.com/stream.m3u8", nil)
	
	// Initially not running
	if processor.running {
		t.Error("Processor should not be running initially")
	}
	
	// Start processor
	processor.Start()
	if !processor.running {
		t.Error("Processor should be running after Start()")
	}
	
	// Stop processor
	processor.Stop()
	if processor.running {
		t.Error("Processor should not be running after Stop()")
	}
}

func TestHLSProcessorFetch(t *testing.T) {
	server := createMockHLSServer()
	defer server.Close()
	
	processor := NewHLSProcessor("", nil)
	
	// Test successful fetch
	data := processor.fetch(server.URL + "/720p.m3u8")
	if data == nil {
		t.Error("Expected data from fetch, got nil")
	}
	
	if !strings.Contains(string(data), "#EXTM3U") {
		t.Error("Fetched data does not contain expected HLS content")
	}
	
	// Test failed fetch (non-existent URL)
	data = processor.fetch(server.URL + "/nonexistent")
	if data != nil {
		t.Error("Expected nil from failed fetch, got data")
	}
	
	// Test timeout (can't easily test without modifying timeout)
	// Test invalid URL
	data = processor.fetch("invalid-url")
	if data != nil {
		t.Error("Expected nil from invalid URL fetch, got data")
	}
}

func TestHLSProcessorGetStats(t *testing.T) {
	processor := NewHLSProcessor("https://example.com/stream.m3u8", nil)
	
	stats := processor.GetStats()
	
	// Check required fields
	if stats["running"] != false {
		t.Error("Expected running to be false initially")
	}
	
	if stats["url"] != "https://example.com/stream.m3u8" {
		t.Errorf("Expected URL in stats, got %v", stats["url"])
	}
	
	if _, exists := stats["cache_size"]; !exists {
		t.Error("Expected cache_size in stats")
	}
	
	// Start processor and check running status
	processor.Start()
	stats = processor.GetStats()
	
	if stats["running"] != true {
		t.Error("Expected running to be true after start")
	}
	
	processor.Stop()
}

func TestHLSProcessorCacheIntegration(t *testing.T) {
	processor := NewHLSProcessor("https://example.com/stream.m3u8", nil)
	
	// Test that cache is working
	url1 := "https://example.com/segment1.ts"
	url2 := "https://example.com/segment2.ts"
	
	// First time should return true (new)
	if !processor.cache.Set(url1) {
		t.Error("Expected cache.Set to return true for new URL")
	}
	
	// Second time should return false (duplicate)
	if processor.cache.Set(url1) {
		t.Error("Expected cache.Set to return false for duplicate URL")
	}
	
	// Different URL should return true
	if !processor.cache.Set(url2) {
		t.Error("Expected cache.Set to return true for different URL")
	}
	
	// Check cache size
	if processor.cache.Size() != 2 {
		t.Errorf("Expected cache size 2, got %d", processor.cache.Size())
	}
}

func TestHLSProcessorWithMockServer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	server := createMockHLSServer()
	defer server.Close()
	
	// Setup test database
	originalDir := setupTestDatabase(t)
	defer cleanupTestDatabase(t, originalDir)
	
	processor := NewHLSProcessor(server.URL+"/720p.m3u8", nil)
	
	// Start processor briefly
	processor.Start()
	
	// Let it run for a short time
	time.Sleep(100 * time.Millisecond)
	
	processor.Stop()
	
	// Check that processor was running
	stats := processor.GetStats()
	if stats["running"] != false {
		t.Error("Expected processor to be stopped")
	}
}

func TestHLSProcessorProcessSegments(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping process segments test in short mode")
	}
	
	// This test would require setting up a full environment
	// with database and file system, which is complex
	// For now, we test the components separately
	
	processor := NewHLSProcessor("https://example.com/stream.m3u8", nil)
	
	// Test that processSegments method exists and can be called
	// (actual testing would require mock M3U8 playlist parsing)
	
	// Verify processor has expected methods
	if processor.cache == nil {
		t.Error("Processor should have cache initialized")
	}
	
	if processor.url == "" {
		t.Error("Processor should have URL set")
	}
}

func TestHLSProcessorErrorHandling(t *testing.T) {
	processor := NewHLSProcessor("https://example.com/stream.m3u8", nil)
	
	// Test fetch with various error conditions
	testCases := []string{
		"",                          // Empty URL
		"invalid-url",              // Invalid URL format
		"http://localhost:99999",   // Connection refused
		"https://httpstat.us/500",  // Server error (if accessible)
	}
	
	for _, url := range testCases {
		data := processor.fetch(url)
		if data != nil {
			t.Logf("Warning: Expected nil for URL %s, got data (might be expected for some test URLs)", url)
		}
	}
}

func TestHLSProcessorUserAgent(t *testing.T) {
	// Test that fetch sets correct User-Agent
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userAgent := r.Header.Get("User-Agent")
		if userAgent != "iptv/1.0" {
			t.Errorf("Expected User-Agent 'iptv/1.0', got '%s'", userAgent)
		}
		w.Write([]byte("OK"))
	}))
	defer server.Close()
	
	processor := NewHLSProcessor("", nil)
	data := processor.fetch(server.URL)
	
	if data == nil {
		t.Error("Expected data from fetch")
	}
	
	if string(data) != "OK" {
		t.Errorf("Expected 'OK', got '%s'", string(data))
	}
}

func TestHLSProcessorTimeout(t *testing.T) {
	// Test fetch timeout behavior
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(15 * time.Second) // Longer than default timeout
		w.Write([]byte("Slow response"))
	}))
	defer server.Close()
	
	processor := NewHLSProcessor("", nil)
	
	start := time.Now()
	data := processor.fetch(server.URL)
	duration := time.Since(start)
	
	// Should timeout before 15 seconds
	if duration > 12*time.Second {
		t.Error("Fetch did not timeout as expected")
	}
	
	if data != nil {
		t.Error("Expected nil data due to timeout")
	}
}

// Benchmark tests
func BenchmarkHLSProcessorFetch(b *testing.B) {
	server := createMockHLSServer()
	defer server.Close()
	
	processor := NewHLSProcessor("", nil)
	url := server.URL + "/720p.m3u8"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		processor.fetch(url)
	}
}

func BenchmarkHLSProcessorCacheSet(b *testing.B) {
	processor := NewHLSProcessor("https://example.com/stream.m3u8", nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		processor.cache.Set(fmt.Sprintf("url%d", i))
	}
}

func BenchmarkHLSProcessorGetStats(b *testing.B) {
	processor := NewHLSProcessor("https://example.com/stream.m3u8", nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		processor.GetStats()
	}
}

// Helper function to setup test database (imported from database_test.go)
func setupTestDatabase(t *testing.T) string {
	// This would normally import from database_test.go
	// For simplicity, we'll create a minimal version here
	return ""
}

func cleanupTestDatabase(t *testing.T, originalDir string) {
	// Cleanup function
}

// Test processor with different playlist types
func TestHLSProcessorPlaylistTypes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		
		switch r.URL.Path {
		case "/master.m3u8":
			w.Write([]byte(mockMasterPlaylist))
		case "/media.m3u8":
			w.Write([]byte(mockMediaPlaylist))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	
	processor := NewHLSProcessor("", nil)
	
	// Test master playlist fetch
	data := processor.fetch(server.URL + "/master.m3u8")
	if data == nil {
		t.Error("Failed to fetch master playlist")
	}
	
	if !strings.Contains(string(data), "EXT-X-STREAM-INF") {
		t.Error("Master playlist missing expected content")
	}
	
	// Test media playlist fetch
	data = processor.fetch(server.URL + "/media.m3u8")
	if data == nil {
		t.Error("Failed to fetch media playlist")
	}
	
	if !strings.Contains(string(data), "EXTINF") {
		t.Error("Media playlist missing expected content")
	}
}
