# Quick Testing Guide - ZFSCloud

## Prerequisites Check

1. **Start Docker Desktop**
   - Open Docker Desktop application
   - Wait for it to be fully running (whale icon in system tray should be steady)
   - Test: Run `docker ps` in PowerShell - should not show errors

2. **Verify Binaries Built**
   ```powershell
   Get-ChildItem bin/ | Where-Object {$_.Name -like "zfscloud-*"}
   ```
   Should show:
   - `zfscloud-server` (Linux binary, ~22 MB)
   - `zfscloud-gui.exe` (Windows binary, ~17 MB)

## Step 1: Build Binaries (if needed)

```powershell
# Build server for Linux
$env:GOOS = "linux"; $env:GOARCH = "amd64"; $env:CGO_ENABLED = "0"
go build -o bin/zfscloud-server ./server/cmd

# Build GUI for Windows
$env:GOOS = "windows"; $env:GOARCH = "amd64"
go build -o bin/zfscloud-gui.exe ./client/cmd
```

## Step 2: Deploy Server in Docker

```powershell
cd docker
docker-compose build
docker-compose up -d
```

Wait for server to initialize (~30 seconds):
```powershell
docker logs -f zfscloud-server
```

You should see:
```
Creating ZFS pool...
ZFS pool created successfully
ZFS Pool Status:
  pool: tank
 state: ONLINE
Server initialized successfully
```

Press Ctrl+C to stop following logs.

## Step 3: Verify Server is Running

```powershell
# Check container status
docker ps

# Should show zfscloud-server running on port 50051

# Check ZFS pool
docker exec zfscloud-server zpool status tank

# Should show ONLINE status

# Check server logs
docker logs --tail 20 zfscloud-server
```

## Step 4: Create Client Configuration

```powershell
cd ..  # Back to project root

# Copy example config
Copy-Item client/config.example.yaml client-config.yaml
```

Edit `client-config.yaml`:
```yaml
client:
  client_id: "test-laptop"
  client_name: "Test Laptop"
  server_address: "localhost:50051"
  watch_directory: "C:\\Users\\<YourName>\\Documents"
  upload_interval: "5m"
  private_key_path: "./client-key.pem"

backup:
  compression: true
  deduplication: true
  batch_size: 100
```

## Step 5: Launch GUI Client

```powershell
.\bin\zfscloud-gui.exe
```

GUI window should open with 3 tabs: Dashboard, Settings, Files

## Step 6: Configure Client

1. Click **Settings** button
2. Fill in:
   - Server Address: `localhost:50051`
   - Client ID: `test-laptop`
   - Client Name: `Test Laptop`
   - Watch Directory: `C:\Users\<YourName>\Documents`
   - Upload Interval: `5m`
3. Click **Test Connection**
   - Should show: "Connected successfully!"
4. Click **Save Settings**

## Step 7: Start Backup

1. Click **Dashboard** button
2. Click **Start Backup** button
3. Watch statistics update:
   - Connection Status: Connected (green)
   - Engine Status: Running
   - Files Scanned: (counting up)
   - Files Uploaded: (counting up)
   - Bytes Transferred: (increasing)

## Step 8: Create Test Files

```powershell
# Create a test file
"Hello ZFSCloud!" | Out-File "$env:USERPROFILE\Documents\test-zfscloud.txt"

# Create another file
"Test file 2" | Out-File "$env:USERPROFILE\Documents\test2.txt"
```

Wait a few seconds, then check Dashboard - Files Uploaded should increase.

## Step 9: Browse Files

1. Click **Files** button
2. Click **Refresh** button
3. You should see your uploaded files in the list
4. Click on a file to select it

## Step 10: Test Restore

1. Select a file from the list
2. Click **Restore Selected** button
3. Dialog appears - enter destination path:
   ```
   C:\Users\<YourName>\Downloads\restored-test.txt
   ```
4. Click **Restore** button
5. Progress shows "Downloading file..."
6. Success message appears

Verify file was restored:
```powershell
Get-Content "$env:USERPROFILE\Downloads\restored-test.txt"
```

## Step 11: Check System Tray

1. Look at system tray (bottom-right of screen)
2. Find ZFSCloud icon
3. Right-click icon
4. Menu shows: Dashboard, Settings, Files, Start/Stop Backup, Quit

## Verification Commands

### Check Server Database
```powershell
docker exec -it zfscloud-server sqlite3 /var/lib/zfscloud/zfscloud.db

# In SQLite:
SELECT * FROM users;
SELECT COUNT(*) FROM files;
.exit
```

### Check ZFS Storage
```powershell
# List datasets
docker exec zfscloud-server zfs list -r tank

# Check files for your client
docker exec zfscloud-server ls -lah /tank/test-laptop/
```

### View Server Logs
```powershell
# Real-time logs
docker logs -f zfscloud-server

# Last 50 lines
docker logs --tail 50 zfscloud-server

# Search for uploads
docker logs zfscloud-server | Select-String "Uploaded"
```

## Cleanup

### Stop Server (Keep Data)
```powershell
cd docker
docker-compose down
cd ..
```

### Clean Everything (Delete All Data)
```powershell
cd docker
docker-compose down -v  # -v removes volumes
cd ..
```

## Troubleshooting

### Docker Desktop Not Running
**Error**: "open //./pipe/dockerDesktopLinuxEngine: The system cannot find the file"

**Solution**: 
1. Open Docker Desktop
2. Wait for it to fully start
3. Run `docker ps` - should work without errors

### Server Won't Start
```powershell
# Check logs
docker logs zfscloud-server

# Common issues:
# - Port 50051 in use: Change port in docker-compose.yml
# - ZFS module: Container must be privileged (already configured)
```

### Client Can't Connect
**Checklist**:
1. Server running? `docker ps | findstr zfscloud`
2. Port exposed? `docker port zfscloud-server`
3. Server address correct? `localhost:50051` in Settings

### Files Not Uploading
1. Check backup engine is Started (Dashboard)
2. Check watch directory exists and has files
3. Check server logs: `docker logs zfscloud-server | Select-String "error"`

## Success Criteria

You've successfully tested ZFSCloud when:

- [x] Server starts and creates ZFS pool
- [x] Client connects to server
- [x] Backup engine starts without errors
- [x] Files upload successfully
- [x] Files appear in Files tab
- [x] File restore works
- [x] System tray icon appears and works
- [x] Statistics update in real-time
- [x] Server logs show activity
- [x] ZFS datasets created for client

## Next Steps

After successful testing:

1. **Review the full documentation**:
   - `DOCKER_TESTING_GUIDE.md` - Comprehensive testing
   - `GUI_USER_GUIDE.md` - Complete GUI manual
   - `docker/README.md` - Advanced Docker usage

2. **Performance testing**:
   - Upload larger files
   - Test with many files
   - Monitor resource usage

3. **Production planning**:
   - Deploy on real hardware
   - Use actual ZFS disks
   - Enable TLS/mTLS
   - Configure backups

---

**Quick Command Reference**:

```powershell
# Start server
cd docker; docker-compose up -d; cd ..

# View logs
docker logs -f zfscloud-server

# Stop server
cd docker; docker-compose down; cd ..

# Launch client
.\bin\zfscloud-gui.exe

# Clean everything
cd docker; docker-compose down -v; cd ..
```

**Status**: Ready for testing! 🚀
