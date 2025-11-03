# ZFSCloud - Encrypted Backup System

A secure client-server backup system built in Go with ZFS backend storage and end-to-end encryption.

## 🚀 Project Status

**Phase 7 Started** 🔨 - GUI implementation in progress (40% complete)

- ✅ Project structure and build system
- ✅ gRPC protocol definitions (28 RPC methods)
- ✅ Cryptographic utilities (Ed25519, RSA, Argon2)
- ✅ Configuration management (YAML-based)
- ✅ ZFS management wrapper
- ✅ Complete server implementation (BackupService + AdminService)
- ✅ Database layer with SQLite (8 tables, 28 operations)
- ✅ Authentication middleware with Ed25519 signatures
- ✅ File I/O layer with hash verification
- ✅ Interactive CLI client
- ✅ File system watcher with real-time change detection
- ✅ Automatic backup engine with compression
- ✅ Docker testing environment
- ✅ Phase 8: Security hardening with TLS/mTLS (ECDSA P-256)
- ✅ Phase 9: Snapshot automation (4-tier scheduling with retention)
- 🔨 Phase 7: GUI with Gio UI framework (dashboard, settings, files)

**Next Steps**: Complete GUI polish → Add system tray → Implement file browser

## Features

- 🔐 End-to-end encrypted backups
- 📸 Automated ZFS snapshots with retention policies
- 🔑 Ed25519 public key authentication
- 💾 Per-user encrypted ZFS datasets (AES-256-GCM)
- 🔄 Incremental deduplication (client & server-side)
- 🖥️ Cross-platform client with system tray UI (Gio)
- 👨‍💼 Admin approval workflow for registration & restores
- 🔒 mTLS for all client-server communication

## Project Structure

```
zfscloud/
├── client/          # Client application (Gio UI)
├── server/          # Server application (ZFS backend)
├── proto/           # gRPC protocol definitions
├── shared/          # Shared libraries
│   ├── crypto/      # Cryptographic utilities
│   └── config/      # Configuration management
└── docs/            # Documentation
```

## Requirements

### Server
- Go 1.23+
- ZFS filesystem
- Linux (for ZFS support)

### Client
- Go 1.23+
- Windows/Linux/macOS

## Getting Started

### Quick Test with Docker

The easiest way to test the system:

```powershell
# 1. Build server for Docker
.\build-docker.ps1

# 2. Start server in Docker
cd docker
docker-compose up -d

# 3. Run client
.\bin\zfscloud-client.exe
```

See **[docker/README.md](./docker/README.md)** for complete Docker testing guide.

### Build

```bash
# Build server
go build -o bin/zfscloud-server ./server/cmd

# Build CLI client
go build -o bin/zfscloud-client ./client/cmd

# Build GUI client (NEW!)
go build -o bin/zfscloud-gui ./client/gui/cmd
```

### GUI Client

The new GUI provides an intuitive interface for managing backups:

```powershell
# Run the GUI
.\bin\zfscloud-gui.exe
```

Features:
- **Dashboard**: Real-time statistics and backup controls
- **Settings**: Configure server, TLS, folders, and advanced options
- **Files**: Browse and restore backed up files (coming soon)

See **[client/gui/README.md](./client/gui/README.md)** for detailed GUI documentation.

### Development

```bash
# Generate protobuf files
make proto

# Run tests
go test ./...

# Run server
./bin/zfscloud-server -config server/config.yaml

# Run client
./bin/zfscloud-client -config client/config.yaml
```

## 📚 Documentation

- **[docker/README.md](./docker/README.md)** - Docker testing guide (START HERE)
- **[SERVER_QUICKSTART.md](./SERVER_QUICKSTART.md)** - Server deployment guide
- **[QUICKREF.md](./QUICKREF.md)** - Quick reference and project overview
- **[SETUP.md](./SETUP.md)** - Complete setup and installation guide
- **[STRUCTURE.md](./STRUCTURE.md)** - Project structure and components
- **[implementation-plan.md](./implementation-plan.md)** - Detailed architecture and design

## 🎯 Quick Start

```powershell
# 1. Install prerequisites (Go 1.23+, protoc)
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 2. Download dependencies
go mod download

# 3. Generate protobuf files (requires protoc)
.\build.ps1 proto

# 4. Build
.\build.ps1 all

# 5. Configure (copy and edit example configs)
cp server/config.example.yaml server/config.yaml
cp client/config.example.yaml client/config.yaml

# 6. Run (server requires Linux with ZFS)
./bin/zfscloud-server -config server/config.yaml
./bin/zfscloud-client -config client/config.yaml
```

See [SETUP.md](./SETUP.md) for detailed instructions.

## 🏗️ Architecture

```
Client (Gio UI)
    ↓ gRPC/mTLS
Server (Go)
    ↓ Ed25519 Auth
ZFS Pool
    └── Per-user encrypted datasets
        ├── Automated snapshots
        └── Deduplication
```

## 📊 Implementation Progress

| Phase | Component | Status |
|-------|-----------|--------|
| 1 | Core Infrastructure | ✅ Complete |
| 2 | Database Layer | ✅ Complete |
| 3 | gRPC Services | ✅ Complete |
| 4 | Basic Client | ✅ Complete |
| 5 | Docker Testing | ✅ Complete |
| 6 | Backup Engine | ✅ Complete |
| 7 | Client UI (Gio) | 🔲 Planned |
| 8 | Snapshot Scheduler | 🔲 Planned |
| 9 | Security Hardening (TLS) | 🔲 Planned |
| 10 | Admin Interface | 🔲 Planned |

## 🛠️ Technology Stack

- **Language**: Go 1.23+
- **RPC**: gRPC with Protocol Buffers
- **Storage**: ZFS with native encryption
- **Crypto**: Ed25519 (auth), RSA (key exchange), Argon2 (KDF)
- **Transport**: TLS 1.3 with mutual authentication
- **UI**: Gio (cross-platform native GUI)
- **Database**: SQLite (development) / PostgreSQL (production)
- **Config**: YAML

## 📝 License

MIT License
