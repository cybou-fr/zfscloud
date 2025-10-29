# ZFSCloud Docker Testing Guide

This guide explains how to run ZFSCloud in Docker for testing client-server interactions.

## Prerequisites

- Docker Desktop with WSL2 (Windows) or Docker Engine (Linux)
- Docker Compose
- Built binaries (`zfscloud-server.exe` and `zfscloud-client.exe`)

**Important for Windows:** The server container requires ZFS kernel modules, which are only available in Linux. You'll need WSL2 or a Linux VM.

## Quick Start

### 1. Build the Server Binary

```powershell
# Build for Linux (required for Docker)
$env:GOOS="linux"
$env:GOARCH="amd64"
go build -o bin/zfscloud-server ./server/cmd

# Build client for Windows (to run on your host)
$env:GOOS="windows"
$env:GOARCH="amd64"
go build -o bin/zfscloud-client.exe ./client/cmd
```

### 2. Start the Server

```bash
cd docker
docker-compose up -d zfscloud-server
```

### 3. Check Server Logs

```bash
docker-compose logs -f zfscloud-server
```

You should see:
```
Creating ZFS pool...
ZFS pool created successfully
ZFS Pool Status:
  pool: tank
 state: ONLINE
ZFSCloud Server v0.1.0 starting...
Server initialized successfully
```

### 4. Run the Client

```powershell
# From project root
.\bin\zfscloud-client.exe
```

## Configuration

### Server Configuration

Edit `docker/config/server.yaml`:

```yaml
server:
  listen_address: "0.0.0.0:50051"  # Listen on all interfaces

security:
  admin_token: "your-secret-admin-token"  # Change this!
```

### Client Configuration

Create `client-config.yaml` in project root:

```yaml
client:
  server_address: "localhost:50051"  # Docker host
  client_id: "test-client-1"

backup:
  folders:
    - "C:\\Users\\YourName\\Documents"
  interval_seconds: 300
```

## Testing Workflow

### Step 1: Register Client

In the client interactive shell:
```
> register
```

This creates a pending registration waiting for admin approval.

### Step 2: List Pending Registrations (Admin)

```bash
docker-compose run --rm zfscloud-admin /scripts/list-registrations.sh
```

Or use grpcurl directly:
```bash
docker exec zfscloud-server sh -c '
  apt-get update && apt-get install -y curl
  curl -sSL https://github.com/fullstorydev/grpcurl/releases/download/v1.8.9/grpcurl_1.8.9_linux_amd64.tar.gz | tar -xz -C /usr/local/bin
'
```

### Step 3: Approve Registration

Copy the `registration_id` from the previous step:

```bash
docker-compose run --rm zfscloud-admin /scripts/approve-registration.sh <registration-id> 10
```

Or manually with grpcurl:
```bash
docker exec -it zfscloud-server grpcurl -plaintext \
  -d '{
    "admin_token": "change-me-in-production",
    "registration_id": "your-registration-id-here",
    "quota_bytes": 10737418240
  }' \
  localhost:50051 \
  backup.AdminService/ApproveRegistration
```

### Step 4: Send Heartbeat

In the client:
```
> heartbeat
```

You should see: "Heartbeat sent successfully!"

### Step 5: Upload a Test File

Create a test file:
```powershell
echo "Hello ZFSCloud!" > test.txt
```

Upload it:
```
> upload test.txt /test.txt
```

### Step 6: List Files

```
> list
```

You should see your uploaded file.

### Step 7: List Users (Admin)

```bash
docker-compose run --rm zfscloud-admin /scripts/list-users.sh
```

### Step 8: Check System Status (Admin)

```bash
docker-compose run --rm zfscloud-admin /scripts/system-status.sh
```

## Debugging

### Access Server Container

```bash
docker exec -it zfscloud-server bash
```

### Check ZFS Status

```bash
docker exec zfscloud-server zpool status tank
docker exec zfscloud-server zfs list -r tank
```

### Check Database

```bash
docker exec -it zfscloud-server sqlite3 /var/lib/zfscloud/zfscloud.db
```

```sql
SELECT * FROM users;
SELECT * FROM registrations;
SELECT * FROM sessions;
.exit
```

### View Server Logs

```bash
docker-compose logs -f zfscloud-server
```

### Check Dataset Contents

```bash
docker exec zfscloud-server ls -la /tank/<user-id>/
```

## Troubleshooting

### ZFS Module Not Loaded

**Error:** `modprobe: FATAL: Module zfs not found`

**Solution:** ZFS kernel modules must be available in the host kernel. On Windows, use WSL2 with ZFS support:

```bash
# In WSL2 Ubuntu
sudo apt-get install zfsutils-linux
```

### Permission Denied

**Error:** `cannot create 'tank': permission denied`

**Solution:** The container must run in privileged mode (already configured in docker-compose.yml).

### Port Already in Use

**Error:** `bind: address already in use`

**Solution:** Change the port mapping in `docker-compose.yml`:

```yaml
ports:
  - "50052:50051"  # Use different host port
```

Then update client config:
```yaml
server_address: "localhost:50052"
```

### Container Keeps Restarting

**Check logs:**
```bash
docker-compose logs zfscloud-server
```

**Common issues:**
- ZFS pool creation failed
- Configuration file syntax error
- Binary not found (rebuild with GOOS=linux)

## Advanced Usage

### Custom ZFS Pool Size

Edit `docker/Dockerfile.server`:

```dockerfile
RUN truncate -s 50G /zfs-pool/disk.img  # Increase to 50GB
```

Rebuild:
```bash
docker-compose build zfscloud-server
```

### Enable Deduplication

Edit `docker/config/server.yaml`:

```yaml
zfs:
  deduplication: true
```

**Warning:** Deduplication requires significant RAM (5GB+ per TB of storage).

### Persistent Storage

The database and ZFS pool are stored in Docker volumes. To back them up:

```bash
# Backup
docker run --rm -v zfscloud-data:/data -v $(pwd):/backup ubuntu tar czf /backup/zfscloud-backup.tar.gz /data

# Restore
docker run --rm -v zfscloud-data:/data -v $(pwd):/backup ubuntu tar xzf /backup/zfscloud-backup.tar.gz -C /
```

### Multiple Clients

Run multiple client instances:

```powershell
# Client 1
.\bin\zfscloud-client.exe -keys client1-keys.dat

# Client 2
.\bin\zfscloud-client.exe -keys client2-keys.dat
```

Each will have a unique client ID and key pair.

## Clean Up

### Stop Services

```bash
docker-compose down
```

### Remove Volumes (Delete All Data)

```bash
docker-compose down -v
```

### Remove Images

```bash
docker rmi zfscloud-server zfscloud-admin
```

## Production Considerations

**DO NOT use this Docker setup for production!**

This is a testing/development environment. For production:

1. ✅ Use real ZFS pools on dedicated disks
2. ✅ Enable TLS/mTLS for gRPC
3. ✅ Use strong admin tokens (generate with `openssl rand -hex 32`)
4. ✅ Run on bare metal or VM with native ZFS support
5. ✅ Implement proper backup strategy
6. ✅ Configure firewall rules
7. ✅ Use proper monitoring and alerting
8. ✅ Set up log aggregation

## Next Steps

- [Testing Guide](../TESTING.md) - Comprehensive test scenarios
- [Server Quick Start](../SERVER_QUICKSTART.md) - Production deployment
- [Architecture](../STRUCTURE.md) - System design documentation
