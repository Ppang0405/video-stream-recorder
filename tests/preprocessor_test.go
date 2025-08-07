package main

import (
	"os"
	"strings"
	"testing"
	"time"
)

// Mock data for testing
var mockYTDLPOutput = `{
	"id": "test123",
	"title": "Test Video",
	"url": "https://example.com/video",
	"duration": 120.5,
	"is_live": false,
	"thumbnail": "https://example.com/thumb.jpg",
	"formats": [
		{
			"format_id": "96",
			"url": "https://example.com/stream.m3u8",
			"protocol": "m3u8_native",
			"ext": "mp4",
			"resolution": "1280x720",
			"width": 1280,
			"height": 720,
			"filesize": 1024000,
			"tbr": 1500.0
		},
		{
			"format_id": "95",
			"url": "https://example.com/stream2.m3u8",
			"protocol": "m3u8",
			"ext": "mp4",
			"resolution": "854x480",
			"width": 854,
			"height": 480,
			"filesize": 512000,
			"tbr": 800.0
		}
	]
}`

func TestNewYTDLPProcessor(t *testing.T) {
	processor := NewYTDLPProcessor("/usr/bin/yt-dlp", "best")
	
	if processor.binaryPath != "/usr/bin/yt-dlp" {
		t.Errorf("Expected binary path '/usr/bin/yt-dlp', got '%s'", processor.binaryPath)
	}
	
	if processor.quality != "best" {
		t.Errorf("Expected quality 'best', got '%s'", processor.quality)
	}
	
	if processor.timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", processor.timeout)
	}
}

func TestIsHLSURL(t *testing.T) {
	testCases := []struct {
		url      string
		expected bool
	}{
		{"https://example.com/stream.m3u8", true},
		{"http://test.com/video.M3U8", true},
		{"https://example.com/playlist.m3u8?token=abc", true},
		{"https://youtube.com/watch?v=123", false},
		{"https://example.com/video.mp4", false},
		{"https://twitch.tv/user", false},
		{"", false},
	}
	
	for _, tc := range testCases {
		result := isHLSURL(tc.url)
		if result != tc.expected {
			t.Errorf("isHLSURL(%s) = %v, expected %v", tc.url, result, tc.expected)
		}
	}
}

func TestNeedsPreprocessing(t *testing.T) {
	testCases := []struct {
		url      string
		expected bool
	}{
		// Should NOT need preprocessing (already HLS)
		{"https://example.com/stream.m3u8", false},
		{"http://test.com/video.M3U8", false},
		
		// Should need preprocessing (platform URLs)
		{"https://youtube.com/watch?v=123", true},
		{"https://www.youtube.com/watch?v=456", true},
		{"https://youtu.be/789", true},
		{"https://twitch.tv/streamer", true},
		{"https://vimeo.com/123456", true},
		{"https://facebook.com/video", true},
		{"https://instagram.com/p/123", true},
		{"https://tiktok.com/@user/video/123", true},
		{"https://twitter.com/user/status/123", true},
		{"https://x.com/user/status/123", true},
		
		// Should NOT need preprocessing (unknown platforms)
		{"https://example.com/video.mp4", false},
		{"https://unknown-site.com/stream", false},
		{"", false},
	}
	
	for _, tc := range testCases {
		result := needsPreprocessing(tc.url)
		if result != tc.expected {
			t.Errorf("needsPreprocessing(%s) = %v, expected %v", tc.url, result, tc.expected)
		}
	}
}

func TestProcessURL(t *testing.T) {
	// Test with preprocessing disabled
	t.Run("PreprocessingDisabled", func(t *testing.T) {
		url := "https://youtube.com/watch?v=123"
		result, err := ProcessURL(nil, url, false)
		
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		
		if result != url {
			t.Errorf("Expected URL unchanged when preprocessing disabled, got %s", result)
		}
	})
	
	// Test with HLS URL (no preprocessing needed)
	t.Run("HLSURLNoPreprocessing", func(t *testing.T) {
		processor := NewYTDLPProcessor("yt-dlp", "best")
		url := "https://example.com/stream.m3u8"
		result, err := ProcessURL(processor, url, true)
		
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		
		if result != url {
			t.Errorf("Expected HLS URL unchanged, got %s", result)
		}
	})
}

func TestStreamInfo(t *testing.T) {
	info := StreamInfo{
		HLSURL:     "https://example.com/stream.m3u8",
		Title:      "Test Stream",
		Duration:   120.5,
		Resolution: "1280x720",
		IsLive:     true,
	}
	
	if info.HLSURL != "https://example.com/stream.m3u8" {
		t.Errorf("Expected HLSURL 'https://example.com/stream.m3u8', got '%s'", info.HLSURL)
	}
	
	if info.Title != "Test Stream" {
		t.Errorf("Expected title 'Test Stream', got '%s'", info.Title)
	}
	
	if info.Duration != 120.5 {
		t.Errorf("Expected duration 120.5, got %f", info.Duration)
	}
	
	if !info.IsLive {
		t.Error("Expected IsLive to be true")
	}
}

func TestYTDLPProcessorExtractHLSURL(t *testing.T) {
	processor := NewYTDLPProcessor("yt-dlp", "best")
	
	// Test with direct HLS URL
	t.Run("DirectHLSURL", func(t *testing.T) {
		url := "https://example.com/stream.m3u8"
		result, err := processor.ExtractHLSURL(url)
		
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		
		if result.HLSURL != url {
			t.Errorf("Expected HLSURL %s, got %s", url, result.HLSURL)
		}
		
		if result.Title != "Direct HLS Stream" {
			t.Errorf("Expected title 'Direct HLS Stream', got '%s'", result.Title)
		}
	})
}

// TestYTDLPAvailability tests if yt-dlp is available on the system
func TestYTDLPAvailability(t *testing.T) {
	// Only run this test if yt-dlp is available
	if _, err := os.Stat("/usr/bin/yt-dlp"); os.IsNotExist(err) {
		if _, err := os.Stat("/usr/local/bin/yt-dlp"); os.IsNotExist(err) {
			if _, err := os.Stat("/opt/homebrew/bin/yt-dlp"); os.IsNotExist(err) {
				t.Skip("yt-dlp not found, skipping test")
			}
		}
	}
	
	// Test with actual yt-dlp binary
	err := TestYTDLP("yt-dlp")
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			t.Skip("yt-dlp not available in PATH, skipping test")
		}
		t.Errorf("yt-dlp test failed: %v", err)
	}
}

// Benchmark tests
func BenchmarkIsHLSURL(b *testing.B) {
	url := "https://example.com/stream.m3u8"
	for i := 0; i < b.N; i++ {
		isHLSURL(url)
	}
}

func BenchmarkNeedsPreprocessing(b *testing.B) {
	url := "https://youtube.com/watch?v=123"
	for i := 0; i < b.N; i++ {
		needsPreprocessing(url)
	}
}

func BenchmarkNewYTDLPProcessor(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewYTDLPProcessor("yt-dlp", "best")
	}
}

// Helper function to create test processor
func createTestProcessor() *YTDLPProcessor {
	return NewYTDLPProcessor("echo", "best") // Use echo as a safe mock
}

// Test error handling
func TestYTDLPProcessorErrors(t *testing.T) {
	// Test with non-existent binary
	processor := NewYTDLPProcessor("/non/existent/binary", "best")
	
	t.Run("NonExistentBinary", func(t *testing.T) {
		_, err := processor.ExtractVideoInfo("https://youtube.com/watch?v=123")
		if err == nil {
			t.Error("Expected error for non-existent binary")
		}
	})
	
	t.Run("TestYTDLPNonExistent", func(t *testing.T) {
		err := TestYTDLP("/non/existent/binary")
		if err == nil {
			t.Error("Expected error for non-existent yt-dlp binary")
		}
	})
}

// Integration test placeholder
func TestPreprocessorIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	// This would test actual yt-dlp integration
	// Only run if yt-dlp is available and we have test URLs
	t.Skip("Integration test requires actual yt-dlp and test URLs")
}
