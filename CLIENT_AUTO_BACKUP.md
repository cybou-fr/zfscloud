# ZFSCloud Client - Automatic Backup Quick Start

## Overview

The ZFSCloud client now includes an automatic backup engine that watches your folders for changes and automatically backs them up to the server.

## Prerequisites

1. **Server Running**: Ensure the ZFSCloud server is running and accessible
2. **Client Registered**: You must register with the server first
3. **Configuration**: Configure backup folders in `client-config.yaml`

## Configuration

Edit `client-config.yaml`:

```yaml
client:
  server_address: "localhost:50051"
  client_id: "your-client-id"  # Optional, auto-generated if empty

backup:
  folders:
    - "C:\\Users\\YourName\\Documents"
    - "C:\\Users\\YourName\\Pictures"
  
  exclude_patterns:
    - "*.tmp"
    - "*.cache"
    - ".git"
    - "node_modules"
    - "__pycache__"
    - "*.swp"
    - "~$*"
  
  schedule:
    check_interval_minutes: 5    # Check for changes every 5 minutes
    backup_on_change: true        # Backup immediately when files change

bandwidth:
  limit_mbps: 0                   # 0 = no limit
  throttle_on_metered: false
```

## Usage

### Start the Client

```powershell
.\bin\zfscloud-client.exe
```

The client will:
1. Load/generate Ed25519 keys
2. Connect to the server
3. Enter interactive mode

### Interactive Commands

#### First Time Setup

```text
> register
```

Registers your client with the server. Wait for admin approval before proceeding.

#### Start Automatic Backups

```text
> auto-start
```

This will:
- Create the backup engine
- Start watching configured folders recursively
- Automatically detect file changes (create, modify, delete)
- Upload changes to server

#### Check Status

```text
> status
```

Shows:
- Client ID and public key
- Backup engine state (idle, scanning, backing up, error, paused)
- Files scanned, uploaded, failed counters
- Total bytes uploaded
- Last backup timestamp
- Any errors

#### Perform Initial Scan

```text
> auto-scan
```

Scans all configured folders and queues files for initial backup. This is useful when starting the engine for the first time.

#### Pause/Resume

```text
> auto-pause    # Pause automatic backups
> auto-resume   # Resume automatic backups
```

Useful when:
- You're on metered connection
- System is under heavy load
- Performing large file operations

#### Stop Automatic Backups

```text
> auto-stop
```

Stops the backup engine completely. Files will no longer be watched.

### Manual Backup Commands

Even with automatic backup running, you can still manually upload files:

```text
> upload C:\path\to\file.txt remotefile.txt
> upload-dir C:\path\to\folder remote-folder
```

### Restore Commands

```text
> list                  # List all backed up files
> list path/to/folder   # List specific folder
> snapshots             # List available snapshots
```

## Workflow Examples

### Example 1: Initial Setup

```text
> register
✓ Registration submitted. Waiting for admin approval.

# Wait for admin to approve...

> auto-start
✓ Backup engine started!
  Watching 2 folder(s)
    - C:\Users\YourName\Documents
    - C:\Users\YourName\Pictures

> auto-scan
Performing initial scan...
✓ Scan complete: found 1,234 files

> status
Client ID: abc-123-def
Backup Engine:
  State: backing_up
  Files Scanned: 1234
  Files Uploaded: 456
  Files Failed: 0
  Bytes Uploaded: 104857600
  Last Backup: 2025-10-29T10:30:00Z
```

### Example 2: Daily Usage

The engine runs automatically in the background:

```text
# Edit a document...
# [Engine detects change after 5 seconds]
# [Engine automatically uploads to server]

> status
  State: idle
  Files Uploaded: 1,235    # Counter increased by 1
  Last Backup: 2025-10-29T14:25:00Z
```

### Example 3: Pausing During Heavy Work

```text
> auto-pause
✓ Backup engine paused

# Work on large files...
# No automatic uploads occur

> auto-resume
✓ Backup engine resumed

# Engine catches up with changes made during pause
```

## How It Works

### Real-Time File Watching

The engine uses `fsnotify` to detect file system events:
- **Create**: New file/folder added
- **Modify**: File content changed
- **Delete**: File removed (logged but not uploaded)
- **Rename**: File renamed (treated as delete + create)

### Intelligent Batching

The engine uses a 5-second debounce timer:
1. File change detected
2. Engine waits 5 seconds
3. If more changes occur, timer resets
4. When timer expires, all queued changes are uploaded

This prevents excessive uploads during rapid edits (e.g., IDE auto-save).

### Deduplication

The engine maintains a change queue using file paths as keys:
- Multiple changes to same file = single upload
- Only latest version is uploaded
- Reduces network traffic and server load

### Compression

Files are compressed before upload using zstd (default):
- Fast compression speed
- Good compression ratio
- Automatically skips already-compressed formats (images, videos, archives)

### Exclude Patterns

Files matching exclude patterns are ignored:
- Wildcards supported: `*.tmp`, `~$*`
- Directory names: `node_modules`, `.git`
- Applied at multiple levels for efficiency

## Monitoring

### Check Status Regularly

```text
> status
```

### Watch for Errors

If `State: error` appears, check `Last Error` field:

```text
  State: error
  Last Error: failed to upload file.txt: connection refused
```

Common errors:
- **Connection refused**: Server is down
- **Permission denied**: File access issue
- **Quota exceeded**: Storage limit reached

## Best Practices

1. **Start Small**: Begin with one folder, verify it works
2. **Use Exclude Patterns**: Exclude temp files, build artifacts
3. **Monitor First Hour**: Watch stats to ensure everything works
4. **Check Quota**: Use server admin tools to monitor space
5. **Pause When Needed**: Pause during large downloads/metered connections

## Troubleshooting

### Engine Won't Start

```text
Error: No backup folders configured
```
**Solution**: Add folders to `backup.folders` in config file.

### Files Not Being Uploaded

```text
State: idle
Files Uploaded: 0
```
**Possible causes**:
- All files match exclude patterns
- No changes detected yet (try editing a file)
- Engine paused (use `auto-resume`)

### High CPU/Memory Usage

**Solutions**:
- Reduce number of watched folders
- Add more exclude patterns
- Increase `check_interval_minutes`

### Connection Errors

```text
Last Error: connection refused
```
**Solutions**:
- Verify server is running: `docker ps` or check server logs
- Check `server_address` in config
- Verify network connectivity

## Advanced Configuration

### Compression

To change compression algorithm, modify the engine code:

```go
// In client/backup/engine.go, NewEngine function
compressionType := CompressionZstd    // Default
// compressionType := CompressionGzip  // Alternative
// compressionType := CompressionNone  // No compression
```

### Debounce Time

To change how long the engine waits before uploading:

```go
// In client/backup/engine.go, NewEngine function
debounceTime: 5 * time.Second,    // Default
// debounceTime: 10 * time.Second, // Wait longer
// debounceTime: 1 * time.Second,  // Upload faster
```

### Backup Interval

In `client-config.yaml`:

```yaml
backup:
  schedule:
    check_interval_minutes: 5    # Check every 5 minutes
    backup_on_change: true       # Also backup on file change
```

## Exit Safely

Always use `exit` or `quit` command to shut down properly:

```text
> exit
Stopping backup engine...
✓ Backup engine stopped
Goodbye!
```

This ensures:
- All pending uploads complete
- Resources are released
- File watchers are closed

## Next Steps

- **Phase 7**: Gio-based GUI with system tray icon
- **Phase 8**: TLS/mTLS for secure connections
- **Phase 9**: Automatic snapshot scheduling

---

**Version**: Phase 6 Complete  
**Status**: Fully Functional ✅  
**Documentation**: [Full Docs](./README.md)
