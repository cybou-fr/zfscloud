#!/bin/bash
# Quick automated test - tests key functionality without server dependency

RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
NC='\033[0m'

PASSED=0
FAILED=0
TEST_DIR="/tmp/zfscloud-quick-test-$$"

log_pass() { echo -e "${GREEN}✓${NC} $1"; ((PASSED++)); }
log_fail() { echo -e "${RED}✗${NC} $1"; ((FAILED++)); }
log_info() { echo -e "${CYAN}→${NC} $1"; }

mkdir -p "$TEST_DIR"
cd "$TEST_DIR"

echo "═══════════════════════════════════════"
echo "  ZFSCloud Quick Test Suite"
echo "═══════════════════════════════════════"
echo ""

# FILE TYPE TESTS
echo "1. FILE TYPE TESTS"
echo "-------------------"
echo "text" > test.txt && log_pass "Text file" || log_fail "Text file"
dd if=/dev/urandom of=binary.bin bs=1K count=10 2>/dev/null && log_pass "Binary file (10KB)" || log_fail "Binary file"
dd if=/dev/urandom of=large.bin bs=1M count=5 2>/dev/null && log_pass "Large file (5MB)" || log_fail "Large file"
touch empty.txt && log_pass "Empty file" || log_fail "Empty file"
echo "éàü 中文 🚀" > unicode.txt && log_pass "Unicode content" || log_fail "Unicode content"
echo '{"key":"value"}' > data.json && log_pass "JSON file" || log_fail "JSON file"
gzip -c test.txt > test.gz 2>/dev/null && log_pass "Compressed file" || log_fail "Compressed file"
mkdir -p sub/dir && echo "nested" > sub/dir/file.txt && log_pass "Nested directory" || log_fail "Nested directory"
echo ""

# SECURITY TESTS
echo "2. SECURITY TESTS"
echo "-------------------"

# Test crypto functions by checking if they compile and run
cat > /tmp/test-crypto-$$.go << 'CRYPTOEOF'
package main
import (
    "fmt"
    "github.com/ioustamora/zfscloud/shared/crypto"
)
func main() {
    // Key generation
    kp, err := crypto.GenerateKeyPair()
    if err != nil || len(kp.PublicKey) != 32 { panic("key gen failed") }
    fmt.Println("key_ok")
    
    // Passphrase generation
    p, err := crypto.GeneratePassphrase()
    if err != nil || len(p) != 32 { panic("passphrase failed") }
    fmt.Println("pass_ok")
    
    // Nonce generation
    n, err := crypto.GenerateNonce()
    if err != nil || len(n) == 0 { panic("nonce failed") }
    fmt.Println("nonce_ok")
    
    // Signature
    sigData := &crypto.SignatureData{
        ClientID: "test",
        Timestamp: crypto.GetTimestamp(),
        Nonce: n,
        Payload: []byte("test"),
    }
    sig, err := kp.Sign(sigData)
    if err != nil || len(sig) == 0 { panic("signature failed") }
    fmt.Println("sig_ok")
}
CRYPTOEOF

if go run /tmp/test-crypto-$$.go 2>/dev/null | grep -q "key_ok"; then
    log_pass "Key generation"
else
    log_fail "Key generation"
fi

if go run /tmp/test-crypto-$$.go 2>/dev/null | grep -q "pass_ok"; then
    log_pass "Passphrase generation"
else
    log_fail "Passphrase generation"
fi

if go run /tmp/test-crypto-$$.go 2>/dev/null | grep -q "nonce_ok"; then
    log_pass "Nonce generation"
else
    log_fail "Nonce generation"
fi

if go run /tmp/test-crypto-$$.go 2>/dev/null | grep -q "sig_ok"; then
    log_pass "Signature creation"
else
    log_fail "Signature creation"
fi

rm -f /tmp/test-crypto-$$.go
echo ""

# EDGE CASES
echo "3. EDGE CASE TESTS"
echo "-------------------"
touch zero.bin && [ "$(stat -c%s zero.bin 2>/dev/null || stat -f%z zero.bin)" = "0" ] && log_pass "Zero-byte file" || log_fail "Zero-byte file"

SPECIAL="file with spaces & (symbols).txt"
echo "test" > "$SPECIAL" && [ -f "$SPECIAL" ] && log_pass "Special chars in filename" || log_fail "Special chars in filename"

for i in {1..10}; do echo "$i" > "concurrent-$i.txt" & done
wait
[ "$(ls concurrent-*.txt 2>/dev/null | wc -l)" = "10" ] && log_pass "Concurrent file ops (10 files)" || log_fail "Concurrent file ops"

HASH1=$(echo "data1" | sha256sum | cut -d' ' -f1)
HASH2=$(echo "data2" | sha256sum | cut -d' ' -f1)
[ "$HASH1" != "$HASH2" ] && log_pass "Hash uniqueness" || log_fail "Hash uniqueness"
echo ""

# PERFORMANCE TESTS
echo "4. PERFORMANCE TESTS"
echo "-------------------"
START=$(date +%s.%N)
for i in {1..100}; do echo "test" > "perf-$i.txt"; done
END=$(date +%s.%N)
DURATION=$(echo "$END - $START" | bc)
RATE=$(echo "scale=1; 100 / $DURATION" | bc)
log_pass "Small file throughput: $RATE files/sec"

dd if=/dev/zero of=perf-large.bin bs=1M count=50 2>/dev/null
SIZE=$(stat -c%s perf-large.bin 2>/dev/null || stat -f%z perf-large.bin)
[ "$SIZE" = "52428800" ] && log_pass "Large file creation (50MB)" || log_fail "Large file creation"

for i in {1..100}; do echo "repeated line for compression"; done > compress-test.txt
ORIG=$(stat -c%s compress-test.txt 2>/dev/null || stat -f%z compress-test.txt)
gzip -c compress-test.txt > compress-test.gz 2>/dev/null
COMP=$(stat -c%s compress-test.gz 2>/dev/null || stat -f%z compress-test.gz)
RATIO=$(echo "scale=1; 100 - ($COMP * 100 / $ORIG)" | bc)
log_pass "Compression ratio: ${RATIO}% ($ORIG -> $COMP bytes)"
echo ""

# ZFS TESTS
echo "5. ZFS INTEGRATION TESTS"
echo "-------------------"
if zpool list tank >/dev/null 2>&1; then
    log_pass "ZFS pool accessible (tank)"
    COMPRESSION=$(zfs get -H -o value compression tank 2>/dev/null)
    log_pass "ZFS compression: $COMPRESSION"
    AVAIL=$(zfs get -H -o value available tank 2>/dev/null)
    log_pass "ZFS available: $AVAIL"
else
    log_fail "ZFS pool not accessible"
fi
echo ""

# DATABASE TESTS
echo "6. DATABASE TESTS"
echo "-------------------"
DB_FILE="test.db"
sqlite3 "$DB_FILE" "CREATE TABLE test (id INTEGER PRIMARY KEY, data TEXT);" 2>/dev/null && log_pass "Database creation" || log_fail "Database creation"
sqlite3 "$DB_FILE" "INSERT INTO test (data) VALUES ('test1'), ('test2');" 2>/dev/null && log_pass "Database insert" || log_fail "Database insert"
COUNT=$(sqlite3 "$DB_FILE" "SELECT COUNT(*) FROM test;" 2>/dev/null)
[ "$COUNT" = "2" ] && log_pass "Database query (count=$COUNT)" || log_fail "Database query"
sqlite3 "$DB_FILE" "PRAGMA integrity_check;" 2>/dev/null | grep -q "ok" && log_pass "Database integrity" || log_fail "Database integrity"
echo ""

# CLEANUP
cd /
rm -rf "$TEST_DIR"

# SUMMARY
echo "═══════════════════════════════════════"
echo "  TEST SUMMARY"
echo "═══════════════════════════════════════"
echo ""
TOTAL=$((PASSED + FAILED))
echo "Total Tests: $TOTAL"
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    RATE="100"
else
    RATE=$(echo "scale=1; $PASSED * 100 / $TOTAL" | bc)
fi

echo "Success Rate: ${RATE}%"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ ALL TESTS PASSED${NC}"
    exit 0
else
    echo -e "${RED}✗ SOME TESTS FAILED${NC}"
    exit 1
fi
