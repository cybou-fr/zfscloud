# ZFSCloud Project Structure

## Directory Layout

```
zfscloud/
├── client/                      # Client application
│   ├── cmd/                     # Client main entry point
│   │   └── main.go
│   ├── ui/                      # Gio UI components (TODO)
│   ├── backup/                  # Backup engine (TODO)
│   └── config.example.yaml      # Example client configuration
│
├── server/                      # Server application
│   ├── cmd/                     # Server main entry point
│   │   └── main.go
│   ├── zfs/                     # ZFS management
│   │   └── manager.go
│   ├── grpc/                    # gRPC service implementations (TODO)
│   ├── db/                      # Database layer (TODO)
│   └── config.example.yaml      # Example server configuration
│
├── proto/                       # Protocol Buffer definitions
│   └── backup.proto             # Main service definition
│
├── shared/                      # Shared libraries
│   ├── crypto/                  # Cryptographic utilities
│   │   ├── keys.go              # Ed25519 key management
│   │   └── passphrase.go        # Passphrase generation and encryption
│   └── config/                  # Configuration management
│       ├── server.go            # Server config
│       └── client.go            # Client config
│
├── bin/                         # Compiled binaries (gitignored)
├── data/                        # Runtime data (gitignored)
├── certs/                       # TLS certificates (gitignored)
│
├── go.mod                       # Go module definition
├── Makefile                     # Build automation (Linux/Mac)
├── build.ps1                    # Build script (Windows)
├── README.md                    # Project documentation
├── implementation-plan.md       # Detailed implementation plan
└── .gitignore                   # Git ignore rules
```

## Component Status

### ✅ Completed
- [x] Project structure
- [x] Go modules initialization
- [x] Protocol Buffer definitions (gRPC API)
- [x] Cryptographic utilities (Ed25519, passphrases)
- [x] Configuration management
- [x] ZFS manager wrapper
- [x] Basic server entry point
- [x] Basic client entry point
- [x] Example configurations
- [x] Build scripts

### 🚧 In Progress
- [ ] gRPC server implementation
- [ ] Database layer (user management)
- [ ] Client UI with Gio

### 📋 TODO (Next Steps)
1. Generate protobuf Go files
2. Implement gRPC server services
3. Create database schema and models
4. Implement registration workflow
5. Build backup engine (file watching, hashing, chunking)
6. Create Gio-based UI
7. Add TLS/mTLS support
8. Implement snapshot scheduling
9. Build admin interface
10. Add tests

## Development Workflow

### Prerequisites
- Go 1.23+
- Protocol Buffers compiler (protoc)
- ZFS (for server, Linux only)

### Setup
```powershell
# Clone and setup
cd c:\Users\yosta\dev\zfscloud

# Install proto generation tools
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Download dependencies
go mod download

# Generate protobuf files
.\build.ps1 proto

# Build everything
.\build.ps1 all
```

### Running

```powershell
# Run server (Linux with ZFS)
.\bin\zfscloud-server.exe -config server\config.example.yaml

# Run client
.\bin\zfscloud-client.exe -config client\config.example.yaml
```

## Next Implementation Steps

1. **Generate Protobuf Files**
   - Install protoc and Go plugins
   - Run: `.\build.ps1 proto`

2. **Implement Database Layer**
   - Create SQLite schema for users, registrations, snapshots
   - Implement CRUD operations
   - Add migration support

3. **Implement gRPC Services**
   - BackupService implementation
   - AdminService implementation
   - Request validation and authentication

4. **Build Backup Engine**
   - File system watcher
   - Hash calculation
   - Content-defined chunking
   - Data compression

5. **Create Client UI**
   - System tray icon with Gio
   - Status display
   - Configuration interface
   - File browser for restore

6. **Add Security Layer**
   - TLS certificate management
   - mTLS setup
   - Key storage (OS keychain integration)

7. **Testing & Documentation**
   - Unit tests
   - Integration tests
   - API documentation
   - User guide
