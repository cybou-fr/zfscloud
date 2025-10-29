# ZFSCloud - Encrypted Backup System

A secure client-server backup system built in Go with ZFS backend storage and end-to-end encryption.

## 🚀 Project Status

**Phase 1 Complete** ✅ - Core infrastructure implemented

- ✅ Project structure and build system
- ✅ gRPC protocol definitions (28 RPC methods)
- ✅ Cryptographic utilities (Ed25519, RSA, Argon2)
- ✅ Configuration management (YAML-based)
- ✅ ZFS management wrapper
- ✅ Basic server and client entry points

**Next Steps**: Generate protobuf files → Database layer → gRPC implementation

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

### Build

```bash
# Build server
go build -o bin/zfscloud-server ./server/cmd

# Build client
go build -o bin/zfscloud-client ./client/cmd
```

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
| 2 | Database Layer | 🔲 Planned |
| 3 | gRPC Services | 🔲 Planned |
| 4 | Backup Engine | 🔲 Planned |
| 5 | Client UI (Gio) | 🔲 Planned |
| 6 | Security Hardening | 🔲 Planned |
| 7 | Snapshot Scheduler | 🔲 Planned |
| 8 | Admin Interface | 🔲 Planned |

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
