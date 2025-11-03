# ZFSCloud Client - Usage Guide

## Running the Client

The ZFSCloud client supports multiple modes of operation:

### 1. GUI Mode (Default) ✅

Launch the graphical user interface:

```powershell
./bin/zfscloud-gui.exe
```

Or explicitly:

```powershell
./bin/zfscloud-gui.exe --gui
```

**Features:**
- ✅ Visual dashboard with backup status
- ✅ Configuration editor
- ✅ File browser
- ✅ Toast notifications
- ✅ Confirmation dialogs for destructive actions
- ✅ Progress bars for long operations
- ✅ Material Design UI

### 2. Interactive CLI Mode

Launch the interactive command-line interface:

```powershell
./bin/zfscloud-gui.exe --interactive
```

**Features:**
- Command prompt for manual operations
- Upload files and directories
- Start/stop automatic backup engine
- View status and statistics
- Register with server
- List files and snapshots

### 3. Command-Line Options

```
Usage: zfscloud-gui.exe [options]

Options:
  --config <path>       Path to configuration file (default: "client-config.yaml")
  --keys <path>         Path to key file (default: "client-keys.dat")
  --gui                 Run in GUI mode (default: true)
  --interactive         Run in interactive CLI mode (default: false)
```

**Examples:**

```powershell
# GUI mode with custom config
./bin/zfscloud-gui.exe --config my-config.yaml

# CLI mode with custom keys
./bin/zfscloud-gui.exe --interactive --keys my-keys.dat

# GUI mode explicitly
./bin/zfscloud-gui.exe --gui --config client-config.yaml
```

## First Run

On first run, the client will:

1. **Generate Keys** - Create Ed25519 key pair if not found
2. **Load Configuration** - Load `client-config.yaml` or use defaults
3. **Connect to Server** - Connect to `localhost:50051` (or configured address)
4. **Launch GUI** - Open the graphical interface

## Configuration

Edit `client-config.yaml`:

```yaml
client:
  server_address: "localhost:50051"
  client_id: "client-1234567890"
  
backup:
  folders:
    - "C:\\Users\\YourName\\Documents"
    - "C:\\Users\\YourName\\Pictures"
  exclude_patterns:
    - "*.tmp"
    - "*.log"
  scan_interval: 300  # seconds
```

## GUI Features

### Dashboard View
- **Connection Status** - Shows server connection state
- **Engine Controls** - Start/Stop/Pause/Resume backup engine
- **Scan Button** - Trigger file scan with progress bar
- **Activity** - Current backup activity with helpful messages
- **Statistics** - Files scanned, uploaded, bytes transferred

### Settings View
- **Server Configuration** - Address, port, TLS settings
- **Backup Folders** - Add/remove directories to backup
- **Exclude Patterns** - Skip files matching patterns
- **Test Connection** - Verify server connectivity

### Files View
- **File Browser** - Browse backed up files
- **Search** - Find files by name
- **Restore** - Download files from backup

## Troubleshooting

### GUI Not Launching

If you see the CLI prompt instead of GUI:

```powershell
# Explicitly launch GUI
./bin/zfscloud-gui.exe --gui
```

### Connection Failed

Check:
- Server is running: `docker-compose up -d` (server side)
- Server address correct in config
- Firewall not blocking port 50051
- TLS certificates valid

### Keys Not Loading

If keys fail to load:
- Check file permissions on `client-keys.dat`
- Verify file format (client ID + base64 private key)
- Delete and regenerate: `Remove-Item client-keys.dat`

## Common Tasks

### Register New Client

1. Start GUI: `./bin/zfscloud-gui.exe`
2. Go to Settings tab
3. Click "Register" button
4. Wait for admin approval on server

### Start Automatic Backup

1. Start GUI
2. Go to Dashboard
3. Click "Start Backup" button
4. Engine will monitor configured folders

### Upload Single File

**CLI Mode:**
```powershell
./bin/zfscloud-gui.exe --interactive
> upload C:\path\to\file.txt remote-file.txt
```

**GUI Mode:**
- Go to Files tab
- Click "Upload" button
- Select file
- Click "Upload"

### View Backup Status

**Dashboard** shows:
- Current activity
- Files scanned/uploaded
- Bytes transferred
- Last backup time
- Connection state

## Performance

### Binary Size
- GUI mode: 17.3 MB
- Includes Gio UI framework
- Zero external dependencies

### Memory Usage
- Idle: ~50 MB
- Active backup: ~100-200 MB
- Depends on file count and size

### CPU Usage
- Idle: <1%
- Scanning: 5-15%
- Uploading: 10-25%

## Security

### Key Storage
- Private keys stored in `client-keys.dat` (0600 permissions)
- Passphrase stored in OS keychain (Windows Credential Manager)
- Communication encrypted with TLS

### Network
- gRPC with TLS 1.3
- Ed25519 signatures for authentication
- AES-256 encryption for data at rest

## Support

For issues or questions:
- Check logs in terminal output
- Review server logs: `docker-compose logs`
- File issue on GitHub: `github.com/ioustamora/zfscloud`

---

**Version:** 0.2.0  
**Platform:** Windows, Linux, macOS  
**License:** MIT
