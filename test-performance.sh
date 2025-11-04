#!/bin/bash
# Performance test script - Tests with larger files and multiple concurrent clients
# Measures upload/download speeds, concurrent operations, and system load

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
NC='\033[0m'

# Configuration
NUM_CLIENTS=${NUM_CLIENTS:-3}
FILE_SIZE_SMALL=${FILE_SIZE_SMALL:-1}    # MB
FILE_SIZE_MEDIUM=${FILE_SIZE_MEDIUM:-10} # MB
FILE_SIZE_LARGE=${FILE_SIZE_LARGE:-50}   # MB
NUM_FILES_PER_CLIENT=${NUM_FILES_PER_CLIENT:-10}
SERVER_CONTAINER="zfscloud-server"
TEST_DIR="perf-test-$(date +%s)"
RESULTS_FILE="perf-results-$(date +%s).txt"

# Cleanup function
cleanup() {
    echo -e "${YELLOW}Cleaning up test data...${NC}"
    rm -rf "$TEST_DIR" 2>/dev/null || true
    for i in $(seq 1 $NUM_CLIENTS); do
        rm -f "client-keys-$i.dat" 2>/dev/null || true
        rm -f "client-config-$i.yaml" 2>/dev/null || true
    done
}

trap cleanup EXIT

# Helper functions
print_header() {
    echo ""
    echo -e "${MAGENTA}╔════════════════════════════════════════════════════╗${NC}"
    echo -e "${MAGENTA}║  $1${NC}"
    echo -e "${MAGENTA}╚════════════════════════════════════════════════════╝${NC}"
}

print_step() {
    echo ""
    echo -e "${BLUE}>>> $1${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${CYAN}→ $1${NC}"
}

print_metric() {
    echo -e "${YELLOW}  $1: ${GREEN}$2${NC}"
}

log_result() {
    echo "$1" >> "$RESULTS_FILE"
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

create_test_file() {
    local size_mb=$1
    local filename=$2
    
    dd if=/dev/urandom of="$filename" bs=1M count=$size_mb 2>/dev/null
}

register_client() {
    local client_id=$1
    local config_file=$2
    
    print_info "Registering client: $client_id"
    
    # Create client config
    cat > "$config_file" <<EOF
client:
  server_address: "localhost:50051"
  keys_dir: "./data/keys-$client_id"

backup:
  chunk_size: 1048576
  folders: []
EOF
    
    # Register (timeout to avoid hanging)
    timeout 10s bash -c "echo 'register' | ./bin/zfscloud-client -config $config_file" >/dev/null 2>&1 || true
    
    sleep 1
    
    # Get registration ID
    local reg_id=$(docker exec $SERVER_CONTAINER sqlite3 /var/lib/zfscloud/zfscloud.db \
        "SELECT registration_id FROM registrations WHERE client_id='$client_id' AND status='pending';" 2>/dev/null | head -n1)
    
    if [ -z "$reg_id" ]; then
        print_error "Registration failed for $client_id"
        return 1
    fi
    
    # Approve registration
    docker exec $SERVER_CONTAINER sqlite3 /var/lib/zfscloud/zfscloud.db <<EOF >/dev/null 2>&1
UPDATE registrations 
SET status='approved', processed_at=datetime('now') 
WHERE registration_id='$reg_id';

INSERT INTO users (user_id, client_id, public_key, encrypted_passphrase, quota_bytes, used_bytes, revoked, zfs_dataset, registered_at)
SELECT 
    '$client_id' as user_id,
    client_id,
    public_key,
    encrypted_passphrase,
    107374182400 as quota_bytes,
    0 as used_bytes,
    0 as revoked,
    'tank/$client_id' as zfs_dataset,
    datetime('now') as registered_at
FROM registrations 
WHERE registration_id='$reg_id';
EOF
    
    # Create ZFS dataset
    docker exec $SERVER_CONTAINER zfs create "tank/$client_id" 2>/dev/null || true
    
    print_success "Client registered: $client_id"
}

measure_upload() {
    local client_id=$1
    local config_file=$2
    local file_path=$3
    local remote_path=$4
    
    local start_time=$(date +%s.%N)
    
    timeout 300s bash -c "echo -e 'upload $file_path $remote_path\nexit' | ./bin/zfscloud-client -config $config_file" >/dev/null 2>&1 || true
    
    local end_time=$(date +%s.%N)
    local duration=$(echo "$end_time - $start_time" | bc)
    
    # Check if file exists
    if docker exec $SERVER_CONTAINER ls "/tank/$client_id$remote_path" >/dev/null 2>&1; then
        local file_size=$(stat -f%z "$file_path" 2>/dev/null || stat -c%s "$file_path" 2>/dev/null)
        local speed=$(echo "scale=2; $file_size / 1024 / 1024 / $duration" | bc)
        echo "$duration|$speed"
    else
        echo "0|0"
    fi
}

measure_download() {
    local client_id=$1
    local config_file=$2
    local remote_path=$3
    local local_path=$4
    
    local start_time=$(date +%s.%N)
    
    timeout 300s bash -c "echo -e 'download $remote_path $local_path\nexit' | ./bin/zfscloud-client -config $config_file" >/dev/null 2>&1 || true
    
    local end_time=$(date +%s.%N)
    local duration=$(echo "$end_time - $start_time" | bc)
    
    if [ -f "$local_path" ]; then
        local file_size=$(stat -f%z "$local_path" 2>/dev/null || stat -c%s "$local_path" 2>/dev/null)
        local speed=$(echo "scale=2; $file_size / 1024 / 1024 / $duration" | bc)
        echo "$duration|$speed"
    else
        echo "0|0"
    fi
}

run_concurrent_uploads() {
    local num_clients=$1
    local file_size_mb=$2
    
    print_step "Starting concurrent uploads ($num_clients clients, ${file_size_mb}MB files)"
    
    local pids=()
    local start_time=$(date +%s)
    
    for i in $(seq 1 $num_clients); do
        local client_id="perf-client-$i-$(date +%s)"
        local config_file="client-config-$i.yaml"
        local test_file="$TEST_DIR/concurrent-${file_size_mb}mb-$i.bin"
        
        (
            create_test_file $file_size_mb "$test_file"
            register_client "$client_id" "$config_file"
            measure_upload "$client_id" "$config_file" "$test_file" "/concurrent-$i.bin"
        ) &
        pids+=($!)
    done
    
    # Wait for all uploads to complete
    for pid in "${pids[@]}"; do
        wait $pid
    done
    
    local end_time=$(date +%s)
    local total_duration=$((end_time - start_time))
    local total_data=$((file_size_mb * num_clients))
    local throughput=$(echo "scale=2; $total_data / $total_duration" | bc)
    
    print_success "Concurrent uploads completed"
    print_metric "Total Duration" "${total_duration}s"
    print_metric "Total Data" "${total_data}MB"
    print_metric "Aggregate Throughput" "${throughput} MB/s"
    
    log_result "Concurrent Upload Test ($num_clients clients, ${file_size_mb}MB each):"
    log_result "  Duration: ${total_duration}s"
    log_result "  Total Data: ${total_data}MB"
    log_result "  Throughput: ${throughput} MB/s"
    log_result ""
}

# Main test execution
main() {
    print_header "ZFSCloud Performance Test Suite"
    
    echo "Configuration:"
    echo -e "  ${CYAN}Number of clients:${NC} $NUM_CLIENTS"
    echo -e "  ${CYAN}Small file size:${NC} ${FILE_SIZE_SMALL}MB"
    echo -e "  ${CYAN}Medium file size:${NC} ${FILE_SIZE_MEDIUM}MB"
    echo -e "  ${CYAN}Large file size:${NC} ${FILE_SIZE_LARGE}MB"
    echo -e "  ${CYAN}Files per client:${NC} $NUM_FILES_PER_CLIENT"
    echo ""
    
    # Initialize results file
    log_result "ZFSCloud Performance Test Results"
    log_result "Date: $(date)"
    log_result "Clients: $NUM_CLIENTS"
    log_result "================================================"
    log_result ""
    
    # Step 1: Build binaries
    print_header "Step 1: Building Binaries"
    
    if [ ! -f "bin/zfscloud-server" ]; then
        print_info "Building server binary..."
        export GOOS=linux
        export GOARCH=amd64
        export CGO_ENABLED=1
        go build -o bin/zfscloud-server ./server/cmd
        print_success "Server binary built"
    fi
    
    if [ ! -f "bin/zfscloud-client" ]; then
        print_info "Building client binary..."
        go build -o bin/zfscloud-client ./client/cmd
        print_success "Client binary built"
    fi
    
    # Step 2: Start server
    print_header "Step 2: Starting Docker Server"
    
    print_info "Stopping any existing containers..."
    cd docker
    docker-compose down -v >/dev/null 2>&1 || true
    
    print_info "Starting fresh server container..."
    docker-compose up -d zfscloud-server
    cd ..
    
    wait_for_server
    
    # Create test directory
    mkdir -p "$TEST_DIR/downloads"
    
    # Step 3: Single client upload performance
    print_header "Step 3: Single Client Upload Performance"
    
    CLIENT_ID="perf-single-$(date +%s)"
    CONFIG_FILE="client-config-single.yaml"
    
    register_client "$CLIENT_ID" "$CONFIG_FILE"
    
    # Test small file
    print_step "Testing small file (${FILE_SIZE_SMALL}MB)"
    create_test_file $FILE_SIZE_SMALL "$TEST_DIR/small.bin"
    RESULT=$(measure_upload "$CLIENT_ID" "$CONFIG_FILE" "$TEST_DIR/small.bin" "/small.bin")
    DURATION=$(echo $RESULT | cut -d'|' -f1)
    SPEED=$(echo $RESULT | cut -d'|' -f2)
    print_metric "Duration" "${DURATION}s"
    print_metric "Speed" "${SPEED} MB/s"
    log_result "Single Upload - Small (${FILE_SIZE_SMALL}MB): ${DURATION}s @ ${SPEED} MB/s"
    
    # Test medium file
    print_step "Testing medium file (${FILE_SIZE_MEDIUM}MB)"
    create_test_file $FILE_SIZE_MEDIUM "$TEST_DIR/medium.bin"
    RESULT=$(measure_upload "$CLIENT_ID" "$CONFIG_FILE" "$TEST_DIR/medium.bin" "/medium.bin")
    DURATION=$(echo $RESULT | cut -d'|' -f1)
    SPEED=$(echo $RESULT | cut -d'|' -f2)
    print_metric "Duration" "${DURATION}s"
    print_metric "Speed" "${SPEED} MB/s"
    log_result "Single Upload - Medium (${FILE_SIZE_MEDIUM}MB): ${DURATION}s @ ${SPEED} MB/s"
    
    # Test large file
    print_step "Testing large file (${FILE_SIZE_LARGE}MB)"
    create_test_file $FILE_SIZE_LARGE "$TEST_DIR/large.bin"
    RESULT=$(measure_upload "$CLIENT_ID" "$CONFIG_FILE" "$TEST_DIR/large.bin" "/large.bin")
    DURATION=$(echo $RESULT | cut -d'|' -f1)
    SPEED=$(echo $RESULT | cut -d'|' -f2)
    print_metric "Duration" "${DURATION}s"
    print_metric "Speed" "${SPEED} MB/s"
    log_result "Single Upload - Large (${FILE_SIZE_LARGE}MB): ${DURATION}s @ ${SPEED} MB/s"
    log_result ""
    
    # Step 4: Single client download performance
    print_header "Step 4: Single Client Download Performance"
    
    # Test small file download
    print_step "Testing small file download (${FILE_SIZE_SMALL}MB)"
    RESULT=$(measure_download "$CLIENT_ID" "$CONFIG_FILE" "/small.bin" "$TEST_DIR/downloads/small.bin")
    DURATION=$(echo $RESULT | cut -d'|' -f1)
    SPEED=$(echo $RESULT | cut -d'|' -f2)
    print_metric "Duration" "${DURATION}s"
    print_metric "Speed" "${SPEED} MB/s"
    log_result "Single Download - Small (${FILE_SIZE_SMALL}MB): ${DURATION}s @ ${SPEED} MB/s"
    
    # Test medium file download
    print_step "Testing medium file download (${FILE_SIZE_MEDIUM}MB)"
    RESULT=$(measure_download "$CLIENT_ID" "$CONFIG_FILE" "/medium.bin" "$TEST_DIR/downloads/medium.bin")
    DURATION=$(echo $RESULT | cut -d'|' -f1)
    SPEED=$(echo $RESULT | cut -d'|' -f2)
    print_metric "Duration" "${DURATION}s"
    print_metric "Speed" "${SPEED} MB/s"
    log_result "Single Download - Medium (${FILE_SIZE_MEDIUM}MB): ${DURATION}s @ ${SPEED} MB/s"
    log_result ""
    
    # Step 5: Multiple files upload
    print_header "Step 5: Multiple Files Upload (${NUM_FILES_PER_CLIENT} files)"
    
    print_info "Creating ${NUM_FILES_PER_CLIENT} test files..."
    for i in $(seq 1 $NUM_FILES_PER_CLIENT); do
        create_test_file $FILE_SIZE_SMALL "$TEST_DIR/multi-$i.bin"
    done
    print_success "Test files created"
    
    START_TIME=$(date +%s)
    for i in $(seq 1 $NUM_FILES_PER_CLIENT); do
        timeout 60s bash -c "echo -e 'upload $TEST_DIR/multi-$i.bin /multi-$i.bin\nexit' | ./bin/zfscloud-client -config $CONFIG_FILE" >/dev/null 2>&1 || true
    done
    END_TIME=$(date +%s)
    
    TOTAL_DURATION=$((END_TIME - START_TIME))
    TOTAL_DATA=$((FILE_SIZE_SMALL * NUM_FILES_PER_CLIENT))
    THROUGHPUT=$(echo "scale=2; $TOTAL_DATA / $TOTAL_DURATION" | bc)
    
    print_metric "Total Duration" "${TOTAL_DURATION}s"
    print_metric "Total Data" "${TOTAL_DATA}MB"
    print_metric "Average Throughput" "${THROUGHPUT} MB/s"
    log_result "Multiple Files Upload (${NUM_FILES_PER_CLIENT} x ${FILE_SIZE_SMALL}MB):"
    log_result "  Duration: ${TOTAL_DURATION}s"
    log_result "  Throughput: ${THROUGHPUT} MB/s"
    log_result ""
    
    # Step 6: Concurrent clients test
    print_header "Step 6: Concurrent Clients Test"
    
    run_concurrent_uploads $NUM_CLIENTS $FILE_SIZE_MEDIUM
    
    # Step 7: System metrics
    print_header "Step 7: System Metrics"
    
    print_info "ZFS pool status:"
    docker exec $SERVER_CONTAINER zpool status tank
    
    print_info "ZFS pool I/O stats:"
    docker exec $SERVER_CONTAINER zpool iostat tank
    
    print_info "Dataset list:"
    docker exec $SERVER_CONTAINER zfs list -r tank
    
    print_info "Compression ratio:"
    docker exec $SERVER_CONTAINER zfs get compressratio tank
    
    # Database stats
    print_step "Database Statistics"
    
    TOTAL_USERS=$(docker exec $SERVER_CONTAINER sqlite3 /var/lib/zfscloud/zfscloud.db \
        "SELECT COUNT(*) FROM users;" 2>/dev/null)
    TOTAL_FILES=$(docker exec $SERVER_CONTAINER sqlite3 /var/lib/zfscloud/zfscloud.db \
        "SELECT COUNT(*) FROM files;" 2>/dev/null)
    TOTAL_SIZE=$(docker exec $SERVER_CONTAINER sqlite3 /var/lib/zfscloud/zfscloud.db \
        "SELECT SUM(size) FROM files;" 2>/dev/null)
    
    print_metric "Total Users" "$TOTAL_USERS"
    print_metric "Total Files" "$TOTAL_FILES"
    print_metric "Total Size" "$TOTAL_SIZE bytes"
    
    log_result "System Metrics:"
    log_result "  Users: $TOTAL_USERS"
    log_result "  Files: $TOTAL_FILES"
    log_result "  Total Size: $TOTAL_SIZE bytes"
    log_result ""
    
    # Final summary
    print_header "Performance Test Complete"
    
    echo ""
    echo -e "${GREEN}Results saved to: $RESULTS_FILE${NC}"
    echo ""
    echo "Summary:"
    cat "$RESULTS_FILE"
    
    echo ""
    echo -e "${YELLOW}Server is still running. To stop:${NC}"
    echo "  cd docker && docker-compose down"
    echo ""
    echo -e "${YELLOW}To clean all data:${NC}"
    echo "  cd docker && docker-compose down -v"
}

# Run main function
main "$@"
