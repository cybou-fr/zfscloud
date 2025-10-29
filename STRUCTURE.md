# ZFSCloud Project Structure

# ZFSCloud Project Structure

## Directory Layout

```text
zfscloud/
├── client/                      # Client application
│   ├── cmd/                     # Client main entry point
│   │   └── main.go              # Interactive CLI (350 lines)
│   ├── client.go                # gRPC client implementation (300 lines)
│   ├── uploader.go              # File upload engine (280 lines)
│   ├── ui/                      # Gio UI components (TODO - Phase 7)
│   ├── backup/                  # Backup engine (TODO - Phase 6)
│   └── config.example.yaml      # Example client configuration
│
├── server/                      # Server application
│   ├── cmd/                     # Server main entry point
│   │   └── main.go              # Server startup with graceful shutdown
│   ├── zfs/                     # ZFS management
│   │   └── manager.go           # ZFS dataset operations
│   ├── grpc/                    # gRPC service implementations ✅
│   │   ├── server.go            # gRPC server setup
│   │   ├── backup_service.go    # BackupService (14 RPCs) ✅
│   │   ├── admin_service.go     # AdminService (10 RPCs) ✅
│   │   └── auth.go              # Request authentication
│   ├── db/                      # Database layer ✅
│   │   ├── database.go          # SQLite setup and schema
│   │   └── operations.go        # 28 database operations
│   ├── storage/                 # File I/O layer ✅
│   │   └── fileio.go            # Chunk storage and retrieval
│   └── config.example.yaml      # Example server configuration
│
├── proto/                       # Protocol Buffer definitions
│   ├── backup.proto             # Main service definition (24 RPCs)
│   └── backup/                  # Generated Go code
│       ├── backup.pb.go         # Protobuf messages
│       └── backup_grpc.pb.go    # gRPC service stubs
│
├── shared/                      # Shared libraries
│   ├── crypto/                  # Cryptographic utilities
│   │   ├── keys.go              # Ed25519 key management
│   │   └── passphrase.go        # Passphrase generation and encryption
│   └── config/                  # Configuration management
│       ├── server.go            # Server config
│       └── client.go            # Client config
│
├── docker/                      # Docker testing environment ✅
│   ├── Dockerfile.server        # Server container (Ubuntu + ZFS)
│   ├── Dockerfile.admin         # Admin helper (Alpine + grpcurl)
│   ├── docker-compose.yml       # Multi-container orchestration
│   ├── config/                  
│   │   └── server.yaml          # Docker server configuration
│   ├── scripts/                 # Admin helper scripts
│   │   ├── list-registrations.sh
│   │   ├── approve-registration.sh
│   │   ├── list-users.sh
│   │   └── system-status.sh
│   └── README.md                # Docker testing guide (350+ lines)
│
├── bin/                         # Compiled binaries (gitignored)
│   ├── zfscloud-server          # Server binary (Linux)
│   ├── zfscloud-client.exe      # Client binary (Windows)
│   └── ...
│
├── data/                        # Runtime data (gitignored)
│   └── zfscloud.db              # SQLite database
│
├── certs/                       # TLS certificates (gitignored)
│
├── go.mod                       # Go module definition
├── Makefile                     # Build automation (Linux/Mac)
├── build.ps1                    # Build script (Windows)
├── build-docker.ps1             # Docker build (PowerShell)
├── build-docker.sh              # Docker build (shell)
├── test-docker.ps1              # Docker test automation
├── .dockerignore                # Docker build optimization
├── README.md                    # Project documentation
├── QUICKREF.md                  # Quick reference guide
├── SETUP.md                     # Setup and installation
├── STRUCTURE.md                 # This file
├── SERVER_QUICKSTART.md         # Server deployment guide
├── implementation-plan.md       # Original implementation plan
└── .gitignore                   # Git ignore rules
```

## Component Status

### ✅ Completed (Phases 1-5)

**Phase 1: Core Infrastructure**
- [x] Project structure
- [x] Go modules initialization
- [x] Protocol Buffer definitions (24 gRPC methods)
- [x] Cryptographic utilities (Ed25519, passphrases)
- [x] Configuration management
- [x] ZFS manager wrapper
- [x] Basic server entry point
- [x] Basic client entry point
- [x] Example configurations
- [x] Build scripts

**Phase 2: Database & gRPC Services**
- [x] SQLite database with 8 tables
- [x] 28 database operations
- [x] BackupService implementation (14 RPCs)
- [x] AdminService implementation (10 RPCs)
- [x] Request authentication
- [x] Session management

**Phase 3: Server Integration**
- [x] File I/O storage layer
- [x] Complete server wiring
- [x] AdminService full implementation
- [x] Graceful shutdown
- [x] Main server entry point

**Phase 4: Basic Client**
- [x] Interactive CLI with command shell
- [x] gRPC client implementation
- [x] File upload with chunking (1MB chunks)
- [x] Directory upload with recursion
- [x] Ed25519 authentication
- [x] SHA-256 hash verification
- [x] 9 commands (register, heartbeat, upload, etc.)

**Phase 5: Docker Testing Environment**
- [x] ZFS-enabled Docker containers
- [x] File-backed 10GB ZFS pool
- [x] Docker Compose orchestration
- [x] Admin helper container with grpcurl
- [x] Build scripts for cross-compilation
- [x] Comprehensive testing guide (350+ lines)

### 🚧 TODO (Next Steps)

**Phase 6: Backup Engine**
- [ ] File system watcher (fsnotify)
- [ ] Automatic backup on file changes
- [ ] Content-defined chunking
- [ ] Compression pipeline (zstd)
- [ ] Background upload queue
- [ ] Bandwidth throttling

**Phase 7: Client UI**
- [ ] Gio system tray integration
- [ ] Status indicators
- [ ] Configuration interface
- [ ] File browser for restore
- [ ] Progress indicators

**Phase 8: Security Hardening**
- [ ] TLS certificate management
- [ ] mTLS implementation
- [ ] OS keychain integration
- [ ] Rate limiting
- [ ] Audit logging

**Phase 9: Snapshot Automation**
- [ ] Cron-based snapshot scheduler
- [ ] Retention policy enforcement
- [ ] Automatic cleanup

**Phase 10: Production Features**
- [ ] Web-based admin dashboard
- [ ] Monitoring and metrics
- [ ] Multi-threaded uploads
- [ ] Delta sync optimizations
- [ ] Comprehensive test suite

## Development Workflow

### Prerequisites
- Go 1.23+
- Protocol Buffers compiler (protoc)
- Docker Desktop (for testing)
- ZFS (for server, Linux only - or use Docker)

### Quick Start with Docker (Recommended)

```powershell
# Clone and navigate to project
cd c:\Users\yosta\dev\zfscloud

# Install dependencies
go mod download

# Build server for Docker (Linux)
.\build-docker.ps1

# Start Docker testing environment
.\test-docker.ps1

# Or manually
cd docker
docker-compose up -d

# Build and run client
.\build.ps1 client
.\bin\zfscloud-client.exe
```

### Native Development Setup

```powershell
# Install proto generation tools
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Download dependencies
go mod download

# Generate protobuf files (if proto changed)
.\build.ps1 proto

# Build everything
.\build.ps1 all

# Run server (Linux with ZFS)
./bin/zfscloud-server -config server/config.example.yaml

# Run client (any OS)
.\bin\zfscloud-client.exe -config client\config.example.yaml
```

### Testing Workflow

See `docker/README.md` for comprehensive testing guide:

1. Build server for Linux: `.\build-docker.ps1`
2. Start containers: `docker-compose up -d`
3. Run client on host: `.\bin\zfscloud-client.exe`
4. Register client: `register`
5. Approve registration: `docker-compose --profile admin run zfscloud-admin approve-registration <id> <quota>`
6. Upload files: `upload <file>` or `upload-dir <directory>`
7. List files: `list`
8. View snapshots: `snapshots`

## Current Implementation Details

### Server Architecture (~6,000 lines)

**Database Layer** (`server/db/`)
- SQLite with 8 tables: users, registrations, sessions, files, chunks, file_hashes, snapshots, restore_requests
- 28 operations: CreateUser, GetUser, ListUsers, CreateSession, GetActiveSession, etc.
- Handles registration workflow, session management, file metadata, restore requests

**gRPC Services** (`server/grpc/`)
- BackupService: 14 RPCs for client operations
  - Register, Heartbeat, GetBackupPolicy
  - UploadChunk, CompleteBackup, ListFiles, DownloadFile, DeleteFile
  - LoadDatasetKey, UnloadDatasetKey
  - ListSnapshots, RequestSnapshotRestore
  - GetQuotaInfo, GetSessionInfo
- AdminService: 10 RPCs for administration
  - ListPendingRegistrations, ApproveRegistration, RejectRegistration
  - ListUsers, UpdateUserQuota
  - ListRestoreRequests, ApproveRestoreRequest, RejectRestoreRequest
  - UpdateSnapshotPolicy, GetSystemStatus
- Authentication via Ed25519 signature verification on every request
- Session management with automatic cleanup

**Storage Layer** (`server/storage/`)
- File I/O operations for chunk storage
- Writes to ZFS datasets at `/tank/users/{userid}/chunks/`
- File metadata tracking
- SHA-256 hash verification

**ZFS Integration** (`server/zfs/`)
- Dataset creation with native encryption (AES-256-GCM)
- Key loading/unloading for session management
- Quota management and enforcement
- Snapshot operations (create, list, destroy)
- Space usage monitoring

### Client Architecture (~1,000 lines)

**Interactive CLI** (`client/cmd/main.go` - 350 lines)
- Command shell with 9 commands
- Ed25519 key auto-generation and management
- Metadata collection (hostname, username, OS)
- Commands: register, heartbeat, upload, upload-dir, list, snapshots, status, help, exit

**gRPC Client** (`client/client.go` - 300 lines)
- Connection management
- Request signing with Ed25519
- Methods for all BackupService RPCs
- Error handling and retries

**File Upload Engine** (`client/uploader.go` - 280 lines)
- Single file upload with 1MB chunking
- Directory upload with recursive walk
- SHA-256 hash calculation
- Progress tracking with percentage display
- Hash verification on completion

### Docker Testing Environment

**Server Container** (`docker/Dockerfile.server`)
- Ubuntu 22.04 with zfsutils-linux
- File-backed 10GB ZFS pool (no physical disks needed)
- Privileged mode for ZFS kernel module access
- Port 50051 exposed for gRPC

**Admin Container** (`docker/Dockerfile.admin`)
- Alpine Linux with grpcurl
- 4 helper scripts for common admin tasks
- Optional profile (--profile admin) for on-demand use

**Orchestration** (`docker/docker-compose.yml`)
- Multi-container setup with shared network
- Persistent volume for database and ZFS data
- Healthcheck for ZFS pool status
- Automated startup with test-docker.ps1

### Configuration

**Server Config** (`server/config.example.yaml`)
```yaml
server:
  listen_address: "0.0.0.0:50051"
  data_dir: "./data"

zfs:
  pool: "tank"
  default_quota: 10737418240  # 10GB

security:
  admin_token: "change-me-in-production"
  session_timeout: "24h"
```

**Client Config** (`client/config.example.yaml`)
```yaml
client:
  server_address: "localhost:50051"
  keys_dir: "./data/keys"

backup:
  chunk_size: 1048576  # 1MB
  folders: []
```
