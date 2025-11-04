#!/bin/bash
# Quick testing script for client-server interaction on Linux

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== ZFSCloud Client-Server Test ===${NC}"
echo ""

# Check if server binary exists
if [ ! -f "bin/zfscloud-server" ]; then
    echo -e "${YELLOW}Server binary not found. Building for Linux...${NC}"
    ./build-docker.sh
    if [ $? -ne 0 ]; then
        echo -e "${RED}Build failed!${NC}"
        exit 1
    fi
fi

echo -e "${GREEN}✓ Server binary ready${NC}"
echo ""

# Start Docker containers
echo -e "${YELLOW}Starting Docker containers...${NC}"
cd docker
docker-compose up -d zfscloud-server
cd ..

if [ $? -ne 0 ]; then
    echo -e "${RED}✗ Failed to start Docker containers!${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Server started${NC}"
echo ""

# Wait for server to be ready
echo -e "${YELLOW}Waiting for server to initialize...${NC}"
sleep 5

# Show server logs
echo ""
echo -e "${CYAN}=== Server Logs ===${NC}"
cd docker
docker-compose logs --tail=20 zfscloud-server
cd ..

echo ""
echo -e "${GREEN}=== Test Instructions ===${NC}"
echo ""
echo -e "${CYAN}1. Run the client:${NC}"
echo -e "   ./bin/zfscloud-client"
echo ""
echo -e "${CYAN}2. Register:${NC}"
echo -e "   > register"
echo ""
echo -e "${CYAN}3. In another terminal, approve the registration:${NC}"
echo -e "   docker exec zfscloud-server sqlite3 /var/lib/zfscloud/zfscloud.db \"SELECT registration_id FROM registrations WHERE status='pending';\""
echo -e "   ${GREEN}# Copy the registration ID, then:${NC}"
echo -e "   docker exec zfscloud-server sqlite3 /var/lib/zfscloud/zfscloud.db \"UPDATE registrations SET status='approved', processed_at=datetime('now') WHERE registration_id='<ID>';\""
echo ""
echo -e "${CYAN}4. Back in client, send heartbeat:${NC}"
echo -e "   > heartbeat"
echo ""
echo -e "${CYAN}5. Upload a file:${NC}"
echo -e "   echo 'Test data' > test.txt"
echo -e "   > upload test.txt /test.txt"
echo ""
echo -e "${CYAN}6. List files:${NC}"
echo -e "   > list"
echo ""
echo -e "${GREEN}=== Useful Commands ===${NC}"
echo ""
echo -e "${CYAN}View server logs:${NC}"
echo -e "   cd docker && docker-compose logs -f zfscloud-server"
echo ""
echo -e "${CYAN}Check ZFS status:${NC}"
echo -e "   docker exec zfscloud-server zpool status tank"
echo -e "   docker exec zfscloud-server zfs list -r tank"
echo ""
echo -e "${CYAN}Check database:${NC}"
echo -e "   docker exec -it zfscloud-server sqlite3 /var/lib/zfscloud/zfscloud.db"
echo ""
echo -e "${CYAN}Stop server:${NC}"
echo -e "   cd docker && docker-compose down"
echo ""
echo -e "${CYAN}Clean everything (removes all data):${NC}"
echo -e "   cd docker && docker-compose down -v"
echo ""
