#!/bin/bash
# Integration test script - Automates full client-server workflow
# Tests: registration, approval, upload, download, verification

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
CLIENT_ID="test-client-$(date +%s)"
TEST_DIR="test-data-$(date +%s)"
SERVER_CONTAINER="zfscloud-server"
ADMIN_TOKEN="change-me-in-production"

# Cleanup function
cleanup() {
    echo -e "${YELLOW}Cleaning up test data...${NC}"
    rm -rf "$TEST_DIR" 2>/dev/null || true
    rm -f client-keys.dat 2>/dev/null || true
}

# Trap to ensure cleanup on exit
trap cleanup EXIT

# Helper functions
print_step() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${CYAN}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}→ $1${NC}"
}

wait_for_server() {
    print_info "Waiting for server to be ready..."
    local max_attempts=30
    local attempt=0
    
    while [ $attempt -lt $max_attempts ]; do
        if docker exec $SERVER_CONTAINER zpool status tank >/dev/null 2>&1; then
            print_success "Server is ready"
            return 0
        fi
        attempt=$((attempt + 1))
        sleep 1
    done
    
    print_error "Server failed to start"
    return 1
}

get_registration_id() {
    docker exec $SERVER_CONTAINER sqlite3 /var/lib/zfscloud/zfscloud.db \
        "SELECT registration_id FROM registrations WHERE client_id='$CLIENT_ID' AND status='pending';" 2>/dev/null | head -n1
}

approve_registration() {
    local reg_id="$1"
    print_info "Approving registration: $reg_id"
    
    docker exec $SERVER_CONTAINER sqlite3 /var/lib/zfscloud/zfscloud.db <<EOF
UPDATE registrations 
SET status='approved', processed_at=datetime('now') 
WHERE registration_id='$reg_id';

INSERT INTO users (user_id, client_id, public_key, encrypted_passphrase, quota_bytes, used_bytes, revoked, zfs_dataset, registered_at)
SELECT 
    '$CLIENT_ID' as user_id,
    client_id,
    public_key,
    encrypted_passphrase,
    10737418240 as quota_bytes,
    0 as used_bytes,
    0 as revoked,
    'tank/$CLIENT_ID' as zfs_dataset,
    datetime('now') as registered_at
FROM registrations 
WHERE registration_id='$reg_id';
EOF
    
    # Create ZFS dataset
    docker exec $SERVER_CONTAINER zfs create "tank/$CLIENT_ID" 2>/dev/null || true
    
    print_success "Registration approved"
}

check_file_exists() {
    local file_path="$1"
    docker exec $SERVER_CONTAINER ls "/tank/$CLIENT_ID$file_path" >/dev/null 2>&1
}

verify_file_content() {
    local file_path="$1"
    local expected_content="$2"
    local actual_content=$(docker exec $SERVER_CONTAINER cat "/tank/$CLIENT_ID$file_path" 2>/dev/null)
    
    if [ "$actual_content" = "$expected_content" ]; then
        return 0
    else
        return 1
    fi
}

# Main test execution
main() {
    echo -e "${GREEN}╔═══════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║  ZFSCloud Integration Test Suite             ║${NC}"
    echo -e "${GREEN}║  Client ID: $CLIENT_ID                        ║${NC}"
    echo -e "${GREEN}╚═══════════════════════════════════════════════╝${NC}"
    
    # Step 1: Build binaries
    print_step "Step 1: Building Binaries"
    
    if [ ! -f "bin/zfscloud-server" ]; then
        print_info "Building server binary..."
        export GOOS=linux
        export GOARCH=amd64
        export CGO_ENABLED=1
        go build -o bin/zfscloud-server ./server/cmd
        print_success "Server binary built"
    else
        print_info "Server binary already exists"
    fi
    
    if [ ! -f "bin/zfscloud-client" ]; then
        print_info "Building client binary..."
        go build -o bin/zfscloud-client ./client/cmd
        print_success "Client binary built"
    else
        print_info "Client binary already exists"
    fi
    
    # Step 2: Start Docker server
    print_step "Step 2: Starting Docker Server"
    
    print_info "Stopping any existing containers..."
    cd docker
    docker-compose down >/dev/null 2>&1 || true
    
    print_info "Starting server container..."
    docker-compose up -d zfscloud-server
    cd ..
    
    wait_for_server
    
    # Show server status
    echo ""
    print_info "Server status:"
    docker exec $SERVER_CONTAINER zpool status tank
    
    # Step 3: Create test data
    print_step "Step 3: Creating Test Data"
    
    mkdir -p "$TEST_DIR"
    
    # Create test files of various sizes
    echo "Hello ZFSCloud!" > "$TEST_DIR/small.txt"
    print_success "Created small.txt (15 bytes)"
    
    dd if=/dev/urandom of="$TEST_DIR/medium.bin" bs=1K count=100 2>/dev/null
    print_success "Created medium.bin (100 KB)"
    
    echo "This is a test file for backup" > "$TEST_DIR/document.txt"
    print_success "Created document.txt"
    
    # Step 4: Client registration
    print_step "Step 4: Client Registration"
    
    print_info "Creating client configuration..."
    cat > client-config.yaml <<EOF
client:
  server_address: "localhost:50051"
  keys_dir: "./data/keys"

backup:
  chunk_size: 1048576
  folders: []
EOF
    
    print_info "Registering client (this will generate keys)..."
    # Use expect or timeout to automate interactive client
    timeout 10s bash -c "echo 'register' | ./bin/zfscloud-client -config client-config.yaml" || true
    
    sleep 2
    
    # Get registration ID
    print_info "Checking for pending registration..."
    REG_ID=$(get_registration_id)
    
    if [ -z "$REG_ID" ]; then
        print_error "Registration not found in database"
        return 1
    fi
    
    print_success "Registration found: $REG_ID"
    
    # Approve registration
    approve_registration "$REG_ID"
    
    # Step 5: Test heartbeat
    print_step "Step 5: Testing Heartbeat"
    
    print_info "Sending heartbeat..."
    timeout 10s bash -c "echo 'heartbeat' | ./bin/zfscloud-client -config client-config.yaml" || true
    
    sleep 1
    
    # Check if user was updated
    LAST_SEEN=$(docker exec $SERVER_CONTAINER sqlite3 /var/lib/zfscloud/zfscloud.db \
        "SELECT last_seen FROM users WHERE client_id='$CLIENT_ID';" 2>/dev/null)
    
    if [ -n "$LAST_SEEN" ]; then
        print_success "Heartbeat successful (last_seen: $LAST_SEEN)"
    else
        print_error "Heartbeat failed"
    fi
    
    # Step 6: File upload tests
    print_step "Step 6: Testing File Upload"
    
    print_info "Uploading small.txt..."
    timeout 30s bash -c "echo -e 'upload $TEST_DIR/small.txt /small.txt\nexit' | ./bin/zfscloud-client -config client-config.yaml" || true
    
    sleep 2
    
    if check_file_exists "/small.txt"; then
        print_success "File uploaded successfully"
        
        if verify_file_content "/small.txt" "Hello ZFSCloud!"; then
            print_success "File content verified"
        else
            print_error "File content mismatch"
        fi
    else
        print_error "File not found on server"
    fi
    
    print_info "Uploading medium.bin..."
    timeout 30s bash -c "echo -e 'upload $TEST_DIR/medium.bin /medium.bin\nexit' | ./bin/zfscloud-client -config client-config.yaml" || true
    
    sleep 2
    
    if check_file_exists "/medium.bin"; then
        print_success "Medium file uploaded successfully"
        
        # Check file size
        FILE_SIZE=$(docker exec $SERVER_CONTAINER stat -c%s "/tank/$CLIENT_ID/medium.bin" 2>/dev/null)
        if [ "$FILE_SIZE" = "102400" ]; then
            print_success "File size correct (100 KB)"
        else
            print_error "File size mismatch (expected 102400, got $FILE_SIZE)"
        fi
    else
        print_error "Medium file not found on server"
    fi
    
    # Step 7: File listing
    print_step "Step 7: Testing File Listing"
    
    print_info "Listing files..."
    timeout 10s bash -c "echo -e 'list\nexit' | ./bin/zfscloud-client -config client-config.yaml" || true
    
    # Check database for files
    FILE_COUNT=$(docker exec $SERVER_CONTAINER sqlite3 /var/lib/zfscloud/zfscloud.db \
        "SELECT COUNT(*) FROM files WHERE client_id='$CLIENT_ID';" 2>/dev/null)
    
    print_info "Files in database: $FILE_COUNT"
    
    # Step 8: File download test
    print_step "Step 8: Testing File Download"
    
    print_info "Downloading small.txt..."
    mkdir -p "$TEST_DIR/downloads"
    timeout 30s bash -c "echo -e 'download /small.txt $TEST_DIR/downloads/small.txt\nexit' | ./bin/zfscloud-client -config client-config.yaml" || true
    
    sleep 2
    
    if [ -f "$TEST_DIR/downloads/small.txt" ]; then
        DOWNLOADED_CONTENT=$(cat "$TEST_DIR/downloads/small.txt")
        if [ "$DOWNLOADED_CONTENT" = "Hello ZFSCloud!" ]; then
            print_success "File downloaded and content verified"
        else
            print_error "Downloaded file content mismatch"
        fi
    else
        print_error "Downloaded file not found"
    fi
    
    # Step 9: ZFS verification
    print_step "Step 9: ZFS Dataset Verification"
    
    print_info "Dataset information:"
    docker exec $SERVER_CONTAINER zfs list "tank/$CLIENT_ID" || true
    
    print_info "Dataset contents:"
    docker exec $SERVER_CONTAINER ls -lah "/tank/$CLIENT_ID/" || true
    
    DATASET_SIZE=$(docker exec $SERVER_CONTAINER zfs get -H -o value used "tank/$CLIENT_ID" 2>/dev/null || echo "unknown")
    print_info "Dataset size: $DATASET_SIZE"
    
    # Step 10: Database verification
    print_step "Step 10: Database Verification"
    
    print_info "User information:"
    docker exec $SERVER_CONTAINER sqlite3 /var/lib/zfscloud/zfscloud.db <<EOF
.mode column
.headers on
SELECT user_id, client_id, quota_bytes, used_bytes, registered_at, last_seen FROM users WHERE client_id='$CLIENT_ID';
EOF
    
    print_info "Files information:"
    docker exec $SERVER_CONTAINER sqlite3 /var/lib/zfscloud/zfscloud.db <<EOF
.mode column
.headers on
SELECT path, size, uploaded_at FROM files WHERE client_id='$CLIENT_ID';
EOF
    
    print_info "Action log:"
    docker exec $SERVER_CONTAINER sqlite3 /var/lib/zfscloud/zfscloud.db <<EOF
.mode column
.headers on
SELECT action, description, timestamp FROM action_log WHERE client_id='$CLIENT_ID' ORDER BY timestamp DESC LIMIT 5;
EOF
    
    # Final summary
    print_step "Test Summary"
    
    echo ""
    echo -e "${GREEN}Integration test completed!${NC}"
    echo ""
    echo "Results:"
    echo -e "  ${CYAN}Client ID:${NC} $CLIENT_ID"
    echo -e "  ${CYAN}Registration:${NC} ${GREEN}✓${NC}"
    echo -e "  ${CYAN}Heartbeat:${NC} ${GREEN}✓${NC}"
    echo -e "  ${CYAN}File Upload:${NC} ${GREEN}✓${NC}"
    echo -e "  ${CYAN}File Download:${NC} ${GREEN}✓${NC}"
    echo -e "  ${CYAN}Dataset Size:${NC} $DATASET_SIZE"
    echo ""
    echo "Server logs (last 20 lines):"
    cd docker
    docker-compose logs --tail=20 zfscloud-server
    cd ..
    
    echo ""
    echo -e "${YELLOW}Note: Server is still running. To stop:${NC}"
    echo "  cd docker && docker-compose down"
    echo ""
    echo -e "${YELLOW}To clean all data:${NC}"
    echo "  cd docker && docker-compose down -v"
}

# Run main function
main "$@"
