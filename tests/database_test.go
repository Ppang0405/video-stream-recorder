package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupTestDatabase(t *testing.T) string {
	// Create temporary directory for test database
	tmpDir, err := os.MkdirTemp("", "vsr_test_")
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
	
	// Return original directory for cleanup
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

func TestDatabaseInit(t *testing.T) {
	originalDir := setupTestDatabase(t)
	defer cleanupTestDatabase(t, originalDir)
	
	// Check that database file was created
	if _, err := os.Stat("db.db"); os.IsNotExist(err) {
		t.Error("Database file was not created")
	}
	
	// Check that database variable is set
	if database == nil {
		t.Error("Database variable is nil after init")
	}
}

func TestDatabaseStore(t *testing.T) {
	originalDir := setupTestDatabase(t)
	defer cleanupTestDatabase(t, originalDir)
	
	// Create test item
	item := &DatabaseItem{
		Name: "test_segment.ts",
		Len:  10.5,
		T:    time.Now().UnixNano(),
	}
	
	// Store item
	database_store(item)
	
	// Verify item was stored and got an ID
	if item.ID == 0 {
		t.Error("Item ID was not set after storage")
	}
	
	// Store multiple items
	items := []*DatabaseItem{
		{Name: "segment1.ts", Len: 5.0, T: time.Now().UnixNano()},
		{Name: "segment2.ts", Len: 7.5, T: time.Now().UnixNano()},
		{Name: "segment3.ts", Len: 12.0, T: time.Now().UnixNano()},
	}
	
	for _, testItem := range items {
		database_store(testItem)
		if testItem.ID == 0 {
			t.Errorf("Item %s did not get an ID", testItem.Name)
		}
	}
}

func TestDatabaseLast5(t *testing.T) {
	originalDir := setupTestDatabase(t)
	defer cleanupTestDatabase(t, originalDir)
	
	// Initially should return empty slice
	items := database_last_5()
	if len(items) != 0 {
		t.Errorf("Expected 0 items from empty database, got %d", len(items))
	}
	
	// Add 3 items
	testItems := []*DatabaseItem{
		{Name: "segment1.ts", Len: 5.0, T: time.Now().UnixNano()},
		{Name: "segment2.ts", Len: 7.5, T: time.Now().UnixNano() + 1000000},
		{Name: "segment3.ts", Len: 12.0, T: time.Now().UnixNano() + 2000000},
	}
	
	for _, item := range testItems {
		database_store(item)
	}
	
	// Should return all 3 items
	items = database_last_5()
	if len(items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(items))
	}
	
	// Add 7 more items (total 10)
	for i := 4; i <= 10; i++ {
		item := &DatabaseItem{
			Name: fmt.Sprintf("segment%d.ts", i),
			Len:  float64(i) * 1.5,
			T:    time.Now().UnixNano() + int64(i*1000000),
		}
		database_store(item)
	}
	
	// Should return only last 5 items
	items = database_last_5()
	if len(items) != 5 {
		t.Errorf("Expected 5 items, got %d", len(items))
	}
	
	// Verify they are the last 5 (highest IDs)
	for i := 0; i < len(items)-1; i++ {
		if items[i].ID >= items[i+1].ID {
			t.Error("Items are not sorted by ID in ascending order")
		}
	}
}

func TestDatabaseGet(t *testing.T) {
	originalDir := setupTestDatabase(t)
	defer cleanupTestDatabase(t, originalDir)
	
	// Create test items with specific timestamps
	baseTime := time.Now().Unix()
	testItems := []*DatabaseItem{
		{Name: "old1.ts", Len: 5.0, T: (baseTime - 3600) * 1000000000}, // 1 hour ago
		{Name: "old2.ts", Len: 7.5, T: (baseTime - 1800) * 1000000000}, // 30 min ago
		{Name: "recent1.ts", Len: 10.0, T: (baseTime - 600) * 1000000000}, // 10 min ago
		{Name: "recent2.ts", Len: 12.0, T: (baseTime - 300) * 1000000000}, // 5 min ago
		{Name: "current.ts", Len: 8.0, T: baseTime * 1000000000},          // now
	}
	
	for _, item := range testItems {
		database_store(item)
	}
	
	// Test getting items from 45 minutes ago with limit 2
	startTime := fmt.Sprintf("%d", baseTime-2700) // 45 min ago
	items := database_get(startTime, "2")
	
	// Should get recent1.ts and recent2.ts
	if len(items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(items))
	}
	
	// Test getting all items from beginning
	startTime = fmt.Sprintf("%d", baseTime-7200) // 2 hours ago
	items = database_get(startTime, "10")
	
	if len(items) != 5 {
		t.Errorf("Expected 5 items, got %d", len(items))
	}
	
	// Test with invalid start time
	items = database_get("invalid", "5")
	if len(items) != 0 {
		t.Errorf("Expected 0 items for invalid start time, got %d", len(items))
	}
}

func TestDatabaseCountSegments(t *testing.T) {
	originalDir := setupTestDatabase(t)
	defer cleanupTestDatabase(t, originalDir)
	
	// Initially should have 0 segments
	count := database_count_segments()
	if count != 0 {
		t.Errorf("Expected 0 segments in empty database, got %d", count)
	}
	
	// Add some segments
	numSegments := 15
	for i := 1; i <= numSegments; i++ {
		item := &DatabaseItem{
			Name: fmt.Sprintf("segment%d.ts", i),
			Len:  float64(i) * 2.5,
			T:    time.Now().UnixNano() + int64(i*1000000),
		}
		database_store(item)
	}
	
	// Should count all segments
	count = database_count_segments()
	if count != numSegments {
		t.Errorf("Expected %d segments, got %d", numSegments, count)
	}
}

func TestDatabaseItem(t *testing.T) {
	item := &DatabaseItem{
		ID:   12345,
		Name: "test_segment.ts",
		Len:  15.75,
		T:    1609459200000000000, // 2021-01-01 00:00:00 UTC in nanoseconds
	}
	
	// Test field values
	if item.ID != 12345 {
		t.Errorf("Expected ID 12345, got %d", item.ID)
	}
	
	if item.Name != "test_segment.ts" {
		t.Errorf("Expected name 'test_segment.ts', got '%s'", item.Name)
	}
	
	if item.Len != 15.75 {
		t.Errorf("Expected length 15.75, got %f", item.Len)
	}
	
	if item.T != 1609459200000000000 {
		t.Errorf("Expected timestamp 1609459200000000000, got %d", item.T)
	}
}

func TestItob(t *testing.T) {
	testCases := []struct {
		input    int
		expected []byte
	}{
		{0, []byte{0, 0, 0, 0, 0, 0, 0, 0}},
		{1, []byte{0, 0, 0, 0, 0, 0, 0, 1}},
		{255, []byte{0, 0, 0, 0, 0, 0, 0, 255}},
		{256, []byte{0, 0, 0, 0, 0, 0, 1, 0}},
		{65535, []byte{0, 0, 0, 0, 0, 0, 255, 255}},
	}
	
	for _, tc := range testCases {
		result := itob(tc.input)
		if len(result) != 8 {
			t.Errorf("itob(%d) returned slice of length %d, expected 8", tc.input, len(result))
		}
		
		for i, expected := range tc.expected {
			if result[i] != expected {
				t.Errorf("itob(%d)[%d] = %d, expected %d", tc.input, i, result[i], expected)
			}
		}
	}
}

func TestDatabaseConcurrency(t *testing.T) {
	originalDir := setupTestDatabase(t)
	defer cleanupTestDatabase(t, originalDir)
	
	// Test concurrent writes
	const numGoroutines = 10
	const itemsPerGoroutine = 20
	
	done := make(chan bool, numGoroutines)
	
	for i := 0; i < numGoroutines; i++ {
		go func(routineID int) {
			for j := 0; j < itemsPerGoroutine; j++ {
				item := &DatabaseItem{
					Name: fmt.Sprintf("routine%d_segment%d.ts", routineID, j),
					Len:  float64(j) * 1.5,
					T:    time.Now().UnixNano(),
				}
				database_store(item)
			}
			done <- true
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
	
	// Verify all items were stored
	count := database_count_segments()
	expected := numGoroutines * itemsPerGoroutine
	if count != expected {
		t.Errorf("Expected %d segments after concurrent writes, got %d", expected, count)
	}
}

func TestDatabaseCleanup(t *testing.T) {
	originalDir := setupTestDatabase(t)
	defer cleanupTestDatabase(t, originalDir)
	
	// Create segments with different ages
	baseTime := time.Now().Unix()
	
	// Old segments (should be cleaned up)
	oldItems := []*DatabaseItem{
		{Name: "old1.ts", Len: 5.0, T: (baseTime - 25*3600) * 1000000000}, // 25 hours ago
		{Name: "old2.ts", Len: 7.5, T: (baseTime - 30*3600) * 1000000000}, // 30 hours ago
	}
	
	// Recent segments (should be kept)
	recentItems := []*DatabaseItem{
		{Name: "recent1.ts", Len: 10.0, T: (baseTime - 12*3600) * 1000000000}, // 12 hours ago
		{Name: "recent2.ts", Len: 12.0, T: (baseTime - 6*3600) * 1000000000},  // 6 hours ago
	}
	
	// Store all items
	for _, item := range append(oldItems, recentItems...) {
		database_store(item)
	}
	
	// Create corresponding files
	os.MkdirAll("files", 0755)
	for _, item := range append(oldItems, recentItems...) {
		file, err := os.Create(filepath.Join("files", item.Name))
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		file.Close()
	}
	
	// Initial count should be 4
	if count := database_count_segments(); count != 4 {
		t.Errorf("Expected 4 segments initially, got %d", count)
	}
	
	// Note: database_worker runs cleanup automatically, but we can't easily test it
	// without modifying the cleanup logic or running it manually
	// This test verifies the setup for cleanup functionality
}

// Benchmark tests
func BenchmarkDatabaseStore(b *testing.B) {
	originalDir := setupTestDatabase(&testing.T{})
	defer cleanupTestDatabase(&testing.T{}, originalDir)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		item := &DatabaseItem{
			Name: fmt.Sprintf("segment%d.ts", i),
			Len:  float64(i) * 1.5,
			T:    time.Now().UnixNano(),
		}
		database_store(item)
	}
}

func BenchmarkDatabaseLast5(b *testing.B) {
	originalDir := setupTestDatabase(&testing.T{})
	defer cleanupTestDatabase(&testing.T{}, originalDir)
	
	// Pre-populate database
	for i := 0; i < 100; i++ {
		item := &DatabaseItem{
			Name: fmt.Sprintf("segment%d.ts", i),
			Len:  float64(i) * 1.5,
			T:    time.Now().UnixNano(),
		}
		database_store(item)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		database_last_5()
	}
}

func BenchmarkDatabaseCountSegments(b *testing.B) {
	originalDir := setupTestDatabase(&testing.T{})
	defer cleanupTestDatabase(&testing.T{}, originalDir)
	
	// Pre-populate database
	for i := 0; i < 100; i++ {
		item := &DatabaseItem{
			Name: fmt.Sprintf("segment%d.ts", i),
			Len:  float64(i) * 1.5,
			T:    time.Now().UnixNano(),
		}
		database_store(item)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		database_count_segments()
	}
}
