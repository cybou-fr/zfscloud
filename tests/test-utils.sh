#!/bin/bash
# Utility functions for ZFSCloud testing

# Colors
export RED='\033[0;31m'
export GREEN='\033[0;32m'
export YELLOW='\033[1;33m'
export CYAN='\033[0;36m'
export BLUE='\033[0;34m'
export MAGENTA='\033[0;35m'
export NC='\033[0m'

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

log_debug() {
    if [ "$DEBUG" = "1" ]; then
        echo -e "${MAGENTA}[DEBUG]${NC} $1"
    fi
}

log_section() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

# Docker helper functions
docker_is_running() {
    docker info > /dev/null 2>&1
    return $?
}

container_is_running() {
    local container=$1
    docker ps --format '{{.Names}}' | grep -q "^${container}$"
    return $?
}

container_is_healthy() {
    local container=$1
    local health=$(docker inspect --format='{{.State.Health.Status}}' "$container" 2>/dev/null)
    [ "$health" = "healthy" ] || [ -z "$health" ]
    return $?
}

wait_for_container() {
    local container=$1
    local timeout=${2:-30}
    local elapsed=0
    
    log_info "Waiting for container $container to be ready..."
    while [ $elapsed -lt $timeout ]; do
        if container_is_running "$container"; then
            log_success "Container $container is running"
            return 0
        fi
        sleep 1
        ((elapsed++))
    done
    
    log_error "Container $container did not start within ${timeout}s"
    return 1
}

# Server helper functions
server_is_listening() {
    local port=${1:-50051}
    nc -z localhost "$port" 2>/dev/null || timeout 1 bash -c "cat < /dev/null > /dev/tcp/localhost/$port" 2>/dev/null
    return $?
}

get_server_logs() {
    local lines=${1:-50}
    docker logs --tail "$lines" zfscloud-server 2>&1
}

check_zfs_pool() {
    local pool=${1:-tank}
    docker exec zfscloud-server zpool status "$pool" > /dev/null 2>&1
    return $?
}

get_zfs_pool_status() {
    local pool=${1:-tank}
    docker exec zfscloud-server zpool status "$pool" 2>/dev/null
}

get_zfs_datasets() {
    local pool=${1:-tank}
    docker exec zfscloud-server zfs list -r "$pool" 2>/dev/null
}

# Database helper functions
db_exists() {
    docker exec zfscloud-server test -f /var/lib/zfscloud/zfscloud.db
    return $?
}

db_query() {
    local query=$1
    docker exec zfscloud-server sqlite3 /var/lib/zfscloud/zfscloud.db "$query" 2>/dev/null
}

db_count_users() {
    db_query "SELECT COUNT(*) FROM users;"
}

db_count_registrations() {
    db_query "SELECT COUNT(*) FROM registrations;"
}

db_get_pending_registrations() {
    db_query "SELECT registration_id, client_id FROM registrations WHERE status='pending';"
}

db_approve_registration() {
    local reg_id=$1
    local quota=${2:-10737418240}  # 10GB default
    db_query "UPDATE registrations SET status='approved', processed_at=datetime('now') WHERE registration_id='$reg_id';"
}

# File helper functions
generate_test_file() {
    local file=$1
    local size=$2  # in bytes
    dd if=/dev/urandom of="$file" bs="$size" count=1 2>/dev/null
}

get_file_size() {
    local file=$1
    stat -f%z "$file" 2>/dev/null || stat -c%s "$file" 2>/dev/null
}

get_file_hash() {
    local file=$1
    sha256sum "$file" 2>/dev/null | awk '{print $1}' || shasum -a 256 "$file" 2>/dev/null | awk '{print $1}'
}

# Performance monitoring
get_container_stats() {
    local container=${1:-zfscloud-server}
    docker stats "$container" --no-stream --format "CPU: {{.CPUPerc}} | Memory: {{.MemUsage}} | Network I/O: {{.NetIO}}" 2>/dev/null
}

measure_time() {
    local start=$1
    local end=$2
    echo "scale=4; $end - $start" | bc
}

get_timestamp() {
    date +%s.%N
}

# Network testing
test_port() {
    local host=${1:-localhost}
    local port=${2:-50051}
    timeout 2 bash -c "cat < /dev/null > /dev/tcp/$host/$port" 2>/dev/null
    return $?
}

# Cleanup functions
cleanup_test_files() {
    local pattern=${1:-"test-*.txt"}
    rm -f $pattern 2>/dev/null || true
}

cleanup_docker() {
    log_info "Stopping Docker containers..."
    cd docker && docker-compose down -v 2>/dev/null || true
}

# Validation functions
validate_prerequisites() {
    local missing=0
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed"
        missing=1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose is not installed"
        missing=1
    fi
    
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed"
        missing=1
    fi
    
    if ! docker_is_running; then
        log_error "Docker daemon is not running"
        missing=1
    fi
    
    return $missing
}

# Report generation
generate_test_report() {
    local report_file=$1
    local test_name=$2
    local passed=$3
    local failed=$4
    
    cat > "$report_file" <<EOF
ZFSCloud Test Report
====================
Test: $test_name
Date: $(date)
Duration: $(date -u -d @${SECONDS} +"%T" 2>/dev/null || echo "${SECONDS}s")

Results:
  Passed: $passed
  Failed: $failed
  Total:  $((passed + failed))
  Success Rate: $(echo "scale=2; $passed * 100 / ($passed + $failed)" | bc)%

Status: $([ $failed -eq 0 ] && echo "PASS" || echo "FAIL")
EOF
}

# Export all functions
export -f log_info log_success log_error log_warning log_debug log_section
export -f docker_is_running container_is_running container_is_healthy wait_for_container
export -f server_is_listening get_server_logs check_zfs_pool get_zfs_pool_status get_zfs_datasets
export -f db_exists db_query db_count_users db_count_registrations db_get_pending_registrations db_approve_registration
export -f generate_test_file get_file_size get_file_hash
export -f get_container_stats measure_time get_timestamp
export -f test_port cleanup_test_files cleanup_docker
export -f validate_prerequisites generate_test_report
