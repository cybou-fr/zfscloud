# ZFSCloud Linux Testing Suite

Comprehensive testing scripts for Docker-based client-server testing on Linux.

## 📋 Overview

This directory contains automated testing scripts for ZFSCloud:

- **test-docker.sh** - Quick manual testing guide
- **test-integration.sh** - Automated integration tests
- **test-performance.sh** - Performance benchmarking
- **test-helpers.sh** - Reusable utility functions

## 🚀 Quick Start

### 1. Make Scripts Executable

```bash
chmod +x test-docker.sh
chmod +x test-integration.sh
chmod +x test-performance.sh
chmod +x test-helpers.sh
```

### 2. Run Quick Test

```bash
./test-docker.sh
```

This will:
- Build server and client binaries
- Start Docker container
- Show manual testing instructions

### 3. Run Integration Tests

```bash
./test-integration.sh
```

Fully automated test suite that:
- ✅ Builds binaries
- ✅ Starts server
- ✅ Registers and approves client
- ✅ Uploads test files
- ✅ Downloads and verifies files
- ✅ Validates ZFS datasets
- ✅ Checks database integrity

### 4. Run Performance Tests

```bash
# Default settings
./test-performance.sh

# Custom configuration
NUM_CLIENTS=5 FILE_SIZE_LARGE=100 ./test-performance.sh
```

Tests:
- Single client upload/download speeds
- Multiple file uploads
- Concurrent client operations
- System resource usage

## 📚 Script Details

### test-docker.sh

Quick interactive testing script. Starts the server and provides manual test instructions.

**Usage:**
```bash
./test-docker.sh
```

**Features:**
- Builds binaries if missing
- Starts Docker server
- Shows server logs
- Provides step-by-step commands

### test-integration.sh

Automated full workflow testing.

**Usage:**
```bash
./test-integration.sh
```

**Tests:**
1. Binary compilation
2. Server startup and health check
3. Client registration
4. Admin approval workflow
5. Heartbeat functionality
6. File upload (small, medium files)
7. File download and verification
8. ZFS dataset validation
9. Database integrity check

**Output:**
- Colored console output
- Success/failure indicators
- Detailed verification steps
- Server logs and metrics

### test-performance.sh

Performance benchmarking with configurable parameters.

**Usage:**
```bash
# Default: 3 clients, 1/10/50 MB files
./test-performance.sh

# Custom settings
NUM_CLIENTS=5 \
FILE_SIZE_SMALL=2 \
FILE_SIZE_MEDIUM=20 \
FILE_SIZE_LARGE=100 \
NUM_FILES_PER_CLIENT=20 \
./test-performance.sh
```

**Environment Variables:**
- `NUM_CLIENTS` - Number of concurrent clients (default: 3)
- `FILE_SIZE_SMALL` - Small file size in MB (default: 1)
- `FILE_SIZE_MEDIUM` - Medium file size in MB (default: 10)
- `FILE_SIZE_LARGE` - Large file size in MB (default: 50)
- `NUM_FILES_PER_CLIENT` - Files per client (default: 10)

**Metrics:**
- Upload/download speeds (MB/s)
- Concurrent operation throughput
- File transfer durations
- System resource usage
- ZFS compression ratios

**Output:**
- Console display with metrics
- Results file: `perf-results-<timestamp>.txt`

### test-helpers.sh

Reusable utility functions for test scripts.

**Usage:**

```bash
# Source in another script
source ./test-helpers.sh

# Use functions
docker_start_server
wait_for_server
register_and_approve_client "test-client-1" "config.yaml"
```

**Or execute directly:**
```bash
# Show available functions
./test-helpers.sh

# Execute a function
./test-helpers.sh show_system_info
./test-helpers.sh db_list_users
./test-helpers.sh zfs_pool_status
```

**Function Categories:**

**Docker Management:**
- `docker_start_server` - Start server container
- `docker_stop_server` - Stop server
- `docker_clean_all` - Remove all data
- `docker_server_logs [lines]` - Show logs
- `wait_for_server [timeout]` - Wait for ready

**ZFS Operations:**
- `zfs_pool_status` - Show pool status
- `zfs_list_datasets` - List all datasets
- `zfs_create_dataset <name>` - Create dataset
- `zfs_dataset_size <name>` - Get dataset size
- `zfs_compression_ratio` - Get compression ratio

**Database Operations:**
- `db_list_users` - List all users
- `db_list_files [client_id]` - List files
- `db_get_registration <client_id>` - Get pending registration
- `db_approve_registration <reg_id> [quota]` - Approve registration
- `db_interactive` - Open SQLite shell

**Client Management:**
- `generate_client_id` - Generate unique ID
- `create_client_config <id> <file> [addr]` - Create config
- `register_and_approve_client <id> <config> [quota]` - Full registration

**File Operations:**
- `create_test_file <size_mb> <filename>` - Create random file
- `verify_file_on_server <client_id> <path>` - Check file exists
- `compare_file_hash <local> <client_id> <remote>` - Compare hashes

**System Info:**
- `show_system_info` - Display system overview
- `show_client_info <client_id>` - Show client details
- `build_all` - Build server and client

## 🔧 Requirements

### System Requirements

- **OS:** Linux (Ubuntu 22.04+ recommended)
- **Docker:** 20.10+
- **Docker Compose:** 1.29+
- **Go:** 1.23+
- **ZFS Utils:** For kernel module

### Install Dependencies

```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install -y docker.io docker-compose golang-go zfsutils-linux

# Fedora/RHEL
sudo dnf install -y docker docker-compose golang zfs

# Arch Linux
sudo pacman -S docker docker-compose go zfs-utils
```

### Load ZFS Kernel Module

```bash
sudo modprobe zfs
```

## 🎯 Test Scenarios

### Scenario 1: Quick Smoke Test

```bash
./test-docker.sh
# Follow manual instructions
```

**Time:** 5-10 minutes  
**Purpose:** Verify basic functionality

### Scenario 2: Full Integration Test

```bash
./test-integration.sh
```

**Time:** 2-3 minutes  
**Purpose:** Automated end-to-end validation

### Scenario 3: Performance Baseline

```bash
./test-performance.sh
```

**Time:** 5-10 minutes  
**Purpose:** Establish performance benchmarks

### Scenario 4: Stress Test

```bash
NUM_CLIENTS=10 FILE_SIZE_LARGE=100 NUM_FILES_PER_CLIENT=50 ./test-performance.sh
```

**Time:** 15-30 minutes  
**Purpose:** Test under heavy load

### Scenario 5: Custom Test

```bash
# Source helpers
source ./test-helpers.sh

# Build and start
build_all
docker_start_server
wait_for_server

# Create clients
CLIENT1=$(generate_client_id)
create_client_config "$CLIENT1" "client1.yaml"
register_and_approve_client "$CLIENT1" "client1.yaml"

# Your custom tests here...

# Cleanup
cleanup_client_data "$CLIENT1"
docker_stop_server
```

## 📊 Understanding Results

### Integration Test Output

```
✓ Server binary built
✓ Container started
✓ Registration approved
✓ File uploaded successfully
✓ File content verified
```

**Green ✓** = Success  
**Red ✗** = Failure  
**Yellow →** = In progress

### Performance Test Results

Sample output:
```
Single Upload - Large (50MB): 2.34s @ 21.37 MB/s
Concurrent Upload (3 clients, 10MB each):
  Duration: 5s
  Total Data: 30MB
  Throughput: 6.00 MB/s
```

**Metrics to monitor:**
- Upload/download speed should be > 10 MB/s for local tests
- Concurrent throughput scales with number of clients
- Compression ratio should be > 1.0 for compressible data

## 🐛 Troubleshooting

### Server Won't Start

```bash
# Check Docker
docker ps -a

# View logs
docker logs zfscloud-server

# Check ZFS module
lsmod | grep zfs
sudo modprobe zfs
```

### Tests Timeout

```bash
# Increase timeouts in scripts
# Or manually check server
docker exec zfscloud-server zpool status tank
```

### Permission Errors

```bash
# Add user to docker group
sudo usermod -aG docker $USER
newgrp docker

# Or run with sudo
sudo ./test-integration.sh
```

### Port Already in Use

```bash
# Check what's using port 50051
sudo netstat -tuln | grep 50051
sudo lsof -i :50051

# Kill process or change port in docker-compose.yml
```

### Out of Disk Space

```bash
# Check disk usage
df -h

# Clean Docker data
docker system prune -a
cd docker && docker-compose down -v
```

### ZFS Module Not Found

```bash
# Install ZFS
sudo apt-get install zfsutils-linux

# Load module
sudo modprobe zfs

# Verify
lsmod | grep zfs
```

## 🧹 Cleanup

### Stop Server (Keep Data)

```bash
cd docker && docker-compose down
```

### Clean All Data

```bash
cd docker && docker-compose down -v
```

### Remove Test Files

```bash
rm -rf test-data-* perf-test-*
rm -f perf-results-*.txt
rm -f client-*.yaml client-keys-*.dat
```

### Full Reset

```bash
cd docker
docker-compose down -v
docker system prune -af
cd ..
rm -rf bin/ data/
```

## 📈 Best Practices

1. **Always start with test-docker.sh** to verify basic setup
2. **Run integration tests** before making changes
3. **Benchmark performance** to establish baselines
4. **Clean data between tests** for consistent results
5. **Monitor server logs** during tests
6. **Check ZFS status** after intensive operations

## 🔗 Related Documentation

- [DOCKER_TESTING_GUIDE.md](../DOCKER_TESTING_GUIDE.md) - Comprehensive Docker guide
- [docker/README.md](../docker/README.md) - Docker configuration
- [README.md](../README.md) - Main project documentation

## 📝 Notes

- Tests create temporary files and datasets
- Server continues running after tests (manual cleanup)
- Performance varies based on hardware
- First run may be slower (Docker image build)
- ZFS compression affects storage metrics

## 🎓 Examples

### Example 1: Quick Daily Test

```bash
./test-integration.sh && echo "Tests passed!" || echo "Tests failed!"
```

### Example 2: Weekly Performance Check

```bash
./test-performance.sh > weekly-perf-$(date +%Y%m%d).log 2>&1
```

### Example 3: CI/CD Integration

```bash
#!/bin/bash
set -e
./test-integration.sh
if [ $? -eq 0 ]; then
    echo "Integration tests passed"
    exit 0
else
    echo "Integration tests failed"
    exit 1
fi
```

### Example 4: Custom Multi-Client Test

```bash
source ./test-helpers.sh

build_all
docker_start_server
wait_for_server

for i in {1..5}; do
    CLIENT="client-$i"
    create_client_config "$CLIENT" "$CLIENT.yaml"
    register_and_approve_client "$CLIENT" "$CLIENT.yaml"
    echo "Client $i ready"
done

show_system_info
```

---

**Happy Testing! 🚀**
