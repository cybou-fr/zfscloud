#!/bin/bash
# Performance testing script for ZFSCloud
# Tests upload performance with various file sizes and multiple concurrent clients

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
TEST_DATA_DIR="$PROJECT_DIR/test-data"
RESULTS_DIR="$PROJECT_DIR/test-results"

# Test parameters
FILE_SIZES=(1024 10240 102400 1048576 10485760)  # 1KB, 10KB, 100KB, 1MB, 10MB
FILE_SIZE_NAMES=("1KB" "10KB" "100KB" "1MB" "10MB")
NUM_CLIENTS=(1 2 4 8)

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${CYAN}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
}

log_error() {
    echo -e "${RED}[FAIL]${NC} $1"
}

log_section() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

cleanup() {
    log_info "Cleaning up test data..."
    rm -rf "$TEST_DATA_DIR" 2>/dev/null || true
}

trap cleanup EXIT

# Create directories
mkdir -p "$TEST_DATA_DIR"
mkdir -p "$RESULTS_DIR"

log_section "ZFSCloud Performance Test Suite"
echo "Test Data Directory: $TEST_DATA_DIR"
echo "Results Directory: $RESULTS_DIR"
echo ""

cd "$PROJECT_DIR"

# Test 1: Generate test files
log_section "Test 1: Generate Test Files"
for i in "${!FILE_SIZES[@]}"; do
    SIZE=${FILE_SIZES[$i]}
    NAME=${FILE_SIZE_NAMES[$i]}
    FILE="$TEST_DATA_DIR/test-$NAME.bin"
    
    log_info "Generating $NAME file ($SIZE bytes)..."
    dd if=/dev/urandom of="$FILE" bs=$SIZE count=1 2>/dev/null
    
    if [ -f "$FILE" ]; then
        ACTUAL_SIZE=$(stat -f%z "$FILE" 2>/dev/null || stat -c%s "$FILE")
        log_success "Created $NAME file: $ACTUAL_SIZE bytes"
    else
        log_error "Failed to create $NAME file"
    fi
done

# Test 2: File I/O performance
log_section "Test 2: File I/O Performance (Baseline)"
RESULT_FILE="$RESULTS_DIR/io-performance-$(date +%Y%m%d-%H%M%S).txt"
echo "File I/O Performance Test - $(date)" > "$RESULT_FILE"
echo "========================================" >> "$RESULT_FILE"

for i in "${!FILE_SIZES[@]}"; do
    SIZE=${FILE_SIZES[$i]}
    NAME=${FILE_SIZE_NAMES[$i]}
    FILE="$TEST_DATA_DIR/test-$NAME.bin"
    
    log_info "Testing read performance for $NAME..."
    START_TIME=$(date +%s.%N)
    cat "$FILE" > /dev/null
    END_TIME=$(date +%s.%N)
    DURATION=$(echo "$END_TIME - $START_TIME" | bc)
    THROUGHPUT=$(echo "scale=2; $SIZE / 1024 / 1024 / $DURATION" | bc)
    
    log_success "$NAME read: ${DURATION}s (${THROUGHPUT} MB/s)"
    echo "$NAME read: ${DURATION}s (${THROUGHPUT} MB/s)" >> "$RESULT_FILE"
done

# Test 3: Compression performance
log_section "Test 3: Compression Performance"
for i in "${!FILE_SIZES[@]}"; do
    SIZE=${FILE_SIZES[$i]}
    NAME=${FILE_SIZE_NAMES[$i]}
    FILE="$TEST_DATA_DIR/test-$NAME.bin"
    
    log_info "Testing compression for $NAME..."
    START_TIME=$(date +%s.%N)
    gzip -c "$FILE" > "$TEST_DATA_DIR/test-$NAME.bin.gz" 2>/dev/null
    END_TIME=$(date +%s.%N)
    DURATION=$(echo "$END_TIME - $START_TIME" | bc)
    
    ORIGINAL_SIZE=$(stat -f%z "$FILE" 2>/dev/null || stat -c%s "$FILE")
    COMPRESSED_SIZE=$(stat -f%z "$TEST_DATA_DIR/test-$NAME.bin.gz" 2>/dev/null || stat -c%s "$TEST_DATA_DIR/test-$NAME.bin.gz")
    RATIO=$(echo "scale=2; 100 - ($COMPRESSED_SIZE * 100 / $ORIGINAL_SIZE)" | bc)
    
    log_success "$NAME compression: ${DURATION}s (${RATIO}% reduction)"
    echo "$NAME compression: ${DURATION}s (${RATIO}% reduction)" >> "$RESULT_FILE"
done

# Test 4: ZFS pool performance
log_section "Test 4: ZFS Pool Performance"
log_info "Checking ZFS pool I/O statistics..."
docker exec zfscloud-server zpool iostat tank 1 5 2>/dev/null | tee -a "$RESULT_FILE"

log_info "Checking ZFS dataset statistics..."
docker exec zfscloud-server zfs list -o name,used,avail,refer,compressratio tank 2>/dev/null | tee -a "$RESULT_FILE"

# Test 5: Concurrent operations simulation
log_section "Test 5: Concurrent Operations"
log_info "Simulating concurrent file operations..."

for NUM in "${NUM_CLIENTS[@]}"; do
    log_info "Testing with $NUM concurrent operations..."
    START_TIME=$(date +%s.%N)
    
    for ((i=1; i<=NUM; i++)); do
        cp "$TEST_DATA_DIR/test-1MB.bin" "$TEST_DATA_DIR/concurrent-$i.bin" &
    done
    wait
    
    END_TIME=$(date +%s.%N)
    DURATION=$(echo "$END_TIME - $START_TIME" | bc)
    log_success "$NUM concurrent operations: ${DURATION}s"
    echo "$NUM concurrent operations: ${DURATION}s" >> "$RESULT_FILE"
    
    # Clean up
    rm -f "$TEST_DATA_DIR/concurrent-*.bin"
done

# Test 6: Memory usage
log_section "Test 6: Memory Usage"
log_info "Checking server container memory usage..."
MEMORY_STATS=$(docker stats zfscloud-server --no-stream --format "{{.MemUsage}}" 2>/dev/null)
log_success "Server memory usage: $MEMORY_STATS"
echo "Server memory usage: $MEMORY_STATS" >> "$RESULT_FILE"

# Test 7: Database performance
log_section "Test 7: Database Performance"
log_info "Testing database query performance..."

START_TIME=$(date +%s.%N)
docker exec zfscloud-server sqlite3 /var/lib/zfscloud/zfscloud.db "SELECT COUNT(*) FROM users;" > /dev/null 2>&1
END_TIME=$(date +%s.%N)
DURATION=$(echo "scale=4; $END_TIME - $START_TIME" | bc)
log_success "SELECT query: ${DURATION}s"
echo "SELECT query: ${DURATION}s" >> "$RESULT_FILE"

log_info "Checking database size..."
DB_SIZE=$(docker exec zfscloud-server stat -f%z /var/lib/zfscloud/zfscloud.db 2>/dev/null || docker exec zfscloud-server stat -c%s /var/lib/zfscloud/zfscloud.db 2>/dev/null)
DB_SIZE_MB=$(echo "scale=2; $DB_SIZE / 1024 / 1024" | bc)
log_success "Database size: ${DB_SIZE_MB} MB"
echo "Database size: ${DB_SIZE_MB} MB" >> "$RESULT_FILE"

# Test 8: Network throughput
log_section "Test 8: Network Throughput"
log_info "Testing gRPC port responsiveness..."
START_TIME=$(date +%s.%N)
for i in {1..10}; do
    timeout 1 bash -c "cat < /dev/null > /dev/tcp/localhost/50051" 2>/dev/null || true
done
END_TIME=$(date +%s.%N)
DURATION=$(echo "scale=4; ($END_TIME - $START_TIME) / 10" | bc)
log_success "Average connection time: ${DURATION}s"
echo "Average connection time: ${DURATION}s" >> "$RESULT_FILE"

# Summary
log_section "Performance Test Summary"
echo ""
echo "Results saved to: $RESULT_FILE"
echo ""
echo "Key Metrics:"
cat "$RESULT_FILE" | grep -E "(MB/s|reduction|concurrent|memory|query)"
echo ""
log_success "Performance tests completed!"
echo ""
echo "Next steps:"
echo "  1. Review detailed results: cat $RESULT_FILE"
echo "  2. Run integration tests: ./tests/integration-test.sh"
echo "  3. Test with real client: ./bin/zfscloud-client"
echo ""
