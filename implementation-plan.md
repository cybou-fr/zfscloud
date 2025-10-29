# ZFSCloud - Encrypted Backup System Implementation Plan

## Overview
A client-server backup system built in Go, featuring a Gio UI-based client (system tray application) and a ZFS-backed server for secure, encrypted, per-user backups.

## Current Status

**Completed: Phases 1-5 of 10** ✅

- ✅ **Phase 1**: Core Infrastructure (proto, crypto, config, ZFS wrapper)
- ✅ **Phase 2**: Database & gRPC Services (SQLite with 8 tables, 24 RPCs)
- ✅ **Phase 3**: Server Integration (file I/O, admin service, complete server)
- ✅ **Phase 4**: Basic Client (interactive CLI, file upload with chunking)
- ✅ **Phase 5**: Docker Testing Environment (ZFS-enabled containers)

**In Progress: Phase 6** 🚧

- ⏳ Automatic backup engine (file watching, compression)

**Planned: Phases 7-10** 📋

- Phase 7: Gio-based UI (system tray, visual interface)
- Phase 8: Security hardening (TLS/mTLS, keychain integration)
- Phase 9: Snapshot automation (cron scheduler, retention policies)
- Phase 10: Production features (monitoring, optimization, testing)

## Working Features

### Server (Fully Functional)
- ✅ 24 gRPC methods (BackupService: 14, AdminService: 10)
- ✅ SQLite database with 8 tables
- ✅ File chunking and storage
- ✅ ZFS dataset management with encryption
- ✅ Session management with automatic cleanup
- ✅ Admin approval workflow
- ✅ Quota enforcement
- ✅ Ed25519 request authentication

### Client (Basic Functionality)
- ✅ Interactive CLI with 9 commands
- ✅ File upload with 1MB chunking
- ✅ Directory upload with recursion
- ✅ SHA-256 hash verification
- ✅ Ed25519 key generation and signing
- ✅ Progress tracking

### Docker Environment
- ✅ File-backed 10GB ZFS pool (no physical disks needed)
- ✅ Multi-container orchestration
- ✅ Admin helper with grpcurl
- ✅ Comprehensive testing guide (350+ lines)

See **[docker/README.md](docker/README.md)** for complete testing workflow.

---

### Technology Stack
- **Language**: Go (Golang)
- **Client UI**: Gio (cross-platform GUI framework)
- **Server Storage**: ZFS with native encryption
- **Communication**: gRPC with mutual TLS (mTLS)
- **Authentication**: Public key cryptography (Ed25519 or ECDSA)

## Core Components

### 1. Client Application

#### UI/UX
- System tray icon with popup menu
- Status indicators (connected, backing up, idle)
- Configuration interface for:
  - Folder selection for backup
  - Backup schedule preferences
  - Server connection settings
  - Restore operations

#### Client Responsibilities
- **Registration & Authentication**
  - Generate Ed25519 key pair on first run
  - Register with server using public key
  - Securely store private key in OS keychain
  - Sign all requests with private key

- **Dataset Encryption**
  - Generate strong encryption passphrase (256-bit) for ZFS dataset
  - Derive key using PBKDF2 or Argon2
  - Send encrypted passphrase to server (encrypted with server's public key)
  - Load passphrase to ZFS on connection, unload on disconnect

- **Backup Operations**
  - Monitor specified folders for changes (file system watcher)
  - Query server for backup interval and policy
  - Perform incremental backups using content-defined chunking
  - Calculate file hashes (SHA256) for deduplication
  - Compress data before transmission (zstd or lz4)
  - Handle bandwidth throttling

- **Restore Operations**
  - Browse current dataset files
  - List available snapshots
  - Restore files/folders from current dataset
  - Request snapshot restore (requires admin approval)

### 2. Server Application

#### Server Responsibilities
- **User Management**
  - Accept client registrations (pending approval)
  - Admin interface for approving/rejecting clients
  - Maintain user database with public keys
  - Revoke client access when needed

- **ZFS Dataset Management**
  - Create isolated, encrypted ZFS dataset per user
  - Use native ZFS encryption (encryption=aes-256-gcm)
  - Store encrypted passphrases securely (encrypted at rest)
  - Load/unload dataset keys on client connect/disconnect
  - Set quotas per user to prevent storage abuse

- **Backup Storage**
  - Accept and store client backup data
  - Leverage ZFS deduplication at block level
  - Maintain file hash index for client-side deduplication
  - Compress datasets (lz4 or zstd compression)

- **Snapshot Management**
  - Create automated snapshots per admin-defined schedule
  - Retention policy configuration (e.g., hourly, daily, weekly, monthly)
  - Snapshot naming convention: `user-{userid}-{timestamp}`
  - Admin approval workflow for snapshot restores

- **Admin Interface**
  - Web dashboard or CLI tool
  - Approve/reject pending registrations
  - Configure backup schedules
  - Set snapshot retention policies
  - Monitor storage usage and client activity
  - Manage user quotas

## Security Architecture

### 1. Authentication & Authorization
- **Client Registration Flow**:
  1. Client generates Ed25519 key pair
  2. Client sends registration request with public key + client metadata
  3. Server stores request in "pending" state
  4. Admin reviews and approves/rejects
  5. On approval, server creates ZFS dataset and sends confirmation

- **Request Authentication**:
  - Every client request signed with private key
  - Server validates signature using stored public key
  - Implement replay attack prevention (nonce + timestamp)

### 2. Data Encryption
- **Transport Layer**: 
  - Mutual TLS (mTLS) for all client-server communication
  - TLS 1.3 with strong cipher suites
  - Certificate pinning on client side

- **At-Rest Encryption**:
  - ZFS native encryption (per-user dataset)
  - Client-generated passphrase (never stored in plaintext on server)
  - Server stores passphrase encrypted with server master key
  - Dataset unlocked only when client is connected

- **End-to-End Encryption**:
  - Optional: Client can encrypt files before transmission
  - Ensures zero-knowledge architecture

### 3. Key Management
- Client private key: OS keychain (Windows: DPAPI, macOS: Keychain, Linux: Secret Service)
- Server master key: Hardware Security Module (HSM) or secure key management service
- ZFS dataset passphrases: Encrypted with server master key
- Automatic key rotation policy

## Backup Process Flow

### Initial Backup
1. Client scans selected folders and builds file inventory
2. Client calculates file hashes (SHA256)
3. Client requests hash list from server
4. Client determines delta (new/modified files)
5. Client chunks large files using content-defined chunking
6. Client compresses and transmits chunks
7. Server stores data in user's ZFS dataset
8. Server updates hash index

### Incremental Backup
1. Client monitors folders with file system watcher
2. On change detection or scheduled interval:
   - Calculate hashes for changed files
   - Query server for existing hashes
   - Transmit only new/modified chunks
3. Server merges changes into dataset

### Deduplication Strategy
- **Client-side**: Hash comparison to avoid transmitting duplicate files
- **Server-side**: ZFS block-level deduplication (optional, consider RAM requirements)
- **Chunk-level**: Content-defined chunking for efficient delta sync

## Snapshot & Restore Flow

### Snapshot Creation (Server-side)
- Automated via cron/scheduler based on admin policy
- ZFS snapshot command: `zfs snapshot pool/users/{userid}@{timestamp}`
- Retention policy enforcement (automatic cleanup)

### Current Data Restore
1. Client browses current dataset
2. Client requests file/folder restore
3. Server validates permissions
4. Server sends requested data
5. Client restores to original or alternate location

### Snapshot Restore
1. Client lists available snapshots
2. Client selects snapshot and requests restore
3. Server creates approval request for admin
4. Admin reviews and approves/rejects
5. On approval:
   - Server clones snapshot or exports data
   - Client receives and restores data

## Configuration & Policies

### Server Configuration
```yaml
server:
  listen_address: "0.0.0.0:8443"
  tls_cert: "/path/to/cert.pem"
  tls_key: "/path/to/key.pem"
  
zfs:
  pool: "tank/backups"
  compression: "zstd"
  deduplication: false  # Consider RAM requirements
  
quotas:
  default_quota: "100GB"
  max_quota: "1TB"
  
snapshots:
  schedule:
    hourly: 24    # Keep 24 hourly
    daily: 7      # Keep 7 daily
    weekly: 4     # Keep 4 weekly
    monthly: 12   # Keep 12 monthly
    
security:
  require_admin_approval: true
  snapshot_restore_requires_approval: true
  session_timeout: "1h"
```

### Client Configuration
```yaml
client:
  server_address: "backup.example.com:8443"
  server_cert_fingerprint: "sha256:..."
  
backup:
  folders:
    - "/home/user/Documents"
    - "/home/user/Pictures"
  
  exclude_patterns:
    - "*.tmp"
    - "*.cache"
    - "node_modules/"
  
  schedule:
    check_interval: "5m"
    backup_on_change: true
  
  bandwidth:
    limit_mbps: 10
    throttle_on_metered: true
```

## Implementation Phases

### ✅ Phase 1: Core Infrastructure (COMPLETE)

**Status**: All tasks complete

- [x] gRPC service definitions (proto/backup.proto)
  - BackupService: 14 RPC methods
  - AdminService: 10 RPC methods
  - Complete message definitions
- [x] mTLS setup and certificate management (basic structure)
- [x] Client key pair generation and storage (Ed25519)
- [x] Server registration system (basic implementation)
- [x] ZFS dataset creation wrapper
- [x] Configuration management (YAML-based)
- [x] Build scripts (Windows + Unix)

**Deliverables**:
- `proto/backup.proto` - Protocol definitions
- `shared/crypto/` - Cryptographic utilities
- `shared/config/` - Configuration management
- `server/zfs/manager.go` - ZFS wrapper
- `build.ps1`, `Makefile` - Build automation

### ✅ Phase 2: Database & gRPC Services (COMPLETE)

**Status**: All tasks complete

- [x] SQLite database schema (8 tables)
  - users, registrations, sessions, files, chunks, file_hashes, snapshots, restore_requests
- [x] Database operations (28 operations)
  - User management, registration workflow, session handling, file metadata
- [x] BackupService implementation (14 RPCs)
  - Register, Heartbeat, GetBackupPolicy, UploadChunk, CompleteBackup
  - ListFiles, DownloadFile, DeleteFile
  - LoadDatasetKey, UnloadDatasetKey
  - ListSnapshots, RequestSnapshotRestore
  - GetQuotaInfo, GetSessionInfo
- [x] AdminService implementation (10 RPCs)
  - ListPendingRegistrations, ApproveRegistration, RejectRegistration
  - ListUsers, UpdateUserQuota
  - ListRestoreRequests, ApproveRestoreRequest, RejectRestoreRequest
  - UpdateSnapshotPolicy, GetSystemStatus
- [x] Request authentication (Ed25519 signature verification)
- [x] Session management

**Deliverables**:
- `server/db/database.go` - Database schema and setup
- `server/db/operations.go` - 28 database operations
- `server/grpc/backup_service.go` - BackupService implementation
- `server/grpc/admin_service.go` - AdminService implementation
- `server/grpc/auth.go` - Authentication middleware

### ✅ Phase 3: Server Integration (COMPLETE)

**Status**: All tasks complete

- [x] File I/O storage layer
  - Chunk writing to ZFS datasets
  - File metadata tracking
  - Hash verification
- [x] Complete server wiring
  - gRPC server setup
  - Service registration
  - Graceful shutdown
- [x] AdminService full implementation
- [x] Main server entry point

**Deliverables**:
- `server/storage/fileio.go` - File I/O operations
- `server/grpc/server.go` - gRPC server setup
- `server/cmd/main.go` - Complete server entry point

### ✅ Phase 4: Basic Client (COMPLETE)

**Status**: All tasks complete

- [x] Interactive CLI application
  - Command shell with 9 commands
  - Ed25519 key auto-generation
  - Metadata collection (hostname, username, OS)
- [x] gRPC client implementation
  - Connection management
  - Request signing
  - All BackupService RPCs
- [x] File upload engine
  - Chunking (1MB default)
  - Directory recursion
  - SHA-256 hashing
  - Progress tracking

**Deliverables**:
- `client/cmd/main.go` - Interactive CLI (350 lines)
- `client/client.go` - gRPC client (300 lines)
- `client/uploader.go` - File upload engine (280 lines)

### ✅ Phase 5: Docker Testing Environment (COMPLETE)

**Status**: All tasks complete

- [x] Server Docker container (Ubuntu + ZFS)
- [x] File-backed ZFS pool (10GB, no physical disks)
- [x] Admin helper container (Alpine + grpcurl)
- [x] Docker Compose orchestration
- [x] Build scripts for cross-compilation
- [x] Comprehensive testing guide

**Deliverables**:
- `docker/Dockerfile.server` - ZFS-enabled server container
- `docker/Dockerfile.admin` - Admin helper with grpcurl
- `docker/docker-compose.yml` - Multi-container setup
- `docker/README.md` - Testing guide (350+ lines)
- `build-docker.ps1`, `test-docker.ps1` - Build/test automation

### ⏳ Phase 6: Backup Engine (IN PROGRESS)

**Status**: Not started

- [ ] File system watcher (fsnotify)
- [ ] Automatic backup on file changes
- [ ] Content-defined chunking for large files
- [ ] Compression pipeline (zstd)
- [ ] Background upload queue
- [ ] Bandwidth throttling
- [ ] Delta sync optimization

### ⏳ Phase 7: Client UI (PLANNED)

**Status**: Not started

- [ ] Gio system tray integration
- [ ] Status indicators (connected, backing up, idle)
- [ ] Folder selection interface
- [ ] Visual restore browser
- [ ] Configuration editor
- [ ] Progress indicators
- [ ] Notifications

### ⏳ Phase 8: Security Hardening (PLANNED)

**Status**: Not started

- [ ] TLS certificate management
- [ ] mTLS implementation for production
- [ ] OS keychain integration (DPAPI/Keychain/Secret Service)
- [ ] Replay attack prevention improvements
- [ ] Rate limiting on API endpoints
- [ ] Audit logging of all operations
- [ ] Security penetration testing

### ⏳ Phase 9: Snapshot Automation (PLANNED)

**Status**: Not started

- [ ] Cron-based snapshot scheduler
- [ ] Retention policy enforcement (hourly/daily/weekly/monthly)
- [ ] Automatic cleanup
- [ ] Snapshot browsing improvements
- [ ] Restore workflow enhancements
- [ ] Append-only snapshots (ransomware protection)

### ⏳ Phase 10: Production Features (PLANNED)

**Status**: Not started

- [ ] Web-based admin dashboard
- [ ] Monitoring and metrics (Prometheus)
- [ ] Distributed tracing (OpenTelemetry)
- [ ] Multi-threaded uploads
- [ ] Delta sync optimizations
- [ ] Client-side encryption option (zero-knowledge)
- [ ] Conflict resolution
- [ ] Comprehensive test suite (unit + integration)
- [ ] Performance tuning
- [ ] Documentation completion

---

### Performance
- Use goroutines for concurrent operations
- Implement connection pooling
- Batch small files to reduce overhead
- Use buffer pools to reduce GC pressure
- Implement exponential backoff for retries

### Reliability
- Transaction logging for backup operations
- Automatic reconnection on network failures
- Checksum verification for data integrity
- Partial upload resume capability
- Health checks and heartbeats

### Scalability
- Support for multiple storage pools
- Horizontal scaling with load balancing
- Database for user metadata (PostgreSQL)
- Message queue for async operations (optional)

### Monitoring & Logging
- Structured logging (zerolog or zap)
- Metrics export (Prometheus)
- Distributed tracing (OpenTelemetry)
- Alert on errors or quota violations

## Security Considerations

### Threat Model
- Unauthorized access to backups
- Man-in-the-middle attacks
- Server compromise scenarios
- Ransomware attacking backup data

### Mitigations
- Zero-knowledge architecture (optional client-side encryption)
- Append-only snapshots (protect against ransomware)
- Rate limiting on API endpoints
- Audit logging of all operations
- Regular security audits
- Principle of least privilege

## Testing Strategy
- Unit tests for core algorithms
- Integration tests for client-server communication
- Load testing for concurrent clients
- Chaos engineering for failure scenarios
- Security penetration testing

## Documentation Requirements
- API documentation (gRPC/protobuf)
- Deployment guide
- User manual
- Admin guide
- Security best practices
- Disaster recovery procedures

