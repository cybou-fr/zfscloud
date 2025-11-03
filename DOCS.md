# ZFSCloud Documentation Index

## Getting Started

### New Users

1. **[README.md](./README.md)** - Project overview and quick start
2. **[QUICKSTART_GUI.md](./QUICKSTART_GUI.md)** - GUI client tutorial (5 minutes)
3. **[SETUP.md](./SETUP.md)** - Installation and configuration

### Deployment

- **[docker/README.md](./docker/README.md)** - Docker testing environment
- **[DEPLOYMENT_PACKAGE.md](./DEPLOYMENT_PACKAGE.md)** - Production deployment
- **[DOCKER_TESTING_GUIDE.md](./DOCKER_TESTING_GUIDE.md)** - Complete Docker guide
- **[SERVER_QUICKSTART.md](./SERVER_QUICKSTART.md)** - Server setup

## User Guides

### GUI Client

- **[GUI_USER_GUIDE.md](./GUI_USER_GUIDE.md)** - Complete GUI manual
- **[GUI_FEATURES.md](./GUI_FEATURES.md)** - Feature documentation

### Command Line

- **[QUICKREF.md](./QUICKREF.md)** - Command reference and examples
- **[QUICK_TEST.md](./QUICK_TEST.md)** - Manual testing procedures

## Technical Documentation

### Architecture

- **[implementation-plan.md](./implementation-plan.md)** - Complete system design
- **[STRUCTURE.md](./STRUCTURE.md)** - Code organization

### Security

- **[SECURITY_FIX_PROGRESS.md](./SECURITY_FIX_PROGRESS.md)** - User-controlled encryption implementation
- **[ZFS_ENCRYPTION_FIX.md](./ZFS_ENCRYPTION_FIX.md)** - Encryption architecture details

## Quick Links

| Task | Document |
|------|----------|
| First time setup | [QUICKSTART_GUI.md](./QUICKSTART_GUI.md) |
| Install server | [SETUP.md](./SETUP.md) |
| Docker testing | [docker/README.md](./docker/README.md) |
| CLI commands | [QUICKREF.md](./QUICKREF.md) |
| Troubleshooting | [GUI_USER_GUIDE.md](./GUI_USER_GUIDE.md#troubleshooting) |
| Security status | [SECURITY_FIX_PROGRESS.md](./SECURITY_FIX_PROGRESS.md) |
| Architecture | [implementation-plan.md](./implementation-plan.md) |

## Documentation by Role

### End Users

Focus on GUI client usage:

1. [QUICKSTART_GUI.md](./QUICKSTART_GUI.md)
2. [GUI_USER_GUIDE.md](./GUI_USER_GUIDE.md)
3. [QUICKREF.md](./QUICKREF.md) (for CLI mode)

### System Administrators

Server deployment and management:

1. [SETUP.md](./SETUP.md)
2. [SERVER_QUICKSTART.md](./SERVER_QUICKSTART.md)
3. [DEPLOYMENT_PACKAGE.md](./DEPLOYMENT_PACKAGE.md)
4. [docker/README.md](./docker/README.md)
5. [QUICKREF.md](./QUICKREF.md) (admin commands)

### Developers

Architecture and implementation:

1. [implementation-plan.md](./implementation-plan.md)
2. [STRUCTURE.md](./STRUCTURE.md)
3. [SECURITY_FIX_PROGRESS.md](./SECURITY_FIX_PROGRESS.md)
4. [ZFS_ENCRYPTION_FIX.md](./ZFS_ENCRYPTION_FIX.md)
5. [GUI_FEATURES.md](./GUI_FEATURES.md)

## Status & Progress

All core features implemented ✅

**Current Focus:** Client-side passphrase encryption (see [SECURITY_FIX_PROGRESS.md](./SECURITY_FIX_PROGRESS.md))

## Contributing

See [implementation-plan.md](./implementation-plan.md) for architecture details.

## License

MIT License - see LICENSE file for details.
