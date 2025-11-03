# VirtualBox Deployment - Ready to Test!

## Summary

All resources have been created for deploying and testing ZFSCloud in a VirtualBox environment with the server on Ubuntu and client on Windows.

## What's Ready

### ✅ Documentation
- **`VIRTUALBOX_QUICKSTART.md`** - Fast 30-minute setup guide (recommended starting point)
- **`VIRTUALBOX_DEPLOYMENT_GUIDE.md`** - Comprehensive deployment manual with detailed instructions

### ✅ Scripts

#### Windows Scripts (PowerShell)
- **`build-server-linux.ps1`** ✅ TESTED
  - Cross-compiles server for Linux (GOOS=linux)
  - Binary: `bin/zfscloud-server` (22.01 MB)
  - Ready to transfer to VM

- **`test-vm-connection.ps1`**
  - Tests network connectivity to VM
  - Verifies port 9090 accessibility
  - Creates test configuration automatically
  - Creates test backup directory and files

#### Ubuntu Scripts (Bash)
- **`vm-setup.sh`**
  - Installs ZFS utilities
  - Creates test ZFS pool
  - Configures firewall (ufw)
  - Sets up data directory
  - Creates server configuration

## Quick Start Steps

### 1. Create Ubuntu VM (10 min)
```
VirtualBox → New VM
- Name: zfscloud-server
- Type: Linux / Ubuntu 64-bit
- RAM: 2048 MB
- Network: NAT + Host-only Adapter
- Install Ubuntu Server 22.04
```

### 2. Setup VM (5 min)
```bash
# Transfer and run setup script
scp vm-setup.sh zfscloud@<vm-ip>:~/
ssh zfscloud@<vm-ip>
chmod +x vm-setup.sh
sudo ./vm-setup.sh
```

### 3. Deploy Server (5 min)
```powershell
# On Windows - Server is already built!
scp bin/zfscloud-server zfscloud@<vm-ip>:~/

# On VM
chmod +x ~/zfscloud-server
sudo ./zfscloud-server --config ~/server-config.yaml
```

### 4. Test Connection (5 min)
```powershell
# On Windows
.\test-vm-connection.ps1 -ServerIP <vm-ip>

# If successful
.\bin\zfscloud-gui.exe
# Configure: <vm-ip>:9090 → Connect
```

## What Gets Tested

### Connection Status Fix (Current PR)
The GUI connection status bug has been fixed:
- ✅ Shows "Disconnected" when server not running
- ✅ Shows "Connected" only when actually connected
- ✅ Uses proper `IsConnected()` method checking `conn != nil`

This VirtualBox setup is perfect for validating this fix!

### Theme System (Recently Completed)
- ✅ Color theming with Material Design
- ✅ Semantic colors (Success=green, Error=red, etc.)
- ✅ Status indicators (Connected=green, Disconnected=red)
- ✅ Loading widgets and progress indicators

### Core Functionality
- Server-client gRPC communication
- File backup and chunking
- ZFS snapshot creation
- File listing and browsing
- Restore operations
- Connection management

## Network Architecture

```
Windows Host (192.168.56.1)
    │
    │ Host-Only Network (192.168.56.x)
    │
    ▼
Ubuntu VM (192.168.56.101)
    │
    ├─ ZFSCloud Server (port 9090)
    ├─ ZFS Pool (tank/backups)
    └─ SQLite Database (client metadata)
```

## Expected Results

### Before Connecting
- GUI Status: "Disconnected from server" (red indicator)
- Dashboard: No stats available
- Files: Browse disabled

### After Connecting
- GUI Status: "Connected to server" (green indicator)
- Dashboard: Shows server stats
- Files: Can browse and restore

### During Backup
- Progress indicator shows percentage
- Current file name displays
- Upload speed visible
- Files scanned/uploaded count updates

### After Backup
- Snapshot created on server
- Can list backed up files
- Can restore individual files
- ZFS snapshot visible on VM

## Verification Commands

### On Windows
```powershell
# Test network
ping <vm-ip>
Test-NetConnection -ComputerName <vm-ip> -Port 9090

# Test client
.\bin\zfscloud-cli.exe --config test-vm-config.yaml status

# Run GUI
.\bin\zfscloud-gui.exe
```

### On Ubuntu VM
```bash
# Check server running
ps aux | grep zfscloud-server
sudo netstat -tlnp | grep 9090

# Check ZFS
sudo zpool status
sudo zfs list -t snapshot

# Check data
ls -la ~/zfscloud-data/clients/
sqlite3 ~/zfscloud-data/zfscloud.db "SELECT * FROM clients;"
```

## Troubleshooting

### Can't ping VM
- Check VirtualBox Host-Only network adapter is enabled
- Verify VM has IP: `ip addr show` on VM
- Windows: `Get-NetAdapter | Where-Object {$_.Name -like "*VirtualBox*"}`

### Port 9090 not accessible
- Ensure server is running: `ps aux | grep zfscloud`
- Check firewall: `sudo ufw status` (should allow 9090/tcp)
- Verify listening: `sudo netstat -tlnp | grep 9090`

### GUI shows "Disconnected"
- ✅ This is expected if server not running (bug is fixed!)
- Start server on VM
- Wait a few seconds for heartbeat
- Status should change to "Connected"

## Performance Expectations

### Host-Only Network
- Speed: ~1 Gbps (local VM network)
- Latency: <1ms
- Good for: Development and testing
- Not suitable for: Production workloads

### Backup Performance
- Small files (< 1MB): ~100-200 files/sec
- Large files (100MB+): ~80-100 MB/sec
- Depends on: Chunk size, file count, ZFS pool type

## File Structure

```
zfscloud/
├── VIRTUALBOX_QUICKSTART.md          ← Start here!
├── VIRTUALBOX_DEPLOYMENT_GUIDE.md    ← Detailed guide
├── VIRTUALBOX_DEPLOYMENT_STATUS.md   ← This file
├── build-server-linux.ps1            ← Build Linux binary
├── test-vm-connection.ps1            ← Test connectivity
├── vm-setup.sh                       ← Setup Ubuntu VM
└── bin/
    ├── zfscloud-server               ← Linux server (22MB)
    ├── zfscloud-gui.exe              ← Windows GUI (25MB)
    └── zfscloud-cli.exe              ← Windows CLI (17MB)
```

## Security Notes

### Current Setup (Testing)
- ⚠️ TLS disabled for simplicity
- ⚠️ File-backed ZFS pool (not production-ready)
- ⚠️ No authentication beyond client ID
- ✅ Safe for isolated VM testing

### Production Recommendations
- Enable TLS with proper certificates
- Use real disk for ZFS pool
- Implement proper authentication
- Add rate limiting
- Setup monitoring and logging
- Use systemd service for auto-start

## Next Steps After Successful Test

1. **Enable TLS**
   - Generate server certificate
   - Configure client to trust certificate
   - Update configs: `security.tls.enabled = true`

2. **Real ZFS Pool**
   - Add second disk to VM
   - Create pool: `sudo zpool create tank /dev/sdb`
   - Better performance and reliability

3. **Systemd Service**
   - Create `/etc/systemd/system/zfscloud-server.service`
   - Enable auto-start on boot
   - Use systemd for log management

4. **Monitoring**
   - Prometheus metrics endpoint
   - Grafana dashboards
   - Alert on errors or low space

5. **Multiple Clients**
   - Test from different machines
   - Verify multi-client concurrency
   - Check ZFS snapshot organization

## Success Checklist

Before considering test complete:

- [ ] VM is accessible from Windows
- [ ] Server starts without errors
- [ ] ZFS pool is healthy
- [ ] GUI connects to server
- [ ] Connection status shows correctly (green when connected)
- [ ] Can create test backup
- [ ] Backup appears in ZFS snapshots
- [ ] Can list backed up files
- [ ] Can restore files
- [ ] Theme colors display correctly
- [ ] Loading indicators work
- [ ] No crashes or errors in logs

## Getting Help

If you encounter issues:

1. Check the detailed guide: `VIRTUALBOX_DEPLOYMENT_GUIDE.md`
2. Review logs:
   - Server: `tail -f ~/server.log` on VM
   - Client: Check console output
3. Verify network: `ping` and `Test-NetConnection`
4. Check processes: `ps aux | grep zfscloud`
5. Inspect ZFS: `sudo zpool status`

## Resources

- VirtualBox: https://www.virtualbox.org/
- Ubuntu Server: https://ubuntu.com/download/server
- ZFS Documentation: https://openzfs.github.io/openzfs-docs/
- Go Cross-Compilation: https://go.dev/doc/install/source#environment

---

**Ready to test!** Start with `VIRTUALBOX_QUICKSTART.md` for the fastest path to a working setup.
