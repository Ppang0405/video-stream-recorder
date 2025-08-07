#!/bin/bash

# Test runner script for VSR (Video Stream Recorder)
# This script runs all tests with proper setup and cleanup

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
GO_DIR="../go"
TEST_DIR="."
COVERAGE_DIR="./coverage"
TIMEOUT="30s"

echo -e "${BLUE}=== VSR Test Suite ===${NC}"
echo "Running comprehensive tests for Video Stream Recorder"
echo

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed or not in PATH${NC}"
    exit 1
fi

# Check if we're in the right directory
if [ ! -d "$GO_DIR" ]; then
    echo -e "${RED}Error: Go source directory not found at $GO_DIR${NC}"
    exit 1
fi

# Create coverage directory
mkdir -p "$COVERAGE_DIR"

# Function to run tests for a specific package
run_test_file() {
    local test_file=$1
    local test_name=$(basename "$test_file" _test.go)
    
    echo -e "${YELLOW}Running $test_name tests...${NC}"
    
    # Copy test file to go directory temporarily
    cp "$test_file" "$GO_DIR/"
    
    # Run the test
    cd "$GO_DIR"
    if go test -v -timeout="$TIMEOUT" -coverprofile="../tests/coverage/${test_name}.out" -run="Test.*" "./${test_name}_test.go" *.go 2>/dev/null; then
        echo -e "${GREEN}✓ $test_name tests passed${NC}"
        test_results["$test_name"]="PASS"
    else
        echo -e "${RED}✗ $test_name tests failed${NC}"
        test_results["$test_name"]="FAIL"
        failed_tests+=("$test_name")
    fi
    
    # Clean up
    rm -f "./${test_name}_test.go"
    cd - > /dev/null
    echo
}

# Function to run benchmarks
run_benchmarks() {
    echo -e "${YELLOW}Running benchmark tests...${NC}"
    
    for test_file in *_test.go; do
        test_name=$(basename "$test_file" _test.go)
        
        # Copy test file to go directory temporarily
        cp "$test_file" "$GO_DIR/"
        
        cd "$GO_DIR"
        echo "Benchmarking $test_name..."
        if go test -v -bench=. -run=^$ "./${test_name}_test.go" *.go 2>/dev/null | grep -E "(Benchmark|PASS|FAIL)"; then
            echo -e "${GREEN}✓ $test_name benchmarks completed${NC}"
        else
            echo -e "${YELLOW}! $test_name benchmarks skipped or failed${NC}"
        fi
        
        # Clean up
        rm -f "./${test_name}_test.go"
        cd - > /dev/null
        echo
    done
}

# Function to generate coverage report
generate_coverage() {
    echo -e "${YELLOW}Generating coverage report...${NC}"
    
    cd "$GO_DIR"
    
    # Combine all coverage files
    echo "mode: set" > "../tests/coverage/combined.out"
    for coverage_file in ../tests/coverage/*.out; do
        if [ -f "$coverage_file" ] && [ "$(basename "$coverage_file")" != "combined.out" ]; then
            tail -n +2 "$coverage_file" >> "../tests/coverage/combined.out" 2>/dev/null || true
        fi
    done
    
    # Generate HTML coverage report
    if [ -f "../tests/coverage/combined.out" ]; then
        go tool cover -html="../tests/coverage/combined.out" -o="../tests/coverage/coverage.html" 2>/dev/null || true
        
        # Show coverage percentage
        coverage_percent=$(go tool cover -func="../tests/coverage/combined.out" 2>/dev/null | tail -1 | awk '{print $3}' || echo "N/A")
        echo -e "${GREEN}Coverage: $coverage_percent${NC}"
        echo "HTML coverage report: tests/coverage/coverage.html"
    fi
    
    cd - > /dev/null
}

# Function to check for yt-dlp availability
check_ytdlp() {
    if command -v yt-dlp &> /dev/null; then
        echo -e "${GREEN}✓ yt-dlp found: $(yt-dlp --version)${NC}"
        return 0
    else
        echo -e "${YELLOW}! yt-dlp not found - some tests may be skipped${NC}"
        return 1
    fi
}

# Initialize arrays for tracking results
declare -A test_results
failed_tests=()

# Pre-flight checks
echo -e "${BLUE}Pre-flight checks:${NC}"
echo "Go version: $(go version)"
check_ytdlp
echo

# Check if test files exist
test_files=(
    "preprocessor_test.go"
    "fifocache_test.go"
    "database_test.go"
    "processor_test.go"
    "server_test.go"
    "postprocessor_test.go"
)

missing_files=()
for test_file in "${test_files[@]}"; do
    if [ ! -f "$test_file" ]; then
        missing_files+=("$test_file")
    fi
done

if [ ${#missing_files[@]} -gt 0 ]; then
    echo -e "${RED}Error: Missing test files:${NC}"
    printf '%s\n' "${missing_files[@]}"
    exit 1
fi

# Run all tests
echo -e "${BLUE}Running unit tests:${NC}"
for test_file in "${test_files[@]}"; do
    run_test_file "$test_file"
done

# Run benchmarks if requested
if [ "$1" = "--bench" ] || [ "$1" = "-b" ]; then
    echo -e "${BLUE}Running benchmarks:${NC}"
    run_benchmarks
fi

# Generate coverage report
generate_coverage

# Summary
echo -e "${BLUE}=== Test Summary ===${NC}"
total_tests=${#test_files[@]}
passed_tests=$((total_tests - ${#failed_tests[@]}))

echo "Total test suites: $total_tests"
echo -e "Passed: ${GREEN}$passed_tests${NC}"
echo -e "Failed: ${RED}${#failed_tests[@]}${NC}"

if [ ${#failed_tests[@]} -eq 0 ]; then
    echo -e "\n${GREEN}🎉 All tests passed!${NC}"
    exit 0
else
    echo -e "\n${RED}❌ Failed tests:${NC}"
    printf '%s\n' "${failed_tests[@]}"
    exit 1
fi
