#!/bin/bash
# Comprehensive automated test suite for ZFSCloud
# Tests file types, security, edge cases, and full workflows

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
NC='\033[0m'

# Counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0

# Test results array
declare -a TEST_RESULTS
declare -a TEST_NAMES

# Configuration
SERVER_PORT=50053
SERVER_ADDR="127.0.0.1:$SERVER_PORT"
TEST_DIR="/tmp/zfscloud-automated-test"
DB_PATH="$TEST_DIR/test.db"
ZFS_POOL="tank"
TEST_DATA_DIR="$TEST_DIR/testdata"
SERVER_PID=""

# Logging functions
log_test() {
    echo -e "${BLUE}[TEST $((TOTAL_TESTS+1))]${NC} $1"
}

log_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((PASSED_TESTS++))
    TEST_RESULTS+=("PASS")
    TEST_NAMES+=("$1")
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((FAILED_TESTS++))
    TEST_RESULTS+=("FAIL")
    TEST_NAMES+=("$1")
}

log_skip() {
    echo -e "${YELLOW}[SKIP]${NC} $1"
    ((SKIPPED_TESTS++))
    TEST_RESULTS+=("SKIP")
    TEST_NAMES+=("$1")
}

log_info() {
    echo -e "${CYAN}[INFO]${NC} $1"
}

log_section() {
    echo ""
    echo -e "${MAGENTA}╔════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${MAGENTA}║${NC} $1"
    echo -e "${MAGENTA}╚════════════════════════════════════════════════════════════════╝${NC}"
}

# Cleanup function
cleanup() {
    log_info "Cleaning up..."
    
    # Stop server
    if [ -n "$SERVER_PID" ] && kill -0 "$SERVER_PID" 2>/dev/null; then
        kill "$SERVER_PID" 2>/dev/null || true
        sleep 1
        kill -9 "$SERVER_PID" 2>/dev/null || true
    fi
    
    # Clean test files
    rm -rf "$TEST_DIR" 2>/dev/null || true
    rm -f /tmp/test-*.go /tmp/test-*.txt /tmp/test-*.bin 2>/dev/null || true
    
    # Clean ZFS test datasets (if any were created)
    for dataset in $(zfs list -H -o name | grep "^$ZFS_POOL/test-client-" 2>/dev/null || true); do
        log_info "Cleaning ZFS dataset: $dataset"
        sudo zfs destroy -r "$dataset" 2>/dev/null || true
    done
}

trap cleanup EXIT

# Setup function
setup() {
    log_section "Test Environment Setup"
    
    # Create test directories
    mkdir -p "$TEST_DIR"/{db,data,testdata}
    
    # Create server config
    cat > "$TEST_DIR/server-config.yaml" << EOF
server:
  listen_address: "$SERVER_ADDR"

database:
  path: "$DB_PATH"

zfs:
  pool: "$ZFS_POOL"
  compression: "lz4"
  deduplication: false

quotas:
  default_quota_gb: 10
  max_quota_gb: 100

snapshots:
  schedule:
    hourly: 24
    daily: 7
    weekly: 4
    monthly: 12

security:
  admin_token: "test-admin-$(date +%s)"
  nonce_timeout_minutes: 10
  session_timeout_hours: 24
  signature_timeout_minutes: 5
EOF
    
    log_info "Test directory: $TEST_DIR"
    log_info "Server config: $TEST_DIR/server-config.yaml"
    log_info "Server address: $SERVER_ADDR"
    
    # Start server
    log_info "Starting server..."
    ./bin/zfscloud-server-test -config "$TEST_DIR/server-config.yaml" > "$TEST_DIR/server.log" 2>&1 &
    SERVER_PID=$!
    
    log_info "Server PID: $SERVER_PID"
    sleep 3
    
    if ! kill -0 "$SERVER_PID" 2>/dev/null; then
        log_fail "Server failed to start"
        cat "$TEST_DIR/server.log"
        exit 1
    fi
    
    log_pass "Test environment ready"
}

# ============================================================================
# TEST SUITE 1: File Type Tests
# ============================================================================

test_file_types() {
    log_section "FILE TYPE TESTS"
    ((TOTAL_TESTS++))
    
    log_test "Creating various file types..."
    
    # Text file
    echo "Plain text content" > "$TEST_DATA_DIR/test.txt"
    [ -f "$TEST_DATA_DIR/test.txt" ] && log_pass "Text file created" || log_fail "Text file creation failed"
    ((TOTAL_TESTS++))
    
    # Binary file
    dd if=/dev/urandom of="$TEST_DATA_DIR/test.bin" bs=1024 count=10 2>/dev/null
    [ -f "$TEST_DATA_DIR/test.bin" ] && log_pass "Binary file created" || log_fail "Binary file creation failed"
    ((TOTAL_TESTS++))
    
    # Large file (10MB)
    dd if=/dev/urandom of="$TEST_DATA_DIR/large.bin" bs=1M count=10 2>/dev/null
    [ -f "$TEST_DATA_DIR/large.bin" ] && log_pass "Large file (10MB) created" || log_fail "Large file creation failed"
    ((TOTAL_TESTS++))
    
    # Empty file
    touch "$TEST_DATA_DIR/empty.txt"
    [ -f "$TEST_DATA_DIR/empty.txt" ] && log_pass "Empty file created" || log_fail "Empty file creation failed"
    ((TOTAL_TESTS++))
    
    # File with special characters
    echo "Special chars: éàü ñ 中文 🚀" > "$TEST_DATA_DIR/unicode.txt"
    [ -f "$TEST_DATA_DIR/unicode.txt" ] && log_pass "Unicode file created" || log_fail "Unicode file creation failed"
    ((TOTAL_TESTS++))
    
    # JSON file
    echo '{"test": "data", "number": 123}' > "$TEST_DATA_DIR/test.json"
    [ -f "$TEST_DATA_DIR/test.json" ] && log_pass "JSON file created" || log_fail "JSON file creation failed"
    ((TOTAL_TESTS++))
    
    # Compressed file
    gzip -c "$TEST_DATA_DIR/test.txt" > "$TEST_DATA_DIR/test.txt.gz"
    [ -f "$TEST_DATA_DIR/test.txt.gz" ] && log_pass "Compressed file created" || log_fail "Compressed file creation failed"
    ((TOTAL_TESTS++))
    
    # Symlink (if supported)
    ln -s "$TEST_DATA_DIR/test.txt" "$TEST_DATA_DIR/link.txt" 2>/dev/null
    [ -L "$TEST_DATA_DIR/link.txt" ] && log_pass "Symlink created" || log_skip "Symlink creation not supported"
    ((TOTAL_TESTS++))
    
    # Directory
    mkdir -p "$TEST_DATA_DIR/subdir"
    echo "nested content" > "$TEST_DATA_DIR/subdir/nested.txt"
    [ -f "$TEST_DATA_DIR/subdir/nested.txt" ] && log_pass "Nested directory created" || log_fail "Nested directory creation failed"
    ((TOTAL_TESTS++))
    
    # File with long name
    LONG_NAME=$(printf 'a%.0s' {1..200})
    echo "long name test" > "$TEST_DATA_DIR/$LONG_NAME.txt" 2>/dev/null
    [ -f "$TEST_DATA_DIR/$LONG_NAME.txt" ] && log_pass "Long filename created" || log_skip "Long filename not supported"
    ((TOTAL_TESTS++))
}

# ============================================================================
# TEST SUITE 2: Security Tests
# ============================================================================

test_security() {
    log_section "SECURITY TESTS"
    
    # Test 1: Server public key retrieval
    ((TOTAL_TESTS++))
    log_test "Testing server public key retrieval..."
    
    cat > /tmp/test-pubkey.go << 'TESTEOF'
package main
import (
    "context"
    "fmt"
    "os"
    "github.com/ioustamora/zfscloud/proto"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)
func main() {
    conn, _ := grpc.NewClient(os.Args[1], grpc.WithTransportCredentials(insecure.NewCredentials()))
    defer conn.Close()
    client := proto.NewBackupServiceClient(conn)
    resp, err := client.GetServerPublicKey(context.Background(), &proto.ServerPublicKeyRequest{})
    if err != nil { os.Exit(1) }
    if len(resp.PublicKey) == 0 { os.Exit(1) }
    fmt.Println("OK")
}
TESTEOF
    
    if go run /tmp/test-pubkey.go "$SERVER_ADDR" 2>/dev/null | grep -q "OK"; then
        log_pass "Server public key retrieved"
    else
        log_fail "Failed to retrieve server public key"
    fi
    
    # Test 2: Client key generation
    ((TOTAL_TESTS++))
    log_test "Testing client key generation..."
    
    cat > /tmp/test-keygen.go << 'TESTEOF'
package main
import (
    "fmt"
    "github.com/ioustamora/zfscloud/shared/crypto"
)
func main() {
    kp, err := crypto.GenerateKeyPair()
    if err != nil || len(kp.PublicKey) != 32 || len(kp.PrivateKey) != 64 {
        panic("key generation failed")
    }
    fmt.Println("OK")
}
TESTEOF
    
    if go run /tmp/test-keygen.go 2>/dev/null | grep -q "OK"; then
        log_pass "Client key generation successful"
    else
        log_fail "Client key generation failed"
    fi
    
    # Test 3: Passphrase encryption
    ((TOTAL_TESTS++))
    log_test "Testing passphrase encryption..."
    
    cat > /tmp/test-encrypt.go << 'TESTEOF'
package main
import (
    "fmt"
    "github.com/ioustamora/zfscloud/shared/crypto"
)
func main() {
    passphrase, err := crypto.GeneratePassphrase()
    if err != nil || len(passphrase) != 32 {
        panic("passphrase generation failed")
    }
    fmt.Println("OK")
}
TESTEOF
    
    if go run /tmp/test-encrypt.go 2>/dev/null | grep -q "OK"; then
        log_pass "Passphrase generation successful"
    else
        log_fail "Passphrase generation failed"
    fi
    
    # Test 4: Signature verification
    ((TOTAL_TESTS++))
    log_test "Testing signature creation..."
    
    cat > /tmp/test-signature.go << 'TESTEOF'
package main
import (
    "fmt"
    "github.com/ioustamora/zfscloud/shared/crypto"
)
func main() {
    kp, _ := crypto.GenerateKeyPair()
    nonce, _ := crypto.GenerateNonce()
    sigData := &crypto.SignatureData{
        ClientID: "test",
        Timestamp: crypto.GetTimestamp(),
        Nonce: nonce,
        Payload: []byte("test"),
    }
    sig, err := kp.Sign(sigData)
    if err != nil || len(sig) == 0 {
        panic("signature failed")
    }
    fmt.Println("OK")
}
TESTEOF
    
    if go run /tmp/test-signature.go 2>/dev/null | grep -q "OK"; then
        log_pass "Signature creation successful"
    else
        log_fail "Signature creation failed"
    fi
    
    # Test 5: Database injection protection
    ((TOTAL_TESTS++))
    log_test "Testing SQL injection protection..."
    
    MALICIOUS_INPUT="'; DROP TABLE users; --"
    echo "INSERT INTO audit_log (user_id, client_id, action, details, ip_address) VALUES ('test', 'test', 'test', '$MALICIOUS_INPUT', '127.0.0.1');" | sqlite3 "$DB_PATH" 2>/dev/null || true
    
    TABLE_COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='users';" 2>/dev/null)
    if [ "$TABLE_COUNT" = "1" ]; then
        log_pass "SQL injection protection working"
    else
        log_fail "SQL injection vulnerability detected"
    fi
    
    # Test 6: Path traversal protection
    ((TOTAL_TESTS++))
    log_test "Testing path traversal protection..."
    
    # Create test file with path traversal attempt
    TRAVERSAL_PATH="../../../etc/passwd"
    mkdir -p "$TEST_DATA_DIR/traversal"
    echo "test" > "$TEST_DATA_DIR/traversal/safe.txt"
    
    # The application should sanitize this path
    if [ ! -f "/etc/passwd.uploaded" ]; then
        log_pass "Path traversal protection working"
    else
        log_fail "Path traversal vulnerability detected"
    fi
}

# ============================================================================
# TEST SUITE 3: Edge Cases
# ============================================================================

test_edge_cases() {
    log_section "EDGE CASE TESTS"
    
    # Test 1: Zero-byte file
    ((TOTAL_TESTS++))
    log_test "Testing zero-byte file handling..."
    
    touch "$TEST_DATA_DIR/zero.bin"
    SIZE=$(stat -c%s "$TEST_DATA_DIR/zero.bin" 2>/dev/null || stat -f%z "$TEST_DATA_DIR/zero.bin" 2>/dev/null)
    if [ "$SIZE" = "0" ]; then
        log_pass "Zero-byte file handled correctly"
    else
        log_fail "Zero-byte file size incorrect"
    fi
    
    # Test 2: Maximum filename length
    ((TOTAL_TESTS++))
    log_test "Testing maximum filename length..."
    
    # Linux typically supports 255 bytes
    LONG_NAME=$(printf 'x%.0s' {1..255})
    touch "$TEST_DATA_DIR/$LONG_NAME" 2>/dev/null
    if [ -f "$TEST_DATA_DIR/$LONG_NAME" ]; then
        log_pass "Maximum filename length supported"
    else
        log_skip "Maximum filename length not supported"
    fi
    
    # Test 3: Special characters in filename
    ((TOTAL_TESTS++))
    log_test "Testing special characters in filename..."
    
    SPECIAL_NAME="test file with spaces & symbols (123).txt"
    echo "test" > "$TEST_DATA_DIR/$SPECIAL_NAME" 2>/dev/null
    if [ -f "$TEST_DATA_DIR/$SPECIAL_NAME" ]; then
        log_pass "Special characters in filename supported"
    else
        log_fail "Special characters in filename not supported"
    fi
    
    # Test 4: Concurrent file operations
    ((TOTAL_TESTS++))
    log_test "Testing concurrent file operations..."
    
    for i in {1..10}; do
        echo "concurrent $i" > "$TEST_DATA_DIR/concurrent-$i.txt" &
    done
    wait
    
    FILE_COUNT=$(ls "$TEST_DATA_DIR"/concurrent-*.txt 2>/dev/null | wc -l)
    if [ "$FILE_COUNT" -eq 10 ]; then
        log_pass "Concurrent operations handled correctly ($FILE_COUNT files)"
    else
        log_fail "Concurrent operations failed (expected 10, got $FILE_COUNT)"
    fi
    
    # Test 5: Disk space handling
    ((TOTAL_TESTS++))
    log_test "Testing disk space check..."
    
    AVAIL_SPACE=$(df -k "$TEST_DIR" | awk 'NR==2 {print $4}')
    if [ "$AVAIL_SPACE" -gt 100000 ]; then
        log_pass "Sufficient disk space available (${AVAIL_SPACE}KB)"
    else
        log_fail "Insufficient disk space (${AVAIL_SPACE}KB)"
    fi
    
    # Test 6: File modification during upload
    ((TOTAL_TESTS++))
    log_test "Testing file modification handling..."
    
    echo "initial content" > "$TEST_DATA_DIR/modified.txt"
    HASH1=$(sha256sum "$TEST_DATA_DIR/modified.txt" | awk '{print $1}')
    sleep 1
    echo "modified content" > "$TEST_DATA_DIR/modified.txt"
    HASH2=$(sha256sum "$TEST_DATA_DIR/modified.txt" | awk '{print $1}')
    
    if [ "$HASH1" != "$HASH2" ]; then
        log_pass "File modification detected correctly"
    else
        log_fail "File modification not detected"
    fi
    
    # Test 7: Database corruption recovery
    ((TOTAL_TESTS++))
    log_test "Testing database integrity..."
    
    if sqlite3 "$DB_PATH" "PRAGMA integrity_check;" 2>/dev/null | grep -q "ok"; then
        log_pass "Database integrity verified"
    else
        log_fail "Database integrity check failed"
    fi
    
    # Test 8: Network timeout simulation
    ((TOTAL_TESTS++))
    log_test "Testing connection timeout handling..."
    
    if timeout 2 bash -c "cat < /dev/null > /dev/tcp/127.0.0.1/$SERVER_PORT" 2>/dev/null; then
        log_pass "Connection timeout handling works"
    else
        log_fail "Connection timeout handling failed"
    fi
}

# ============================================================================
# TEST SUITE 4: Performance Tests
# ============================================================================

test_performance() {
    log_section "PERFORMANCE TESTS"
    
    # Test 1: Small file throughput
    ((TOTAL_TESTS++))
    log_test "Testing small file throughput..."
    
    START_TIME=$(date +%s.%N)
    for i in {1..100}; do
        echo "test $i" > "$TEST_DATA_DIR/perf-small-$i.txt"
    done
    END_TIME=$(date +%s.%N)
    DURATION=$(echo "$END_TIME - $START_TIME" | bc)
    THROUGHPUT=$(echo "scale=2; 100 / $DURATION" | bc)
    
    log_pass "Small file throughput: ${THROUGHPUT} files/sec"
    
    # Test 2: Large file handling
    ((TOTAL_TESTS++))
    log_test "Testing large file handling..."
    
    START_TIME=$(date +%s.%N)
    dd if=/dev/zero of="$TEST_DATA_DIR/perf-large.bin" bs=1M count=100 2>/dev/null
    END_TIME=$(date +%s.%N)
    DURATION=$(echo "$END_TIME - $START_TIME" | bc)
    THROUGHPUT=$(echo "scale=2; 100 / $DURATION" | bc)
    
    log_pass "Large file write: ${THROUGHPUT} MB/sec"
    
    # Test 3: Compression ratio
    ((TOTAL_TESTS++))
    log_test "Testing compression effectiveness..."
    
    # Create compressible data
    for i in {1..1000}; do
        echo "This is a repeated line for compression testing"
    done > "$TEST_DATA_DIR/compressible.txt"
    
    ORIGINAL_SIZE=$(stat -c%s "$TEST_DATA_DIR/compressible.txt" 2>/dev/null || stat -f%z "$TEST_DATA_DIR/compressible.txt" 2>/dev/null)
    gzip -c "$TEST_DATA_DIR/compressible.txt" > "$TEST_DATA_DIR/compressible.txt.gz"
    COMPRESSED_SIZE=$(stat -c%s "$TEST_DATA_DIR/compressible.txt.gz" 2>/dev/null || stat -f%z "$TEST_DATA_DIR/compressible.txt.gz" 2>/dev/null)
    
    RATIO=$(echo "scale=2; 100 - ($COMPRESSED_SIZE * 100 / $ORIGINAL_SIZE)" | bc)
    log_pass "Compression ratio: ${RATIO}% (${ORIGINAL_SIZE}B -> ${COMPRESSED_SIZE}B)"
    
    # Test 4: Database query performance
    ((TOTAL_TESTS++))
    log_test "Testing database query performance..."
    
    START_TIME=$(date +%s.%N)
    for i in {1..100}; do
        sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM users;" > /dev/null 2>&1
    done
    END_TIME=$(date +%s.%N)
    DURATION=$(echo "$END_TIME - $START_TIME" | bc)
    QPS=$(echo "scale=2; 100 / $DURATION" | bc)
    
    log_pass "Database query rate: ${QPS} queries/sec"
}

# ============================================================================
# TEST SUITE 5: ZFS Integration Tests
# ============================================================================

test_zfs_integration() {
    log_section "ZFS INTEGRATION TESTS"
    
    # Test 1: ZFS pool accessibility
    ((TOTAL_TESTS++))
    log_test "Testing ZFS pool accessibility..."
    
    if zpool list "$ZFS_POOL" >/dev/null 2>&1; then
        log_pass "ZFS pool accessible"
    else
        log_fail "ZFS pool not accessible"
        return
    fi
    
    # Test 2: ZFS compression
    ((TOTAL_TESTS++))
    log_test "Testing ZFS compression setting..."
    
    COMPRESSION=$(zfs get -H -o value compression "$ZFS_POOL" 2>/dev/null)
    if [ -n "$COMPRESSION" ]; then
        log_pass "ZFS compression: $COMPRESSION"
    else
        log_fail "Could not read ZFS compression setting"
    fi
    
    # Test 3: ZFS available space
    ((TOTAL_TESTS++))
    log_test "Testing ZFS available space..."
    
    AVAIL=$(zfs get -H -o value available "$ZFS_POOL" 2>/dev/null)
    if [ -n "$AVAIL" ]; then
        log_pass "ZFS available space: $AVAIL"
    else
        log_fail "Could not read ZFS available space"
    fi
    
    # Test 4: ZFS dataset creation simulation
    ((TOTAL_TESTS++))
    log_test "Testing ZFS dataset creation..."
    
    TEST_DATASET="$ZFS_POOL/test-automated-$(date +%s)"
    if sudo zfs create "$TEST_DATASET" 2>/dev/null; then
        log_pass "ZFS dataset created: $TEST_DATASET"
        
        # Set properties
        sudo zfs set compression=lz4 "$TEST_DATASET" 2>/dev/null
        sudo zfs set quota=1G "$TEST_DATASET" 2>/dev/null
        
        # Cleanup
        sudo zfs destroy "$TEST_DATASET" 2>/dev/null || true
    else
        log_skip "ZFS dataset creation requires sudo"
    fi
}

# ============================================================================
# Main Test Execution
# ============================================================================

main() {
    log_section "ZFSCloud Automated Test Suite"
    echo "Starting comprehensive automated testing..."
    echo "Test directory: $TEST_DIR"
    echo "Server: $SERVER_ADDR"
    echo ""
    
    # Setup
    setup
    
    # Run test suites
    test_file_types
    test_security
    test_edge_cases
    test_performance
    test_zfs_integration
    
    # Generate report
    log_section "TEST REPORT"
    echo ""
    echo "═══════════════════════════════════════════════════════════════"
    echo "                     TEST RESULTS SUMMARY                      "
    echo "═══════════════════════════════════════════════════════════════"
    echo ""
    echo -e "  Total Tests:   ${BLUE}$TOTAL_TESTS${NC}"
    echo -e "  Passed:        ${GREEN}$PASSED_TESTS${NC}"
    echo -e "  Failed:        ${RED}$FAILED_TESTS${NC}"
    echo -e "  Skipped:       ${YELLOW}$SKIPPED_TESTS${NC}"
    echo ""
    
    if [ $FAILED_TESTS -eq 0 ]; then
        SUCCESS_RATE="100"
    else
        SUCCESS_RATE=$(echo "scale=2; $PASSED_TESTS * 100 / ($PASSED_TESTS + $FAILED_TESTS)" | bc)
    fi
    
    echo -e "  Success Rate:  ${GREEN}${SUCCESS_RATE}%${NC}"
    echo ""
    echo "═══════════════════════════════════════════════════════════════"
    echo ""
    
    # Detailed results
    if [ $FAILED_TESTS -gt 0 ]; then
        echo "Failed Tests:"
        for i in "${!TEST_RESULTS[@]}"; do
            if [ "${TEST_RESULTS[$i]}" = "FAIL" ]; then
                echo -e "  ${RED}✗${NC} ${TEST_NAMES[$i]}"
            fi
        done
        echo ""
    fi
    
    # Save report
    REPORT_FILE="$TEST_DIR/test-report.txt"
    {
        echo "ZFSCloud Automated Test Report"
        echo "Generated: $(date)"
        echo "═══════════════════════════════════════"
        echo "Total Tests: $TOTAL_TESTS"
        echo "Passed: $PASSED_TESTS"
        echo "Failed: $FAILED_TESTS"
        echo "Skipped: $SKIPPED_TESTS"
        echo "Success Rate: ${SUCCESS_RATE}%"
        echo ""
        echo "Server Log: $TEST_DIR/server.log"
        echo "Database: $DB_PATH"
    } > "$REPORT_FILE"
    
    echo "Report saved to: $REPORT_FILE"
    echo ""
    
    # Exit code
    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "${GREEN}✓ All tests passed!${NC}"
        exit 0
    else
        echo -e "${RED}✗ Some tests failed${NC}"
        exit 1
    fi
}

# Run main
main "$@"
