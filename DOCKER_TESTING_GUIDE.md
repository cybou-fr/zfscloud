# ZFSCloud Docker Deployment & Testing Guide

**Quick Start**: Run `.\deploy-and-test.ps1` to deploy everything automatically!

---

## 📋 Overview

This guide walks through deploying the ZFSCloud server in Docker with Debian + ZFS, and testing it with the GUI client.

**Architecture:**
```
┌─────────────────┐         ┌─────────────────────────┐
│  Windows Host   │         │   Docker Container      │
│                 │         │   (Ubuntu 22.04)        │
│  GUI Client ────┼────────>│   ZFSCloud Server       │
│  (Port 50051)   │  gRPC   │   + ZFS Pool (tank)     │
│                 │         │   + SQLite Database     │
└─────────────────┘         └─────────────────────────┘
```

---

## 🚀 Method 1: Automated Deployment (Recommended)

### One-Command Deployment

```powershell
.\deploy-and-test.ps1
```

This will:
1. ✅ Check prerequisites (Docker, Go)
2. ✅ Build server binary for Linux
3. ✅ Build GUI client for Windows
4. ✅ Create Docker image with ZFS
5. ✅ Start server container
6. ✅ Wait for server initialization
7. ✅ Create client configuration
8. ✅ Launch GUI client

### Script Options

```powershell
# Clean everything and start fresh
.\deploy-and-test.ps1 -Clean

# Skip building binaries (use existing)
.\deploy-and-test.ps1 -NoBuild

# Only deploy server (no client)
.\deploy-and-test.ps1 -NoClient

# Only build binaries (no deployment)
.\deploy-and-test.ps1 -NoServer

# Show detailed information
.\deploy-and-test.ps1 -Verbose
```

---

## 🔧 Method 2: Manual Deployment

### Step 1: Build Binaries

```powershell
# Build server for Linux
$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"
go build -o bin/zfscloud-server ./server/cmd

# Build GUI client for Windows
$env:GOOS = "windows"
go build -o bin/zfscloud-gui.exe ./client/cmd
```

### Step 2: Deploy Server

```powershell
cd docker
docker-compose build
docker-compose up -d
```

### Step 3: Check Server Status

```powershell
# View logs
docker logs -f zfscloud-server

# Check ZFS pool
docker exec zfscloud-server zpool status tank

# Check ZFS datasets
docker exec zfscloud-server zfs list -r tank
```

You should see:
```
ZFS Pool Status:
  pool: tank
 state: ONLINE
  scan: none requested
config:
        NAME              STATE
        tank              ONLINE
          /zfs-pool/disk.img  ONLINE

Server initialized successfully
Server listening on 0.0.0.0:50051
```

### Step 4: Configure Client

Edit `client-config.yaml`:
```yaml
client:
  client_id: "my-laptop"
  client_name: "My Laptop"
  server_address: "localhost:50051"
  watch_directory: "C:\\Users\\YourName\\Documents"
  upload_interval: "5m"
  private_key_path: "./client-key.pem"

backup:
  compression: true
  deduplication: true
```

### Step 5: Launch GUI Client

```powershell
cd ..  # Back to project root
.\bin\zfscloud-gui.exe
```

---

## 🧪 Testing Workflow

### Phase 1: Initial Connection Test

1. **Launch GUI Client**
   ```powershell
   .\bin\zfscloud-gui.exe
   ```

2. **Configure Settings**
   - Click **"Settings"** button
   - Server Address: `localhost:50051`
   - Client ID: `test-client-1`
   - Client Name: `Test Client`
   - Watch Directory: `C:\Users\YourName\Documents`
   - Upload Interval: `5m`

3. **Test Connection**
   - Click **"Test Connection"** button
   - You should see: ✅ "Connected successfully!"

4. **Save Settings**
   - Click **"Save Settings"** button
   - Config saved to `client-config.yaml`

### Phase 2: Backup Engine Test

1. **Start Backup**
   - Click **"Dashboard"** button
   - Click **"Start Backup"** button
   - Status should change to: 🟢 "Running"

2. **Monitor Statistics**
   - Watch files scanned counter increment
   - Check bytes transferred
   - View last backup timestamp
   - Statistics update every second

3. **Create Test File**
   ```powershell
   # Create a test file in watch directory
   "Hello ZFSCloud!" | Out-File "$env:USERPROFILE\Documents\test-zfscloud.txt"
   ```

4. **Wait for Upload**
   - Watch "Files Uploaded" counter increase
   - Check server logs:
     ```powershell
     docker logs zfscloud-server | Select-String "Uploaded"
     ```

### Phase 3: File Browser Test

1. **Browse Files**
   - Click **"Files"** button
   - Click **"Refresh"** button
   - You should see uploaded files

2. **View File Details**
   - Click on a file in the list
   - Check path, size, and date
   - File should be highlighted

### Phase 4: Restore Test

1. **Restore File**
   - Select a file from the list
   - Click **"Restore Selected"** button
   - Dialog appears

2. **Choose Destination**
   - Browse to destination folder
   - Or manually enter path
   - Example: `C:\Users\YourName\Downloads\restored.txt`

3. **Download File**
   - Click **"Restore"** button
   - Progress shows: "Downloading file..."
   - Success message appears with path

4. **Verify Restore**
   ```powershell
   # Check restored file exists
   Get-Content "C:\Users\YourName\Downloads\restored.txt"
   ```

### Phase 5: System Tray Test

1. **Check System Tray**
   - Look for ZFSCloud icon in system tray (bottom-right)
   - Right-click icon to see menu

2. **Test Menu Actions**
   - Click **"Dashboard"** - switches to dashboard view
   - Click **"Settings"** - switches to settings view
   - Click **"Files"** - switches to files view
   - Click **"Stop Backup"** - stops backup engine
   - Click **"Start Backup"** - starts backup engine

3. **Background Operation**
   - Minimize window
   - Backup continues in background
   - Use tray menu to control

---

## 🔍 Verification & Debugging

### Check Server Database

```powershell
# Access database
docker exec -it zfscloud-server sqlite3 /var/lib/zfscloud/zfscloud.db

# List users
SELECT * FROM users;

# List files
SELECT client_id, path, size FROM files LIMIT 10;

# Check sessions
SELECT * FROM sessions WHERE expires_at > datetime('now');

# Exit
.exit
```

### Check ZFS Storage

```powershell
# List all datasets
docker exec zfscloud-server zfs list -r tank

# Check dataset for specific client
docker exec zfscloud-server ls -lah /tank/test-client-1/

# Show file contents
docker exec zfscloud-server cat /tank/test-client-1/test-zfscloud.txt
```

### View Server Logs

```powershell
# Follow logs in real-time
docker logs -f zfscloud-server

# Last 50 lines
docker logs --tail 50 zfscloud-server

# Search for errors
docker logs zfscloud-server 2>&1 | Select-String "error|failed|ERROR"

# Search for uploads
docker logs zfscloud-server 2>&1 | Select-String "Uploaded file"
```

### Check Client Key

```powershell
# Verify client key was created
Get-Content client-key.pem

# Should see Ed25519 private key
```

### Network Connectivity

```powershell
# Test gRPC endpoint
Test-NetConnection localhost -Port 50051

# Should show: TcpTestSucceeded : True
```

---

## 🐛 Troubleshooting

### Server Won't Start

**Problem**: Container keeps restarting

**Solution**: Check logs
```powershell
docker logs zfscloud-server
```

**Common Issues:**
- ZFS module not loaded → Docker must run in privileged mode (already configured)
- Port 50051 already in use → Change port in docker-compose.yml
- Binary not found → Rebuild with `$env:GOOS="linux"`

### Client Can't Connect

**Problem**: "Connection refused" or "Failed to connect"

**Checklist:**
1. ✅ Server is running: `docker ps | findstr zfscloud`
2. ✅ Port is exposed: `docker port zfscloud-server`
3. ✅ Server address correct: `localhost:50051`
4. ✅ Firewall allows port 50051

**Test with grpcurl:**
```powershell
# Install grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# List services
grpcurl -plaintext localhost:50051 list

# Should show: backup.BackupService, backup.AdminService
```

### Files Not Uploading

**Problem**: Files scanned but not uploaded

**Checklist:**
1. ✅ Backup engine started
2. ✅ Watch directory exists and has files
3. ✅ Client has valid key: check `client-key.pem` exists
4. ✅ Server logs show activity

**Check server logs:**
```powershell
docker logs zfscloud-server | Select-String "Uploaded|Failed"
```

### Restore Failed

**Problem**: "Failed to restore file"

**Checklist:**
1. ✅ File exists in server: check Files tab
2. ✅ Destination path is writable
3. ✅ Server is running
4. ✅ Network connectivity OK

**Check client logs:**
- GUI shows error message in dialog
- Check console output if running from terminal

### ZFS Pool Issues

**Problem**: "cannot open 'tank': no such pool"

**Solution**: Recreate pool
```powershell
docker exec -it zfscloud-server bash

# Inside container:
zpool create -f tank /zfs-pool/disk.img
zfs set compression=lz4 tank
exit
```

---

## 📊 Performance Testing

### Upload Performance

```powershell
# Create large test file (10 MB)
$random = New-Object byte[] 10485760
(New-Object Random).NextBytes($random)
[System.IO.File]::WriteAllBytes("$env:USERPROFILE\Documents\test-10mb.bin", $random)

# Watch upload in GUI
# Monitor: Dashboard → Bytes Transferred
```

### Database Performance

```powershell
# Count total files
docker exec zfscloud-server sqlite3 /var/lib/zfscloud/zfscloud.db "SELECT COUNT(*) FROM files;"

# Database size
docker exec zfscloud-server du -h /var/lib/zfscloud/zfscloud.db
```

### ZFS Pool Performance

```powershell
# Show pool I/O stats
docker exec zfscloud-server zpool iostat tank

# Show dataset sizes
docker exec zfscloud-server zfs list -r tank

# Compression ratio
docker exec zfscloud-server zfs get compressratio tank
```

---

## 🧹 Cleanup

### Stop Server (Keep Data)

```powershell
cd docker
docker-compose down
```

### Clean Everything (Delete Data)

```powershell
cd docker
docker-compose down -v  # -v removes volumes
```

### Remove Images

```powershell
docker rmi zfscloud-server zfscloud-admin
```

### Full Clean with Script

```powershell
.\deploy-and-test.ps1 -Clean
```

This removes:
- All containers
- All volumes (database + ZFS pool)
- All images

---

## 📈 Advanced Testing

### Multiple Clients

```powershell
# Client 1
.\bin\zfscloud-gui.exe

# Client 2 (new terminal)
# Edit config first to use different client_id
Copy-Item client-config.yaml client2-config.yaml
# Edit client2-config.yaml: change client_id to "test-client-2"
# Then run GUI and point to new config
```

### Load Testing

```powershell
# Create 100 small files
1..100 | ForEach-Object {
    "Test file $_" | Out-File "$env:USERPROFILE\Documents\test-$_.txt"
}

# Watch upload progress in Dashboard
```

### Failure Recovery

```powershell
# Stop server mid-upload
docker stop zfscloud-server

# Restart server
docker start zfscloud-server

# Client should reconnect automatically
# Check Dashboard → Connection Status
```

---

## 📝 Testing Checklist

Use this checklist to verify all functionality:

### Server Deployment
- [ ] Docker image builds successfully
- [ ] Container starts without errors
- [ ] ZFS pool creates successfully
- [ ] Server listens on port 50051
- [ ] Database initializes

### Client Connection
- [ ] GUI launches successfully
- [ ] Settings can be configured
- [ ] Connection test succeeds
- [ ] Settings persist to YAML
- [ ] Private key generates

### Backup Engine
- [ ] Backup starts from Dashboard
- [ ] Files are scanned
- [ ] Files are uploaded
- [ ] Statistics update in real-time
- [ ] Backup can be stopped

### File Management
- [ ] Files tab shows uploaded files
- [ ] File list can be refreshed
- [ ] Files can be selected
- [ ] File details display correctly

### Restore Functionality
- [ ] Restore dialog opens
- [ ] Destination path can be set
- [ ] File downloads successfully
- [ ] Success message displays
- [ ] Restored file is correct

### System Tray
- [ ] Tray icon appears
- [ ] Menu opens on right-click
- [ ] Navigation from tray works
- [ ] Start/Stop from tray works
- [ ] Quit from tray works

### Error Handling
- [ ] Connection errors show dialog
- [ ] Upload errors logged
- [ ] Restore errors show message
- [ ] Invalid settings rejected

---

## 🎓 Next Steps

After successful testing:

1. **Review Logs**
   - Check for any errors or warnings
   - Verify all operations logged correctly

2. **Performance Tuning**
   - Adjust upload interval
   - Enable/disable compression
   - Configure deduplication

3. **Production Planning**
   - Deploy on real hardware
   - Use real ZFS disks (not loop device)
   - Enable TLS/mTLS
   - Configure proper backup retention
   - Set up monitoring

4. **Documentation**
   - Document your specific configuration
   - Create runbooks for operations
   - Plan disaster recovery procedures

---

## 📞 Support

### Documentation
- **Quick Start**: QUICKSTART_GUI.md
- **User Guide**: GUI_USER_GUIDE.md
- **Docker Guide**: docker/README.md
- **Server Deployment**: SERVER_QUICKSTART.md

### Logs
```powershell
# Server logs
docker logs zfscloud-server

# Client console (if run from terminal)
.\bin\zfscloud-gui.exe  # Watch console output
```

---

**Version**: 1.0.0  
**Last Updated**: November 3, 2025  
**Status**: Production Ready ✅

**Happy Testing! 🚀**
