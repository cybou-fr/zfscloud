# ZFSCloud - Encrypted Backup System Implementation Plan

## Overview
A client-server backup system built in Go, featuring a Gio UI-based client (system tray application) and a ZFS-backed server for secure, encrypted, per-user backups.

## Architecture

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

### Phase 1: Core Infrastructure
- [ ] gRPC service definitions
- [ ] mTLS setup and certificate management
- [ ] Client key pair generation and storage
- [ ] Server registration and approval system
- [ ] Basic ZFS dataset creation

### Phase 2: Backup Engine
- [ ] File system monitoring
- [ ] Hash calculation and comparison
- [ ] Chunking algorithm
- [ ] Compression pipeline
- [ ] Data transmission protocol
- [ ] ZFS dataset writing

### Phase 3: Client UI
- [ ] Gio system tray application
- [ ] Folder selection interface
- [ ] Status display
- [ ] Configuration management
- [ ] Basic restore functionality

### Phase 4: Snapshot Management
- [ ] Automated snapshot scheduling
- [ ] Retention policy enforcement
- [ ] Snapshot listing API
- [ ] Restore request workflow
- [ ] Admin approval system

### Phase 5: Admin Interface
- [ ] User approval dashboard
- [ ] Storage monitoring
- [ ] Policy configuration
- [ ] Activity logs

### Phase 6: Advanced Features
- [ ] Bandwidth optimization
- [ ] Delta sync improvements
- [ ] Multi-threaded uploads
- [ ] Client-side encryption option
- [ ] Conflict resolution
- [ ] Versioning support

## Technical Considerations

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

