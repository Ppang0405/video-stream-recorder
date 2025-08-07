package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewPostProcessor(t *testing.T) {
	config := PostProcessorConfig{
		Enabled:       true,
		CloudUpload:   true,
		VideoStitch:   true,
		CleanupAfter:  24 * time.Hour,
		OutputFormats: []string{"mp4", "mkv"},
	}
	
	pp := NewPostProcessor(config)
	
	if !pp.enabled {
		t.Error("Expected post-processor to be enabled")
	}
	
	if !pp.cloudUpload {
		t.Error("Expected cloud upload to be enabled")
	}
	
	if !pp.videoStitch {
		t.Error("Expected video stitch to be enabled")
	}
	
	if pp.cleanupAfter != 24*time.Hour {
		t.Errorf("Expected cleanup after 24h, got %v", pp.cleanupAfter)
	}
	
	if len(pp.outputFormats) != 2 {
		t.Errorf("Expected 2 output formats, got %d", len(pp.outputFormats))
	}
	
	expectedFormats := map[string]bool{"mp4": true, "mkv": true}
	for _, format := range pp.outputFormats {
		if !expectedFormats[format] {
			t.Errorf("Unexpected output format: %s", format)
		}
	}
}

func TestNewPostProcessorDisabled(t *testing.T) {
	config := PostProcessorConfig{
		Enabled:       false,
		CloudUpload:   false,
		VideoStitch:   false,
		CleanupAfter:  0,
		OutputFormats: []string{},
	}
	
	pp := NewPostProcessor(config)
	
	if pp.enabled {
		t.Error("Expected post-processor to be disabled")
	}
	
	if pp.cloudUpload {
		t.Error("Expected cloud upload to be disabled")
	}
	
	if pp.videoStitch {
		t.Error("Expected video stitch to be disabled")
	}
}

func TestPostProcessorProcessSegmentDisabled(t *testing.T) {
	config := PostProcessorConfig{Enabled: false}
	pp := NewPostProcessor(config)
	
	item := &DatabaseItem{
		Name: "test_segment.ts",
		Len:  10.0,
		T:    time.Now().UnixNano(),
	}
	
	// Should return nil without processing when disabled
	err := pp.ProcessSegment("./files/test_segment.ts", item)
	if err != nil {
		t.Errorf("Expected no error from disabled processor, got: %v", err)
	}
}

func TestPostProcessorProcessSegmentEnabled(t *testing.T) {
	config := PostProcessorConfig{
		Enabled:     true,
		CloudUpload: true,
		VideoStitch: false,
	}
	pp := NewPostProcessor(config)
	
	item := &DatabaseItem{
		Name: "test_segment.ts",
		Len:  10.0,
		T:    time.Now().UnixNano(),
	}
	
	// Should process without error (even though cloud upload is mock)
	err := pp.ProcessSegment("./files/test_segment.ts", item)
	if err != nil {
		t.Errorf("Expected no error from enabled processor, got: %v", err)
	}
}

func TestPostProcessorProcessCompleteDisabled(t *testing.T) {
	config := PostProcessorConfig{Enabled: false}
	pp := NewPostProcessor(config)
	
	segments := []*DatabaseItem{
		{Name: "segment1.ts", Len: 10.0, T: time.Now().UnixNano()},
		{Name: "segment2.ts", Len: 10.0, T: time.Now().UnixNano()},
	}
	
	err := pp.ProcessComplete(segments, "./output/video.mp4")
	if err != nil {
		t.Errorf("Expected no error from disabled processor, got: %v", err)
	}
}

func TestPostProcessorProcessCompleteEnabled(t *testing.T) {
	config := PostProcessorConfig{
		Enabled:     true,
		VideoStitch: true,
	}
	pp := NewPostProcessor(config)
	
	segments := []*DatabaseItem{
		{Name: "segment1.ts", Len: 10.0, T: time.Now().UnixNano()},
		{Name: "segment2.ts", Len: 10.0, T: time.Now().UnixNano()},
		{Name: "segment3.ts", Len: 10.0, T: time.Now().UnixNano()},
	}
	
	err := pp.ProcessComplete(segments, "./output/video.mp4")
	if err != nil {
		t.Errorf("Expected no error from enabled processor, got: %v", err)
	}
}

func TestPostProcessorCleanupOldFilesDisabled(t *testing.T) {
	config := PostProcessorConfig{Enabled: false}
	pp := NewPostProcessor(config)
	
	err := pp.CleanupOldFiles("./files", 24)
	if err != nil {
		t.Errorf("Expected no error from disabled cleanup, got: %v", err)
	}
}

func TestPostProcessorCleanupOldFiles(t *testing.T) {
	// Create temporary directory structure
	tmpDir, err := os.MkdirTemp("", "postprocessor_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	filesDir := filepath.Join(tmpDir, "files")
	err = os.MkdirAll(filesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create files dir: %v", err)
	}
	
	// Create test files with different ages
	now := time.Now()
	
	// Old file (should be deleted)
	oldFile := filepath.Join(filesDir, "old_segment.ts")
	err = os.WriteFile(oldFile, []byte("old content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create old file: %v", err)
	}
	
	// Set old modification time (25 hours ago)
	oldTime := now.Add(-25 * time.Hour)
	err = os.Chtimes(oldFile, oldTime, oldTime)
	if err != nil {
		t.Fatalf("Failed to set old file time: %v", err)
	}
	
	// Recent file (should be kept)
	recentFile := filepath.Join(filesDir, "recent_segment.ts")
	err = os.WriteFile(recentFile, []byte("recent content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create recent file: %v", err)
	}
	
	// Set recent modification time (12 hours ago)
	recentTime := now.Add(-12 * time.Hour)
	err = os.Chtimes(recentFile, recentTime, recentTime)
	if err != nil {
		t.Fatalf("Failed to set recent file time: %v", err)
	}
	
	config := PostProcessorConfig{Enabled: true}
	pp := NewPostProcessor(config)
	
	// Run cleanup with 24 hour retention
	err = pp.CleanupOldFiles(filesDir, 24)
	if err != nil {
		t.Errorf("Cleanup failed: %v", err)
	}
	
	// Check that old file was deleted
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Error("Old file should have been deleted")
	}
	
	// Check that recent file still exists
	if _, err := os.Stat(recentFile); os.IsNotExist(err) {
		t.Error("Recent file should still exist")
	}
}

func TestPostProcessorCleanupOldFilesNonExistentDir(t *testing.T) {
	config := PostProcessorConfig{Enabled: true}
	pp := NewPostProcessor(config)
	
	// Should handle non-existent directory gracefully
	err := pp.CleanupOldFiles("/non/existent/directory", 24)
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}
}

func TestPostProcessorGetJobStatus(t *testing.T) {
	config := PostProcessorConfig{Enabled: true}
	pp := NewPostProcessor(config)
	
	// Should return error since job tracking is not implemented
	job, err := pp.GetJobStatus("test-job-123")
	if err == nil {
		t.Error("Expected error for unimplemented job status")
	}
	
	if job != nil {
		t.Error("Expected nil job for unimplemented feature")
	}
}

func TestPostProcessorListJobs(t *testing.T) {
	config := PostProcessorConfig{Enabled: true}
	pp := NewPostProcessor(config)
	
	// Should return error since job listing is not implemented
	jobs, err := pp.ListJobs()
	if err == nil {
		t.Error("Expected error for unimplemented job listing")
	}
	
	if jobs != nil {
		t.Error("Expected nil jobs for unimplemented feature")
	}
}

func TestPostProcessJob(t *testing.T) {
	job := &PostProcessJob{
		ID:         "test-job-123",
		Type:       "upload",
		Files:      []string{"segment1.ts", "segment2.ts"},
		StartTime:  time.Now(),
		EndTime:    time.Now().Add(5 * time.Minute),
		OutputPath: "/output/video.mp4",
		Status:     "completed",
		Metadata:   map[string]interface{}{"quality": "720p"},
	}
	
	// Test field values
	if job.ID != "test-job-123" {
		t.Errorf("Expected ID 'test-job-123', got '%s'", job.ID)
	}
	
	if job.Type != "upload" {
		t.Errorf("Expected type 'upload', got '%s'", job.Type)
	}
	
	if len(job.Files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(job.Files))
	}
	
	if job.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", job.Status)
	}
	
	if job.Metadata["quality"] != "720p" {
		t.Errorf("Expected quality '720p', got %v", job.Metadata["quality"])
	}
}

func TestPostProcessorConfig(t *testing.T) {
	config := PostProcessorConfig{
		Enabled:       true,
		CloudUpload:   true,
		VideoStitch:   false,
		CleanupAfter:  48 * time.Hour,
		OutputFormats: []string{"mp4", "mkv", "avi"},
	}
	
	// Test configuration values
	if !config.Enabled {
		t.Error("Expected config to be enabled")
	}
	
	if !config.CloudUpload {
		t.Error("Expected cloud upload to be enabled")
	}
	
	if config.VideoStitch {
		t.Error("Expected video stitch to be disabled")
	}
	
	if config.CleanupAfter != 48*time.Hour {
		t.Errorf("Expected cleanup after 48h, got %v", config.CleanupAfter)
	}
	
	if len(config.OutputFormats) != 3 {
		t.Errorf("Expected 3 output formats, got %d", len(config.OutputFormats))
	}
}

func TestPostProcessorUploadToCloud(t *testing.T) {
	config := PostProcessorConfig{Enabled: true, CloudUpload: true}
	pp := NewPostProcessor(config)
	
	item := &DatabaseItem{
		Name: "test_segment.ts",
		Len:  10.0,
		T:    time.Now().UnixNano(),
	}
	
	// Test cloud upload (currently a placeholder)
	err := pp.uploadToCloud("./files/test_segment.ts", item)
	if err != nil {
		t.Errorf("Expected no error from mock cloud upload, got: %v", err)
	}
}

func TestPostProcessorStitchVideo(t *testing.T) {
	config := PostProcessorConfig{Enabled: true, VideoStitch: true}
	pp := NewPostProcessor(config)
	
	segments := []*DatabaseItem{
		{Name: "segment1.ts", Len: 10.0, T: time.Now().UnixNano()},
		{Name: "segment2.ts", Len: 10.0, T: time.Now().UnixNano()},
	}
	
	// Test video stitching (currently a placeholder)
	err := pp.stitchVideo(segments, "./output/video.mp4")
	if err != nil {
		t.Errorf("Expected no error from mock video stitching, got: %v", err)
	}
}

func TestPostProcessorCleanupEdgeCases(t *testing.T) {
	config := PostProcessorConfig{Enabled: true}
	pp := NewPostProcessor(config)
	
	// Test with zero retention (should delete everything)
	tmpDir, err := os.MkdirTemp("", "postprocessor_edge_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	filesDir := filepath.Join(tmpDir, "files")
	err = os.MkdirAll(filesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create files dir: %v", err)
	}
	
	// Create test file
	testFile := filepath.Join(filesDir, "test_segment.ts")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Run cleanup with 0 hour retention
	err = pp.CleanupOldFiles(filesDir, 0)
	if err != nil {
		t.Errorf("Cleanup failed: %v", err)
	}
	
	// File should be deleted
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("File should have been deleted with 0 hour retention")
	}
}

// Benchmark tests
func BenchmarkPostProcessorProcessSegment(b *testing.B) {
	config := PostProcessorConfig{Enabled: true, CloudUpload: true}
	pp := NewPostProcessor(config)
	
	item := &DatabaseItem{
		Name: "benchmark_segment.ts",
		Len:  10.0,
		T:    time.Now().UnixNano(),
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pp.ProcessSegment("./files/benchmark_segment.ts", item)
	}
}

func BenchmarkPostProcessorProcessComplete(b *testing.B) {
	config := PostProcessorConfig{Enabled: true, VideoStitch: true}
	pp := NewPostProcessor(config)
	
	segments := []*DatabaseItem{
		{Name: "segment1.ts", Len: 10.0, T: time.Now().UnixNano()},
		{Name: "segment2.ts", Len: 10.0, T: time.Now().UnixNano()},
		{Name: "segment3.ts", Len: 10.0, T: time.Now().UnixNano()},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pp.ProcessComplete(segments, "./output/video.mp4")
	}
}

func BenchmarkNewPostProcessor(b *testing.B) {
	config := PostProcessorConfig{
		Enabled:       true,
		CloudUpload:   true,
		VideoStitch:   true,
		CleanupAfter:  24 * time.Hour,
		OutputFormats: []string{"mp4", "mkv"},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewPostProcessor(config)
	}
}

// Test with subdirectories
func TestPostProcessorCleanupSubdirectories(t *testing.T) {
	// Create temporary directory structure with subdirectories
	tmpDir, err := os.MkdirTemp("", "postprocessor_subdir_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	filesDir := filepath.Join(tmpDir, "files")
	subDir := filepath.Join(filesDir, "subdir")
	err = os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	
	// Create files in main directory and subdirectory
	now := time.Now()
	oldTime := now.Add(-25 * time.Hour)
	
	// Old file in main directory
	oldMainFile := filepath.Join(filesDir, "old_main.ts")
	err = os.WriteFile(oldMainFile, []byte("old main content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create old main file: %v", err)
	}
	err = os.Chtimes(oldMainFile, oldTime, oldTime)
	if err != nil {
		t.Fatalf("Failed to set old main file time: %v", err)
	}
	
	// Old file in subdirectory
	oldSubFile := filepath.Join(subDir, "old_sub.ts")
	err = os.WriteFile(oldSubFile, []byte("old sub content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create old sub file: %v", err)
	}
	err = os.Chtimes(oldSubFile, oldTime, oldTime)
	if err != nil {
		t.Fatalf("Failed to set old sub file time: %v", err)
	}
	
	config := PostProcessorConfig{Enabled: true}
	pp := NewPostProcessor(config)
	
	// Run cleanup
	err = pp.CleanupOldFiles(filesDir, 24)
	if err != nil {
		t.Errorf("Cleanup failed: %v", err)
	}
	
	// Both old files should be deleted
	if _, err := os.Stat(oldMainFile); !os.IsNotExist(err) {
		t.Error("Old main file should have been deleted")
	}
	
	if _, err := os.Stat(oldSubFile); !os.IsNotExist(err) {
		t.Error("Old sub file should have been deleted")
	}
	
	// Subdirectory should still exist (even if empty)
	if _, err := os.Stat(subDir); os.IsNotExist(err) {
		t.Error("Subdirectory should still exist")
	}
}
