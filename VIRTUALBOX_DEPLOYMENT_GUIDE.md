# VirtualBox Server Deployment Guide

This guide walks through deploying the ZFSCloud server in VirtualBox and testing client-server interaction from your Windows host.

## Prerequisites

- VirtualBox installed on Windows
- Ubuntu Server ISO (recommended: Ubuntu 22.04 LTS)
- At least 2GB RAM allocated to VM
- Network connectivity between host and VM

## Part 1: VirtualBox VM Setup

### 1.1 Create Ubuntu Server VM

1. **Create New VM:**
   - Name: `zfscloud-server`
   - Type: Linux
   - Version: Ubuntu (64-bit)
   - Memory: 2048 MB (minimum)
   - Hard Disk: 20 GB VDI (dynamically allocated)

2. **Network Configuration:**
   - Adapter 1: NAT (for internet access)
   - Adapter 2: Host-only Adapter (for host-VM communication)
   
   **Or use Bridged Adapter for simpler setup:**
   - Adapter 1: Bridged Adapter (VM gets IP on your local network)

3. **Install Ubuntu Server:**
   - Boot from Ubuntu Server ISO
   - Follow installation wizard
   - Create user account (e.g., `zfscloud`)
   - Install OpenSSH server when prompted

### 1.2 Configure Network (Host-Only Adapter)

If using Host-Only Adapter:

```bash
# On Ubuntu VM - Check network interfaces
ip addr show

# Configure static IP (example for enp0s8)
sudo nano /etc/netplan/01-netcfg.yaml
```

Add configuration:
```yaml
network:
  version: 2
  ethernets:
    enp0s8:
      dhcp4: no
      addresses:
        - 192.168.56.10/24
```

Apply configuration:
```bash
sudo netplan apply
```

### 1.3 Install Dependencies

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install ZFS utilities
sudo apt install -y zfsutils-linux

# Install development tools (if building on VM)
sudo apt install -y wget tar

# Install Go (if building on VM)
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/bin/go/bin' >> ~/.bashrc
source ~/.bashrc
```

## Part 2: Deploy ZFSCloud Server

### Option A: Build on VM (Recommended)

```bash
# Clone repository or transfer files
git clone <your-repo-url> ~/zfscloud
cd ~/zfscloud

# Build server
go build -o bin/zfscloud-server ./server/cmd

# Create necessary directories
mkdir -p ~/zfscloud-data
```

### Option B: Cross-Compile on Windows and Transfer

On Windows PowerShell:
```powershell
# Set environment for Linux build
$env:GOOS="linux"
$env:CGO_ENABLED="0"

# Build server
go build -o bin/zfscloud-server ./server/cmd

# Reset environment
Remove-Item Env:\GOOS
Remove-Item Env:\CGO_ENABLED
```

Transfer to VM using SCP:
```powershell
scp bin/zfscloud-server zfscloud@192.168.56.10:~/
```

### 2.1 Configure Server

```bash
cd ~/zfscloud
cp server/config.example.yaml server/config.yaml
nano server/config.yaml
```

Minimal configuration:
```yaml
server:
  address: "0.0.0.0:9090"  # Listen on all interfaces
  data_dir: "/home/zfscloud/zfscloud-data"
  
database:
  path: "/home/zfscloud/zfscloud-data/zfscloud.db"

security:
  tls:
    enabled: false  # Start with TLS disabled for testing

zfs:
  pool_name: "tank"  # Will create test pool
  base_path: "backups"
```

### 2.2 Setup ZFS Pool (Test Environment)

For testing without real ZFS:
```bash
# Create a file-backed ZFS pool for testing
sudo truncate -s 1G /tmp/zfs-test.img
sudo zpool create tank /tmp/zfs-test.img

# Verify pool
sudo zfs list
```

For production with real disk:
```bash
# List available disks
sudo fdisk -l

# Create pool on real disk (example)
sudo zpool create tank /dev/sdb
```

### 2.3 Start Server

```bash
# Make executable
chmod +x bin/zfscloud-server

# Run server
sudo ./bin/zfscloud-server

# Or run in background
sudo nohup ./bin/zfscloud-server > server.log 2>&1 &
```

Check if running:
```bash
# Check process
ps aux | grep zfscloud-server

# Check port
sudo netstat -tlnp | grep 9090

# View logs
tail -f server.log
```

## Part 3: Configure Windows Client

### 3.1 Get VM IP Address

On Ubuntu VM:
```bash
# Find IP address
ip addr show

# Test connectivity
ping 192.168.56.1  # Host IP
```

On Windows:
```powershell
# Find host-only network adapter IP
ipconfig

# Test connectivity to VM
ping 192.168.56.10
```

### 3.2 Configure Client

Edit `client/config.example.yaml` or create new config:
```yaml
server:
  address: "192.168.56.10:9090"  # VM IP address
  
client:
  id: "windows-client-01"
  hostname: "WIN-DESKTOP"
  
paths:
  backup_dirs:
    - "C:\\Users\\yosta\\Documents\\TestBackup"
  exclude_patterns:
    - "*.tmp"
    - "*.log"

storage:
  chunk_size: 4194304  # 4MB

security:
  tls:
    enabled: false  # Match server config
```

### 3.3 Test Connection

```powershell
# On Windows - Run client
.\bin\zfscloud-cli.exe --config client-config.yaml status

# Or use GUI
.\bin\zfscloud-gui.exe
```

## Part 4: Testing Scenarios

### Test 1: Basic Connection
```powershell
# Windows client
.\bin\zfscloud-cli.exe --config client-config.yaml status
```

Expected output:
```
Client ID: windows-client-01
Server: 192.168.56.10:9090
Status: Connected
```

### Test 2: File Backup
```powershell
# Create test file
New-Item -Path "C:\Users\yosta\Documents\TestBackup\test.txt" -ItemType File -Force
Set-Content -Path "C:\Users\yosta\Documents\TestBackup\test.txt" -Value "Test backup data"

# Start backup
.\bin\zfscloud-cli.exe --config client-config.yaml backup
```

### Test 3: Verify on Server
```bash
# On Ubuntu VM
sudo zfs list -t all

# Check uploaded files
ls -la ~/zfscloud-data/clients/

# Query database
./bin/zfscloud-server --config server/config.yaml admin list-clients
```

### Test 4: File Restore
```powershell
# List available backups
.\bin\zfscloud-cli.exe --config client-config.yaml list

# Restore file
.\bin\zfscloud-cli.exe --config client-config.yaml restore --snapshot <snapshot-name>
```

### Test 5: GUI Testing
```powershell
# Run GUI
.\bin\zfscloud-gui.exe

# Test operations:
# 1. Configure server address (192.168.56.10:9090)
# 2. Connect to server
# 3. Check connection status (should show "Connected")
# 4. Browse files
# 5. Perform backup
# 6. Monitor progress
```

## Part 5: Monitoring and Debugging

### On Server (Ubuntu VM)

**View logs:**
```bash
tail -f server.log

# Or if running in foreground, check console output
```

**Check ZFS status:**
```bash
sudo zpool status
sudo zfs list -t all
```

**Monitor network:**
```bash
# Watch connections
sudo netstat -an | grep 9090

# Monitor traffic
sudo tcpdump -i enp0s8 port 9090
```

**Database inspection:**
```bash
# List clients
sqlite3 ~/zfscloud-data/zfscloud.db "SELECT * FROM clients;"

# List snapshots
sqlite3 ~/zfscloud-data/zfscloud.db "SELECT * FROM snapshots;"
```

### On Client (Windows)

**Test connectivity:**
```powershell
# Ping VM
ping 192.168.56.10

# Test port
Test-NetConnection -ComputerName 192.168.56.10 -Port 9090

# Telnet (if enabled)
telnet 192.168.56.10 9090
```

**View client logs:**
```powershell
# GUI logs (if implemented)
Get-Content -Path "$env:APPDATA\zfscloud\logs\gui.log" -Tail 50

# CLI logs
# Check console output
```

## Part 6: Firewall Configuration

### Ubuntu VM Firewall
```bash
# Check firewall status
sudo ufw status

# Allow server port
sudo ufw allow 9090/tcp

# Enable firewall
sudo ufw enable
```

### Windows Firewall
```powershell
# Allow outbound connections to VM (usually allowed by default)
# If needed, create rule:
New-NetFirewallRule -DisplayName "ZFSCloud Client" -Direction Outbound -Action Allow -Protocol TCP -RemotePort 9090
```

## Part 7: Common Issues

### Issue 1: Connection Refused
**Symptoms:** Client can't connect to server

**Solutions:**
```bash
# On VM - Check if server is running
ps aux | grep zfscloud-server

# Check if port is listening
sudo netstat -tlnp | grep 9090

# Verify firewall
sudo ufw status

# Check server config
cat server/config.yaml | grep address
```

### Issue 2: Connection Timeout
**Symptoms:** Client times out when connecting

**Solutions:**
```powershell
# On Windows - Test network connectivity
ping 192.168.56.10
Test-NetConnection -ComputerName 192.168.56.10 -Port 9090
```

```bash
# On VM - Check network interface
ip addr show
```

### Issue 3: Permission Denied (ZFS)
**Symptoms:** Server can't create ZFS datasets

**Solutions:**
```bash
# Run server with sudo
sudo ./bin/zfscloud-server

# Or configure ZFS permissions
sudo zfs allow zfscloud create,mount,snapshot tank
```

### Issue 4: TLS Certificate Errors
**Symptoms:** Certificate verification failed

**Solutions:**
- Ensure TLS is disabled in both client and server configs for testing
- Or generate proper certificates if using TLS

## Part 8: Performance Testing

### Network Performance
```bash
# On VM - Install iperf3
sudo apt install -y iperf3
iperf3 -s
```

```powershell
# On Windows - Download iperf3
iperf3 -c 192.168.56.10
```

### Backup Speed Test
```powershell
# Create large test file
fsutil file createnew C:\Users\yosta\Documents\TestBackup\largefile.bin 104857600  # 100MB

# Time backup
Measure-Command { .\bin\zfscloud-cli.exe --config client-config.yaml backup }
```

## Part 9: Snapshot Management

### Create Manual Snapshot
```bash
# On VM
sudo zfs snapshot tank/backups/windows-client-01@manual-test
sudo zfs list -t snapshot
```

### Verify Snapshot on Client
```powershell
.\bin\zfscloud-cli.exe --config client-config.yaml list-snapshots
```

## Part 10: Cleanup

### Stop Server
```bash
# Find process
ps aux | grep zfscloud-server

# Kill process
sudo kill <PID>

# Or if using systemd service (future)
sudo systemctl stop zfscloud-server
```

### Remove Test Data
```bash
# Destroy ZFS pool (careful!)
sudo zpool destroy tank

# Remove data directory
rm -rf ~/zfscloud-data
```

## Next Steps

1. **Enable TLS:** Configure proper certificates for secure communication
2. **Systemd Service:** Create service file for automatic startup
3. **Monitoring:** Setup Prometheus/Grafana for metrics
4. **Backup Automation:** Configure scheduled backups
5. **Multi-Client Testing:** Connect multiple clients to same server

## Useful Commands Reference

**Server Management:**
```bash
# Start server
sudo ./bin/zfscloud-server

# Check status
ps aux | grep zfscloud-server

# View logs
tail -f server.log

# Stop server
sudo pkill zfscloud-server
```

**ZFS Commands:**
```bash
# List pools
sudo zpool list

# List datasets
sudo zfs list

# List snapshots
sudo zfs list -t snapshot

# Get pool info
sudo zpool status tank
```

**Network Testing:**
```bash
# Test from VM to Windows
ping 192.168.56.1

# Test from Windows to VM
ping 192.168.56.10

# Check open ports
sudo netstat -tlnp
```
