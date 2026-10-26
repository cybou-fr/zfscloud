# ZFSCloud — Encrypted Backup System

**Étude de cas pour clients :** [architecture Go / gRPC / ZFS et compromis](docs/CASE_STUDY_FR.md) · [Portfolio de l'architecte](https://github.com/cybou-fr)

> **Engineering case:** a GUI/CLI client, RPC contracts, Linux backend and ZFS storage working as one architecture. **Status:** personal R&D, not a client delivery.

**🇫🇷 [Étude de cas — Go, gRPC et ZFS](docs/CASE_STUDY_FR.md)** — Architecture client–serveur et compromis techniques.

A secure client-server backup system built in Go with ZFS backend storage and user-controlled encryption.

## 🚀 Status

**Advanced R&D prototype — not qualified for production deployment without further validation.** Features listed below are documented in source and guides; their presence does not establish complete release qualification.

**Core Features:**
- ✅ gRPC protocol with 28 RPC methods
- ✅ User-controlled ZFS dataset encryption (AES-256-GCM)
- ✅ Ed25519 authentication + RSA passphrase encryption
- ✅ Server implementation (BackupService + AdminService)
- ✅ Database layer with SQLite
- ✅ TLS/mTLS transport security
- ✅ Automated snapshot scheduling (4-tier retention)
- ✅ GUI client (Gio UI framework)
- ✅ CLI client with interactive mode
- ✅ File system watcher with incremental backups
- ✅ Docker deployment support

**Security status:** design and test scenarios are documented; OS keychain integration for secret storage is still pending (see [SECURITY_FIX_PROGRESS.md](./SECURITY_FIX_PROGRESS.md)). No independent security audit is claimed.

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

```text
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

**On Linux:**
```bash
# Interactive menu (recommended)
./test-menu.sh

# Or run specific tests
./test-integration.sh           # Automated full workflow
./test-performance.sh           # Performance benchmarks
```

**On Windows:**
```powershell
# 1. Build server for Docker
.\build-docker.ps1

# 2. Start server in Docker
cd docker
docker-compose up -d

# 3. Run client
.\bin\zfscloud-client.exe
```

See **[LINUX_TESTING_GUIDE.md](./LINUX_TESTING_GUIDE.md)** for Linux testing or **[docker/README.md](./docker/README.md)** for complete Docker guide.

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

**📋 [DOCS.md](./DOCS.md)** - Complete documentation index

### Quick Start

- **[QUICKSTART_GUI.md](./QUICKSTART_GUI.md)** - 5-minute GUI tutorial ⭐
- **[SETUP.md](./SETUP.md)** - Installation guide
- **[QUICKREF.md](./QUICKREF.md)** - Command reference

### Testing Guides

- **[LINUX_TESTING_GUIDE.md](./LINUX_TESTING_GUIDE.md)** - Linux testing suite ⭐
- **[docker/README.md](./docker/README.md)** - Docker deployment
- **[DOCKER_TESTING_GUIDE.md](./DOCKER_TESTING_GUIDE.md)** - Windows Docker testing

### Key Documents

- **[GUI_USER_GUIDE.md](./GUI_USER_GUIDE.md)** - GUI manual
- **[SECURITY_FIX_PROGRESS.md](./SECURITY_FIX_PROGRESS.md)** - Encryption status
- **[implementation-plan.md](./implementation-plan.md)** - Architecture

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

```text
┌─────────────────────────────────────┐
│  Client (GUI/CLI)                   │
│  - Ed25519 authentication           │
│  - RSA passphrase encryption        │
│  - File watcher & backup engine     │
└──────────────┬──────────────────────┘
               │ gRPC over mTLS
┌──────────────▼──────────────────────┐
│  Server (Go)                        │
│  - Authentication middleware        │
│  - Admin approval workflow          │
│  - Snapshot scheduler               │
└──────────────┬──────────────────────┘
               │ ZFS commands
┌──────────────▼──────────────────────┐
│  ZFS Pool                           │
│  └── Per-user encrypted datasets    │
│      ├── AES-256-GCM encryption     │
│      ├── Automated snapshots        │
│      └── Deduplication              │
└─────────────────────────────────────┘
```

## 📊 Implementation Status

| Component | Status |
|-----------|--------|
| Core Infrastructure | ✅ Complete |
| Database Layer | ✅ Complete |
| gRPC Services (28 methods) | ✅ Complete |
| Backup Engine | ✅ Complete |
| GUI Client (Gio) | ✅ Complete |
| CLI Client | ✅ Complete |
| Snapshot Scheduler | ✅ Complete |
| TLS/mTLS Security | ✅ Complete |
| Docker Deployment | ✅ Complete |
| Passphrase Encryption | ✅ Complete (server + client) |
| OS Keychain Storage | ⏳ Pending (manual passphrase backup required) |

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
