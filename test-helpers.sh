#!/bin/bash
# Helper utilities for Docker testing
# Common functions used across test scripts

# Colors
export RED='\033[0;31m'
export GREEN='\033[0;32m'
export YELLOW='\033[1;33m'
export CYAN='\033[0;36m'
export BLUE='\033[0;34m'
export MAGENTA='\033[0;35m'
export NC='\033[0m'

# Configuration
export SERVER_CONTAINER="${SERVER_CONTAINER:-zfscloud-server}"
export ADMIN_TOKEN="${ADMIN_TOKEN:-change-me-in-production}"
export DEFAULT_QUOTA="${DEFAULT_QUOTA:-10737418240}" # 10GB

# Logging functions
log_info() {
    echo -e "${CYAN}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_step() {
    echo ""
    echo -e "${BLUE}═══════════════════════════════════════${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}═══════════════════════════════════════${NC}"
}

# Docker management functions
docker_start_server() {
    log_info "Starting Docker server..."
    cd docker
    docker-compose up -d zfscloud-server
    local result=$?
    cd ..
    return $result
}

docker_stop_server() {
    log_info "Stopping Docker server..."
    cd docker
    docker-compose down
    cd ..
}

docker_clean_all() {
    log_warning "Cleaning all Docker data (volumes will be removed)..."
    cd docker
    docker-compose down -v
    cd ..
    log_success "All data cleaned"
}

docker_restart_server() {
    log_info "Restarting Docker server..."
    docker_stop_server
    sleep 2
    docker_start_server
}

docker_server_logs() {
    local lines="${1:-50}"
    docker logs --tail="$lines" "$SERVER_CONTAINER"
}

docker_follow_logs() {
    docker logs -f "$SERVER_CONTAINER"
}

# Server health checks
wait_for_server() {
    log_info "Waiting for server to be ready..."
    local max_attempts="${1:-30}"
    local attempt=0
    
    while [ $attempt -lt $max_attempts ]; do
        if docker exec "$SERVER_CONTAINER" zpool status tank >/dev/null 2>&1; then
            log_success "Server is ready"
            return 0
        fi
        attempt=$((attempt + 1))
        sleep 1
        if [ $((attempt % 5)) -eq 0 ]; then
            log_info "Still waiting... ($attempt/$max_attempts)"
        fi
    done
    
    log_error "Server failed to start within timeout"
    return 1
}

check_server_running() {
    if docker ps | grep -q "$SERVER_CONTAINER"; then
        return 0
    else
        return 1
    fi
}

# ZFS management functions
zfs_pool_status() {
    docker exec "$SERVER_CONTAINER" zpool status tank
}

zfs_list_datasets() {
    docker exec "$SERVER_CONTAINER" zfs list -r tank
}

zfs_dataset_size() {
    local dataset="$1"
    docker exec "$SERVER_CONTAINER" zfs get -H -o value used "tank/$dataset" 2>/dev/null
}

zfs_compression_ratio() {
    docker exec "$SERVER_CONTAINER" zfs get -H -o value compressratio tank 2>/dev/null
}

zfs_pool_iostat() {
    docker exec "$SERVER_CONTAINER" zpool iostat tank
}

zfs_create_dataset() {
    local dataset="$1"
    log_info "Creating ZFS dataset: tank/$dataset"
    docker exec "$SERVER_CONTAINER" zfs create "tank/$dataset" 2>/dev/null
    if [ $? -eq 0 ]; then
        log_success "Dataset created: tank/$dataset"
        return 0
    else
        log_error "Failed to create dataset: tank/$dataset"
        return 1
    fi
}

zfs_destroy_dataset() {
    local dataset="$1"
    log_warning "Destroying ZFS dataset: tank/$dataset"
    docker exec "$SERVER_CONTAINER" zfs destroy -r "tank/$dataset" 2>/dev/null
}

zfs_list_files() {
    local dataset="$1"
    local path="${2:-.}"
    docker exec "$SERVER_CONTAINER" ls -lah "/tank/$dataset/$path" 2>/dev/null
}

# Database management functions
db_exec() {
    local query="$1"
    docker exec "$SERVER_CONTAINER" sqlite3 /var/lib/zfscloud/zfscloud.db "$query" 2>/dev/null
}

db_get_user() {
    local client_id="$1"
    db_exec "SELECT * FROM users WHERE client_id='$client_id';"
}

db_list_users() {
    db_exec ".mode column
.headers on
SELECT user_id, client_id, quota_bytes, used_bytes, registered_at FROM users;"
}

db_list_pending_registrations() {
    db_exec ".mode column
.headers on
SELECT registration_id, client_id, hostname, status, requested_at FROM registrations WHERE status='pending';"
}

db_get_registration() {
    local client_id="$1"
    db_exec "SELECT registration_id FROM registrations WHERE client_id='$client_id' AND status='pending' LIMIT 1;" | head -n1
}

db_approve_registration() {
    local reg_id="$1"
    local quota="${2:-$DEFAULT_QUOTA}"
    
    log_info "Approving registration: $reg_id"
    
    # Get client_id from registration
    local client_id=$(db_exec "SELECT client_id FROM registrations WHERE registration_id='$reg_id';" | head -n1)
    
    if [ -z "$client_id" ]; then
        log_error "Registration not found: $reg_id"
        return 1
    fi
    
    # Update registration status
    db_exec "UPDATE registrations SET status='approved', processed_at=datetime('now') WHERE registration_id='$reg_id';"
    
    # Create user
    db_exec "INSERT INTO users (user_id, client_id, public_key, encrypted_passphrase, quota_bytes, used_bytes, revoked, zfs_dataset, registered_at)
SELECT 
    '$client_id' as user_id,
    client_id,
    public_key,
    encrypted_passphrase,
    $quota as quota_bytes,
    0 as used_bytes,
    0 as revoked,
    'tank/$client_id' as zfs_dataset,
    datetime('now') as registered_at
FROM registrations 
WHERE registration_id='$reg_id';"
    
    # Create ZFS dataset
    zfs_create_dataset "$client_id"
    
    log_success "Registration approved: $client_id"
}

db_list_files() {
    local client_id="${1:-}"
    if [ -n "$client_id" ]; then
        db_exec ".mode column
.headers on
SELECT path, size, uploaded_at FROM files WHERE client_id='$client_id';"
    else
        db_exec ".mode column
.headers on
SELECT client_id, path, size, uploaded_at FROM files;"
    fi
}

db_file_count() {
    local client_id="${1:-}"
    if [ -n "$client_id" ]; then
        db_exec "SELECT COUNT(*) FROM files WHERE client_id='$client_id';"
    else
        db_exec "SELECT COUNT(*) FROM files;"
    fi
}

db_total_size() {
    local client_id="${1:-}"
    if [ -n "$client_id" ]; then
        db_exec "SELECT COALESCE(SUM(size), 0) FROM files WHERE client_id='$client_id';"
    else
        db_exec "SELECT COALESCE(SUM(size), 0) FROM files;"
    fi
}

db_action_log() {
    local client_id="${1:-}"
    local limit="${2:-20}"
    if [ -n "$client_id" ]; then
        db_exec ".mode column
.headers on
SELECT action, description, timestamp FROM action_log WHERE client_id='$client_id' ORDER BY timestamp DESC LIMIT $limit;"
    else
        db_exec ".mode column
.headers on
SELECT client_id, action, description, timestamp FROM action_log ORDER BY timestamp DESC LIMIT $limit;"
    fi
}

db_interactive() {
    log_info "Starting interactive database session..."
    docker exec -it "$SERVER_CONTAINER" sqlite3 /var/lib/zfscloud/zfscloud.db
}

# Client management functions
generate_client_id() {
    echo "client-$(date +%s)-$(( RANDOM % 1000 ))"
}

create_client_config() {
    local client_id="$1"
    local config_file="$2"
    local server_address="${3:-localhost:50051}"
    
    cat > "$config_file" <<EOF
client:
  server_address: "$server_address"
  keys_dir: "./data/keys-$client_id"

backup:
  chunk_size: 1048576
  folders: []
EOF
    
    log_success "Client config created: $config_file"
}

register_and_approve_client() {
    local client_id="$1"
    local config_file="$2"
    local quota="${3:-$DEFAULT_QUOTA}"
    
    log_info "Registering client: $client_id"
    
    # Register client (with timeout)
    timeout 15s bash -c "echo 'register' | ./bin/zfscloud-client -config $config_file" >/dev/null 2>&1 || true
    
    sleep 2
    
    # Get registration ID
    local reg_id=$(db_get_registration "$client_id")
    
    if [ -z "$reg_id" ]; then
        log_error "Registration not found for client: $client_id"
        return 1
    fi
    
    # Approve registration
    db_approve_registration "$reg_id" "$quota"
    
    log_success "Client registered and approved: $client_id"
}

# File operations
create_test_file() {
    local size_mb="$1"
    local filename="$2"
    
    dd if=/dev/urandom of="$filename" bs=1M count="$size_mb" 2>/dev/null
    
    if [ $? -eq 0 ]; then
        log_success "Created test file: $filename (${size_mb}MB)"
        return 0
    else
        log_error "Failed to create test file: $filename"
        return 1
    fi
}

verify_file_on_server() {
    local client_id="$1"
    local file_path="$2"
    
    if docker exec "$SERVER_CONTAINER" ls "/tank/$client_id$file_path" >/dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

get_file_size_on_server() {
    local client_id="$1"
    local file_path="$2"
    
    docker exec "$SERVER_CONTAINER" stat -c%s "/tank/$client_id$file_path" 2>/dev/null
}

compare_file_hash() {
    local local_file="$1"
    local client_id="$2"
    local remote_path="$3"
    
    local local_hash=$(sha256sum "$local_file" | awk '{print $1}')
    local remote_hash=$(docker exec "$SERVER_CONTAINER" sha256sum "/tank/$client_id$remote_path" 2>/dev/null | awk '{print $1}')
    
    if [ "$local_hash" = "$remote_hash" ]; then
        log_success "File hashes match"
        return 0
    else
        log_error "File hashes do not match"
        return 1
    fi
}

# System information
show_system_info() {
    log_step "System Information"
    
    echo "Docker Container:"
    docker ps | grep "$SERVER_CONTAINER" || echo "  Not running"
    
    echo ""
    echo "ZFS Pool:"
    zfs_pool_status
    
    echo ""
    echo "Datasets:"
    zfs_list_datasets
    
    echo ""
    echo "Database Stats:"
    echo "  Users: $(db_exec 'SELECT COUNT(*) FROM users;')"
    echo "  Files: $(db_file_count)"
    echo "  Total Size: $(db_total_size) bytes"
    echo "  Pending Registrations: $(db_exec "SELECT COUNT(*) FROM registrations WHERE status='pending';")"
}

show_client_info() {
    local client_id="$1"
    
    log_step "Client Information: $client_id"
    
    echo "User Details:"
    db_get_user "$client_id"
    
    echo ""
    echo "Files:"
    db_list_files "$client_id"
    
    echo ""
    echo "Dataset Size:"
    zfs_dataset_size "$client_id"
    
    echo ""
    echo "Action Log:"
    db_action_log "$client_id" 10
}

# Cleanup functions
cleanup_client_data() {
    local client_id="$1"
    
    log_warning "Cleaning up client data: $client_id"
    
    # Delete from database
    db_exec "DELETE FROM files WHERE client_id='$client_id';"
    db_exec "DELETE FROM users WHERE client_id='$client_id';"
    db_exec "DELETE FROM registrations WHERE client_id='$client_id';"
    db_exec "DELETE FROM action_log WHERE client_id='$client_id';"
    
    # Destroy ZFS dataset
    zfs_destroy_dataset "$client_id"
    
    log_success "Client data cleaned: $client_id"
}

# Build functions
build_server() {
    log_info "Building server binary for Linux..."
    export GOOS=linux
    export GOARCH=amd64
    export CGO_ENABLED=1
    go build -o bin/zfscloud-server ./server/cmd
    
    if [ $? -eq 0 ]; then
        log_success "Server binary built"
        return 0
    else
        log_error "Server build failed"
        return 1
    fi
}

build_client() {
    log_info "Building client binary..."
    go build -o bin/zfscloud-client ./client/cmd
    
    if [ $? -eq 0 ]; then
        log_success "Client binary built"
        return 0
    else
        log_error "Client build failed"
        return 1
    fi
}

build_all() {
    log_step "Building All Binaries"
    build_server && build_client
}

check_binaries() {
    local missing=0
    
    if [ ! -f "bin/zfscloud-server" ]; then
        log_warning "Server binary not found"
        missing=1
    fi
    
    if [ ! -f "bin/zfscloud-client" ]; then
        log_warning "Client binary not found"
        missing=1
    fi
    
    if [ $missing -eq 1 ]; then
        log_info "Run: ./test-helpers.sh build_all"
        return 1
    else
        log_success "All binaries present"
        return 0
    fi
}

# Main function to allow direct execution
if [ "${BASH_SOURCE[0]}" -eq "${0}" ]; then
    # Script is being executed directly
    if [ $# -eq 0 ]; then
        echo "ZFSCloud Test Helper Utilities"
        echo ""
        echo "Usage: source $0  (to import functions)"
        echo "   or: $0 <function_name> [args...]  (to execute directly)"
        echo ""
        echo "Available functions:"
        echo "  Docker: docker_start_server, docker_stop_server, docker_clean_all"
        echo "  ZFS: zfs_pool_status, zfs_list_datasets, zfs_create_dataset"
        echo "  Database: db_list_users, db_list_files, db_interactive"
        echo "  System: show_system_info, show_client_info"
        echo "  Build: build_server, build_client, build_all"
        exit 0
    fi
    
    # Execute the requested function
    "$@"
fi
