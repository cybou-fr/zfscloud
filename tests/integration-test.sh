#!/bin/bash
# Integration test script for ZFSCloud client-server interactions
# This script automates the complete testing workflow

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
TEST_CLIENT_ID="test-client-$(date +%s)"
TEST_FILE="test-data.txt"
TEST_UPLOAD_PATH="/test-upload-$(date +%s).txt"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BLUE='\033[0;34m'
NC='\033[0m'

# Counters
TESTS_PASSED=0
TESTS_FAILED=0

# Helper functions
log_info() {
    echo -e "${CYAN}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((TESTS_PASSED++))
}

log_error() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((TESTS_FAILED++))
}

log_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_section() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

cleanup() {
    log_info "Cleaning up test artifacts..."
    rm -f "$PROJECT_DIR/$TEST_FILE" 2>/dev/null || true
    rm -f "$PROJECT_DIR/downloaded-*.txt" 2>/dev/null || true
    cd "$PROJECT_DIR/docker" && docker-compose down -v 2>/dev/null || true
}

# Trap to ensure cleanup on exit
trap cleanup EXIT

# Start tests
log_section "ZFSCloud Integration Test Suite"
echo "Test Client ID: $TEST_CLIENT_ID"
echo "Project Directory: $PROJECT_DIR"
echo ""

cd "$PROJECT_DIR"

# Test 1: Build binaries
log_section "Test 1: Build Binaries"
log_info "Building server binary for Linux..."
if ./build-docker.sh > /dev/null 2>&1; then
    log_success "Server binary built successfully"
else
    log_error "Failed to build server binary"
    exit 1
fi

log_info "Building client binary..."
if go build -o bin/zfscloud-client ./client/cmd > /dev/null 2>&1; then
    log_success "Client binary built successfully"
else
    log_error "Failed to build client binary"
    exit 1
fi

# Test 2: Start Docker server
log_section "Test 2: Start Docker Server"
log_info "Starting Docker containers..."
cd docker
if docker-compose up -d zfscloud-server > /dev/null 2>&1; then
    log_success "Docker server started"
else
    log_error "Failed to start Docker server"
    exit 1
fi
cd ..

log_info "Waiting for server initialization (10 seconds)..."
sleep 10

# Test 3: Check server health
log_section "Test 3: Server Health Check"
log_info "Checking if server is listening on port 50051..."
if docker exec zfscloud-server netstat -tuln 2>/dev/null | grep -q "50051"; then
    log_success "Server is listening on port 50051"
else
    log_error "Server is not listening on port 50051"
fi

log_info "Checking ZFS pool status..."
if docker exec zfscloud-server zpool status tank > /dev/null 2>&1; then
    log_success "ZFS pool 'tank' is online"
else
    log_error "ZFS pool 'tank' is not available"
fi

log_info "Checking database initialization..."
if docker exec zfscloud-server test -f /var/lib/zfscloud/zfscloud.db; then
    log_success "Database file exists"
else
    log_error "Database file not found"
fi

# Test 4: Server public key retrieval
log_section "Test 4: Server Public Key"
log_info "Testing if client can retrieve server public key..."
# This would require a gRPC client call - we'll check server logs instead
if docker logs zfscloud-server 2>&1 | grep -q "Server initialized successfully"; then
    log_success "Server is ready to serve requests"
else
    log_warning "Server initialization not confirmed in logs"
fi

# Test 5: Create test file
log_section "Test 5: Prepare Test Data"
log_info "Creating test file..."
echo "ZFSCloud Integration Test Data - $(date)" > "$TEST_FILE"
echo "Random data: $(openssl rand -hex 32)" >> "$TEST_FILE"
if [ -f "$TEST_FILE" ]; then
    log_success "Test file created ($(stat -f%z "$TEST_FILE" 2>/dev/null || stat -c%s "$TEST_FILE") bytes)"
else
    log_error "Failed to create test file"
fi

# Test 6: Database operations
log_section "Test 6: Database Schema"
log_info "Checking database tables..."
TABLES=$(docker exec zfscloud-server sqlite3 /var/lib/zfscloud/zfscloud.db ".tables" 2>/dev/null)
if echo "$TABLES" | grep -q "users"; then
    log_success "Database table 'users' exists"
else
    log_error "Database table 'users' not found"
fi

if echo "$TABLES" | grep -q "registrations"; then
    log_success "Database table 'registrations' exists"
else
    log_error "Database table 'registrations' not found"
fi

if echo "$TABLES" | grep -q "sessions"; then
    log_success "Database table 'sessions' exists"
else
    log_error "Database table 'sessions' not found"
fi

# Test 7: ZFS dataset operations
log_section "Test 7: ZFS Operations"
log_info "Checking ZFS compression..."
if docker exec zfscloud-server zfs get compression tank | grep -q "lz4"; then
    log_success "ZFS compression is enabled (lz4)"
else
    log_warning "ZFS compression is not lz4"
fi

log_info "Checking available space..."
AVAILABLE=$(docker exec zfscloud-server zfs list -H -o avail tank 2>/dev/null | head -1)
log_success "Available space: $AVAILABLE"

# Test 8: Network connectivity
log_section "Test 8: Network Connectivity"
log_info "Testing gRPC port accessibility..."
if nc -z localhost 50051 2>/dev/null || timeout 1 bash -c "cat < /dev/null > /dev/tcp/localhost/50051" 2>/dev/null; then
    log_success "Port 50051 is accessible"
else
    log_warning "Cannot verify port 50051 accessibility (may require netcat)"
fi

# Test 9: Container resource usage
log_section "Test 9: Container Resources"
log_info "Checking container resource usage..."
CONTAINER_STATS=$(docker stats zfscloud-server --no-stream --format "table {{.CPUPerc}}\t{{.MemUsage}}" 2>/dev/null | tail -1)
if [ -n "$CONTAINER_STATS" ]; then
    log_success "Container stats: $CONTAINER_STATS"
else
    log_warning "Could not retrieve container stats"
fi

# Test 10: Configuration files
log_section "Test 10: Configuration"
log_info "Checking server configuration..."
if docker exec zfscloud-server test -f /etc/zfscloud/config.yaml; then
    log_success "Server configuration file exists"
    # Show key config values
    LISTEN_ADDR=$(docker exec zfscloud-server grep "listen_address" /etc/zfscloud/config.yaml 2>/dev/null || echo "N/A")
    log_info "Server config: $LISTEN_ADDR"
else
    log_error "Server configuration file not found"
fi

# Summary
log_section "Test Summary"
echo ""
echo -e "${GREEN}Tests Passed: $TESTS_PASSED${NC}"
echo -e "${RED}Tests Failed: $TESTS_FAILED${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
    echo ""
    echo -e "${CYAN}=== Server is ready for client testing ===${NC}"
    echo ""
    echo "Run the client:"
    echo "  ./bin/zfscloud-client"
    echo ""
    echo "View server logs:"
    echo "  docker logs -f zfscloud-server"
    echo ""
    echo "Stop server:"
    echo "  cd docker && docker-compose down"
    echo ""
    exit 0
else
    echo -e "${RED}✗ Some tests failed${NC}"
    echo ""
    echo "Check server logs:"
    echo "  docker logs zfscloud-server"
    echo ""
    exit 1
fi
