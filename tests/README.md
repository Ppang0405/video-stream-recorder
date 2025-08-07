# VSR Test Suite

Comprehensive test suite for the Video Stream Recorder (VSR) project. This test suite covers all major components and functionalities of the VSR application.

## 📋 Test Coverage

### Component Tests

| Component | Test File | Coverage |
|-----------|-----------|----------|
| **Preprocessor** | `preprocessor_test.go` | yt-dlp integration, URL detection, stream extraction |
| **FIFO Cache** | `fifocache_test.go` | Cache operations, concurrency, memory management |
| **Database** | `database_test.go` | BoltDB operations, segment storage, retrieval |
| **HLS Processor** | `processor_test.go` | Stream processing, segment downloading, cache integration |
| **HTTP Server** | `server_test.go` | API endpoints, routing, response handling |
| **Post-processor** | `postprocessor_test.go` | File cleanup, upload simulation, video stitching |

### Test Types

- **Unit Tests**: Individual function and method testing
- **Integration Tests**: Component interaction testing
- **Benchmark Tests**: Performance measurement
- **Concurrent Tests**: Thread safety and race condition testing
- **Error Handling Tests**: Edge cases and error scenarios

## 🚀 Running Tests

### Quick Start

```bash
# Run all tests
cd tests
./run_tests.sh

# Run tests with benchmarks
./run_tests.sh --bench

# Run specific test file manually
cd ../go
go test -v preprocessor_test.go *.go
```

### Manual Test Execution

```bash
# Run individual test suites
cd ../go

# Preprocessor tests
go test -v -run="TestPreprocessor.*" preprocessor_test.go *.go

# Cache tests
go test -v -run="TestFIFOCache.*" fifocache_test.go *.go

# Database tests
go test -v -run="TestDatabase.*" database_test.go *.go

# Server tests
go test -v -run="TestHTTPServer.*" server_test.go *.go

# Processor tests
go test -v -run="TestHLSProcessor.*" processor_test.go *.go

# Post-processor tests
go test -v -run="TestPostProcessor.*" postprocessor_test.go *.go
```

### Benchmark Tests

```bash
# Run all benchmarks
go test -bench=. -run=^$ *_test.go *.go

# Run specific benchmark
go test -bench=BenchmarkFIFOCache fifocache_test.go *.go
```

### Coverage Reports

```bash
# Generate coverage report
go test -coverprofile=coverage.out *.go
go tool cover -html=coverage.out -o coverage.html

# View coverage in terminal
go tool cover -func=coverage.out
```

## 📊 Test Details

### Preprocessor Tests (`preprocessor_test.go`)

**Features Tested:**
- ✅ yt-dlp processor initialization
- ✅ URL detection and classification
- ✅ HLS stream extraction
- ✅ Video metadata parsing
- ✅ Error handling for invalid URLs
- ✅ Platform-specific URL processing
- ✅ Quality selection logic

**Mock Data:**
- Sample JSON responses from yt-dlp
- Various URL formats (YouTube, Twitch, direct HLS)
- Error scenarios and edge cases

### FIFO Cache Tests (`fifocache_test.go`)

**Features Tested:**
- ✅ Cache initialization and configuration
- ✅ Set/Get operations
- ✅ FIFO eviction policy
- ✅ Concurrent access safety
- ✅ Memory bounds enforcement
- ✅ Edge cases (zero size, duplicates)

**Stress Tests:**
- High-concurrency operations (10 goroutines × 50 items)
- Memory usage validation
- Performance benchmarks

### Database Tests (`database_test.go`)

**Features Tested:**
- ✅ BoltDB initialization
- ✅ Segment storage and retrieval
- ✅ Time-based queries
- ✅ Cleanup operations
- ✅ Concurrent access
- ✅ Data integrity

**Test Environment:**
- Temporary databases for isolation
- Automated cleanup
- Various data scenarios

### HLS Processor Tests (`processor_test.go`)

**Features Tested:**
- ✅ HLS stream fetching
- ✅ M3U8 playlist parsing
- ✅ Segment downloading
- ✅ Cache integration
- ✅ Error handling
- ✅ Statistics reporting

**Mock Services:**
- HTTP test servers
- Sample HLS playlists
- Segment data simulation

### HTTP Server Tests (`server_test.go`)

**Features Tested:**
- ✅ Route handling
- ✅ Live stream endpoints
- ✅ VOD playlist generation
- ✅ File serving
- ✅ API endpoints
- ✅ Error responses

**Test Scenarios:**
- Valid and invalid requests
- Content type verification
- Status code validation
- Response body checking

### Post-processor Tests (`postprocessor_test.go`)

**Features Tested:**
- ✅ Configuration management
- ✅ File cleanup operations
- ✅ Cloud upload simulation
- ✅ Video stitching placeholders
- ✅ Job management interfaces

**File System Tests:**
- Temporary file creation
- Age-based cleanup
- Directory traversal
- Permission handling

## 🔧 Test Configuration

### Environment Variables

```bash
# Test timeout
TIMEOUT=30s

# Skip integration tests
SHORT=true

# Debug mode
DEBUG=true
```

### Dependencies

**Required:**
- Go 1.13+ (for compatibility testing)
- BoltDB
- Gorilla Mux

**Optional:**
- yt-dlp (for preprocessor integration tests)
- FFmpeg (for future post-processing tests)

## 📈 Performance Benchmarks

### Expected Performance Metrics

| Operation | Benchmark | Target |
|-----------|-----------|---------|
| Cache Set | `BenchmarkFIFOCacheSet` | < 1000 ns/op |
| Database Store | `BenchmarkDatabaseStore` | < 10000 ns/op |
| HTTP Request | `BenchmarkHTTPServerHandleRoot` | < 5000 ns/op |
| URL Processing | `BenchmarkIsHLSURL` | < 100 ns/op |

### Memory Usage

| Component | Memory Target | Notes |
|-----------|---------------|-------|
| FIFO Cache | < 10MB per 1000 items | With configurable bounds |
| Database | < 1MB per 1000 segments | BoltDB overhead |
| HTTP Server | < 50MB base | Request handling |

## 🐛 Debugging Tests

### Common Issues

1. **Database Lock Errors**
   ```bash
   # Ensure proper cleanup
   defer cleanupTestDatabase(t, originalDir)
   ```

2. **Port Conflicts**
   ```bash
   # Use test servers instead of fixed ports
   server := httptest.NewServer(handler)
   ```

3. **File Permissions**
   ```bash
   # Check write permissions in test directories
   os.MkdirAll("./files", 0755)
   ```

### Debug Mode

```bash
# Run tests with verbose output
go test -v -debug *.go

# Run with race detection
go test -race *.go

# Run with CPU profiling
go test -cpuprofile=cpu.prof *.go
```

## 🔄 Continuous Integration

### Test Pipeline

1. **Static Analysis**
   - `go vet`
   - `golint`
   - `go fmt` check

2. **Unit Tests**
   - All component tests
   - Coverage reporting

3. **Integration Tests**
   - Mock service interactions
   - End-to-end scenarios

4. **Performance Tests**
   - Benchmark comparisons
   - Memory leak detection

### Quality Gates

- **Coverage**: > 80% line coverage
- **Performance**: No significant regressions
- **Compatibility**: Go 1.13+ support
- **Dependencies**: Security vulnerability scans

## 📚 Adding New Tests

### Test File Template

```go
package main

import (
    "testing"
)

func TestNewFeature(t *testing.T) {
    // Setup
    
    // Test
    
    // Assert
    
    // Cleanup
}

func BenchmarkNewFeature(b *testing.B) {
    // Setup
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        // Benchmark code
    }
}
```

### Best Practices

1. **Isolation**: Each test should be independent
2. **Cleanup**: Always clean up resources
3. **Mocking**: Use mocks for external dependencies
4. **Coverage**: Aim for comprehensive edge case testing
5. **Documentation**: Clear test names and comments

## 🔍 Test Results Analysis

### Coverage Report

After running tests, check the coverage report:

```bash
# Open coverage report
open tests/coverage/coverage.html
```

### Performance Analysis

```bash
# Analyze benchmark results
go test -bench=. -benchmem *.go | grep Benchmark
```

### Continuous Monitoring

The test suite provides foundation for:
- Automated testing in CI/CD
- Performance regression detection
- Code quality monitoring
- Security vulnerability scanning

---

**Note**: This test suite is designed to be comprehensive yet maintainable. It provides confidence in code changes while being fast enough for regular development workflows.
