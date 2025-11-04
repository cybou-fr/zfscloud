#!/bin/bash
# Admin Helper Script for grpcurl commands

SERVER="127.0.0.1:50052"
ADMIN_TOKEN="test-admin-token-12345"

case "$1" in
    list-pending)
        echo "Listing pending registrations..."
        grpcurl -plaintext -d "{\"admin_token\":\"$ADMIN_TOKEN\"}" \
            $SERVER zfscloud.v1.AdminService/ListPendingRegistrations | jq '.'
        ;;
    
    approve)
        if [ -z "$2" ]; then
            echo "Usage: $0 approve <registration_id> [quota_gb]"
            exit 1
        fi
        REG_ID="$2"
        QUOTA_GB="${3:-10}"
        QUOTA_BYTES=$((QUOTA_GB * 1024 * 1024 * 1024))
        
        echo "Approving registration: $REG_ID with ${QUOTA_GB}GB quota..."
        grpcurl -plaintext -d "{\"admin_token\":\"$ADMIN_TOKEN\",\"registration_id\":\"$REG_ID\",\"quota_bytes\":$QUOTA_BYTES}" \
            $SERVER zfscloud.v1.AdminService/ApproveRegistration | jq '.'
        ;;
    
    list-users)
        echo "Listing all users..."
        grpcurl -plaintext -d "{\"admin_token\":\"$ADMIN_TOKEN\",\"include_revoked\":false}" \
            $SERVER zfscloud.v1.AdminService/ListUsers | jq '.'
        ;;
    
    status)
        echo "Getting system status..."
        grpcurl -plaintext -d "{\"admin_token\":\"$ADMIN_TOKEN\"}" \
            $SERVER zfscloud.v1.AdminService/GetSystemStatus | jq '.'
        ;;
    
    snapshots)
        echo "Listing all snapshots..."
        grpcurl -plaintext -d "{\"admin_token\":\"$ADMIN_TOKEN\"}" \
            $SERVER zfscloud.v1.AdminService/ListAllSnapshots | jq '.'
        ;;
    
    *)
        echo "ZFSCloud Admin Helper"
        echo ""
        echo "Usage: $0 <command> [args]"
        echo ""
        echo "Commands:"
        echo "  list-pending              - List pending registrations"
        echo "  approve <reg_id> [quota]  - Approve registration (quota in GB, default: 10)"
        echo "  list-users                - List all users"
        echo "  status                    - Get system status"
        echo "  snapshots                 - List all snapshots"
        echo ""
        echo "Examples:"
        echo "  $0 list-pending"
        echo "  $0 approve abc123 20"
        echo "  $0 list-users"
        exit 1
        ;;
esac
