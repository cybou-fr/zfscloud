# Client-Server Integration Test - Quick Reference

## 🎯 Test Objective
Test complete workflow: client registration → admin approval → encrypted dataset creation → file upload from `/home/user/Téléchargements`

## 📋 Test Setup

### Current Status
- ✅ Server running on `127.0.0.1:50052`
- ✅ Test config created: `test-client-config-50052.yaml`
- ✅ Test directory: `/home/user/Téléchargements` (~2.2GB, 12 files)
- ✅ Admin token: `test-admin-token-12345`

### Files Created
1. **test-client-config-50052.yaml** - Client configuration
2. **test-full-client-server.sh** - Automated test script
3. **MANUAL_TEST_GUIDE.sh** - Step-by-step manual guide
4. **admin-helper.sh** - Admin command shortcuts

---

## 🚀 Quick Start (3 Options)

### Option 1: Automated Test (Recommended)
```bash
./test-full-client-server.sh
```
This will guide you through all steps interactively.

### Option 2: Manual Interactive Test
```bash
# Terminal 1: Start client
./bin/test-client -config test-client-config-50052.yaml -interactive

# In client prompt:
zfscloud> register
# SAVE THE PASSPHRASE!

# Terminal 2: Approve registration
./admin-helper.sh list-pending
./admin-helper.sh approve <registration_id> 20

# Terminal 1: Upload files
zfscloud> heartbeat
zfscloud> status
zfscloud> upload /home/user/Téléchargements/FZXAC7RI test.txt
zfscloud> upload-dir /home/user/Téléchargements telechargements
zfscloud> list
```

### Option 3: View Manual Guide
```bash
./MANUAL_TEST_GUIDE.sh
```
Shows detailed step-by-step instructions.

---

## 📝 Key Commands

### Client Commands
```bash
# Start client
./bin/test-client -config test-client-config-50052.yaml -interactive

# Inside client:
register              # Register with server
heartbeat             # Send heartbeat
status                # Show status
upload <file> <dest>  # Upload single file
upload-dir <dir> <d>  # Upload directory
list [path]           # List backed up files
snapshots             # List snapshots
help                  # Show help
exit                  # Exit
```

### Admin Commands (use admin-helper.sh)
```bash
./admin-helper.sh list-pending          # List pending registrations
./admin-helper.sh approve <id> [quota]  # Approve (quota in GB)
./admin-helper.sh list-users            # List all users
./admin-helper.sh status                # System status
./admin-helper.sh snapshots             # List snapshots
```

### Raw grpcurl (if needed)
```bash
# List pending
grpcurl -plaintext -d '{"admin_token":"test-admin-token-12345"}' \
    127.0.0.1:50052 zfscloud.v1.AdminService/ListPendingRegistrations

# Approve
grpcurl -plaintext -d '{"admin_token":"test-admin-token-12345","registration_id":"<ID>","quota_bytes":21474836480}' \
    127.0.0.1:50052 zfscloud.v1.AdminService/ApproveRegistration
```

---

## 🔍 Verification Commands

### Check ZFS
```bash
# List datasets
zfs list -r tank/users

# Check encryption
zfs get encryption,keystatus tank/users/<USER_ID>

# Check usage
zfs list -o name,used,avail,refer,quota tank/users/<USER_ID>

# View files
ls -lh /tank/users/<USER_ID>/chunks/
```

### Check Database
```bash
# Open database
sqlite3 /tmp/zfscloud-test/zfscloud.db

# Queries
SELECT * FROM users;
SELECT * FROM registrations;
SELECT path, size FROM files LIMIT 10;
SELECT COUNT(*) FROM chunks;
```

### Check Server Status
```bash
# Server logs (Terminal where server is running)
# Watch for:
# - "Registration request received"
# - "Dataset created successfully"
# - "Chunk uploaded"

# Pool status
zpool status tank
zpool list tank
```

---

## 🐛 Debug Checklist

If something fails:

### 1. Server Connection
```bash
# Test connection
nc -zv 127.0.0.1 50052
timeout 2 bash -c "echo > /dev/tcp/127.0.0.1/50052"

# Check server is running
ps aux | grep zfscloud-server-test
```

### 2. Registration Issues
```bash
# Check pending registrations
./admin-helper.sh list-pending

# Check if already registered
sqlite3 /tmp/zfscloud-test/zfscloud.db "SELECT * FROM registrations;"
sqlite3 /tmp/zfscloud-test/zfscloud.db "SELECT * FROM users;"
```

### 3. Dataset Issues
```bash
# Check if dataset exists
zfs list tank/users/<USER_ID>

# Check encryption
zfs get encryption,keyformat,keystatus tank/users/<USER_ID>

# Check if key is loaded
zfs get keystatus tank/users/<USER_ID>
# Should show "available" when loaded
```

### 4. Upload Issues
```bash
# Check file exists
ls -lh /home/user/Téléchargements/FZXAC7RI

# Check permissions
stat /home/user/Téléchargements/FZXAC7RI

# Check quota
zfs get quota,used tank/users/<USER_ID>
```

### 5. Common Errors

**"authentication failed"**
- Registration not approved yet
- Run: `./admin-helper.sh list-pending` and approve

**"dataset key not loaded"**
- Need to load dataset key with passphrase
- Use passphrase from registration step

**"quota exceeded"**
- Increase quota when approving
- Or use: `./admin-helper.sh approve <id> 50` for 50GB

**"connection refused"**
- Check server is running on 127.0.0.1:50052
- Check config: `cat test-client-config-50052.yaml`

---

## ✅ Expected Test Results

After successful test:

1. **Client registered** → registration_id created
2. **Admin approved** → user_id assigned, dataset created
3. **Dataset created** → `tank/users/<USER_ID>` with AES-256-GCM
4. **Files uploaded** → chunked into 1MB pieces
5. **Database populated** → files, chunks, hashes recorded
6. **Verification works** → can list files from client

### Success Indicators
```bash
# Dataset exists and is encrypted
$ zfs list tank/users/<USER_ID>
NAME                    USED  AVAIL  REFER  MOUNTPOINT
tank/users/<USER_ID>    2.2G  7.8G   2.2G   /tank/users/<USER_ID>

$ zfs get encryption tank/users/<USER_ID>
NAME                    PROPERTY     VALUE        SOURCE
tank/users/<USER_ID>    encryption   aes-256-gcm  -

# Files in database
$ sqlite3 /tmp/zfscloud-test/zfscloud.db "SELECT COUNT(*) FROM files;"
12

# Files on disk
$ find /tank/users/<USER_ID> -type f | wc -l
# Should match uploaded file count
```

---

## 📊 Test Metrics

### Files to Upload
- Total files: 12
- Total size: ~2.2GB
- Largest file: 1.6GB (proxmox-ve_9.0-1.iso)
- Chunk size: 1MB
- Expected chunks: ~2,200

### Expected Performance
- Small files (<10MB): < 1 second
- Medium files (100MB): ~10-30 seconds
- Large files (1GB+): ~2-5 minutes
- Total upload time: ~5-10 minutes (depends on disk speed)

---

## 🎓 Test Workflow Summary

```
┌─────────────────────────────────────────────────┐
│ 1. Start Client (interactive mode)             │
│    ./bin/test-client -config ...                │
└──────────────────┬──────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────┐
│ 2. Register Client                              │
│    > register                                   │
│    SAVE PASSPHRASE!                             │
└──────────────────┬──────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────┐
│ 3. Approve Registration (admin)                 │
│    ./admin-helper.sh approve <id> 20            │
│    → Creates encrypted ZFS dataset              │
└──────────────────┬──────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────┐
│ 4. Verify Dataset Created                       │
│    zfs list tank/users/<USER_ID>                │
│    zfs get encryption,keystatus ...             │
└──────────────────┬──────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────┐
│ 5. Upload Files                                 │
│    > upload <file> <remote>                     │
│    > upload-dir /home/user/Téléchargements ...  │
└──────────────────┬──────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────┐
│ 6. Verify Upload                                │
│    > list                                       │
│    Check database & disk                        │
└─────────────────────────────────────────────────┘
```

---

## 🎯 Next Steps After Test

If test successful:
1. ✅ Client-server communication works
2. ✅ Encryption pipeline works
3. ✅ File upload/chunking works
4. ✅ ZFS integration works

You can then:
- Test file restore/download
- Test snapshot creation/restore
- Test with GUI client
- Test automatic backup engine
- Performance testing with larger datasets

---

**Ready to start testing!** Choose one of the 3 options above to begin.
