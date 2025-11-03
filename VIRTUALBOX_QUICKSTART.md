# VirtualBox Testing - Quick Start

This guide provides the fastest path to test ZFSCloud with a server in VirtualBox and client on Windows.

## Overview

```
┌─────────────────────┐         Network          ┌─────────────────────┐
│   Windows Host      │    (192.168.56.x)        │   Ubuntu VM         │
│                     │ ◄────────────────────► │                     │
│  - ZFSCloud Client  │                          │  - ZFSCloud Server  │
│  - GUI/CLI          │                          │  - ZFS Pool         │
│  - Backup Source    │                          │  - Database         │
└─────────────────────┘                          └─────────────────────┘
```

## Prerequisites

- ✅ VirtualBox installed
- ✅ Ubuntu Server 22.04 ISO
- ✅ 2GB+ RAM for VM
- ✅ Go installed on Windows (already have)

## Quick Setup (30 minutes)

### Step 1: Create Ubuntu VM (10 minutes)

1. **VirtualBox - New VM:**
   - Name: `zfscloud-server`
   - Type: Linux / Ubuntu (64-bit)
   - RAM: 2048 MB
   - Disk: 20 GB

2. **Network:**
   - Adapter 1: NAT (internet)
   - Adapter 2: Host-only Adapter (VM ↔ Windows)
   
   **Quick tip:** Find VirtualBox → File → Host Network Manager → Verify `vboxnet0` exists

3. **Install Ubuntu:**
   - Boot from ISO
   - Username: `zfscloud`
   - Enable OpenSSH when prompted

### Step 2: Configure VM Network (5 minutes)

**Get VM IP:**
```bash
# On VM - Login and check IP
ip addr show

# You should see something like:
# enp0s8: inet 192.168.56.101/24  ← This is your VM IP
```

**If no IP on second adapter, configure it:**
```bash
sudo nano /etc/netplan/01-netcfg.yaml
```

Add:
```yaml
network:
  version: 2
  ethernets:
    enp0s8:
      dhcp4: yes
```

Apply:
```bash
sudo netplan apply
ip addr show  # Verify IP appears
```

### Step 3: Setup Server on VM (5 minutes)

**On VM - Run automated setup:**
```bash
# Transfer setup script from Windows:
# (On Windows PowerShell, from project directory)
scp vm-setup.sh zfscloud@<vm-ip>:~/

# On VM - Run setup
chmod +x vm-setup.sh
sudo ./vm-setup.sh
```

This installs ZFS, creates test pool, and configures firewall.

### Step 4: Build and Deploy Server (5 minutes)

**On Windows:**
```powershell
# Build Linux server binary
.\build-server-linux.ps1

# Transfer to VM (replace with your VM IP)
scp bin/zfscloud-server zfscloud@192.168.56.101:~/
```

**On VM:**
```bash
# Make executable
chmod +x ~/zfscloud-server

# Start server
sudo ./zfscloud-server --config ~/server-config.yaml
```

**Verify server is running:**
```bash
# In another terminal
ps aux | grep zfscloud-server
sudo netstat -tlnp | grep 9090
```

### Step 5: Test Client Connection (5 minutes)

**On Windows:**
```powershell
# Run connection test
.\test-vm-connection.ps1 -ServerIP 192.168.56.101

# If successful, test CLI
.\bin\zfscloud-cli.exe --config test-vm-config.yaml status

# Or launch GUI
.\bin\zfscloud-gui.exe
# Configure: Server → 192.168.56.101:9090 → Connect
```

## Testing Checklist

### ✅ Basic Tests

- [ ] **Network:** `ping 192.168.56.101` from Windows
- [ ] **Port:** `Test-NetConnection -ComputerName 192.168.56.101 -Port 9090`
- [ ] **Server Running:** `ps aux | grep zfscloud-server` on VM
- [ ] **Connection:** GUI shows "Connected to server" (green)

### ✅ Functional Tests

- [ ] **Backup:** Create test file and run backup
- [ ] **List:** View backed up files
- [ ] **Snapshot:** Verify snapshot created on VM
- [ ] **Restore:** Restore a file from backup

### ✅ GUI Tests

- [ ] Connection status updates correctly
- [ ] Dashboard shows stats
- [ ] File browser works
- [ ] Settings save properly
- [ ] Theme colors display correctly

## Common Issues & Fixes

### Issue: Can't ping VM
```powershell
# On Windows - Check VirtualBox Host-Only network
Get-NetAdapter | Where-Object {$_.Name -like "*VirtualBox*"}
# Should show adapter with status "Up"
```

### Issue: Port 9090 not accessible
```bash
# On VM - Check firewall
sudo ufw status
sudo ufw allow 9090/tcp

# Check if server is listening
sudo netstat -tlnp | grep 9090
# Should show: tcp  0.0.0.0:9090  LISTEN
```

### Issue: ZFS commands fail
```bash
# On VM - Check ZFS pool
sudo zpool status
sudo zfs list

# If pool doesn't exist, recreate
sudo ./vm-setup.sh  # Reruns setup
```

### Issue: Connection refused in GUI
1. Verify server IP is correct in settings
2. Ensure server is running: `ps aux | grep zfscloud`
3. Check firewall allows port 9090
4. Test with CLI first: `.\bin\zfscloud-cli.exe --config test-vm-config.yaml status`

## File Transfer Test Scenario

**Complete backup and restore workflow:**

```powershell
# On Windows - Create test data
New-Item -Path "C:\Users\$env:USERNAME\Documents\TestBackup\Project" -ItemType Directory -Force
1..5 | ForEach-Object {
    "Document $_" | Out-File "C:\Users\$env:USERNAME\Documents\TestBackup\Project\doc$_.txt"
}

# Run backup
.\bin\zfscloud-cli.exe --config test-vm-config.yaml backup

# Verify on server
```

```bash
# On VM - Check what was backed up
sudo zfs list -t snapshot
ls -la ~/zfscloud-data/clients/

# Query database
sqlite3 ~/zfscloud-data/zfscloud.db "SELECT * FROM snapshots;"
```

```powershell
# On Windows - List backups
.\bin\zfscloud-cli.exe --config test-vm-config.yaml list

# Restore a file
.\bin\zfscloud-cli.exe --config test-vm-config.yaml restore --file "Project/doc1.txt"
```

## Monitoring

### Server Logs (VM)
```bash
# If running in background
tail -f ~/server.log

# Or check systemd logs (if using service)
sudo journalctl -u zfscloud-server -f
```

### ZFS Status (VM)
```bash
# Pool health
sudo zpool status

# Space usage
sudo zfs list

# All snapshots
sudo zfs list -t snapshot

# Snapshot details
sudo zfs get all tank/backups/test-client-01@snapshot1
```

### Network Traffic (VM)
```bash
# Monitor connections
watch 'sudo netstat -an | grep 9090'

# Capture traffic (debugging)
sudo tcpdump -i enp0s8 port 9090 -w capture.pcap
```

## Performance Testing

### Network Speed
```bash
# On VM
sudo apt install -y iperf3
iperf3 -s
```

```powershell
# On Windows (download iperf3.exe first)
.\iperf3.exe -c 192.168.56.101
# Should show ~1 Gbps for Host-Only network
```

### Backup Speed
```powershell
# Create 100MB test file
fsutil file createnew C:\Users\$env:USERNAME\Documents\TestBackup\large.bin 104857600

# Time the backup
Measure-Command {
    .\bin\zfscloud-cli.exe --config test-vm-config.yaml backup
}
```

## Next Steps

After successful testing:

1. **Enable TLS:** Generate certificates for secure communication
2. **Systemd Service:** Auto-start server on VM boot
3. **Real ZFS Pool:** Use actual disk instead of file-backed pool
4. **Multiple Clients:** Test with multiple machines
5. **Scheduled Backups:** Configure automatic backup intervals
6. **Monitoring:** Setup Prometheus/Grafana dashboards

## Scripts Reference

| Script | Purpose | Run On |
|--------|---------|--------|
| `build-server-linux.ps1` | Cross-compile server for Linux | Windows |
| `test-vm-connection.ps1` | Verify connectivity and setup test config | Windows |
| `vm-setup.sh` | Install dependencies and configure VM | Ubuntu VM |

## Quick Commands

**Windows (PowerShell):**
```powershell
# Build everything
.\build.ps1

# Build just Linux server
.\build-server-linux.ps1

# Test connection
.\test-vm-connection.ps1 -ServerIP 192.168.56.101

# Run CLI
.\bin\zfscloud-cli.exe --config test-vm-config.yaml status

# Run GUI
.\bin\zfscloud-gui.exe
```

**VM (Bash):**
```bash
# Start server (foreground)
sudo ./zfscloud-server --config ~/server-config.yaml

# Start server (background)
sudo nohup ./zfscloud-server --config ~/server-config.yaml > ~/server.log 2>&1 &

# Stop server
sudo pkill zfscloud-server

# Check status
ps aux | grep zfscloud-server
sudo netstat -tlnp | grep 9090
tail -f ~/server.log

# ZFS commands
sudo zpool status
sudo zfs list -t all
```

## Troubleshooting

Run these diagnostic commands if you encounter issues:

**Windows:**
```powershell
# Network diagnostics
ping 192.168.56.101
Test-NetConnection -ComputerName 192.168.56.101 -Port 9090
Get-NetAdapter | Where-Object {$_.Name -like "*VirtualBox*"}

# Check if files exist
Test-Path bin/zfscloud-cli.exe
Test-Path test-vm-config.yaml
```

**VM:**
```bash
# Network diagnostics
ip addr show
ping 192.168.56.1  # Windows host
sudo netstat -tlnp | grep 9090

# Service diagnostics
ps aux | grep zfscloud-server
sudo systemctl status zfscloud-server  # If using systemd

# ZFS diagnostics
sudo zpool status
sudo zfs list
lsmod | grep zfs  # Check if ZFS module loaded
```

## Success Criteria

You'll know everything is working when:

1. ✅ VM has IP address (e.g., 192.168.56.101)
2. ✅ Can ping VM from Windows
3. ✅ Port 9090 is accessible
4. ✅ Server process is running on VM
5. ✅ GUI shows "Connected to server" (green)
6. ✅ Can browse files in GUI
7. ✅ Backup completes successfully
8. ✅ Snapshot appears on VM (`sudo zfs list -t snapshot`)
9. ✅ Can restore files

---

**For detailed information, see:** `VIRTUALBOX_DEPLOYMENT_GUIDE.md`
