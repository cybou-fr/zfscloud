# Server Quick Start Guide

## Prerequisites

- Go 1.23 or later
- ZFS filesystem (Linux or FreeBSD)
- SQLite support (CGo required for go-sqlite3)
- Protocol Buffers compiler (protoc)

## Building the Server

### Windows (PowerShell)
```powershell
# Build everything (proto + server + client)
.\build.ps1 all

# Build server only
.\build.ps1 server

# Or use Go directly
go build -o bin/zfscloud-server.exe ./server/cmd
```

### Linux/macOS
```bash
# Build everything
make all

# Build server only
make server

# Or use Go directly
go build -o bin/zfscloud-server ./server/cmd
```

## Configuration

Create a `config.yaml` file (see `server/config.example.yaml` for template):

```yaml
server:
  listen_address: ":50051"

database:
  path: "./zfscloud.db"

zfs:
  pool: "tank"
  compression: "lz4"
  deduplication: false

quotas:
  default_quota_gb: 100
  max_quota_gb: 1000

snapshots:
  schedule:
    hourly: 24
    daily: 7
    weekly: 4
    monthly: 12

security:
  admin_token: "your-secure-admin-token-here"
  nonce_timeout_minutes: 10
  session_timeout_hours: 24
  signature_timeout_minutes: 5
```

## Running the Server

### Basic Usage
```bash
# With default config file (config.yaml)
./bin/zfscloud-server

# With custom config file
./bin/zfscloud-server -config /path/to/config.yaml
```

### Expected Output
```
ZFSCloud Server v0.1.0 starting...
Configuration loaded from config.yaml
Listen address: :50051
ZFS pool: tank
Initializing database...
Database initialized successfully
Initializing ZFS manager...
ZFS manager initialized successfully
gRPC services registered successfully
Starting gRPC server on :50051
Server initialized successfully
```

## Admin Operations

All admin operations require the admin token from config.

### Using grpcurl

#### List Pending Registrations
```bash
grpcurl -plaintext \
  -d '{"admin_token": "your-admin-token"}' \
  localhost:50051 \
  backup.AdminService/ListPendingRegistrations
```

#### Approve Registration
```bash
grpcurl -plaintext \
  -d '{
    "admin_token": "your-admin-token",
    "registration_id": "reg-uuid",
    "quota_bytes": 107374182400
  }' \
  localhost:50051 \
  backup.AdminService/ApproveRegistration
```

#### List Users
```bash
grpcurl -plaintext \
  -d '{
    "admin_token": "your-admin-token",
    "include_revoked": false
  }' \
  localhost:50051 \
  backup.AdminService/ListUsers
```

#### Get System Status
```bash
grpcurl -plaintext \
  -d '{"admin_token": "your-admin-token"}' \
  localhost:50051 \
  backup.AdminService/GetSystemStatus
```

## ZFS Setup

### Create ZFS Pool (if not exists)
```bash
# Create pool from disk
sudo zpool create tank /dev/sdX

# Or from file (testing)
truncate -s 10G /tmp/zfs-disk.img
sudo zpool create tank /tmp/zfs-disk.img

# Enable compression
sudo zfs set compression=lz4 tank

# Verify
zpool status tank
zfs list tank
```

### Dataset Structure

The server creates datasets per user:
```
tank/
├── user-uuid-1/         # Encrypted dataset for user 1
│   ├── file1.txt
│   └── dir/
│       └── file2.txt
└── user-uuid-2/         # Encrypted dataset for user 2
    └── backup/
```

Each dataset has:
- Native ZFS encryption (AES-256-GCM)
- Per-user passphrase
- Quota enforcement
- Automatic snapshots

## Troubleshooting

### Server won't start

**Database locked:**
```
Error: database is locked
```
Solution: Close other connections or delete `zfscloud.db.lock`

**ZFS pool not found:**
```
Error: pool "tank" not found
```
Solution: Create the pool or update `config.yaml` with correct pool name

**Port already in use:**
```
Error: bind: address already in use
```
Solution: Change `listen_address` in config or stop conflicting service

### Permission Issues

**Cannot create dataset:**
```
Error: permission denied
```
Solution: Run server with sudo or add user to ZFS admin group:
```bash
sudo usermod -a -G zfs $USER
```

**Cannot access database:**
```
Error: unable to open database file
```
Solution: Check file permissions and ensure directory exists:
```bash
mkdir -p data
chmod 755 data
```

## Monitoring

### Check Server Logs
Server logs to stdout. Redirect to file:
```bash
./bin/zfscloud-server 2>&1 | tee server.log
```

### Check Database
```bash
sqlite3 zfscloud.db
> SELECT * FROM users;
> SELECT * FROM registrations;
> SELECT * FROM audit_log ORDER BY created_at DESC LIMIT 10;
> .exit
```

### Check ZFS Datasets
```bash
# List all datasets
zfs list -r tank

# Check quotas
zfs get quota,used,available tank/user-uuid

# Check encryption status
zfs get encryption,keystatus tank/user-uuid

# View snapshots
zfs list -t snapshot -r tank
```

## Security Best Practices

1. **Use Strong Admin Token:**
   ```bash
   # Generate random token
   openssl rand -hex 32
   ```

2. **Enable TLS (TODO):**
   Currently server runs without TLS. Production should use TLS/mTLS.

3. **Restrict Network Access:**
   ```bash
   # Firewall: Only allow specific IPs
   sudo ufw allow from 192.168.1.0/24 to any port 50051
   ```

4. **Regular Backups:**
   ```bash
   # Backup database
   sqlite3 zfscloud.db ".backup zfscloud-backup.db"
   
   # Backup ZFS pool
   zfs snapshot -r tank@backup-$(date +%Y%m%d)
   ```

5. **Monitor Disk Space:**
   ```bash
   # Check pool capacity
   zpool list tank
   
   # Alert if >90% full
   ```

## Graceful Shutdown

Server handles SIGINT (Ctrl+C) and SIGTERM gracefully:

1. Stops accepting new connections
2. Waits for in-flight requests to complete
3. Closes database connection
4. Exits cleanly

```bash
# Send termination signal
kill -TERM $(pgrep zfscloud-server)

# Or use Ctrl+C in terminal
```

## Development Mode

### Enable Verbose Logging
Modify `server/cmd/main.go`:
```go
log.SetFlags(log.LstdFlags | log.Lshortfile)
```

### Use In-Memory Database (Testing)
In `config.yaml`:
```yaml
database:
  path: ":memory:"  # SQLite in-memory database
```

### Hot Reload (TODO)
Currently requires restart for config changes.

## Performance Tuning

### Database Optimization
```sql
-- Add indexes (already in schema)
CREATE INDEX IF NOT EXISTS idx_users_client_id ON users(client_id);
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
```

### ZFS Tuning
```bash
# Enable LZ4 compression (fast)
sudo zfs set compression=lz4 tank

# Adjust recordsize for workload
sudo zfs set recordsize=1M tank  # Large files
sudo zfs set recordsize=128K tank  # Mixed workload

# Enable deduplication (use with caution - RAM intensive)
sudo zfs set dedup=on tank
```

### gRPC Tuning
Modify server initialization for high-load:
```go
grpcServer := grpc_server.NewServer(
    grpc_server.MaxConcurrentStreams(100),
    grpc_server.MaxRecvMsgSize(100 * 1024 * 1024), // 100MB
)
```

## Next Steps

1. **Test Registration Flow:**
   - Run server
   - Use client or grpcurl to register
   - Admin approves via AdminService
   - Verify dataset created

2. **Test Backup Upload:**
   - Client uploads files
   - Verify files in ZFS dataset
   - Check storage quota enforcement

3. **Setup Automated Snapshots:**
   - Implement snapshot scheduler (TODO)
   - Test snapshot creation and retention

4. **Add TLS:**
   - Generate certificates
   - Configure gRPC for TLS
   - Test secure connections

5. **Build Client:**
   - Implement client registration
   - Implement backup engine
   - Create UI

## Resources

- [gRPC Documentation](https://grpc.io/docs/)
- [ZFS Administration Guide](https://openzfs.github.io/openzfs-docs/)
- [SQLite Documentation](https://www.sqlite.org/docs.html)
- [Protocol Buffers Guide](https://protobuf.dev/)
