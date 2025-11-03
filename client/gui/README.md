# ZFSCloud GUI Client

A graphical user interface for the ZFSCloud backup client built with Gio UI framework.

## Features

- **Dashboard**: Monitor backup status, statistics, and control the backup engine
- **Settings**: Configure server connection, TLS, backup folders, and advanced options
- **Files** (Coming Soon): Browse backed up files and trigger restores
- **Real-time Updates**: Statistics update automatically every second
- **Material Design**: Clean, modern interface with Material Design components

## Building

### Prerequisites
- Go 1.24 or later
- Windows, Linux, or macOS

### Build Command
```powershell
# Windows
go build -o bin/zfscloud-gui.exe ./client/gui/cmd

# Linux/macOS
go build -o bin/zfscloud-gui ./client/gui/cmd
```

## Running

### Prerequisites
1. Create a `client-config.yaml` file (see Configuration section below)
2. Ensure the ZFSCloud server is running and accessible

### Start the GUI
```powershell
# Windows
.\bin\zfscloud-gui.exe

# Linux/macOS
./bin/zfscloud-gui
```

## Configuration

Create a `client-config.yaml` file in the same directory as the executable:

```yaml
client:
  server_address: "localhost:50051"
  client_id: "my-laptop"
  key_path: "./data/client.key"
  tls:
    enabled: false
    ca_file: "./certs/ca.crt"
    cert_file: "./certs/client.crt"
    key_file: "./certs/client.key"
    server_name: "zfscloud.local"
    insecure_skip_verify: false

backup:
  folders:
    - "C:\\Users\\username\\Documents"
    - "C:\\Users\\username\\Pictures"
  exclude_patterns:
    - "*.tmp"
    - "*.log"
    - ".git"
    - "node_modules"
  schedule:
    check_interval_minutes: 5
    backup_on_change: true

bandwidth:
  limit_mbps: 0  # 0 = unlimited
  throttle_on_metered: false
```

### Configuration Options

#### Client Settings
- `server_address`: Server hostname or IP with port (e.g., "localhost:50051")
- `client_id`: Unique identifier for this client
- `key_path`: Path to store Ed25519 private key

#### TLS Settings
- `enabled`: Enable TLS/mTLS encryption (recommended for production)
- `ca_file`: Path to Certificate Authority certificate
- `cert_file`: Path to client certificate
- `key_file`: Path to client private key
- `server_name`: Server name for TLS verification
- `insecure_skip_verify`: Skip certificate verification (testing only!)

#### Backup Settings
- `folders`: List of directories to back up
- `exclude_patterns`: File patterns to skip (supports wildcards)
- `check_interval_minutes`: How often to scan for changes
- `backup_on_change`: Enable real-time backup on file changes

#### Bandwidth Settings
- `limit_mbps`: Upload speed limit in Mbps (0 = unlimited)
- `throttle_on_metered`: Reduce bandwidth on metered connections

## Using the GUI

### Dashboard
The dashboard shows your backup status at a glance:

**Connection Status**: Shows whether you're connected to the server

**Control Buttons**:
- **Start Backup**: Begin the backup engine
- **Stop**: Stop the backup engine completely
- **Pause/Resume**: Temporarily pause uploads
- **Scan Now**: Trigger an immediate file scan

**Statistics**:
- Files Scanned: Total files found in watched folders
- Files Uploaded: Successfully backed up files
- Files Failed: Files that failed to upload
- Data Uploaded: Total bytes transferred (human-readable format)

**Current Activity**: Shows the file currently being backed up and last backup time

### Settings
Configure all aspects of the backup client:

**Server Connection**:
- Server address and port

**TLS/Security**:
- Enable/disable TLS encryption
- Configure certificate paths

**Backup Configuration**:
- Add/remove backup folders (one per line)
- Set exclude patterns (wildcards supported)
- Configure scan interval

**Advanced Settings**:
- Max upload workers (parallel uploads)
- Chunk size for large files
- Retry attempts for failed uploads

**Actions**:
- **Test Connection**: Verify server connectivity
- **Save Settings**: Persist configuration changes (TODO)

### Files (Coming Soon)
- Browse backed up files
- View backup history
- Restore individual files or folders

## Keyboard Shortcuts (Planned)
- `Ctrl+Q`: Quit
- `Ctrl+S`: Open settings
- `F5`: Refresh
- `Ctrl+R`: Start/stop backup

## Troubleshooting

### "Failed to load config"
- Ensure `client-config.yaml` exists in the current directory
- Check YAML syntax (use a validator)
- Verify file permissions

### "Failed to connect"
- Check server address and port
- Verify server is running
- Check firewall settings
- If using TLS, verify certificate paths

### "Failed to generate key pair"
- Check write permissions for key_path directory
- Ensure directory exists

### GUI doesn't start
- Verify all dependencies are installed: `go mod tidy`
- Check for error messages in the console
- Ensure graphics drivers are up to date

## Development

### Project Structure
```
client/gui/
├── cmd/
│   └── main.go          # Entry point
├── app.go               # Main application
├── dashboard.go         # Dashboard view
├── settings.go          # Settings view
└── files.go             # Files view (placeholder)
```

### Adding a New View
1. Create a new file (e.g., `newview.go`)
2. Define a struct with widget state
3. Implement `Layout(gtx layout.Context) layout.Dimensions`
4. Add view enum to `app.go`
5. Initialize in `NewApp()`
6. Add case in `layoutContent()`

### Gio UI Resources
- [Gio Documentation](https://gioui.org/doc)
- [Gio Examples](https://git.sr.ht/~eliasnaur/gio-example)
- [Material Design](https://gioui.org/doc/material)

## Known Issues / Limitations

- ❌ No navigation UI to switch views (hardcoded to Dashboard)
- ❌ Settings save/load not implemented
- ❌ File browser only shows placeholder
- ❌ No error dialogs or toast notifications
- ❌ No system tray integration
- ❌ No progress bars for uploads
- ❌ No keyboard shortcuts

See `PHASE7_GUI.md` for implementation status and roadmap.

## License

Same as ZFSCloud project

## Support

For issues, questions, or contributions, see the main project README.
