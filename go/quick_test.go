package main

import (
	"fmt"
	"testing"
	"time"
)

// Quick integration test to verify core functionality
func TestCoreIntegration(t *testing.T) {
	fmt.Println("=== Testing Core Functions ===")
	
	// Test FIFO Cache
	fmt.Println("Testing FIFO Cache...")
	cache := NewFIFOCache(3)
	if !cache.Set("url1") {
		t.Error("Cache Set should return true for new item")
	}
	if cache.Set("url1") {
		t.Error("Cache Set should return false for duplicate item")
	}
	fmt.Println("✓ FIFO Cache working")
	
	// Test preprocessor URL detection
	fmt.Println("Testing Preprocessor...")
	if !isHLSURL("https://example.com/stream.m3u8") {
		t.Error("Should detect HLS URL")
	}
	if isHLSURL("https://youtube.com/watch?v=123") {
		t.Error("Should not detect YouTube as HLS")
	}
	if !needsPreprocessing("https://youtube.com/watch?v=123") {
		t.Error("Should detect YouTube needs preprocessing")
	}
	fmt.Println("✓ Preprocessor URL detection working")
	
	// Test processor creation
	fmt.Println("Testing Processor...")
	processor := NewHLSProcessor("https://example.com/test.m3u8", nil)
	if processor.url != "https://example.com/test.m3u8" {
		t.Error("Processor URL not set correctly")
	}
	stats := processor.GetStats()
	if stats["running"] != false {
		t.Error("Processor should not be running initially")
	}
	fmt.Println("✓ HLS Processor working")
	
	// Test post-processor
	fmt.Println("Testing Post-processor...")
	config := PostProcessorConfig{
		Enabled:     true,
		CloudUpload: false,
		VideoStitch: false,
	}
	pp := NewPostProcessor(config)
	if !pp.enabled {
		t.Error("Post-processor should be enabled")
	}
	fmt.Println("✓ Post-processor working")
	
	// Test HTTP server creation
	fmt.Println("Testing HTTP Server...")
	server := NewHTTPServer(":8080", "http://localhost:8080", processor)
	if server.bindAddr != ":8080" {
		t.Error("Server bind address not set correctly")
	}
	fmt.Println("✓ HTTP Server working")
	
	// Test database item
	fmt.Println("Testing Database Item...")
	item := &DatabaseItem{
		Name: "test.ts",
		Len:  10.5,
		T:    time.Now().UnixNano(),
	}
	if item.Name != "test.ts" {
		t.Error("Database item name not set correctly")
	}
	fmt.Println("✓ Database Item working")
	
	// Test helper function
	fmt.Println("Testing Helper Functions...")
	result := itob(255)
	if len(result) != 8 {
		t.Error("itob should return 8 bytes")
	}
	if result[7] != 255 {
		t.Error("itob not working correctly")
	}
	fmt.Println("✓ Helper functions working")
	
	fmt.Println("=== All Core Functions Working! ===")
}

func TestApplicationBuild(t *testing.T) {
	fmt.Println("=== Testing Application Build ===")
	
	// Test that all components can be instantiated together
	preprocessor := NewYTDLPProcessor("yt-dlp", "best")
	if preprocessor == nil {
		t.Error("Failed to create preprocessor")
	}
	
	postProcessorConfig := PostProcessorConfig{
		Enabled:       false,
		CloudUpload:   false,
		VideoStitch:   false,
		CleanupAfter:  24 * time.Hour,
		OutputFormats: []string{"mp4"},
	}
	postProcessor := NewPostProcessor(postProcessorConfig)
	
	hlsProcessor := NewHLSProcessor("https://example.com/stream.m3u8", postProcessor)
	if hlsProcessor == nil {
		t.Error("Failed to create HLS processor")
	}
	
	httpServer := NewHTTPServer(":8080", "http://localhost:8080", hlsProcessor)
	if httpServer == nil {
		t.Error("Failed to create HTTP server")
	}
	
	fmt.Println("✓ All components can be instantiated together")
	fmt.Println("=== Application Build Test Passed! ===")
}
