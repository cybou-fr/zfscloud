# ZFSCloud - Encrypted Backup System

A secure client-server backup system built in Go with ZFS backend storage and end-to-end encryption.

## 🚀 Project Status

**Phase 7 Complete** ✅ - GUI implementation 100% finished!

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
- ✅ Phase 7: Complete GUI with Gio UI framework
  - Dashboard with real-time stats
  - Settings management (YAML persistence)
  - File browser with restore functionality
  - System tray integration
  - Error handling and dialogs

**Status**: 🟢 Production Ready! All core features implemented.

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

# Build GUI client (Recommended for end users)
go build -o bin/zfscloud-gui ./client/cmd
```

### GUI Client (NEW! ✨)

The GUI provides an intuitive desktop interface for managing backups:

```powershell
# Build and run the GUI
go build -o bin/zfscloud-gui.exe ./client/cmd
.\bin\zfscloud-gui.exe
```

**Features:**
- **Dashboard**: Real-time statistics, connection status, backup controls (Start/Stop/Pause)
- **Settings**: Configure server connection, client details, watch directory, upload interval
- **Files**: Browse backed up files, view file info, restore to local disk
- **Error Dialogs**: User-friendly error messages and notifications
- **Auto-refresh**: Statistics update every second automatically

**Documentation:**
- **[QUICKSTART_GUI.md](./QUICKSTART_GUI.md)** - Quick start guide (5 minutes)
- **[GUI_USER_GUIDE.md](./GUI_USER_GUIDE.md)** - Complete user manual with troubleshooting
- **[GUI_FEATURES.md](./GUI_FEATURES.md)** - Technical documentation and architecture

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

**📋 [GUI_DOCUMENTATION_INDEX.md](./GUI_DOCUMENTATION_INDEX.md)** - Complete GUI documentation index ⭐

### Getting Started
- **[QUICKSTART_GUI.md](./QUICKSTART_GUI.md)** - GUI quick start (5 minutes) ⭐ **START HERE**
- **[docker/README.md](./docker/README.md)** - Docker testing guide
- **[SERVER_QUICKSTART.md](./SERVER_QUICKSTART.md)** - Server deployment guide

### User Guides
- **[GUI_USER_GUIDE.md](./GUI_USER_GUIDE.md)** - Complete GUI user manual with troubleshooting
- **[QUICKREF.md](./QUICKREF.md)** - Quick reference and project overview
- **[SETUP.md](./SETUP.md)** - Complete setup and installation guide

### Technical Documentation
- **[GUI_FEATURES.md](./GUI_FEATURES.md)** - GUI architecture and features
- **[GUI_SUMMARY.md](./GUI_SUMMARY.md)** - Development summary and statistics
- **[GUI_RELEASE_NOTES.md](./GUI_RELEASE_NOTES.md)** - Release notes v1.0.0
- **[PHASE7_COMPLETE.md](./PHASE7_COMPLETE.md)** - Implementation completion report
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
