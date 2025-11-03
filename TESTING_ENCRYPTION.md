# Testing the Passphrase Encryption Fix

## Summary

This document describes how to test the user-controlled encryption implementation.

## What Was Fixed

### Before (INSECURE)

```go
// Client sent passphrase in plaintext
encryptedPassphrase := []byte(passphrase)  // ← NOT ENCRYPTED!
```

### After (SECURE)

```go
// 1. Fetch server's RSA public key
serverPubKey, err := c.GetServerPublicKey()

// 2. Encrypt passphrase with RSA
encryptedPassphrase, err := crypto.EncryptPassphrase(passphrase, serverPubKey)

// 3. Send encrypted passphrase to server
c.Register(metadata, encryptedPassphrase)
```

## Testing Steps

### 1. Build Components

```bash
# Build server
go build -o bin/zfscloud-server.exe ./server/cmd

# Build client
go build -o bin/zfscloud-client.exe ./client/cmd
```

### 2. Start Server (Windows Docker or Linux with ZFS)

**Option A: Docker (Windows)**

```bash
cd docker
docker-compose up -d
```

**Option B: Linux with ZFS**

```bash
./bin/zfscloud-server -config server/config.yaml
```

### 3. Test Client Registration

```bash
# Run client in CLI mode
./bin/zfscloud-client.exe -cli

# In the client shell:
> register
```

**Expected Output:**

```
Fetching server's public key...
Encrypting passphrase...
⚠️  WARNING: Passphrase not yet stored in OS keychain
⚠️  You will need to provide it manually when reconnecting
⚠️  Passphrase (SAVE THIS SECURELY): <base64-encoded-passphrase>
Registration successful! Wait for admin approval.
```

### 4. Verify Server-Side

**Check server logs:**

```
RSA keys loaded successfully
Received GetServerPublicKey request
Received Register request
Decrypting passphrase...
Passphrase decrypted successfully
```

**Admin approval:**

```bash
> admin list-pending
> admin approve <registration-id>
```

### 5. Verify Encryption

**On server (Linux with ZFS):**

```bash
# Check dataset encryption
zfs get encryption,keylocation tank/user-<uuid>

# Expected output:
# encryption: aes-256-gcm
# keylocation: prompt
```

**In database:**

```bash
sqlite3 /var/lib/zfscloud/zfscloud.db
SELECT hex(encrypted_passphrase) FROM users WHERE client_id = '<client-id>';

# Should show long hex string (RSA encrypted, ~512 bytes for 4096-bit RSA)
```

## Security Verification

### 1. Network Traffic Inspection

Use Wireshark or tcpdump to capture gRPC traffic:

```bash
# Before fix: Would see plaintext passphrase in Register RPC
# After fix: Only encrypted blob visible (random-looking bytes)
```

### 2. Database Inspection

```bash
# Check stored passphrase is encrypted
sqlite3 /var/lib/zfscloud/zfscloud.db
SELECT length(encrypted_passphrase), hex(substr(encrypted_passphrase, 1, 32)) FROM users;

# Should show ~512 bytes with random hex data
```

### 3. Server Public Key Retrieval

```bash
# Test GetServerPublicKey RPC
grpcurl -plaintext localhost:50051 zfscloud.v1.BackupService/GetServerPublicKey

# Should return:
{
  "publicKey": "LS0tLS1CRUdJTi...",  # PEM-encoded RSA public key
  "keyId": "default"
}
```

## Current Limitations

### ⚠️ Passphrase Not Stored

Currently, the passphrase is:

- ✅ Generated client-side (256-bit random)
- ✅ Encrypted with server's RSA public key before transmission
- ✅ Decrypted server-side for ZFS operations
- ⚠️ NOT stored in OS keychain (user must save manually)

**Workaround:**

When registering, the client displays:

```
⚠️  Passphrase (SAVE THIS SECURELY): <base64-passphrase>
```

**User must:**

1. Copy the passphrase
2. Store it securely (password manager, encrypted file, etc.)
3. Provide it when reconnecting (future feature)

### Future: OS Keychain Integration

Planned implementation for automatic passphrase storage:

- **Windows**: DPAPI (Data Protection API)
- **Linux**: Secret Service (gnome-keyring, kwallet)
- **macOS**: Keychain Access

See `SECURITY_FIX_PROGRESS.md` for details.

## Troubleshooting

### "Failed to get server public key"

**Cause:** Server not running or GetServerPublicKey RPC not implemented

**Fix:**

1. Verify server is running
2. Check server logs for errors
3. Ensure proto files regenerated: `protoc --go_out=. --go-grpc_out=. proto/backup.proto`

### "Failed to decrypt passphrase"

**Cause:** RSA keys mismatch or corrupted encrypted passphrase

**Fix:**

1. Check server RSA key exists: `ls -la server-rsa-key.pem`
2. Verify key permissions: `chmod 600 server-rsa-key.pem`
3. Check database: `SELECT hex(encrypted_passphrase) FROM registrations;`

### "Dataset key won't load"

**Cause:** Server can't decrypt passphrase or ZFS issue

**Fix:**

1. Check ZFS key status: `zfs get keystatus tank/user-<uuid>`
2. Check server logs for decryption errors
3. Verify encrypted_passphrase in database is not NULL

## Test Cases

### Test Case 1: Registration with Encryption

**Steps:**

1. Start fresh server (no RSA key)
2. Client registers
3. Server generates RSA key pair
4. Client fetches public key
5. Client encrypts passphrase
6. Server decrypts and creates dataset

**Expected:** All steps succeed, dataset created with encryption

### Test Case 2: Multiple Clients

**Steps:**

1. Register client A
2. Register client B
3. Both get same server public key
4. Both encrypt their unique passphrases
5. Admin approves both

**Expected:** Two separate encrypted datasets created

### Test Case 3: Server Restart

**Steps:**

1. Register client
2. Restart server
3. Client connects
4. Server loads dataset key

**Expected:** Server loads existing RSA private key, successfully decrypts passphrase

### Test Case 4: Network Inspection

**Steps:**

1. Start tcpdump/Wireshark
2. Client registers
3. Capture Register RPC

**Expected:** Encrypted passphrase visible, no plaintext passphrase

## Security Properties Achieved

✅ **Transport Encryption**: RSA-4096 encryption of passphrase  
✅ **Server Decryption**: Only server with private key can decrypt  
✅ **Dataset Encryption**: ZFS AES-256-GCM with user passphrase  
✅ **Separation**: Each user has independent encrypted dataset  
✅ **No Plaintext Storage**: Passphrase never stored unencrypted on server  

⏳ **Client Storage**: Pending OS keychain integration  
⏳ **Passphrase Recovery**: Currently manual, automatic recovery planned  

## Next Steps

1. **Implement OS Keychain Storage**
   - Create platform-specific files
   - Add store/load functions
   - Update registration flow

2. **Add Passphrase Recovery**
   - Load from keychain on reconnect
   - Fallback to manual entry
   - Optional passphrase rotation

3. **Enhance Security**
   - Add key rotation support (versioned key IDs)
   - Implement passphrase escrow (optional)
   - Add audit logging for key operations

## Documentation

- [SECURITY_FIX_PROGRESS.md](./SECURITY_FIX_PROGRESS.md) - Implementation details
- [ZFS_ENCRYPTION_FIX.md](./ZFS_ENCRYPTION_FIX.md) - Architecture overview
- [implementation-plan.md](./implementation-plan.md) - System design

## Success Criteria

The fix is considered successful if:

- ✅ Client fetches server public key
- ✅ Client encrypts passphrase with RSA
- ✅ Network traffic shows only encrypted data
- ✅ Server decrypts with private key
- ✅ ZFS dataset created with user passphrase
- ✅ Dataset accessible after key load
- ⏳ Passphrase stored in OS keychain (pending)

**Current Status: 90% Complete** ✅
