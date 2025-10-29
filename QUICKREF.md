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

### ✅ Phase 1 Complete: Core Infrastructure

The following components have been implemented:

1. **Project Structure**
   - Professional directory layout
   - Go modules configuration
   - Build scripts (PowerShell + Makefile)
   - Documentation suite

2. **Protocol Definitions** (`proto/backup.proto`)
   - BackupService (14 RPC methods)
   - AdminService (10 RPC methods)
   - Complete message definitions
   - Registration, backup, restore, snapshot, and admin workflows

3. **Cryptographic Utilities** (`shared/crypto/`)
   - Ed25519 key pair generation and management
   - Request signing and verification
   - RSA encryption for dataset passphrases
   - Argon2 key derivation
   - Secure random generation

4. **Configuration Management** (`shared/config/`)
   - YAML-based configuration for server and client
   - Type-safe config structures
   - Sensible defaults
   - Example configurations

5. **ZFS Management** (`server/zfs/manager.go`)
   - Dataset creation with encryption
   - Key loading/unloading
   - Snapshot operations
   - Quota management
   - Space usage monitoring

6. **Application Entry Points**
   - Server: `server/cmd/main.go`
   - Client: `client/cmd/main.go`
   - Configuration loading
   - Graceful shutdown handling

## File Inventory

### Core Files
```
implementation-plan.md    Comprehensive architecture & design
README.md                 Project overview
SETUP.md                  Setup and installation guide
STRUCTURE.md             Component documentation
build.ps1                 Windows build script
Makefile                  Unix build script
go.mod                   Go module dependencies
.gitignore               Git ignore rules
```

### Source Code
```
proto/backup.proto                   gRPC service definitions
shared/crypto/keys.go                Ed25519 cryptography
shared/crypto/passphrase.go          Passphrase management
shared/config/server.go              Server configuration
shared/config/client.go              Client configuration
server/zfs/manager.go                ZFS operations wrapper
server/cmd/main.go                   Server entry point
client/cmd/main.go                   Client entry point
```

### Configuration
```
server/config.example.yaml           Server config template
client/config.example.yaml           Client config template
```

## Quick Commands

### Build
```powershell
.\build.ps1 all      # Build everything
.\build.ps1 server   # Build server only
.\build.ps1 client   # Build client only
.\build.ps1 proto    # Generate protobuf files
```

### Run
```powershell
# Server (Linux with ZFS)
./bin/zfscloud-server -config server/config.yaml

# Client
./bin/zfscloud-client -config client/config.yaml
```

### Test
```powershell
go test ./...                          # Run tests
go test -coverprofile=coverage.out ./...   # With coverage
```

## Next Implementation Steps

### Immediate (Phase 2)
1. **Generate Protobuf Files**
   ```powershell
   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
   go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
   .\build.ps1 proto
   ```

2. **Database Layer** (`server/db/`)
   - Schema design (SQLite)
   - User management
   - Registration tracking
   - Session management
   - Restore request workflow

3. **gRPC Services** (`server/grpc/`)
   - Implement BackupService
   - Implement AdminService
   - Authentication middleware
   - Request validation

### Near-term (Phase 3)
4. **Backup Engine** (`client/backup/`)
   - File system watcher
   - Content-defined chunking
   - SHA256 hashing
   - Compression (zstd)
   - Network transmission

5. **Client UI** (`client/ui/`)
   - Gio system tray integration
   - Status indicators
   - Configuration interface
   - File browser for restore

### Medium-term (Phase 4)
6. **Security Hardening**
   - TLS certificate management
   - mTLS implementation
   - OS keychain integration
   - Replay attack prevention

7. **Snapshot Scheduler** (`server/scheduler/`)
   - Cron-based scheduling
   - Retention policy enforcement
   - Cleanup automation

8. **Admin Interface** (`server/admin/`)
   - Web dashboard or CLI
   - User approval
   - Monitoring
   - Configuration

## API Overview

### BackupService RPCs
- `Register` - Client registration
- `GetBackupPolicy` - Get backup configuration
- `UploadChunk` - Stream file data to server
- `ListFiles` - Browse backed up files
- `DownloadFile` - Restore files
- `ListSnapshots` - View available snapshots
- `RequestSnapshotRestore` - Request snapshot restore
- `LoadDatasetKey` / `UnloadDatasetKey` - Session management
- `Heartbeat` - Keep session alive

### AdminService RPCs
- `ListPendingRegistrations` - View pending clients
- `ApproveRegistration` / `RejectRegistration` - Manage clients
- `ListUsers` - View all users
- `ApproveRestoreRequest` - Approve snapshot restores
- `UpdateSnapshotPolicy` - Configure snapshots
- `GetSystemStatus` - Monitor system health

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

### Core
- `google.golang.org/grpc` - gRPC framework
- `google.golang.org/protobuf` - Protocol Buffers
- `golang.org/x/crypto` - Cryptography (Ed25519, Argon2)
- `gopkg.in/yaml.v3` - YAML configuration

### Planned
- `gioui.org` - Cross-platform UI
- Database driver (SQLite or PostgreSQL)
- File system notification library

## Resources

- **Documentation**: `implementation-plan.md`, `SETUP.md`, `STRUCTURE.md`
- **Examples**: `server/config.example.yaml`, `client/config.example.yaml`
- **Source**: All source in `client/`, `server/`, `shared/`, `proto/`

## Development Notes

### Current Limitations
- Proto files need to be generated (requires protoc)
- Database layer not implemented
- gRPC services not implemented
- UI not implemented
- No tests yet

### Design Decisions
- **Go**: Performance, concurrency, ZFS tooling
- **gRPC**: Efficient, strongly-typed, streaming support
- **Ed25519**: Fast, secure, small keys
- **ZFS**: Built-in encryption, snapshots, deduplication
- **Gio**: Native performance, cross-platform

### Code Quality
- Type-safe configuration
- Error handling with wrapped errors
- Graceful shutdown patterns
- Logging ready (TODO: structured logging)

## Contribution Guidelines

When extending the codebase:

1. Follow existing patterns
2. Add error handling
3. Document public APIs
4. Write tests
5. Update relevant .md files
6. Use TODO comments for future work

## Project Metrics

- **Lines of Protocol Buffer**: ~450
- **Lines of Go Code**: ~800+
- **Documentation**: ~2000 lines
- **Configuration Examples**: Complete
- **Build Scripts**: Windows + Unix

---

**Status**: Phase 1 Complete ✅  
**Next**: Generate protobuf files, implement database layer  
**Timeline**: Core features targeting Phase 6 completion
