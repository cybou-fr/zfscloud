#!/bin/bash
# Quick Manual Test Guide - Step by Step
# For debugging client-server interaction

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

cat << 'EOF'
================================================================================
  ZFSCloud Client-Server Manual Test Guide
================================================================================

PREREQUISITES:
✓ Server running on 127.0.0.1:50052
✓ Directory: /home/user/Téléchargements (with files to upload)

================================================================================
STEP 1: START CLIENT
================================================================================

Run the client in interactive mode:

    ./bin/test-client -config test-client-config-50052.yaml -interactive

You should see:
    ZFSCloud Client v0.2.0
    Client ID: <uuid>
    Connecting to server at 127.0.0.1:50052...
    Connected successfully!
    zfscloud>

================================================================================
STEP 2: REGISTER CLIENT
================================================================================

At the prompt, type:

    register

Expected output:
    Registering with server...
    Fetching server's public key...
    Encrypting passphrase...
    ⚠️  WARNING: Passphrase not yet stored in OS keychain
    ⚠️  Passphrase (SAVE THIS): <base64-string>
    Registration successful! Wait for admin approval.

IMPORTANT: Copy and save the passphrase!

================================================================================
STEP 3: APPROVE REGISTRATION (New Terminal)
================================================================================

Open a NEW terminal and run:

    # List pending registrations
    grpcurl -plaintext -d '{"admin_token":"test-admin-token-12345"}' \
        127.0.0.1:50052 zfscloud.v1.AdminService/ListPendingRegistrations

Copy the registration_id from output, then approve:

    # Replace <REG_ID> with actual registration_id
    grpcurl -plaintext -d '{"admin_token":"test-admin-token-12345","registration_id":"<REG_ID>","quota_bytes":10737418240}' \
        127.0.0.1:50052 zfscloud.v1.AdminService/ApproveRegistration

Expected: {"success":true, "user_id":"<USER_ID>"}

================================================================================
STEP 4: VERIFY ZFS DATASET (Terminal 2)
================================================================================

Check that the encrypted dataset was created:

    zfs list -r tank/users

You should see:
    NAME                    USED  AVAIL  REFER  MOUNTPOINT
    tank/users/<USER_ID>    ...   ...    ...    /tank/users/<USER_ID>

Check encryption status:

    zfs get encryption,keystatus tank/users/<USER_ID>

Expected:
    encryption  aes-256-gcm
    keystatus   unavailable    # Key not loaded yet

================================================================================
STEP 5: SEND HEARTBEAT (Back to Client Terminal)
================================================================================

At the client prompt:

    heartbeat

Expected:
    Heartbeat sent successfully!

================================================================================
STEP 6: CHECK STATUS
================================================================================

    status

Expected to show:
    - Client ID
    - Server address
    - Connection status
    - Session info

================================================================================
STEP 7: UPLOAD A SINGLE FILE
================================================================================

Upload a small file first to test:

    upload /home/user/Téléchargements/FZXAC7RI test-file.txt

Expected:
    Uploading /home/user/Téléchargements/FZXAC7RI to test-file.txt...
    Upload progress: 100%
    Upload complete!
    Hash verified successfully

DEBUG: If upload fails, check:
    1. Is the dataset key loaded? (keystatus should be "available")
    2. Is the file path correct?
    3. Check server logs for errors

================================================================================
STEP 8: UPLOAD ENTIRE DIRECTORY
================================================================================

Upload all files from Téléchargements:

    upload-dir /home/user/Téléchargements telechargements

Expected:
    Uploading directory /home/user/Téléchargements to telechargements...
    Uploading: <file1> (progress shown)
    Uploading: <file2> (progress shown)
    ...
    Directory upload complete!

NOTE: This will take time depending on file sizes (you have ~2.2GB)

================================================================================
STEP 9: LIST UPLOADED FILES
================================================================================

    list

Expected:
    Files in backup:
    - telechargements/FZXAC7RI
    - telechargements/burpsuite_community_linux_v2025_10_1.sh
    - telechargements/code_1.105.1-1760482543_amd64.deb
    ...

Or list specific path:

    list telechargements

================================================================================
STEP 10: VERIFY ON SERVER (Terminal 2)
================================================================================

Check database:

    sqlite3 /tmp/zfscloud-test/zfscloud.db "SELECT path, size FROM files LIMIT 10;"

Check files on disk:

    ls -lh /tank/users/<USER_ID>/chunks/

Count files:

    find /tank/users/<USER_ID> -type f | wc -l

Check dataset usage:

    zfs list tank/users/<USER_ID>

================================================================================
STEP 11: LIST SNAPSHOTS
================================================================================

Back in client:

    snapshots

Expected (after hourly snapshot runs):
    Available snapshots:
    - <snapshot-name> (created: <timestamp>)

================================================================================
DEBUGGING TIPS
================================================================================

If something fails:

1. CHECK SERVER LOGS:
   Look at the terminal where server is running for error messages

2. CHECK CLIENT CONNECTION:
   - Verify server address in config: 127.0.0.1:50052
   - Test with: nc -zv 127.0.0.1 50052

3. CHECK ZFS:
   - Pool exists: zpool list tank
   - Datasets: zfs list -r tank
   - Encryption: zfs get encryption,keystatus tank/users/<USER_ID>

4. CHECK DATABASE:
   sqlite3 /tmp/zfscloud-test/zfscloud.db
   > .tables
   > SELECT * FROM users;
   > SELECT * FROM files;

5. CHECK PERMISSIONS:
   - Can client write to test-data/keys/?
   - Can server write to /tmp/zfscloud-test/?
   - Can server access ZFS pool?

6. COMMON ISSUES:
   
   ERROR: "failed to load dataset key"
   FIX: Make sure you used the correct passphrase from registration
   
   ERROR: "quota exceeded"
   FIX: Increase quota when approving registration
   
   ERROR: "authentication failed"
   FIX: Check that registration was approved
   
   ERROR: "file not found"
   FIX: Check file path, use absolute paths

================================================================================
EXPECTED RESULTS
================================================================================

After successful test:

✓ Client registered and approved
✓ ZFS encrypted dataset created: tank/users/<USER_ID>
✓ Dataset encrypted with AES-256-GCM
✓ Files uploaded and chunked (1MB chunks)
✓ Files stored in: /tank/users/<USER_ID>/chunks/
✓ File metadata in database
✓ Can list files from client
✓ Automatic snapshots scheduled

================================================================================

EOF

echo ""
echo -e "${GREEN}Quick Start:${NC}"
echo -e "${CYAN}  ./bin/test-client -config test-client-config-50052.yaml -interactive${NC}"
echo ""
