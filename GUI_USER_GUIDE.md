# ZFSCloud GUI - User Guide

## Table of Contents
1. [Introduction](#introduction)
2. [Installation](#installation)
3. [Getting Started](#getting-started)
4. [Dashboard](#dashboard)
5. [Settings](#settings)
6. [File Browser](#file-browser)
7. [Troubleshooting](#troubleshooting)
8. [FAQ](#faq)

---

## Introduction

ZFSCloud GUI is a desktop application for managing your ZFSCloud backups. It provides an intuitive interface to:
- Monitor backup progress in real-time
- Configure backup settings
- Browse backed up files
- Restore files to your computer

### System Requirements
- **Operating System**: Windows 10/11, Linux, or macOS
- **Memory**: 100 MB RAM minimum
- **Disk Space**: 50 MB for application
- **Network**: Connection to ZFSCloud server

---

## Installation

### Windows

#### Option 1: Pre-built Binary
1. Download `zfscloud-gui.exe` from releases
2. Place it in a folder of your choice (e.g., `C:\Program Files\ZFSCloud\`)
3. Double-click to run

#### Option 2: Build from Source
```powershell
# Clone the repository
git clone https://github.com/ioustamora/zfscloud.git
cd zfscloud

# Build the GUI
go build -o bin/zfscloud-gui.exe ./client/cmd

# Run
./bin/zfscloud-gui.exe
```

### Linux

```bash
# Clone the repository
git clone https://github.com/ioustamora/zfscloud.git
cd zfscloud

# Build the GUI
go build -o bin/zfscloud-gui ./client/cmd

# Run
./bin/zfscloud-gui
```

### macOS

```bash
# Clone the repository
git clone https://github.com/ioustamora/zfscloud.git
cd zfscloud

# Build the GUI
go build -o bin/zfscloud-gui ./client/cmd

# Run
./bin/zfscloud-gui
```

---

## Getting Started

### First Launch

When you first launch ZFSCloud GUI, you need to configure your connection:

1. **Launch the application**
   - Double-click `zfscloud-gui.exe` (Windows)
   - Run `./zfscloud-gui` (Linux/macOS)

2. **Navigate to Settings** (click the "Settings" button)

3. **Enter your server details**:
   - **Server Address**: Your ZFSCloud server IP or hostname
   - **Server Port**: Usually `50051`
   - **Client ID**: Your unique client identifier
   - **Client Name**: A friendly name for this computer
   - **Watch Directory**: The folder you want to backup
   - **Upload Interval**: How often to check for new files (e.g., `1h`, `30m`)

4. **Test the connection**
   - Click "Test Connection"
   - Wait for confirmation message

5. **Save your settings**
   - Click "Save Settings"

6. **Return to Dashboard** (click the "Dashboard" button)

### Configuration File

Settings are stored in `client-config.yaml` in the same directory as the executable.

Example configuration:
```yaml
client:
  client_id: "my-laptop"
  client_name: "John's Laptop"
  server_address: "backup.example.com:50051"
  watch_directory: "C:\\Users\\John\\Documents"
  upload_interval: "1h"
  private_key_path: "./client-key.pem"
```

---

## Dashboard

The Dashboard is your main view for monitoring backup operations.

### Statistics Display

The dashboard shows real-time statistics:

- **Files Scanned**: Total files detected in watch directory
- **Files Uploaded**: Successfully backed up files
- **Files Failed**: Files that failed to upload
- **Bytes Uploaded**: Total data transferred (human-readable)
- **Last Backup**: Timestamp of most recent backup

### Connection Status

- **Red indicator**: Disconnected from server
- **Green indicator**: Connected to server

### Backup Controls

#### Start Backup
1. Click "Start Backup" button
2. The engine will scan your watch directory
3. New/modified files will be uploaded automatically

#### Stop Backup
1. Click "Stop Backup" button while backup is running
2. Current file upload will complete
3. Backup engine will stop

#### Pause/Resume
1. Click "Pause/Resume" to temporarily pause backups
2. Click again to resume
3. Note: Implementation status depends on engine support

### Auto-Refresh

Statistics update automatically every second. No manual refresh needed.

---

## Settings

The Settings view allows you to configure the backup client.

### Server Configuration

#### Server Address
- Format: `hostname:port` or `ip:port`
- Example: `backup.example.com:50051` or `192.168.1.100:50051`
- No `http://` or `https://` prefix needed

#### Server Port
- Default: `50051`
- Must match your ZFSCloud server configuration

### Client Configuration

#### Client ID
- **Required**: Unique identifier for this client
- **Format**: Alphanumeric, hyphens, underscores
- **Example**: `johns-laptop`, `office-pc-01`
- **Important**: Don't change this after initial backup (server tracks by ID)

#### Client Name
- **Optional**: Human-readable name
- **Example**: "John's Laptop", "Office Desktop"
- Displayed in server logs

#### Watch Directory
- The folder to backup
- **Windows**: Use format `C:\Users\John\Documents`
- **Linux/macOS**: Use format `/home/john/documents`
- Must be an existing directory
- All files and subdirectories will be backed up

#### Upload Interval
- How often to check for changes
- **Format**: `30s`, `5m`, `1h`, `24h`
- **Recommended**: `1h` for personal computers, `30m` for servers
- Shorter intervals = more frequent backups = more CPU/network usage

### Testing Connection

Before starting backups:

1. Click "Test Connection" button
2. Wait up to 5 seconds
3. Check for success/error message
4. If error, verify:
   - Server is running
   - Server address is correct
   - Firewall allows connection
   - Client key file exists (if using encryption)

### Saving Settings

1. Click "Save Settings" button
2. Settings are written to `client-config.yaml`
3. Restart the application to apply network changes

---

## File Browser

The File Browser lets you view and restore backed up files.

### Viewing Files

1. Navigate to "Files" view (click "Files" button)
2. Click "Refresh" to load file list from server
3. Scroll through the list

### File Information

For each file, you can see:
- **Path**: Full path on the backed up system
- **Size**: File size (KB, MB, GB format)
- **Modified**: Last modification timestamp
- **Directory**: Folders have a trailing `/`

### Selecting Files

- Click on any file to select it
- Selected file appears in **bold**
- Only one file can be selected at a time

### Restoring Files

To restore a backed up file:

1. **Select a file** by clicking on it in the list
2. **Click "Restore..." button**
3. **Restore Dialog appears**:
   - Shows selected file path and size
   - Enter restore destination path
   - Default: `C:\Restored\`
4. **Edit destination** if needed:
   - Type full path including filename
   - Example: `C:\Restored\mydocument.txt`
5. **Click "Restore"** to start
6. Wait for completion message
7. **Click "Cancel"** to close dialog without restoring

### Restore States

- **Normal**: Ready to restore
- **Restoring**: Progress message shown, button disabled
- **Complete**: Success or error message displayed
- **Note**: Actual file download is not yet implemented in this version

### Refreshing File List

- Click "Refresh" button to reload from server
- Useful after uploading new files
- Shows loading indicator while fetching

---

## Troubleshooting

### Application Won't Start

**Problem**: Double-clicking executable does nothing

**Solutions**:
- Check if antivirus is blocking the file
- Run from command line to see error messages:
  ```powershell
  ./zfscloud-gui.exe
  ```
- Verify you have permission to execute the file

### Can't Connect to Server

**Problem**: "Test Connection" fails or shows error

**Solutions**:
1. **Verify server is running**:
   - Check server status
   - Try pinging server: `ping backup.example.com`

2. **Check server address**:
   - Format: `hostname:port` (no http://)
   - Port usually `50051`
   - Try IP address if hostname fails

3. **Firewall issues**:
   - Ensure port 50051 is open
   - Check Windows Firewall settings
   - Check corporate firewall/proxy

4. **Certificate issues**:
   - Ensure `client-key.pem` exists
   - Check file permissions on key file

5. **Network issues**:
   - Verify internet connection
   - Try from another computer
   - Check VPN configuration

### Files Not Showing in Browser

**Problem**: File list is empty or shows "No files to display"

**Solutions**:
1. **Verify backups are running**:
   - Check Dashboard shows "Files Uploaded" > 0
   - Start a backup if none running

2. **Check server has files**:
   - Look at server logs
   - Verify server storage

3. **Refresh the list**:
   - Click "Refresh" button
   - Wait for loading to complete

4. **Connection issues**:
   - Verify green status indicator on Dashboard
   - Test connection in Settings

### Settings Won't Save

**Problem**: Clicking "Save Settings" doesn't work

**Solutions**:
1. **Check file permissions**:
   - Ensure write permission on `client-config.yaml`
   - Try running as administrator (Windows)

2. **Check disk space**:
   - Ensure drive has available space

3. **Verify configuration syntax**:
   - Check for valid values in all fields
   - No special characters in paths

### Backup Performance Issues

**Problem**: Backup is slow or CPU usage is high

**Solutions**:
1. **Increase upload interval**:
   - Change from `30m` to `1h` or `2h`
   - Reduces scanning frequency

2. **Reduce watch directory size**:
   - Backup only essential folders
   - Exclude large media files

3. **Check network bandwidth**:
   - Uploads limited by connection speed
   - Consider running backups overnight

### Error Dialog Appears

**Problem**: Error dialog shows up with message

**Solutions**:
1. **Read the error message** carefully
2. **Common errors**:
   - "Client not initialized": Restart application
   - "Failed to list files": Check server connection
   - "Invalid file selection": Select a file first
3. **Click "OK"** to dismiss
4. **Check application logs** for details

---

## FAQ

### Q: Do I need to keep the GUI open for backups?
**A**: Yes, the backup engine runs while the GUI is open. Minimize it or implement system tray (future feature).

### Q: Can I backup multiple directories?
**A**: Currently only one watch directory is supported. Backup a parent folder to include multiple subdirectories.

### Q: How do I know if my backups are working?
**A**: Check the Dashboard for "Files Uploaded" count and "Last Backup" timestamp.

### Q: Can I pause backups temporarily?
**A**: Use the "Pause/Resume" button on the Dashboard. Note: Implementation depends on engine support.

### Q: Where are my backed up files stored?
**A**: On the ZFSCloud server's ZFS storage pool. Check server configuration for storage location.

### Q: How do I restore an entire directory?
**A**: Currently you restore files one at a time. Batch restore is planned for future versions.

### Q: Can I backup to multiple servers?
**A**: No, each client connects to one server. Run multiple instances with different configs for multiple servers.

### Q: What happens if my connection drops during backup?
**A**: The client will retry automatically. Interrupted uploads will resume when connection is restored.

### Q: Can I schedule backups?
**A**: Use the "Upload Interval" setting. For precise schedules, use OS task scheduler to start/stop the application.

### Q: Is my data encrypted?
**A**: Yes, if using client key files. Check your configuration and server setup for encryption details.

### Q: How much disk space do I need?
**A**: The GUI itself uses ~50 MB. No local storage needed for backups (stored on server).

### Q: Can I see backup history?
**A**: Not in current version. Future releases will include snapshot history viewing.

### Q: What file types are supported?
**A**: All file types are supported. Binary and text files are backed up identically.

### Q: Can I exclude certain files?
**A**: Not in current GUI version. Configure exclusions in server configuration.

### Q: How do I update the GUI?
**A**: Download new version and replace the executable. Your `client-config.yaml` will be preserved.

---

## Keyboard Shortcuts (Planned)

The following shortcuts are planned for future versions:

- **Ctrl+Q**: Quit application
- **Ctrl+S**: Navigate to Settings
- **F5**: Refresh current view
- **Ctrl+R**: Start/Stop backup
- **Esc**: Close dialog

---

## Support

For additional help:

1. **Check logs**: Application logs are printed to console
2. **GitHub Issues**: Report bugs at https://github.com/ioustamora/zfscloud/issues
3. **Documentation**: See README.md and other docs in repository
4. **Server admin**: Contact your ZFSCloud server administrator

---

## Version Information

- **GUI Version**: 1.0 (Phase 7)
- **UI Framework**: Gio v0.9.0
- **Protocol**: gRPC
- **Build Date**: Check executable properties

---

## License

Same as main ZFSCloud project. See LICENSE file in repository.
