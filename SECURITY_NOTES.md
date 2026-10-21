# Credential safety and local keys

ZFSCloud is a development project; the repository does not provide evidence of independent production security qualification.

Two historical local key files (`client-keys.dat` and `test-client-keys.json`) were removed from the current Git tree as a precaution. They were not established to be live credentials. **They still exist in the public Git history.** If any key was used beyond local testing, rotate/revoke it and generate new keys. Do not treat deleting a file in a new commit as scrubbing the history.

The old Linux/macOS fallback in `client/keychain_*.go` writes passphrase bytes to `storeEncryptedDataToFile` without encrypting them; the function name does not make the data encrypted. A separately guarded keychain change is required; **never store a real passphrase in this fallback**. If you previously used it, preserve the dataset recovery key and migrate it before deleting any local legacy data.

The standalone root Go utilities are for explicit `go run file.go`, not part of normal module compilation. Their build tags are being corrected separately. This note is not a security audit.
