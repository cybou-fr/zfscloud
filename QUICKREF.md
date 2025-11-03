# ZFSCloud - Quick Reference# ZFSCloud - Quick Reference



## Overview## What is ZFSCloud?



Secure backup system with ZFS backend and user-controlled encryption.Secure, encrypted backup system with:



**Key Features:**- Client-server architecture using gRPC

- ZFS backend with native encryption (AES-256-GCM)

- Client-server gRPC architecture- Per-user encrypted datasets

- Per-user encrypted ZFS datasets (AES-256-GCM)- Ed25519 + RSA authentication

- Ed25519 authentication + RSA passphrase encryption- TLS/mTLS transport security

- TLS/mTLS transport security- Automated snapshots with retention policies

- Automated snapshots with 4-tier retention- GUI and CLI clients

- GUI client (Gio) + CLI client

## Quick Commands

## Quick Start

### Build

### Build

```bash

```bash# Build everything

# Build all componentsgo build -o bin/zfscloud-server ./server/cmd

go build -o bin/zfscloud-server ./server/cmdgo build -o bin/zfscloud-client ./client/cmd

go build -o bin/zfscloud-client ./client/cmd

# Or use PowerShell script

# Or use build script.\build.ps1 all

.\build.ps1 all```

```

### Run Server

### Run

```bash

```bash./bin/zfscloud-server -config server/config.yaml

# Server (Linux with ZFS required)```

./bin/zfscloud-server -config server/config.yaml

### Run GUI Client

# GUI Client

./bin/zfscloud-client```bash

./bin/zfscloud-client

# CLI Client# Automatically loads client-config.yaml

./bin/zfscloud-client -cli```

```

### Run CLI Client

### Docker Testing

```bash

```bash./bin/zfscloud-client -cli

# Build and start```

.\build-docker.ps1

cd docker**Phase 5: Docker Testing Environment**

docker-compose up -d- ZFS-enabled Docker containers

- File-backed 10GB ZFS pool (no physical disks needed)

# Run client- Docker Compose orchestration

.\bin\zfscloud-client.exe- Admin helper container with grpcurl

```- Comprehensive testing guide

- Build scripts for cross-compilation

## CLI Commands

## File Inventory

### Registration & Authentication

### Core Documentation

```bash```text

register                    # Register new client (admin approval required)implementation-plan.md       Original architecture & design

status                      # Check registration statusREADME.md                   Project overview & quick start

heartbeat                   # Send keepaliveSETUP.md                    Setup and installation guide

```STRUCTURE.md               Component documentation

QUICKREF.md                This quick reference

### Backup OperationsSERVER_QUICKSTART.md       Server deployment guide

docker/README.md           Docker testing guide (350+ lines)

```bash```

upload <file>               # Upload single file

upload-dir <dir>            # Upload directory recursively### Build & Deployment

list [path]                 # List backed up files```text

download <file> <dest>      # Download filebuild.ps1                  Windows build script

```Makefile                  Unix build script

build-docker.ps1          Docker server build (Linux)

### Snapshotsbuild-docker.sh           Docker server build (shell)

test-docker.ps1           Docker test automation

```bashgo.mod                    Go module dependencies

snapshots                   # List available snapshots.gitignore                Git ignore rules

restore-request <snapshot>  # Request snapshot restore (admin approval required).dockerignore             Docker build optimization

``````



### Admin Commands### Server Source Code (~6,000 lines)

```text

```bashproto/backup.proto                          gRPC service definitions (24 RPCs)

admin list-pending          # List pending registrationsproto/backup/backup.pb.go                   Generated protobuf code

admin approve <id>          # Approve registrationproto/backup/backup_grpc.pb.go             Generated gRPC code

admin reject <id> <reason>  # Reject registration

admin list-users            # List all usersshared/crypto/keys.go                       Ed25519 cryptography

admin revoke <id> <reason>  # Revoke user accessshared/crypto/passphrase.go                 Passphrase management

```shared/config/server.go                     Server configuration

shared/config/client.go                     Client configuration

## Configuration

server/cmd/main.go                          Server entry point

### Server Config (server/config.yaml)server/zfs/manager.go                       ZFS operations wrapper

server/db/database.go                       SQLite database layer

```yamlserver/db/operations.go                     28 database operations

server:server/grpc/server.go                       gRPC server setup

  listen_address: "0.0.0.0:50051"server/grpc/backup_service.go               BackupService (14 RPCs)

  admin_token: "your-admin-token"server/grpc/admin_service.go                AdminService (10 RPCs)

  rsa_key_path: "server-rsa-key.pem"  # Auto-generated if missingserver/grpc/auth.go                         Request authentication

  tls:server/storage/fileio.go                    File I/O layer

    enabled: true```

    cert_file: "certs/server-cert.pem"

    key_file: "certs/server-key.pem"### Client Source Code (~1,000 lines)

    ca_file: "certs/ca-cert.pem"```text

client/cmd/main.go                          Interactive CLI (350 lines)

database:client/client.go                            gRPC client (300 lines)

  path: "/var/lib/zfscloud/zfscloud.db"client/uploader.go                          File upload engine (280 lines)

```

zfs:

  pool: "tank"### Docker Environment

  compression: "lz4"```text

docker/Dockerfile.server                    Server container (Ubuntu + ZFS)

snapshots:docker/Dockerfile.admin                     Admin helper (Alpine + grpcurl)

  enabled: truedocker/docker-compose.yml                   Multi-container orchestration

  schedule: "0 * * * *"  # Every hourdocker/config/server.yaml                   Docker server config

  retention:docker/scripts/list-registrations.sh        Admin helper scripts

    hourly: 24docker/scripts/approve-registration.sh

    daily: 7docker/scripts/list-users.sh

    weekly: 4docker/scripts/system-status.sh

    monthly: 12```

```

### Configuration Templates

### Client Config (client-config.yaml)```text

server/config.example.yaml                  Server config template

```yamlclient/config.example.yaml                  Client config template

client:docker/config/server.yaml                   Docker-specific config

  server_address: "localhost:50051"```

  client_id: "auto-generated-uuid"

  metadata:## Quick Commands

    hostname: "my-computer"

    username: "user"### Build

```powershell

  tls:# Build everything

    enabled: true.\build.ps1 all

    cert_file: "certs/client-cert.pem"

    key_file: "certs/client-key.pem"# Build individual components

    ca_file: "certs/ca-cert.pem".\build.ps1 server   # Server only

.\build.ps1 client   # Client only

backup:.\build.ps1 proto    # Generate protobuf files

  folders:

    - "C:\\Users\\user\\Documents"# Build for Docker (Linux)

  check_interval_seconds: 300.\build-docker.ps1   # Cross-compile server for Linux

  max_workers: 4```



bandwidth:### Docker Testing (Recommended)

  limit_mbps: 0  # 0 = unlimited```powershell

```# Quick start

.\test-docker.ps1

## Project Structure

# Or manually

```text.\build-docker.ps1

zfscloud/cd docker

├── server/              # Server applicationdocker-compose up -d

│   ├── cmd/             # Server entry pointdocker-compose logs -f zfscloud-server

│   ├── db/              # Database layer (SQLite)

│   ├── grpc/            # gRPC service implementations# Admin operations

│   ├── storage/         # File I/O operationsdocker-compose --profile admin run zfscloud-admin list-registrations

│   ├── snapshot/        # Snapshot schedulerdocker-compose --profile admin run zfscloud-admin approve-registration <client-id> 10737418240

│   ├── security/        # TLS/crypto utilitiesdocker-compose --profile admin run zfscloud-admin list-users

│   └── zfs/             # ZFS command wrapper```

├── client/              # Client application

│   ├── cmd/             # Client entry point### Run Native (Linux with ZFS required for server)

│   ├── gui/             # Gio UI implementation```bash

│   └── uploader.go      # Backup engine# Server (Linux only)

├── proto/               # gRPC protocol definitions./bin/zfscloud-server -config server/config.yaml

├── shared/              # Shared libraries

│   ├── crypto/          # Ed25519, RSA, Argon2# Client (Windows/Linux/Mac)

│   └── config/          # YAML config management./bin/zfscloud-client -config client/config.yaml

└── docker/              # Docker deployment```

```

### Client Commands (Interactive Mode)

## Architecture```text

register              - Register with server

```textheartbeat            - Send heartbeat to maintain session

Client (GUI/CLI)upload <file>        - Upload single file

  │ Generate 256-bit passphraseupload-dir <dir>     - Upload entire directory

  │ Encrypt with server RSA public keylist                 - List backed up files

  │ Ed25519 signature authenticationsnapshots            - List available snapshots

  ↓ gRPC over mTLS (TLS 1.3)status               - Show client status

Serverhelp                 - Show help

  │ Decrypt passphrase with RSA private keyexit                 - Exit client

  │ Verify Ed25519 signature```

  ↓ ZFS commands

ZFS Pool### Test

  └─ tank/user-<uuid>  (encrypted with user passphrase)```powershell

     ├─ Snapshots (hourly/daily/weekly/monthly)go test ./...                                   # Run unit tests

     └─ Backup filesgo test -coverprofile=coverage.out ./...       # With coverage

```go tool cover -html=coverage.out              # View coverage

```

## Security Model

## Remaining Implementation Steps

### Authentication

### ✅ Phase 6: Backup Engine (COMPLETE)

- **Ed25519 signatures**: All client requests signed with private key- [x] File system watcher (fsnotify)

- **RSA encryption**: User passphrases encrypted with server public key (4096-bit)- [x] Automatic backup on file changes

- **Admin tokens**: Admin operations require secret token- [x] Content-defined chunking for large files

- [x] Compression pipeline (zstd)

### Encryption Layers- [x] Background upload queue

- [x] Bandwidth throttling

1. **Transport**: TLS 1.3 with mTLS (optional)

2. **Dataset**: ZFS native encryption (AES-256-GCM)### Phase 7: Client UI (Gio) - Next Priority

3. **Passphrase**: User-controlled, stored encrypted- [ ] System tray icon

- [ ] Status display (connected, backing up, idle)

### User-Controlled Encryption- [ ] Folder selection interface

- [ ] Visual restore browser

- Client generates random 256-bit passphrase- [ ] Configuration editor

- Client encrypts passphrase with server RSA public key- [ ] Progress indicators

- Server decrypts only when creating dataset or loading key

- Server never stores plaintext passphrase### Phase 8: Security Hardening

- Client stores passphrase in OS keychain- [ ] TLS certificate management

- [ ] mTLS implementation

**Status**: Server-side complete ✅, client-side pending ⏳- [ ] OS keychain integration (DPAPI/Keychain/Secret Service)

- [ ] Replay attack prevention improvements

See [SECURITY_FIX_PROGRESS.md](./SECURITY_FIX_PROGRESS.md) for details.- [ ] Rate limiting

- [ ] Audit logging

## Common Tasks

### Phase 9: Snapshot Automation

### Generate Certificates- [ ] Cron-based snapshot scheduler

- [ ] Retention policy enforcement

```bash- [ ] Automatic cleanup

# Auto-generate (server does this on first run if tls.auto_generate = true)- [ ] Snapshot browsing improvements

# Or manually:

cd certs### Phase 10: Production Features

openssl genrsa -out ca-key.pem 4096- [ ] Web-based admin dashboard

openssl req -new -x509 -days 3650 -key ca-key.pem -out ca-cert.pem- [ ] Monitoring and metrics (Prometheus)

openssl genrsa -out server-key.pem 4096- [ ] Multi-threaded uploads

openssl req -new -key server-key.pem -out server-req.pem- [ ] Delta sync optimizations

openssl x509 -req -in server-req.pem -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial -out server-cert.pem -days 365- [ ] Client-side encryption option

```- [ ] Conflict resolution

- [ ] Comprehensive test suite

### Check ZFS Encryption

## API Overview

```bash

# List datasets### BackupService RPCs (14 methods - all implemented ✅)

zfs list -r tank- `Register` - Client registration with public key

- `Heartbeat` - Keep session alive, auto-creates session

# Check encryption status- `GetBackupPolicy` - Get backup configuration

zfs get encryption,keylocation,keystatus tank/user-<uuid>- `UploadChunk` - Stream file data to server (chunked)

- `CompleteBackup` - Finalize backup with hash verification

# Should show:- `ListFiles` - Browse backed up files

# encryption: aes-256-gcm- `DownloadFile` - Restore files (streaming)

# keylocation: prompt- `DeleteFile` - Remove backed up file

# keystatus: available (if loaded) or unavailable (if unloaded)- `ListSnapshots` - View available snapshots

```- `RequestSnapshotRestore` - Request snapshot restore (requires approval)

- `LoadDatasetKey` - Unlock ZFS dataset with passphrase

### View Snapshots- `UnloadDatasetKey` - Lock ZFS dataset

- `GetQuotaInfo` - Check storage usage

```bash- `GetSessionInfo` - Query session status

zfs list -t snapshot -r tank/user-<uuid>

```### AdminService RPCs (10 methods - all implemented ✅)

- `ListPendingRegistrations` - View pending client registrations

### Database Inspection- `ApproveRegistration` - Approve client with quota allocation

- `RejectRegistration` - Reject client registration

```bash- `ListUsers` - View all registered users

sqlite3 /var/lib/zfscloud/zfscloud.db- `UpdateUserQuota` - Adjust user storage quota

- `ListRestoreRequests` - View pending snapshot restore requests

# Useful queries- `ApproveRestoreRequest` - Approve snapshot restore

SELECT * FROM users;- `RejectRestoreRequest` - Reject snapshot restore

SELECT * FROM registrations WHERE status = 'pending';- `UpdateSnapshotPolicy` - Configure automated snapshots

SELECT * FROM snapshots ORDER BY created_at DESC LIMIT 10;- `GetSystemStatus` - Monitor system health and storage

```

## Security Features

## Troubleshooting

### Authentication

### Client can't connect- Ed25519 public key cryptography

- Request signing with timestamp + nonce

1. Check server is running: `systemctl status zfscloud-server`- Replay attack prevention (5-minute window)

2. Check firewall: `sudo ufw allow 50051/tcp`

3. Verify TLS certificates are valid### Encryption

4. Check client config has correct server address- **Transport**: TLS 1.3 with mTLS

- **At-Rest**: ZFS native encryption (AES-256-GCM)

### Registration stuck pending- **Dataset Keys**: Client-generated, RSA-encrypted

- **Zero-Knowledge**: Optional client-side encryption

1. Check admin approvals: `admin list-pending`

2. Approve: `admin approve <registration-id>`### Access Control

3. Check server logs for errors- Admin approval for registration

- Admin approval for snapshot restores

### Backup not starting- Per-user dataset isolation

- Quota enforcement

1. Verify registration is approved: `status`

2. Check watch folders exist and are readable## Architecture Highlights

3. View logs in GUI error panel

4. Test connection: Settings → Test Connection### Client-Server Communication

```

### Dataset key won't loadClient → gRPC/mTLS → Server

         ↓

1. Check ZFS key status: `zfs get keystatus tank/user-<uuid>`    Ed25519 Auth

2. Verify encrypted passphrase in database         ↓

3. Check server RSA private key exists and is readable    Request Processing

4. See server logs for decryption errors         ↓

    ZFS Dataset (encrypted)

## Documentation```



- **[README.md](./README.md)** - Project overview### Dataset Lifecycle

- **[SETUP.md](./SETUP.md)** - Detailed setup guide```

- **[GUI_USER_GUIDE.md](./GUI_USER_GUIDE.md)** - GUI manual1. Client generates passphrase

- **[SECURITY_FIX_PROGRESS.md](./SECURITY_FIX_PROGRESS.md)** - Encryption implementation status2. Client encrypts with server RSA public key

- **[implementation-plan.md](./implementation-plan.md)** - Complete architecture3. Server creates ZFS dataset (encryption=on)

- **[DOCKER_TESTING_GUIDE.md](./DOCKER_TESTING_GUIDE.md)** - Docker deployment4. Client connects → server loads key

5. Backup operations

## License6. Client disconnects → server unloads key

```

MIT License

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
