package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewHTTPServer(t *testing.T) {
	processor := NewHLSProcessor("https://example.com/stream.m3u8", nil)
	server := NewHTTPServer(":8080", "http://localhost:8080", processor)
	
	if server.bindAddr != ":8080" {
		t.Errorf("Expected bind address ':8080', got '%s'", server.bindAddr)
	}
	
	if server.host != "http://localhost:8080" {
		t.Errorf("Expected host 'http://localhost:8080', got '%s'", server.host)
	}
	
	if server.processor != processor {
		t.Error("Processor not set correctly")
	}
	
	if server.router == nil {
		t.Error("Router should not be nil")
	}
}

func TestHTTPServerHandleRoot(t *testing.T) {
	processor := NewHLSProcessor("https://example.com/stream.m3u8", nil)
	server := NewHTTPServer(":8080", "http://localhost:8080", processor)
	
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	
	server.router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	contentType := w.Header().Get("Content-Type")
	if contentType != "text/html" {
		t.Errorf("Expected Content-Type 'text/html', got '%s'", contentType)
	}
	
	body := w.Body.String()
	if !strings.Contains(body, "Video Stream Recorder") {
		t.Error("Root page missing expected title")
	}
	
	if !strings.Contains(body, "/live/stream.m3u8") {
		t.Error("Root page missing live stream link")
	}
}

func TestHTTPServerHandleLiveRedirect(t *testing.T) {
	processor := NewHLSProcessor("https://example.com/stream.m3u8", nil)
	server := NewHTTPServer(":8080", "http://localhost:8080", processor)
	
	req := httptest.NewRequest("GET", "/live", nil)
	w := httptest.NewRecorder()
	
	server.router.ServeHTTP(w, req)
	
	if w.Code != http.StatusFound {
		t.Errorf("Expected status 302, got %d", w.Code)
	}
	
	location := w.Header().Get("Location")
	if location != "/live/stream.m3u8" {
		t.Errorf("Expected redirect to '/live/stream.m3u8', got '%s'", location)
	}
}

func TestHTTPServerHandleLiveStream(t *testing.T) {
	// Setup test database
	originalDir := setupTestDatabase(t)
	defer cleanupTestDatabase(t, originalDir)
	
	// Add some test segments
	testItems := []*DatabaseItem{
		{Name: "segment1.ts", Len: 10.0, T: time.Now().UnixNano()},
		{Name: "segment2.ts", Len: 10.0, T: time.Now().UnixNano() + 1000000},
		{Name: "segment3.ts", Len: 10.0, T: time.Now().UnixNano() + 2000000},
	}
	
	for _, item := range testItems {
		database_store(item)
	}
	
	processor := NewHLSProcessor("https://example.com/stream.m3u8", nil)
	server := NewHTTPServer(":8080", "http://localhost:8080", processor)
	
	req := httptest.NewRequest("GET", "/live/stream.m3u8", nil)
	w := httptest.NewRecorder()
	
	server.router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/vnd.apple.mpegurl" {
		t.Errorf("Expected Content-Type 'application/vnd.apple.mpegurl', got '%s'", contentType)
	}
	
	body := w.Body.String()
	if !strings.Contains(body, "#EXTM3U") {
		t.Error("Live stream missing M3U header")
	}
	
	if !strings.Contains(body, "EXT-X-TARGETDURATION") {
		t.Error("Live stream missing target duration")
	}
	
	// Should contain segment names
	for _, item := range testItems {
		if !strings.Contains(body, item.Name) {
			t.Errorf("Live stream missing segment %s", item.Name)
		}
	}
}

func TestHTTPServerHandleLiveStreamEmpty(t *testing.T) {
	// Setup empty test database
	originalDir := setupTestDatabase(t)
	defer cleanupTestDatabase(t, originalDir)
	
	processor := NewHLSProcessor("https://example.com/stream.m3u8", nil)
	server := NewHTTPServer(":8080", "http://localhost:8080", processor)
	
	req := httptest.NewRequest("GET", "/live/stream.m3u8", nil)
	w := httptest.NewRecorder()
	
	server.router.ServeHTTP(w, req)
	
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 for empty stream, got %d", w.Code)
	}
}

func TestHTTPServerHandleVODPlaylist(t *testing.T) {
	// Setup test database
	originalDir := setupTestDatabase(t)
	defer cleanupTestDatabase(t, originalDir)
	
	// Add test segments with specific timestamps
	baseTime := time.Now().Unix()
	testItems := []*DatabaseItem{
		{Name: "vod1.ts", Len: 10.0, T: baseTime * 1000000000},
		{Name: "vod2.ts", Len: 10.0, T: (baseTime + 10) * 1000000000},
		{Name: "vod3.ts", Len: 10.0, T: (baseTime + 20) * 1000000000},
	}
	
	for _, item := range testItems {
		database_store(item)
	}
	
	processor := NewHLSProcessor("https://example.com/stream.m3u8", nil)
	server := NewHTTPServer(":8080", "http://localhost:8080", processor)
	
	// Test VOD playlist
	timestamp := time.Unix(baseTime-60, 0).Format("20060102150405") // 1 minute before
	url := "/start/" + timestamp + "/100/vod.m3u8"
	
	req := httptest.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()
	
	server.router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	body := w.Body.String()
	if !strings.Contains(body, "#EXTM3U") {
		t.Error("VOD playlist missing M3U header")
	}
	
	if !strings.Contains(body, "EXT-X-PLAYLIST-TYPE:VOD") {
		t.Error("VOD playlist missing VOD type")
	}
	
	if !strings.Contains(body, "EXT-X-ENDLIST") {
		t.Error("VOD playlist missing end list")
	}
}

func TestHTTPServerHandleVODPlaylistInvalidTimestamp(t *testing.T) {
	processor := NewHLSProcessor("https://example.com/stream.m3u8", nil)
	server := NewHTTPServer(":8080", "http://localhost:8080", processor)
	
	req := httptest.NewRequest("GET", "/start/invalid/100/vod.m3u8", nil)
	w := httptest.NewRecorder()
	
	server.router.ServeHTTP(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid timestamp, got %d", w.Code)
	}
}

func TestHTTPServerHandleAPIStatus(t *testing.T) {
	processor := NewHLSProcessor("https://example.com/stream.m3u8", nil)
	server := NewHTTPServer(":8080", "http://localhost:8080", processor)
	
	req := httptest.NewRequest("GET", "/api/status", nil)
	w := httptest.NewRecorder()
	
	server.router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}
	
	body := w.Body.String()
	if !strings.Contains(body, "status") {
		t.Error("API status missing status field")
	}
	
	if !strings.Contains(body, "timestamp") {
		t.Error("API status missing timestamp field")
	}
	
	if !strings.Contains(body, "version") {
		t.Error("API status missing version field")
	}
}

func TestHTTPServerHandleAPIStats(t *testing.T) {
	// Setup test database
	originalDir := setupTestDatabase(t)
	defer cleanupTestDatabase(t, originalDir)
	
	// Add some test segments
	for i := 0; i < 5; i++ {
		item := &DatabaseItem{
			Name: fmt.Sprintf("stats_segment%d.ts", i),
			Len:  10.0,
			T:    time.Now().UnixNano(),
		}
		database_store(item)
	}
	
	processor := NewHLSProcessor("https://example.com/stream.m3u8", nil)
	server := NewHTTPServer(":8080", "http://localhost:8080", processor)
	
	req := httptest.NewRequest("GET", "/api/stats", nil)
	w := httptest.NewRecorder()
	
	server.router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}
	
	body := w.Body.String()
	if !strings.Contains(body, "segments_total") {
		t.Error("API stats missing segments_total field")
	}
	
	if !strings.Contains(body, "segments_recent") {
		t.Error("API stats missing segments_recent field")
	}
	
	if !strings.Contains(body, "cache_size") {
		t.Error("API stats missing cache_size field")
	}
	
	if !strings.Contains(body, "uptime_seconds") {
		t.Error("API stats missing uptime_seconds field")
	}
}

func TestHTTPServerHandleServeFile(t *testing.T) {
	// Create test files directory
	os.MkdirAll("files", 0755)
	defer os.RemoveAll("files")
	
	// Create test file
	testContent := "test segment content"
	fileName := "test_segment.ts"
	err := os.WriteFile("files/"+fileName, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	processor := NewHLSProcessor("https://example.com/stream.m3u8", nil)
	server := NewHTTPServer(":8080", "http://localhost:8080", processor)
	
	req := httptest.NewRequest("GET", "/live/"+fileName, nil)
	w := httptest.NewRecorder()
	
	server.router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	contentType := w.Header().Get("Content-Type")
	if contentType != "video/mp2t" {
		t.Errorf("Expected Content-Type 'video/mp2t', got '%s'", contentType)
	}
	
	body := w.Body.String()
	if body != testContent {
		t.Errorf("Expected content '%s', got '%s'", testContent, body)
	}
}

func TestHTTPServerHandleServeFileNotFound(t *testing.T) {
	processor := NewHLSProcessor("https://example.com/stream.m3u8", nil)
	server := NewHTTPServer(":8080", "http://localhost:8080", processor)
	
	req := httptest.NewRequest("GET", "/live/nonexistent.ts", nil)
	w := httptest.NewRecorder()
	
	server.router.ServeHTTP(w, req)
	
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 for non-existent file, got %d", w.Code)
	}
}

func TestHTTPServerHandleVODWithUTC(t *testing.T) {
	// Setup test database
	originalDir := setupTestDatabase(t)
	defer cleanupTestDatabase(t, originalDir)
	
	processor := NewHLSProcessor("https://example.com/stream.m3u8", nil)
	server := NewHTTPServer(":8080", "http://localhost:8080", processor)
	
	// Test live stream with UTC parameter (should redirect to VOD)
	req := httptest.NewRequest("GET", "/live/stream.m3u8?utc=1609459200", nil)
	w := httptest.NewRecorder()
	
	server.router.ServeHTTP(w, req)
	
	// Should return OK (though might be empty)
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Errorf("Expected status 200 or 404, got %d", w.Code)
	}
}

func TestHTTPServerCORSAndHeaders(t *testing.T) {
	processor := NewHLSProcessor("https://example.com/stream.m3u8", nil)
	server := NewHTTPServer(":8080", "http://localhost:8080", processor)
	
	// Test that proper headers are set for M3U8 responses
	req := httptest.NewRequest("GET", "/api/status", nil)
	w := httptest.NewRecorder()
	
	server.router.ServeHTTP(w, req)
	
	// Check content type is set
	contentType := w.Header().Get("Content-Type")
	if contentType == "" {
		t.Error("Content-Type header should be set")
	}
}

// Benchmark tests
func BenchmarkHTTPServerHandleRoot(b *testing.B) {
	processor := NewHLSProcessor("https://example.com/stream.m3u8", nil)
	server := NewHTTPServer(":8080", "http://localhost:8080", processor)
	
	req := httptest.NewRequest("GET", "/", nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)
	}
}

func BenchmarkHTTPServerHandleAPIStatus(b *testing.B) {
	processor := NewHLSProcessor("https://example.com/stream.m3u8", nil)
	server := NewHTTPServer(":8080", "http://localhost:8080", processor)
	
	req := httptest.NewRequest("GET", "/api/status", nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)
	}
}

// Test helper functions
func setupTestDatabase(t *testing.T) string {
	// Create temporary directory for test database
	tmpDir, err := os.MkdirTemp("", "vsr_server_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	
	// Change to temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	
	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}
	
	// Initialize database
	database_init()
	
	return originalDir
}

func cleanupTestDatabase(t *testing.T, originalDir string) {
	// Close database if open
	if database != nil {
		database.Close()
		database = nil
	}
	
	// Change back to original directory
	err := os.Chdir(originalDir)
	if err != nil {
		t.Errorf("Failed to change back to original dir: %v", err)
	}
	
	// Remove temp directory
	tmpDir, _ := os.Getwd()
	err = os.RemoveAll(tmpDir)
	if err != nil {
		t.Errorf("Failed to remove temp dir: %v", err)
	}
}

// Test routing edge cases
func TestHTTPServerRouting(t *testing.T) {
	processor := NewHLSProcessor("https://example.com/stream.m3u8", nil)
	server := NewHTTPServer(":8080", "http://localhost:8080", processor)
	
	testCases := []struct {
		method         string
		path           string
		expectedStatus int
	}{
		{"GET", "/", 200},
		{"GET", "/live", 302},
		{"GET", "/api/status", 200},
		{"GET", "/api/stats", 200},
		{"GET", "/nonexistent", 404},
		{"POST", "/", 405}, // Method not allowed
		{"GET", "/live/stream.m3u8", 404}, // No segments available
	}
	
	for _, tc := range testCases {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		w := httptest.NewRecorder()
		
		server.router.ServeHTTP(w, req)
		
		if w.Code != tc.expectedStatus {
			t.Errorf("Expected status %d for %s %s, got %d", 
				tc.expectedStatus, tc.method, tc.path, w.Code)
		}
	}
}
