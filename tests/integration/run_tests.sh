#!/bin/bash

# GoRDP Integration Test Runner
# This script runs comprehensive integration tests using the mock RDP server

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
TEST_TIMEOUT=300s
COVERAGE_OUTPUT="coverage.out"
COVERAGE_HTML="coverage.html"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  GoRDP Integration Test Suite${NC}"
echo -e "${BLUE}========================================${NC}"

# Function to print section headers
print_section() {
    echo -e "\n${YELLOW}$1${NC}"
    echo -e "${YELLOW}$(printf '=%.0s' {1..50})${NC}"
}

# Function to check if port is available
check_port() {
    local port=$1
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        echo -e "${RED}Error: Port $port is already in use${NC}"
        echo "Please stop any services using port $port and try again"
        exit 1
    fi
}

# Function to run tests with coverage
run_tests_with_coverage() {
    local test_pattern=$1
    local test_name=$2
    
    echo -e "\n${GREEN}Running $test_name...${NC}"
    
    go test -v \
        -timeout $TEST_TIMEOUT \
        -coverprofile=$COVERAGE_OUTPUT \
        -covermode=atomic \
        -run "$test_pattern" \
        ./tests/integration/...
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ $test_name completed successfully${NC}"
    else
        echo -e "${RED}✗ $test_name failed${NC}"
        return 1
    fi
}

# Function to run benchmarks
run_benchmarks() {
    echo -e "\n${GREEN}Running benchmarks...${NC}"
    
    go test -v \
        -timeout $TEST_TIMEOUT \
        -bench=. \
        -benchmem \
        -run "^$" \
        ./tests/integration/...
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Benchmarks completed successfully${NC}"
    else
        echo -e "${RED}✗ Benchmarks failed${NC}"
        return 1
    fi
}

# Function to generate coverage report
generate_coverage_report() {
    if [ -f "$COVERAGE_OUTPUT" ]; then
        echo -e "\n${GREEN}Generating coverage report...${NC}"
        
        # Generate HTML coverage report
        go tool cover -html=$COVERAGE_OUTPUT -o $COVERAGE_HTML
        
        # Show coverage summary
        echo -e "\n${BLUE}Coverage Summary:${NC}"
        go tool cover -func=$COVERAGE_OUTPUT | tail -1
        
        echo -e "\n${GREEN}HTML coverage report generated: $COVERAGE_HTML${NC}"
    else
        echo -e "${YELLOW}No coverage data available${NC}"
    fi
}

# Function to clean up
cleanup() {
    echo -e "\n${BLUE}Cleaning up...${NC}"
    
    # Kill any remaining test processes
    pkill -f "go test" 2>/dev/null || true
    
    # Remove coverage files if they exist
    rm -f $COVERAGE_OUTPUT $COVERAGE_HTML 2>/dev/null || true
    
    echo -e "${GREEN}Cleanup completed${NC}"
}

# Set up trap to clean up on exit
trap cleanup EXIT

# Check prerequisites
print_section "Checking Prerequisites"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo -e "${GREEN}Go version: $GO_VERSION${NC}"

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    echo -e "${RED}Error: go.mod not found. Please run this script from the project root${NC}"
    exit 1
fi

# Check if integration tests exist
if [ ! -d "tests/integration" ]; then
    echo -e "${RED}Error: Integration tests directory not found${NC}"
    exit 1
fi

# Check for required ports
print_section "Checking Port Availability"
check_port 3389
check_port 3388
check_port 3387

echo -e "${GREEN}All required ports are available${NC}"

# Build the project
print_section "Building Project"
go build ./...
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Project built successfully${NC}"
else
    echo -e "${RED}✗ Build failed${NC}"
    exit 1
fi

# Run unit tests first
print_section "Running Unit Tests"
go test -v -timeout 60s ./...
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Unit tests passed${NC}"
else
    echo -e "${RED}✗ Unit tests failed${NC}"
    exit 1
fi

# Run integration tests
print_section "Running Integration Tests"

# Test basic connection
run_tests_with_coverage "TestIntegration_BasicConnection" "Basic Connection Tests"

# Test full session workflow
run_tests_with_coverage "TestIntegration_FullSessionWorkflow" "Full Session Workflow Tests"

# Test input handling
run_tests_with_coverage "TestIntegration_InputHandling" "Input Handling Tests"

# Test multi-monitor
run_tests_with_coverage "TestIntegration_MultiMonitor" "Multi-Monitor Tests"

# Test virtual channels
run_tests_with_coverage "TestIntegration_VirtualChannels" "Virtual Channel Tests"

# Test error handling
run_tests_with_coverage "TestIntegration_ErrorHandling" "Error Handling Tests"

# Test performance
run_tests_with_coverage "TestIntegration_Performance" "Performance Tests"

# Test concurrent connections
run_tests_with_coverage "TestIntegration_ConcurrentConnections" "Concurrent Connection Tests"

# Run benchmarks
print_section "Running Benchmarks"
run_benchmarks

# Generate coverage report
print_section "Generating Coverage Report"
generate_coverage_report

# Final summary
print_section "Test Summary"
echo -e "${GREEN}✓ All integration tests completed successfully${NC}"
echo -e "${GREEN}✓ Benchmarks completed successfully${NC}"
echo -e "${GREEN}✓ Coverage report generated${NC}"

echo -e "\n${BLUE}Integration test suite completed successfully!${NC}"
echo -e "${BLUE}Check $COVERAGE_HTML for detailed coverage information${NC}" 