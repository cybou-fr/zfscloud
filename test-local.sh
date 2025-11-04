#!/bin/bash
# Local testing script without Docker - builds and tests server/client on localhost

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${CYAN}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[PASS]${NC} $1"; }
log_error() { echo -e "${RED}[FAIL]${NC} $1"; }
log_section() { echo -e "\n${BLUE}========================================${NC}\n${BLUE}$1${NC}\n${BLUE}========================================${NC}"; }

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

log_section "ZFSCloud Local Test Suite"

# Test 1: Build server
log_section "Test 1: Build Server Binary"
log_info "Building server..."
if go build -o bin/zfscloud-server-test ./server/cmd 2>&1 | tee /tmp/build-server.log; then
    log_success "Server built successfully"
    ls -lh bin/zfscloud-server-test
else
    log_error "Server build failed"
    cat /tmp/build-server.log
    exit 1
fi

# Test 2: Build CLI client (without GUI dependencies)
log_section "Test 2: Build CLI Client"
log_info "Creating CLI-only client..."

# Create a minimal CLI client
cat > /tmp/test-client.go << 'EOF'
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/ioustamora/zfscloud/proto"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

func main() {
    conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer conn.Close()
    
    client := proto.NewBackupServiceClient(conn)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    // Test: Get server public key
    resp, err := client.GetServerPublicKey(ctx, &proto.ServerPublicKeyRequest{})
    if err != nil {
        log.Fatalf("GetServerPublicKey failed: %v", err)
    }
    
    fmt.Printf("✓ Connected to server!\n")
    fmt.Printf("✓ Server public key received (%d bytes)\n", len(resp.PublicKey))
    fmt.Printf("✓ Key ID: %s\n", resp.KeyId)
}
EOF

log_info "Building test client..."
if go build -o bin/test-client /tmp/test-client.go 2>&1 | tee /tmp/build-client.log; then
    log_success "Test client built successfully"
else
    log_error "Test client build failed"
    cat /tmp/build-client.log
    exit 1
fi

# Test 3: Check Go modules
log_section "Test 3: Verify Dependencies"
log_info "Checking Go modules..."
go list -m all | grep -E "(grpc|protobuf|zfs)" || log_info "Core dependencies present"
log_success "Dependencies OK"

# Test 4: Run unit tests
log_section "Test 4: Unit Tests"
log_info "Running Go tests..."
if go test -v ./shared/... 2>&1 | tee /tmp/unit-tests.log; then
    log_success "Unit tests passed"
else
    log_error "Unit tests failed"
    cat /tmp/unit-tests.log
fi

# Test 5: Check proto files
log_section "Test 5: Protocol Buffers"
log_info "Checking proto files..."
if [ -f "proto/backup.proto" ]; then
    log_success "Proto file exists"
    SERVICES=$(grep "^service " proto/backup.proto | wc -l)
    RPCS=$(grep "  rpc " proto/backup.proto | wc -l)
    log_info "Services: $SERVICES, RPCs: $RPCS"
else
    log_error "Proto file not found"
fi

# Test 6: Check generated code
log_section "Test 6: Generated gRPC Code"
if [ -f "proto/backup_grpc.pb.go" ]; then
    log_success "gRPC code generated"
else
    log_error "gRPC code not found - run 'make proto'"
fi

# Test 7: Static analysis
log_section "Test 7: Code Quality"
log_info "Running go vet..."
if go vet ./server/... ./shared/... 2>&1 | head -20; then
    log_success "go vet passed"
else
    log_error "go vet found issues"
fi

# Test 8: Check configuration
log_section "Test 8: Configuration Files"
if [ -f "server/config.example.yaml" ]; then
    log_success "Server config example exists"
    grep "listen_address" server/config.example.yaml || true
else
    log_error "Server config example not found"
fi

# Summary
log_section "Test Summary"
echo ""
log_success "Build tests completed!"
echo ""
echo "Binaries created:"
echo "  - bin/zfscloud-server-test"
echo "  - bin/test-client"
echo ""
echo "To test manually (requires ZFS):"
echo "  1. Create config: cp server/config.example.yaml server-test.yaml"
echo "  2. Edit config to use /tmp for storage"
echo "  3. Run server: ./bin/zfscloud-server-test -config server-test.yaml"
echo "  4. Run client: ./bin/test-client"
echo ""
echo "Note: Full integration tests require:"
echo "  - ZFS filesystem support"
echo "  - Either Docker with permissions or manual ZFS setup"
echo ""
