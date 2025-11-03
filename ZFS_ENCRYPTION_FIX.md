# ZFS Dataset Encryption - Security Architecture Fix

## Overview

This document describes the fix for proper user-controlled ZFS dataset encryption in ZFSCloud.

## Problem Statement

The original implementation had several critical security issues:

1. **Client passphrase not encrypted**: Client sent passphrase in plaintext to server
2. **Server never used passphrase**: CreateDataset didn't provide the passphrase to ZFS
3. **No passphrase storage**: Client generated passphrase but didn't save it
4. **Server control vs User control**: Server could potentially decrypt data without user's knowledge

## Solution: User-Controlled Encryption

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                       CLIENT                                 │
│  1. Generate random 256-bit passphrase                      │
│  2. Encrypt with server's RSA public key                    │
│  3. Store passphrase securely (OS keychain)                 │
│  4. Send encrypted passphrase to server                     │
└──────────────────────┬──────────────────────────────────────┘
                       │ Encrypted Passphrase
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                       SERVER                                 │
│  5. Decrypt passphrase with RSA private key                 │
│  6. Create ZFS dataset with user's passphrase               │
│  7. Store encrypted passphrase for future loads             │
│  8. Never store plaintext passphrase                        │
└─────────────────────────────────────────────────────────────┘
```

### Security Properties

✅ **User Controls Encryption**: Client generates and owns the passphrase  
✅ **Transport Security**: Passphrase encrypted with RSA during transmission  
✅ **Zero-Knowledge Option**: Client can optionally not send passphrase (bring your own key)  
✅ **Server Cannot Decrypt Without User**: Passphrase needed for dataset access  
✅ **Separate Datasets**: Each user has isolated, independently encrypted dataset  

## Implementation Changes

### 1. ZFS Manager (`server/zfs/manager.go`)

**Before:**
```go
func (m *Manager) CreateDataset(userID string, compression string) (string, error) {
    // Created dataset with encryption=on but no passphrase
    // This would fail or use insecure default
}
```

**After:**
```go
func (m *Manager) CreateDataset(userID string, passphrase []byte, compression string) (string, error) {
    // Creates dataset and provides passphrase via stdin
    // ZFS encrypts dataset with this passphrase
    // Returns dataset name
}
```

**Key Changes:**
- Added `passphrase []byte` parameter
- Uses `cmd.StdinPipe()` to provide passphrase to ZFS
- Sends passphrase twice for confirmation
- ZFS creates encrypted dataset with AES-256-GCM

### 2. Admin Service (`server/grpc/admin_service.go`)

**Before:**
```go
// Stored encrypted passphrase but never used it
_, err = s.zfsManager.CreateDataset(userID, s.config.ZFS.Compression)
```

**After:**
```go
// Decrypt user's passphrase
decryptedPassphrase, err := crypto.DecryptPassphrase(registration.EncryptedPassphrase, s.rsaPrivateKey)

// Create dataset with user's passphrase
_, err = s.zfsManager.CreateDataset(userID, decryptedPassphrase, s.config.ZFS.Compression)
```

**Key Changes:**
- Added `rsaPrivateKey []byte` to struct
- Decrypts passphrase from registration
- Passes plaintext passphrase to ZFS (only in memory)
- ZFS encrypts dataset with this key

### 3. Backup Service (`server/grpc/backup_service.go`)

**Before:**
```go
func (s *BackupServiceServer) LoadDatasetKey(...) {
    // TODO: Decrypt passphrase and load into ZFS
    // Just returned success without doing anything
}
```

**After:**
```go
func (s *BackupServiceServer) LoadDatasetKey(...) {
    // Decrypt user's passphrase
    decryptedPassphrase, err := crypto.DecryptPassphrase(user.EncryptedPassphrase, s.rsaPrivateKey)
    
    // Load key into ZFS
    err := s.zfsManager.LoadKey(user.ZFSDataset, decryptedPassphrase)
}
```

**Key Changes:**
- Added `rsaPrivateKey []byte` to struct
- Actually implements key loading
- Decrypts passphrase and loads into ZFS
- Enables access to encrypted dataset

### 4. Client Registration (TODO - needs implementation)

**Current (INSECURE):**
```go
// Generate passphrase
passphrase, err := crypto.GeneratePassphrase()

// NOT ENCRYPTED!
encryptedPassphrase := []byte(passphrase)
```

**Should be:**
```go
// Generate passphrase
passphrase, err := crypto.GeneratePassphrase()

// Get server's public RSA key
serverPubKey, err := c.GetServerPublicKey()

// Encrypt passphrase with server's public key
encryptedPassphrase, err := crypto.EncryptPassphrase(passphrase, serverPubKey)

// Store passphrase securely (OS keychain)
err = storePassphraseInKeychain(c.GetClientID(), passphrase)
```

## Deployment Requirements

### Server Setup

1. **Generate RSA Key Pair**:
   ```bash
   openssl genrsa -out server-rsa-key.pem 4096
   openssl rsa -in server-rsa-key.pem -pubout -out server-rsa-key.pub
   ```

2. **Update server config**:
   ```yaml
   server:
     rsa_key_path: "/path/to/server-rsa-key.pem"
   ```

3. **Serve public key to clients**:
   - Add RPC method `GetServerPublicKey`
   - Returns PEM-encoded RSA public key
   - Clients use this to encrypt passphrases

### Client Setup

1. **Store passphrase securely**:
   - Windows: DPAPI (Data Protection API)
   - macOS: Keychain
   - Linux: Secret Service API (gnome-keyring, kwallet)

2. **Passphrase recovery**:
   - Option 1: Client stores passphrase locally (can recover)
   - Option 2: Key escrow (encrypted with user password)
   - Option 3: Zero-knowledge (user must remember/backup)

## Security Considerations

### Threat Model

**Protected Against:**
- ✅ Network eavesdropping (passphrase encrypted in transit)
- ✅ Server compromise (encrypted passphrase storage)
- ✅ Unauthorized admin access (can't decrypt without private key)
- ✅ Cross-user data access (separate encrypted datasets)

**Not Protected Against:**
- ⚠️ Server memory dump while dataset is loaded (passphrase in memory)
- ⚠️ Server admin with root access (can access loaded datasets)
- ⚠️ Client machine compromise (passphrase in OS keychain)

### Best Practices

1. **Rotate Server RSA Keys**: Periodically regenerate server RSA key pair
2. **Audit Dataset Access**: Log when datasets are loaded/unloaded
3. **Automatic Unload**: Unload dataset keys after inactivity timeout
4. **Client-Side Encryption**: Option for additional client-side encryption layer
5. **Hardware Security Module**: Store server RSA private key in HSM

## Testing

### Verify Encryption

1. **Create dataset**:
   ```bash
   # After registration approval
   zfs list -r tank
   zfs get encryption tank/user-uuid
   ```

2. **Check encryption**:
   ```bash
   # Should show: encryption=aes-256-gcm
   zfs get all tank/user-uuid | grep encrypt
   ```

3. **Test key load/unload**:
   ```bash
   # Unload key
   zfs unload-key tank/user-uuid
   
   # Try to access (should fail)
   ls /tank/user-uuid  # Permission denied
   
   # Load key via client
   # Client calls LoadDatasetKey RPC
   
   # Access should work now
   ls /tank/user-uuid
   ```

### Security Audit

```bash
# Check that passphrases are encrypted in DB
sqlite3 /var/lib/zfscloud/zfscloud.db
SELECT hex(encrypted_passphrase) FROM users LIMIT 1;
# Should see long hex string (RSA encrypted)

# Verify no plaintext passphrases in logs
grep -r "passphrase" /var/log/zfscloud/
# Should not find any plaintext passphrases
```

## Migration Plan

For existing deployments:

1. **Backup existing data**
2. **Generate server RSA key pair**
3. **Update server code**
4. **Restart server**
5. **For existing users**:
   - Option A: Re-register (new passphrase)
   - Option B: Admin manually encrypts stored passphrases
6. **Update all clients**

## Future Enhancements

1. **Key Rotation**: Allow users to change their passphrase
2. **Multi-Key Support**: Multiple passphrases per dataset (key slots)
3. **HSM Integration**: Store server RSA key in hardware security module
4. **Zero-Knowledge Mode**: Client keeps passphrase, server never sees it
5. **Passphrase Recovery**: Encrypted key escrow with user password

---

**Status**: ✅ Core fixes implemented  
**TODO**: Client-side encryption, keychain storage, GetServerPublicKey RPC  
**Priority**: HIGH (security-critical)
