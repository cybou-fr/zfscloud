# ZFSCloud Setup Guide

## Quick Start (Docker - Recommended)

The fastest way to test ZFSCloud is using Docker, which provides a complete testing environment without requiring native ZFS installation.

### Prerequisites
- Go 1.23+
- Docker Desktop (Windows/Mac) or Docker Engine (Linux)
- Docker Compose

### Steps

1. **Clone and Build**
   ```powershell
   cd c:\Users\yosta\dev\zfscloud
   go mod download
   
   # Build server for Docker (Linux)
   .\build-docker.ps1
   
   # Build client for Windows
   .\build.ps1 client
   ```

2. **Start Docker Environment**
   ```powershell
   # Automated startup
   .\test-docker.ps1
   
   # Or manually
   cd docker
   docker-compose up -d
   docker-compose logs -f zfscloud-server
   ```

3. **Test Client-Server Communication**
   ```powershell
   # Run client
   .\bin\zfscloud-client.exe
   
   # In client shell:
   register
   
   # Approve registration (in another terminal)
   cd docker
   docker-compose --profile admin run zfscloud-admin list-registrations
   docker-compose --profile admin run zfscloud-admin approve-registration <client-id> 10737418240
   
   # Upload files
   upload <file-path>
   upload-dir <directory-path>
   
   # List files
   list
   ```

For detailed Docker testing guide, see **[docker/README.md](docker/README.md)** (350+ lines).

---

## Native Setup (Linux Server with ZFS)

If you want to run the server natively on Linux with ZFS (not in Docker):

## Step 1: Install Prerequisites

1. **Install Go 1.23+**
   - Download from: https://go.dev/dl/
   - Verify: `go version`

2. **Install Protocol Buffers Compiler**
   ```powershell
   # Using chocolatey
   choco install protoc
   
   # Or download from: https://github.com/protocolbuffers/protobuf/releases
   ```

3. **Install Go Protocol Buffer Plugins**
   ```powershell
   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
   go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
   ```

4. **Install Git**
   - Download from: https://git-scm.com/

### Linux Server Setup (with ZFS)

1. **Install Go 1.23+**
   ```bash
   # Ubuntu/Debian
   sudo apt update
   sudo apt install golang-1.23
   
   # Or download from go.dev
   ```

2. **Install ZFS**
   ```bash
   # Ubuntu/Debian
   sudo apt install zfsutils-linux
   
   # Verify
   zfs version
   ```

3. **Create ZFS Pool** (if not exists)
   ```bash
   # Example: Create pool on /dev/sdb
   sudo zpool create tank /dev/sdb
   
   # Create backup dataset
   sudo zfs create tank/backups
   ```

4. **Install Protocol Buffers**
   ```bash
   sudo apt install protobuf-compiler
   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
   go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
   ```

## Step 2: Clone and Build

```powershell
# Navigate to project directory
cd c:\Users\yosta\dev\zfscloud

# Download Go dependencies
go mod download

# Generate protobuf files
.\build.ps1 proto

# Build both client and server
.\build.ps1 all

# Or build individually
.\build.ps1 server
.\build.ps1 client
```

## Step 3: Generate Certificates (Development)

For development, you'll need self-signed certificates:

```powershell
# Create certs directory
mkdir certs

# Generate server RSA key (for passphrase encryption)
openssl genrsa -out certs/server-rsa.key 4096
openssl rsa -in certs/server-rsa.key -pubout -out certs/server-rsa-public.pem

# Generate TLS certificate for gRPC
openssl req -x509 -newkey rsa:4096 -keyout certs/server.key -out certs/server.crt -days 365 -nodes -subj "/CN=localhost"
```

## Step 4: Configure Server

```bash
# Copy example config
cp server/config.example.yaml server/config.yaml

# Edit configuration
nano server/config.yaml

# Update these fields:
# - server.listen_address (default: 0.0.0.0:8443)
# - zfs.pool (e.g., tank/backups)
# - security.admin_token (generate a secure random token)
```

### Generate Admin Token
```bash
# Linux
openssl rand -hex 32

# PowerShell
[System.Convert]::ToBase64String((1..32 | ForEach-Object { Get-Random -Minimum 0 -Maximum 256 }))
```

## Step 5: Configure Client

```powershell
# Copy example config
cp client/config.example.yaml client/config.yaml

# Edit configuration
notepad client/config.yaml

# Update these fields:
# - client.server_address (e.g., localhost:8443 for testing)
# - backup.folders (your folders to backup)
```

## Step 6: Initialize Database

The server will automatically create the SQLite database on first run. Make sure the data directory exists:

```bash
mkdir -p data
```

## Step 7: Run Server (Linux)

```bash
# Run server
./bin/zfscloud-server -config server/config.yaml

# Or with sudo if ZFS requires it
sudo ./bin/zfscloud-server -config server/config.yaml
```

## Step 8: Run Client (Windows/Linux)

```powershell
# Windows
.\bin\zfscloud-client.exe -config client\config.yaml

# Linux
./bin/zfscloud-client -config client/config.yaml
```

## Development Workflow

### Making Changes

1. **Modify code**
2. **Rebuild**: `.\build.ps1 all`
3. **Test changes**
4. **Commit to git**

### Updating Protocol Buffers

If you modify `proto/backup.proto`:

```powershell
# Regenerate Go files
.\build.ps1 proto

# Rebuild applications
.\build.ps1 all
```

### Running Tests

```powershell
# Run all tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Project Status

ZFSCloud has completed 5 of 10 planned phases:

- ✅ **Phase 1**: Core Infrastructure (proto, crypto, config, ZFS wrapper)
- ✅ **Phase 2**: Database & gRPC Services (SQLite, 24 RPCs)
- ✅ **Phase 3**: Server Integration (file I/O, admin service)
- ✅ **Phase 4**: Basic Client (interactive CLI, file upload)
- ✅ **Phase 5**: Docker Testing Environment (ZFS containers)
- ⏳ **Phase 6**: Backup Engine (automatic file watching)
- ⏳ **Phase 7**: Client UI (Gio-based GUI)
- ⏳ **Phase 8**: Security Hardening (TLS/mTLS)
- ⏳ **Phase 9**: Snapshot Automation
- ⏳ **Phase 10**: Production Features

**Current capabilities:**
- Full client-server communication via gRPC
- File upload with chunking and verification
- Directory upload with recursion
- Admin approval workflow
- Session management
- Docker testing environment

**Limitations:**
- Manual backup only (no automatic file watching yet)
- CLI only (no GUI)
- No TLS/mTLS (development mode only)
- No automatic snapshots (manual ZFS commands)

---

## Troubleshooting

### "protoc: command not found"

Install Protocol Buffers compiler (see Step 1).

### "cannot find package"

Run: `go mod download` and `go mod tidy`

### ZFS permission denied

Run server with sudo or configure ZFS delegation:
```bash
sudo zfs allow -u username create,mount,snapshot,destroy,load-key tank/backups
```

### TLS certificate errors

Regenerate certificates or disable certificate verification for development (not recommended for production).

## Security Notes for Production

⚠️ **Before deploying to production:**

1. Use proper CA-signed TLS certificates
2. Generate strong admin tokens
3. Enable firewall rules
4. Set up proper user permissions for ZFS
5. Configure backup retention policies
6. Set appropriate quotas
7. Enable audit logging
8. Review security settings in config files
9. Use strong passphrases for ZFS encryption
10. Implement monitoring and alerting

## Additional Resources

- [Implementation Plan](implementation-plan.md) - Detailed architecture
- [Project Structure](STRUCTURE.md) - Component overview
- [Protocol Buffer Docs](https://protobuf.dev/)
- [gRPC Go Tutorial](https://grpc.io/docs/languages/go/)
- [ZFS Documentation](https://openzfs.github.io/openzfs-docs/)
- [Gio UI Framework](https://gioui.org/)

## Getting Help

For issues or questions:

1. Check the implementation plan for design details
2. Review logs in `./data/` directory
3. Check ZFS status: `zpool status` and `zfs list`
4. Enable debug logging in config files

## License

MIT License - See LICENSE file
