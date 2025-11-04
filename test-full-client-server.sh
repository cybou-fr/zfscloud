#!/bin/bash
# Comprehensive Client-Server Test Script
# Tests complete workflow: registration, approval, dataset creation, file upload

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
SERVER_ADDRESS="127.0.0.1:50052"
CLIENT_BINARY="./bin/test-client"
CONFIG_FILE="test-client-config-50052.yaml"
KEYS_FILE="test-data/keys/client-keys.dat"
TEST_DIR="/home/user/Téléchargements"
ADMIN_TOKEN="test-admin-token-12345"

# Global variables
CLIENT_ID=""
REGISTRATION_ID=""
USER_ID=""

# Helper functions
print_banner() {
    echo -e "${BLUE}======================================${NC}"
    echo -e "${CYAN}$1${NC}"
    echo -e "${BLUE}======================================${NC}"
}

print_step() {
    echo ""
    echo -e "${MAGENTA}→ $1${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${CYAN}ℹ $1${NC}"
}

press_enter() {
    echo ""
    echo -e "${YELLOW}Press ENTER to continue...${NC}"
    read
}

# Cleanup function
cleanup_test_data() {
    print_step "Cleaning up previous test data..."
    rm -rf test-data/keys 2>/dev/null || true
    print_success "Cleanup complete"
}

# Check server is running
check_server() {
    print_step "Checking if server is running on $SERVER_ADDRESS..."
    
    if timeout 2 bash -c "echo > /dev/tcp/127.0.0.1/50052" 2>/dev/null; then
        print_success "Server is responding"
        return 0
    else
        print_error "Server is not responding on $SERVER_ADDRESS"
        print_info "Make sure the server is running:"
        echo "    ./bin/zfscloud-server-test -config server-test-50052.yaml"
        exit 1
    fi
}

# Check test directory
check_test_directory() {
    print_step "Checking test directory: $TEST_DIR"
    
    if [ ! -d "$TEST_DIR" ]; then
        print_error "Directory not found: $TEST_DIR"
        exit 1
    fi
    
    local file_count=$(find "$TEST_DIR" -type f 2>/dev/null | wc -l)
    local total_size=$(du -sh "$TEST_DIR" 2>/dev/null | cut -f1)
    
    print_success "Directory found"
    print_info "Files: $file_count"
    print_info "Total size: $total_size"
    
    echo ""
    echo -e "${YELLOW}Sample files:${NC}"
    find "$TEST_DIR" -type f 2>/dev/null | head -5 | while read file; do
        local size=$(du -h "$file" | cut -f1)
        echo "  - $(basename "$file") ($size)"
    done
}

# Step 1: Client Registration
test_registration() {
    print_banner "STEP 1: Client Registration"
    
    print_step "Starting client in interactive mode..."
    print_info "We will register the client with the server"
    
    # Create a temporary expect script for automated interaction
    cat > /tmp/test_register.exp << 'EOF'
#!/usr/bin/expect -f
set timeout 30

spawn ./bin/test-client -config test-client-config-50052.yaml -interactive

expect "zfscloud>" { send "register\r" }
expect "zfscloud>" { send "exit\r" }
expect eof
EOF
    
    chmod +x /tmp/test_register.exp
    
    print_info "Running: register"
    
    if command -v expect >/dev/null 2>&1; then
        /tmp/test_register.exp
    else
        print_info "Expect not installed, running manually..."
        echo ""
        echo -e "${YELLOW}Please type the following commands:${NC}"
        echo -e "${CYAN}  1. register${NC}"
        echo -e "${CYAN}  2. exit${NC}"
        echo ""
        $CLIENT_BINARY -config $CONFIG_FILE -interactive
    fi
    
    # Extract client ID from keys file
    if [ -f "$KEYS_FILE" ]; then
        # The client ID is in the first line of the keys file
        print_success "Client registered successfully"
        print_info "Keys saved to: $KEYS_FILE"
    else
        print_error "Registration failed - keys file not created"
        exit 1
    fi
    
    # Try to get client ID from config or keys
    if [ -f "test-data/client-id.txt" ]; then
        CLIENT_ID=$(cat test-data/client-id.txt)
        print_info "Client ID: $CLIENT_ID"
    fi
    
    press_enter
}

# Step 2: Admin Approval
test_approval() {
    print_banner "STEP 2: Admin Approval"
    
    print_step "Listing pending registrations..."
    
    # Use grpcurl to list pending registrations
    if ! command -v grpcurl >/dev/null 2>&1; then
        print_error "grpcurl not found. Installing..."
        go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
    fi
    
    local response=$(grpcurl -plaintext -d "{\"admin_token\":\"$ADMIN_TOKEN\"}" \
        $SERVER_ADDRESS zfscloud.v1.AdminService/ListPendingRegistrations 2>&1)
    
    echo "$response" | grep -q "registrations" && print_success "Found pending registrations" || print_error "No pending registrations found"
    
    echo ""
    echo -e "${CYAN}Pending Registrations:${NC}"
    echo "$response" | jq '.' 2>/dev/null || echo "$response"
    
    # Extract registration ID (assuming it's the first one)
    REGISTRATION_ID=$(echo "$response" | grep -oP '"registration_id":\s*"\K[^"]+' | head -1)
    CLIENT_ID=$(echo "$response" | grep -oP '"client_id":\s*"\K[^"]+' | head -1)
    
    if [ -z "$REGISTRATION_ID" ]; then
        print_error "Could not extract registration ID"
        echo ""
        print_info "Please provide the registration ID manually:"
        read -p "Registration ID: " REGISTRATION_ID
    else
        print_info "Registration ID: $REGISTRATION_ID"
        print_info "Client ID: $CLIENT_ID"
    fi
    
    echo ""
    print_step "Approving registration..."
    read -p "Quota in GB (default: 10): " quota_gb
    quota_gb=${quota_gb:-10}
    quota_bytes=$((quota_gb * 1024 * 1024 * 1024))
    
    local approval_response=$(grpcurl -plaintext -d "{\"admin_token\":\"$ADMIN_TOKEN\",\"registration_id\":\"$REGISTRATION_ID\",\"quota_bytes\":$quota_bytes}" \
        $SERVER_ADDRESS zfscloud.v1.AdminService/ApproveRegistration 2>&1)
    
    echo "$approval_response" | grep -q '"success":\s*true' && print_success "Registration approved!" || print_error "Approval failed"
    
    echo ""
    echo -e "${CYAN}Approval Response:${NC}"
    echo "$approval_response" | jq '.' 2>/dev/null || echo "$approval_response"
    
    # Extract user ID
    USER_ID=$(echo "$approval_response" | grep -oP '"user_id":\s*"\K[^"]+' | head -1)
    if [ -n "$USER_ID" ]; then
        print_info "User ID: $USER_ID"
    fi
    
    press_enter
}

# Step 3: Verify ZFS Dataset
verify_dataset() {
    print_banner "STEP 3: Verify ZFS Dataset Creation"
    
    print_step "Checking ZFS datasets..."
    
    echo ""
    echo -e "${CYAN}Current ZFS datasets in tank pool:${NC}"
    zfs list -r tank 2>&1 | head -20
    
    echo ""
    if [ -n "$USER_ID" ]; then
        local dataset="tank/users/$USER_ID"
        if zfs list "$dataset" >/dev/null 2>&1; then
            print_success "Dataset created: $dataset"
            
            echo ""
            echo -e "${CYAN}Dataset Details:${NC}"
            zfs list -o name,used,avail,refer,mountpoint,encryption,keystatus "$dataset"
            
            echo ""
            echo -e "${CYAN}Dataset Properties:${NC}"
            zfs get encryption,keyformat,keylocation,compression "$dataset"
        else
            print_error "Dataset not found: $dataset"
        fi
    else
        print_info "User ID not available, showing all user datasets:"
        zfs list -r tank/users 2>/dev/null || print_info "No user datasets yet"
    fi
    
    press_enter
}

# Step 4: Load Dataset Key
test_key_loading() {
    print_banner "STEP 4: Load Dataset Key"
    
    print_step "Client needs to load the dataset key to start uploading..."
    
    print_info "The client will need to provide the passphrase"
    print_info "This passphrase was displayed during registration"
    
    echo ""
    echo -e "${YELLOW}In the client, run: ${NC}"
    echo -e "${CYAN}  load-key${NC}"
    echo ""
    
    press_enter
}

# Step 5: Upload Files
test_upload() {
    print_banner "STEP 5: Upload Files"
    
    print_step "Now we'll test uploading files from $TEST_DIR"
    
    # Get first few files
    local test_files=($(find "$TEST_DIR" -type f 2>/dev/null | head -3))
    
    if [ ${#test_files[@]} -eq 0 ]; then
        print_error "No files found in $TEST_DIR"
        return 1
    fi
    
    print_info "We'll upload ${#test_files[@]} test files"
    
    echo ""
    echo -e "${YELLOW}Files to upload:${NC}"
    for file in "${test_files[@]}"; do
        local size=$(du -h "$file" | cut -f1)
        echo "  - $(basename "$file") ($size)"
    done
    
    echo ""
    print_info "You can now use the client to upload files:"
    echo ""
    echo -e "${CYAN}Single file upload:${NC}"
    echo "  upload <local-path> <remote-path>"
    echo ""
    echo -e "${CYAN}Directory upload:${NC}"
    echo "  upload-dir /home/user/Téléchargements telechargements"
    echo ""
    
    press_enter
}

# Step 6: Verify Upload
verify_upload() {
    print_banner "STEP 6: Verify Uploaded Files"
    
    print_step "Checking uploaded files in database..."
    
    local db_path="/tmp/zfscloud-test/zfscloud.db"
    
    if [ ! -f "$db_path" ]; then
        print_error "Database not found: $db_path"
        return 1
    fi
    
    echo ""
    echo -e "${CYAN}Files in database:${NC}"
    sqlite3 "$db_path" "SELECT path, size, hash FROM files WHERE client_id='$CLIENT_ID' LIMIT 10;" 2>/dev/null || print_info "No files found yet"
    
    echo ""
    print_step "Checking files on disk..."
    
    if [ -n "$USER_ID" ]; then
        local dataset_path="/tank/users/$USER_ID"
        if [ -d "$dataset_path" ]; then
            echo -e "${CYAN}Files in dataset:${NC}"
            find "$dataset_path" -type f 2>/dev/null | head -10 | while read file; do
                local size=$(du -h "$file" | cut -f1)
                echo "  - $file ($size)"
            done
        else
            print_info "Dataset not mounted or path not found: $dataset_path"
        fi
    fi
    
    press_enter
}

# Interactive Client Session
interactive_session() {
    print_banner "INTERACTIVE CLIENT SESSION"
    
    print_info "Starting interactive client session..."
    print_info "You can now use all client commands"
    
    echo ""
    echo -e "${YELLOW}Available commands:${NC}"
    echo -e "${CYAN}  register${NC}              - Register with server (already done)"
    echo -e "${CYAN}  heartbeat${NC}             - Send heartbeat"
    echo -e "${CYAN}  status${NC}                - Show client status"
    echo -e "${CYAN}  upload <file> [remote]${NC} - Upload a file"
    echo -e "${CYAN}  upload-dir <dir> [remote]${NC} - Upload directory"
    echo -e "${CYAN}  list [path]${NC}           - List backed up files"
    echo -e "${CYAN}  snapshots${NC}             - List snapshots"
    echo -e "${CYAN}  help${NC}                  - Show help"
    echo -e "${CYAN}  exit${NC}                  - Exit client"
    
    echo ""
    echo -e "${GREEN}Example: Upload entire Téléchargements directory:${NC}"
    echo -e "${CYAN}  upload-dir /home/user/Téléchargements telechargements${NC}"
    
    echo ""
    press_enter
    
    $CLIENT_BINARY -config $CONFIG_FILE -interactive
}

# Main test flow
main() {
    clear
    print_banner "ZFSCloud Client-Server Integration Test"
    
    echo -e "${CYAN}This script will test the complete workflow:${NC}"
    echo "  1. Client registration"
    echo "  2. Admin approval"
    echo "  3. ZFS dataset creation"
    echo "  4. Dataset key loading"
    echo "  5. File upload from $TEST_DIR"
    echo "  6. Upload verification"
    
    echo ""
    read -p "Continue? (y/n): " confirm
    if [ "$confirm" != "y" ]; then
        echo "Aborted."
        exit 0
    fi
    
    # Run tests
    cleanup_test_data
    check_server
    check_test_directory
    test_registration
    test_approval
    verify_dataset
    test_key_loading
    test_upload
    
    # Interactive session
    interactive_session
    
    # Final verification
    verify_upload
    
    print_banner "TEST COMPLETE"
    print_success "All steps completed!"
}

# Run main
main
