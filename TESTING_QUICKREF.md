# ZFSCloud Linux Testing - Quick Reference

## 🚀 Quick Start

```bash
# Interactive menu (easiest)
./test-menu.sh

# Automated integration test
./test-integration.sh

# Performance benchmark
./test-performance.sh
```

## 📋 Test Scripts

| Script | Purpose | Duration |
|--------|---------|----------|
| `test-menu.sh` | Interactive menu | Interactive |
| `test-docker.sh` | Manual testing guide | 5-10 min |
| `test-integration.sh` | Full automated test | 2-3 min |
| `test-performance.sh` | Performance benchmarks | 5-10 min |
| `test-helpers.sh` | Utility functions | - |

## 🔧 Common Commands

### Server Management

```bash
# Start server
cd docker && docker-compose up -d

# Stop server (keep data)
cd docker && docker-compose down

# Clean all data
cd docker && docker-compose down -v

# View logs
docker logs -f zfscloud-server

# Check status
docker ps | grep zfscloud
```

### Helper Functions

```bash
# Show available functions
./test-helpers.sh

# Execute directly
./test-helpers.sh show_system_info
./test-helpers.sh db_list_users
./test-helpers.sh zfs_pool_status

# Source in script
source ./test-helpers.sh
docker_start_server
wait_for_server
```

### Performance Testing

```bash
# Default settings
./test-performance.sh

# Custom configuration
NUM_CLIENTS=5 \
FILE_SIZE_LARGE=100 \
./test-performance.sh
```

## 🔍 Verification Commands

```bash
# Check ZFS status
docker exec zfscloud-server zpool status tank
docker exec zfscloud-server zfs list -r tank

# Check database
docker exec zfscloud-server sqlite3 /var/lib/zfscloud/zfscloud.db "SELECT * FROM users;"

# Check files
docker exec zfscloud-server ls -la /tank/<client-id>/

# Interactive database
docker exec -it zfscloud-server sqlite3 /var/lib/zfscloud/zfscloud.db
```

## 🎯 Common Tasks

### Register and Approve Client

```bash
# Method 1: Using helper
source ./test-helpers.sh
CLIENT_ID=$(generate_client_id)
create_client_config "$CLIENT_ID" "client.yaml"
register_and_approve_client "$CLIENT_ID" "client.yaml"

# Method 2: Manual
./bin/zfscloud-client -config client.yaml
> register
# Then approve in another terminal
docker exec zfscloud-server sqlite3 /var/lib/zfscloud/zfscloud.db \
  "SELECT registration_id FROM registrations WHERE status='pending';"
docker exec zfscloud-server sqlite3 /var/lib/zfscloud/zfscloud.db \
  "UPDATE registrations SET status='approved' WHERE registration_id='<ID>';"
```

### Upload and Verify File

```bash
# Create test file
echo "Hello ZFSCloud" > test.txt

# Upload
./bin/zfscloud-client -config client.yaml
> upload test.txt /test.txt

# Verify on server
docker exec zfscloud-server cat /tank/<client-id>/test.txt
```

### Download File

```bash
./bin/zfscloud-client -config client.yaml
> download /test.txt ./restored.txt

# Verify
cat restored.txt
```

## 📊 Monitoring

### Real-time Monitoring

```bash
# Follow server logs
docker logs -f zfscloud-server

# Watch ZFS I/O
watch -n 1 'docker exec zfscloud-server zpool iostat tank 1 1'

# Monitor disk usage
watch -n 5 'docker exec zfscloud-server zfs list -r tank'
```

### System Info

```bash
# Using helpers
./test-helpers.sh show_system_info
./test-helpers.sh show_client_info <client-id>

# Manual
docker exec zfscloud-server zpool status tank
docker exec zfscloud-server sqlite3 /var/lib/zfscloud/zfscloud.db \
  "SELECT COUNT(*) FROM users;"
```

## 🐛 Troubleshooting

### Server Won't Start

```bash
# Check Docker
docker ps -a
docker logs zfscloud-server

# Check ZFS module
lsmod | grep zfs
sudo modprobe zfs
```

### Connection Refused

```bash
# Check port
netstat -tuln | grep 50051

# Check server listening
docker exec zfscloud-server netstat -tuln | grep 50051
```

### Permission Denied

```bash
# Add to docker group
sudo usermod -aG docker $USER
newgrp docker
```

## 🧹 Cleanup

```bash
# Stop server, keep data
cd docker && docker-compose down

# Remove all data
cd docker && docker-compose down -v

# Clean test files
rm -rf test-data-* perf-test-*
rm -f perf-results-*.txt
rm -f client-*.yaml
```

## 📈 Environment Variables

### Performance Test

```bash
NUM_CLIENTS=5              # Number of concurrent clients (default: 3)
FILE_SIZE_SMALL=2          # Small file size MB (default: 1)
FILE_SIZE_MEDIUM=20        # Medium file size MB (default: 10)
FILE_SIZE_LARGE=100        # Large file size MB (default: 50)
NUM_FILES_PER_CLIENT=20    # Files per client (default: 10)
```

### Helper Functions

```bash
SERVER_CONTAINER=zfscloud-server    # Container name
ADMIN_TOKEN=your-token              # Admin token
DEFAULT_QUOTA=10737418240           # Default quota (10GB)
```

## 📚 Documentation

- [LINUX_TESTING_GUIDE.md](./LINUX_TESTING_GUIDE.md) - Complete guide
- [docker/README.md](./docker/README.md) - Docker configuration
- [README.md](./README.md) - Main documentation

## 💡 Tips

1. Always use `./test-menu.sh` for interactive testing
2. Run `./test-integration.sh` before making changes
3. Use `source ./test-helpers.sh` for custom scripts
4. Monitor logs with `docker logs -f zfscloud-server`
5. Clean data between tests: `docker-compose down -v`

## 🎓 Example Workflows

### Daily Testing

```bash
./test-integration.sh && echo "✓ Tests passed"
```

### Performance Baseline

```bash
./test-performance.sh > baseline-$(date +%Y%m%d).log
```

### Custom Multi-Client

```bash
source ./test-helpers.sh
build_all
docker_start_server
wait_for_server

for i in {1..3}; do
    CLIENT="client-$i"
    create_client_config "$CLIENT" "$CLIENT.yaml"
    register_and_approve_client "$CLIENT" "$CLIENT.yaml"
done

show_system_info
```

---

**For detailed information, see [LINUX_TESTING_GUIDE.md](./LINUX_TESTING_GUIDE.md)**
