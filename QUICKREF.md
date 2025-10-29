# ZFSCloud - Quick Reference

## Project Overview

**ZFSCloud** is a secure, encrypted backup system with:
- Client-server architecture using gRPC
- ZFS backend with native encryption
- Per-user encrypted datasets
- Ed25519 authentication
- Incremental deduplication
- Automated snapshots
- Admin approval workflow

## Project Status

### ✅ Phases 1-5 Complete: Functional Client-Server System

The following components have been implemented:

**Phase 1: Core Infrastructure**
- Professional directory layout
- Go modules configuration
- Build scripts (PowerShell + shell)
- Protocol Buffer definitions
- Cryptographic utilities (Ed25519, RSA, Argon2)
- Configuration management
- ZFS wrapper

**Phase 2: Database & gRPC Services**
- SQLite database with 8 tables
- 28 database operations
- BackupService: 14 RPC methods implemented
- AdminService: 10 RPC methods implemented
- Complete gRPC server implementation
- Request authentication and validation

**Phase 3: Server Integration**
- AdminService implementation (10 RPCs)
- File I/O storage layer
- Complete server wiring
- Session management
- Main server entry point with graceful shutdown

**Phase 4: Basic Client**
- Interactive CLI with command shell
- File upload with chunking (1MB chunks)
- Directory upload with recursion
- Ed25519 authentication
- SHA-256 hash verification
- 9 commands: register, heartbeat, upload, upload-dir, list, snapshots, status, help, exit

**Phase 5: Docker Testing Environment**
- ZFS-enabled Docker containers
- File-backed 10GB ZFS pool (no physical disks needed)
- Docker Compose orchestration
- Admin helper container with grpcurl
- Comprehensive testing guide
- Build scripts for cross-compilation

## File Inventory

### Core Documentation
```text
implementation-plan.md       Original architecture & design
README.md                   Project overview & quick start
SETUP.md                    Setup and installation guide
STRUCTURE.md               Component documentation
QUICKREF.md                This quick reference
SERVER_QUICKSTART.md       Server deployment guide
docker/README.md           Docker testing guide (350+ lines)
```

### Build & Deployment
```text
build.ps1                  Windows build script
Makefile                  Unix build script
build-docker.ps1          Docker server build (Linux)
build-docker.sh           Docker server build (shell)
test-docker.ps1           Docker test automation
go.mod                    Go module dependencies
.gitignore                Git ignore rules
.dockerignore             Docker build optimization
```

### Server Source Code (~6,000 lines)
```text
proto/backup.proto                          gRPC service definitions (24 RPCs)
proto/backup/backup.pb.go                   Generated protobuf code
proto/backup/backup_grpc.pb.go             Generated gRPC code

shared/crypto/keys.go                       Ed25519 cryptography
shared/crypto/passphrase.go                 Passphrase management
shared/config/server.go                     Server configuration
shared/config/client.go                     Client configuration

server/cmd/main.go                          Server entry point
server/zfs/manager.go                       ZFS operations wrapper
server/db/database.go                       SQLite database layer
server/db/operations.go                     28 database operations
server/grpc/server.go                       gRPC server setup
server/grpc/backup_service.go               BackupService (14 RPCs)
server/grpc/admin_service.go                AdminService (10 RPCs)
server/grpc/auth.go                         Request authentication
server/storage/fileio.go                    File I/O layer
```

### Client Source Code (~1,000 lines)
```text
client/cmd/main.go                          Interactive CLI (350 lines)
client/client.go                            gRPC client (300 lines)
client/uploader.go                          File upload engine (280 lines)
```

### Docker Environment
```text
docker/Dockerfile.server                    Server container (Ubuntu + ZFS)
docker/Dockerfile.admin                     Admin helper (Alpine + grpcurl)
docker/docker-compose.yml                   Multi-container orchestration
docker/config/server.yaml                   Docker server config
docker/scripts/list-registrations.sh        Admin helper scripts
docker/scripts/approve-registration.sh
docker/scripts/list-users.sh
docker/scripts/system-status.sh
```

### Configuration Templates
```text
server/config.example.yaml                  Server config template
client/config.example.yaml                  Client config template
docker/config/server.yaml                   Docker-specific config
```

## Quick Commands

### Build
```powershell
# Build everything
.\build.ps1 all

# Build individual components
.\build.ps1 server   # Server only
.\build.ps1 client   # Client only
.\build.ps1 proto    # Generate protobuf files

# Build for Docker (Linux)
.\build-docker.ps1   # Cross-compile server for Linux
```

### Docker Testing (Recommended)
```powershell
# Quick start
.\test-docker.ps1

# Or manually
.\build-docker.ps1
cd docker
docker-compose up -d
docker-compose logs -f zfscloud-server

# Admin operations
docker-compose --profile admin run zfscloud-admin list-registrations
docker-compose --profile admin run zfscloud-admin approve-registration <client-id> 10737418240
docker-compose --profile admin run zfscloud-admin list-users
```

### Run Native (Linux with ZFS required for server)
```bash
# Server (Linux only)
./bin/zfscloud-server -config server/config.yaml

# Client (Windows/Linux/Mac)
./bin/zfscloud-client -config client/config.yaml
```

### Client Commands (Interactive Mode)
```text
register              - Register with server
heartbeat            - Send heartbeat to maintain session
upload <file>        - Upload single file
upload-dir <dir>     - Upload entire directory
list                 - List backed up files
snapshots            - List available snapshots
status               - Show client status
help                 - Show help
exit                 - Exit client
```

### Test
```powershell
go test ./...                                   # Run unit tests
go test -coverprofile=coverage.out ./...       # With coverage
go tool cover -html=coverage.out              # View coverage
```

## Remaining Implementation Steps

### ✅ Phase 6: Backup Engine (COMPLETE)
- [x] File system watcher (fsnotify)
- [x] Automatic backup on file changes
- [x] Content-defined chunking for large files
- [x] Compression pipeline (zstd)
- [x] Background upload queue
- [x] Bandwidth throttling

### Phase 7: Client UI (Gio) - Next Priority
- [ ] System tray icon
- [ ] Status display (connected, backing up, idle)
- [ ] Folder selection interface
- [ ] Visual restore browser
- [ ] Configuration editor
- [ ] Progress indicators

### Phase 8: Security Hardening
- [ ] TLS certificate management
- [ ] mTLS implementation
- [ ] OS keychain integration (DPAPI/Keychain/Secret Service)
- [ ] Replay attack prevention improvements
- [ ] Rate limiting
- [ ] Audit logging

### Phase 9: Snapshot Automation
- [ ] Cron-based snapshot scheduler
- [ ] Retention policy enforcement
- [ ] Automatic cleanup
- [ ] Snapshot browsing improvements

### Phase 10: Production Features
- [ ] Web-based admin dashboard
- [ ] Monitoring and metrics (Prometheus)
- [ ] Multi-threaded uploads
- [ ] Delta sync optimizations
- [ ] Client-side encryption option
- [ ] Conflict resolution
- [ ] Comprehensive test suite

## API Overview

### BackupService RPCs (14 methods - all implemented ✅)
- `Register` - Client registration with public key
- `Heartbeat` - Keep session alive, auto-creates session
- `GetBackupPolicy` - Get backup configuration
- `UploadChunk` - Stream file data to server (chunked)
- `CompleteBackup` - Finalize backup with hash verification
- `ListFiles` - Browse backed up files
- `DownloadFile` - Restore files (streaming)
- `DeleteFile` - Remove backed up file
- `ListSnapshots` - View available snapshots
- `RequestSnapshotRestore` - Request snapshot restore (requires approval)
- `LoadDatasetKey` - Unlock ZFS dataset with passphrase
- `UnloadDatasetKey` - Lock ZFS dataset
- `GetQuotaInfo` - Check storage usage
- `GetSessionInfo` - Query session status

### AdminService RPCs (10 methods - all implemented ✅)
- `ListPendingRegistrations` - View pending client registrations
- `ApproveRegistration` - Approve client with quota allocation
- `RejectRegistration` - Reject client registration
- `ListUsers` - View all registered users
- `UpdateUserQuota` - Adjust user storage quota
- `ListRestoreRequests` - View pending snapshot restore requests
- `ApproveRestoreRequest` - Approve snapshot restore
- `RejectRestoreRequest` - Reject snapshot restore
- `UpdateSnapshotPolicy` - Configure automated snapshots
- `GetSystemStatus` - Monitor system health and storage

## Security Features

### Authentication
- Ed25519 public key cryptography
- Request signing with timestamp + nonce
- Replay attack prevention (5-minute window)

### Encryption
- **Transport**: TLS 1.3 with mTLS
- **At-Rest**: ZFS native encryption (AES-256-GCM)
- **Dataset Keys**: Client-generated, RSA-encrypted
- **Zero-Knowledge**: Optional client-side encryption

### Access Control
- Admin approval for registration
- Admin approval for snapshot restores
- Per-user dataset isolation
- Quota enforcement

## Architecture Highlights

### Client-Server Communication
```
Client → gRPC/mTLS → Server
         ↓
    Ed25519 Auth
         ↓
    Request Processing
         ↓
    ZFS Dataset (encrypted)
```

### Dataset Lifecycle
```
1. Client generates passphrase
2. Client encrypts with server RSA public key
3. Server creates ZFS dataset (encryption=on)
4. Client connects → server loads key
5. Backup operations
6. Client disconnects → server unloads key
```

### Backup Flow
```
1. Client watches folders
2. Calculate file hashes
3. Query server for existing hashes
4. Chunk new/modified files
5. Compress chunks
6. Stream to server
7. Server writes to ZFS dataset
```

## Dependencies

### Core (Implemented)
- `google.golang.org/grpc` - gRPC framework
- `google.golang.org/protobuf` - Protocol Buffers
- `golang.org/x/crypto` - Cryptography (Ed25519, Argon2)
- `gopkg.in/yaml.v3` - YAML configuration
- `modernc.org/sqlite` - Pure Go SQLite driver
- `github.com/mattn/go-sqlite3` - CGo SQLite driver (alternative)

### Planned (Phase 6+)
- `gioui.org` - Cross-platform UI
- `github.com/fsnotify/fsnotify` - File system watcher
- `github.com/klauspost/compress` - Compression (zstd)
- `github.com/prometheus/client_golang` - Metrics

## Resources

- **Documentation**: 
  - `implementation-plan.md` - Original architecture
  - `SETUP.md` - Installation guide
  - `STRUCTURE.md` - Component details
  - `SERVER_QUICKSTART.md` - Server deployment
  - `docker/README.md` - Docker testing guide (350+ lines)
- **Examples**: 
  - `server/config.example.yaml` - Server config template
  - `client/config.example.yaml` - Client config template
  - `docker/config/server.yaml` - Docker config
- **Source**: 
  - `client/` - Client application (1,000 lines)
  - `server/` - Server application (6,000 lines)
  - `shared/` - Shared libraries (crypto, config)
  - `proto/` - Protocol definitions + generated code
  - `docker/` - Docker testing environment

## Design Decisions

- **Go**: Performance, concurrency, excellent ZFS tooling
- **gRPC**: Efficient binary protocol, strongly-typed, streaming support
- **Ed25519**: Fast signatures, secure, small keys (32 bytes)
- **SQLite**: Zero-config database, perfect for single-server deployments
- **ZFS**: Built-in encryption, snapshots, deduplication, compression
- **Docker**: Easy testing without native ZFS installation
- **Gio** (planned): Native performance, cross-platform UI

## Testing Strategy

### Docker Testing (Current)
- File-backed ZFS pool (10GB) - no physical disks needed
- Privileged container for kernel module access
- Admin helper with grpcurl for API testing
- Comprehensive testing guide with step-by-step workflow

### Native Testing
- Requires Linux with ZFS installed
- Direct binary execution
- Useful for development and debugging

## Project Metrics

- **Total Lines of Code**: ~13,500 (including generated protobuf)
- **Server Code**: ~6,000 lines (database, gRPC, storage, ZFS)
- **Client Code**: ~1,000 lines (CLI, uploader, gRPC client)
- **Shared Libraries**: ~500 lines (crypto, config)
- **Generated Code**: ~5,500 lines (protobuf)
- **Documentation**: ~3,000+ lines across 7 markdown files
- **Build Scripts**: 5 scripts (Windows + Unix + Docker)
- **Docker Environment**: 8 files (Dockerfiles, compose, scripts, config)

## Current Capabilities

### Working Features ✅
- Full client-server communication via gRPC
- Client registration workflow (pending → approved)
- File upload with chunking (1MB chunks)
- Directory upload with recursion
- SHA-256 hash verification
- Session management with heartbeat
- Admin approval operations
- ZFS dataset management
- Docker testing environment
- Interactive CLI with 9 commands

### Limitations
- Manual backup (no automatic file watching yet)
- No GUI (CLI only)
- No TLS/mTLS (development only)
- No automatic snapshots (manual ZFS commands)
- No bandwidth throttling
- Basic error handling (needs improvement)

---

**Status**: Phase 5 Complete ✅ (Phases 1-5 of 10)  
**Next**: Phase 6 - Backup Engine (automatic file watching, compression)  
**Testing**: Docker environment ready (see `docker/README.md`)
