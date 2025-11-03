#!/bin/bash
# ZFSCloud Server Setup Script for Ubuntu VM
# Run this script on the Ubuntu VM to set up the server environment

set -e

echo "============================================"
echo "ZFSCloud Server Setup for Ubuntu VM"
echo "============================================"
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}Please run as root (use sudo)${NC}"
    exit 1
fi

# Get the regular user (who invoked sudo)
REGULAR_USER=${SUDO_USER:-$USER}
USER_HOME=$(eval echo ~$REGULAR_USER)

echo -e "${YELLOW}[1/7] Updating system packages...${NC}"
apt update && apt upgrade -y
echo -e "${GREEN}✓ System updated${NC}"

echo -e "\n${YELLOW}[2/7] Installing ZFS utilities...${NC}"
apt install -y zfsutils-linux
echo -e "${GREEN}✓ ZFS installed${NC}"

echo -e "\n${YELLOW}[3/7] Installing additional tools...${NC}"
apt install -y wget curl git sqlite3 net-tools
echo -e "${GREEN}✓ Tools installed${NC}"

echo -e "\n${YELLOW}[4/7] Creating ZFS test pool...${NC}"
# Check if pool already exists
if zpool list tank &> /dev/null; then
    echo -e "${YELLOW}  Pool 'tank' already exists, skipping...${NC}"
else
    # Create file-backed pool for testing
    truncate -s 2G /tmp/zfs-test.img
    zpool create tank /tmp/zfs-test.img
    echo -e "${GREEN}✓ ZFS pool 'tank' created${NC}"
fi

zpool status tank
zfs list

echo -e "\n${YELLOW}[5/7] Creating data directory...${NC}"
mkdir -p "$USER_HOME/zfscloud-data"
chown -R $REGULAR_USER:$REGULAR_USER "$USER_HOME/zfscloud-data"
echo -e "${GREEN}✓ Data directory created: $USER_HOME/zfscloud-data${NC}"

echo -e "\n${YELLOW}[6/7] Creating server configuration...${NC}"
cat > "$USER_HOME/server-config.yaml" << EOF
server:
  address: "0.0.0.0:9090"
  data_dir: "$USER_HOME/zfscloud-data"
  max_clients: 100
  
database:
  path: "$USER_HOME/zfscloud-data/zfscloud.db"

security:
  tls:
    enabled: false
  rate_limit:
    enabled: true
    requests_per_minute: 100

zfs:
  pool_name: "tank"
  base_path: "backups"
  
logging:
  level: "info"
  file: "$USER_HOME/zfscloud-data/server.log"
EOF

chown $REGULAR_USER:$REGULAR_USER "$USER_HOME/server-config.yaml"
echo -e "${GREEN}✓ Configuration created: $USER_HOME/server-config.yaml${NC}"

echo -e "\n${YELLOW}[7/7] Configuring firewall...${NC}"
# Install ufw if not present
if ! command -v ufw &> /dev/null; then
    apt install -y ufw
fi

ufw allow 22/tcp    # SSH
ufw allow 9090/tcp  # ZFSCloud server
ufw --force enable
echo -e "${GREEN}✓ Firewall configured${NC}"

# Display IP address
echo -e "\n============================================"
echo -e "${GREEN}Setup Complete!${NC}"
echo -e "============================================"
echo ""
echo -e "VM IP Address:"
ip addr show | grep "inet " | grep -v 127.0.0.1 | awk '{print "  " $2}' | cut -d/ -f1
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo -e "1. Transfer server binary from Windows:"
echo -e "   ${GREEN}scp bin/zfscloud-server $REGULAR_USER@<vm-ip>:~/${NC}"
echo ""
echo -e "2. Make it executable:"
echo -e "   ${GREEN}chmod +x ~/zfscloud-server${NC}"
echo ""
echo -e "3. Start the server:"
echo -e "   ${GREEN}sudo ./zfscloud-server --config ~/server-config.yaml${NC}"
echo ""
echo -e "4. Or run in background:"
echo -e "   ${GREEN}sudo nohup ./zfscloud-server --config ~/server-config.yaml > ~/server.log 2>&1 &${NC}"
echo ""
echo -e "5. Check status:"
echo -e "   ${GREEN}ps aux | grep zfscloud-server${NC}"
echo -e "   ${GREEN}sudo netstat -tlnp | grep 9090${NC}"
echo ""
