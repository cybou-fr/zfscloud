# ZFSCloud GUI - Quick Start Guide

## Installation (Windows)

### Option 1: Use Pre-built Binary (Recommended)

1. Download `zfscloud-gui.exe` from the releases page
2. Create a folder: `C:\ZFSCloud\`
3. Copy `zfscloud-gui.exe` to this folder
4. Double-click to run

### Option 2: Build from Source

**Prerequisites:**
- Go 1.21 or later installed
- Git installed

**Steps:**
```powershell
# Clone repository
git clone https://github.com/ioustamora/zfscloud.git
cd zfscloud

# Build GUI
go build -o bin/zfscloud-gui.exe ./client/cmd

# Run
.\bin\zfscloud-gui.exe
```

---

## First-Time Setup

### Step 1: Launch Application
Double-click `zfscloud-gui.exe`

### Step 2: Configure Settings

1. Click **"Settings"** button at top
2. Fill in your details:

   ```
   Server Address: your-server.com:50051
   Server Port: 50051
   Client ID: my-laptop
   Client Name: My Laptop
   Watch Directory: C:\Users\YourName\Documents
   Upload Interval: 1h
   ```

3. Click **"Test Connection"** to verify
4. Click **"Save Settings"** to save

### Step 3: Start Backup

1. Click **"Dashboard"** button
2. Click **"Start Backup"**
3. Watch statistics update in real-time

---

## Usage

### Monitor Backups
- **Dashboard** shows live statistics
- **Green** dot = Connected
- **Red** dot = Disconnected

### Browse Files
1. Click **"Files"** button
2. Click **"Refresh"** to load file list
3. Click on a file to select it

### Restore Files
1. Select a file in the Files view
2. Click **"Restore..."** button
3. Enter destination path
4. Click **"Restore"**

---

## Troubleshooting

### Can't Connect?
- Check server address is correct
- Verify server is running
- Test network: `ping your-server.com`
- Check firewall allows port 50051

### No Files Showing?
- Run a backup first (Dashboard → Start Backup)
- Click Refresh in Files view
- Check server has files

### Settings Won't Save?
- Run as Administrator
- Check file permissions
- Verify disk space available

---

## Support

- **User Guide**: See `GUI_USER_GUIDE.md` for detailed documentation
- **Issues**: https://github.com/ioustamora/zfscloud/issues
- **Logs**: Check console output for errors

---

## Keyboard Shortcuts (Planned)

- `Ctrl+Q` - Quit
- `Ctrl+S` - Settings
- `F5` - Refresh

---

## Next Steps

1. ✅ Configure settings
2. ✅ Test connection
3. ✅ Start backup
4. ✅ Monitor progress
5. ✅ Browse backed up files

For complete documentation, see **GUI_USER_GUIDE.md**
