package db

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

// ============================================================================
// User Operations
// ============================================================================

// CreateUser creates a new user in the database
func (db *Database) CreateUser(user *User) error {
	query := `
		INSERT INTO users (
			user_id, client_id, public_key, encrypted_passphrase, zfs_dataset,
			quota_bytes, hostname, username, os, client_version
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.conn.Exec(query,
		user.UserID, user.ClientID, user.PublicKey, user.EncryptedPassphrase,
		user.ZFSDataset, user.QuotaBytes, user.Hostname, user.Username,
		user.OS, user.ClientVersion,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	user.ID = id

	return nil
}

// GetUserByClientID retrieves a user by client ID
func (db *Database) GetUserByClientID(clientID string) (*User, error) {
	query := `
		SELECT id, user_id, client_id, public_key, encrypted_passphrase, zfs_dataset,
			quota_bytes, used_bytes, revoked, created_at, last_seen,
			hostname, username, os, client_version
		FROM users WHERE client_id = ?
	`

	user := &User{}
	err := db.conn.QueryRow(query, clientID).Scan(
		&user.ID, &user.UserID, &user.ClientID, &user.PublicKey,
		&user.EncryptedPassphrase, &user.ZFSDataset, &user.QuotaBytes,
		&user.UsedBytes, &user.Revoked, &user.CreatedAt, &user.LastSeen,
		&user.Hostname, &user.Username, &user.OS, &user.ClientVersion,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetUserByUserID retrieves a user by user ID
func (db *Database) GetUserByUserID(userID string) (*User, error) {
	query := `
		SELECT id, user_id, client_id, public_key, encrypted_passphrase, zfs_dataset,
			quota_bytes, used_bytes, revoked, created_at, last_seen,
			hostname, username, os, client_version
		FROM users WHERE user_id = ?
	`

	user := &User{}
	err := db.conn.QueryRow(query, userID).Scan(
		&user.ID, &user.UserID, &user.ClientID, &user.PublicKey,
		&user.EncryptedPassphrase, &user.ZFSDataset, &user.QuotaBytes,
		&user.UsedBytes, &user.Revoked, &user.CreatedAt, &user.LastSeen,
		&user.Hostname, &user.Username, &user.OS, &user.ClientVersion,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// ListUsers retrieves all users
func (db *Database) ListUsers(includeRevoked bool) ([]*User, error) {
	query := `
		SELECT id, user_id, client_id, public_key, encrypted_passphrase, zfs_dataset,
			quota_bytes, used_bytes, revoked, created_at, last_seen,
			hostname, username, os, client_version
		FROM users
	`

	if !includeRevoked {
		query += " WHERE revoked = FALSE"
	}

	query += " ORDER BY created_at DESC"

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		err := rows.Scan(
			&user.ID, &user.UserID, &user.ClientID, &user.PublicKey,
			&user.EncryptedPassphrase, &user.ZFSDataset, &user.QuotaBytes,
			&user.UsedBytes, &user.Revoked, &user.CreatedAt, &user.LastSeen,
			&user.Hostname, &user.Username, &user.OS, &user.ClientVersion,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

// UpdateUserLastSeen updates the last seen timestamp
func (db *Database) UpdateUserLastSeen(userID string) error {
	_, err := db.conn.Exec(`
		UPDATE users SET last_seen = CURRENT_TIMESTAMP WHERE user_id = ?
	`, userID)
	return err
}

// UpdateUserUsedSpace updates the used space for a user
func (db *Database) UpdateUserUsedSpace(userID string, usedBytes int64) error {
	_, err := db.conn.Exec(`
		UPDATE users SET used_bytes = ? WHERE user_id = ?
	`, usedBytes, userID)
	return err
}

// RevokeUser revokes a user's access
func (db *Database) RevokeUser(userID string) error {
	_, err := db.conn.Exec(`
		UPDATE users SET revoked = TRUE WHERE user_id = ?
	`, userID)
	return err
}

// ============================================================================
// Registration Operations
// ============================================================================

// CreateRegistration creates a new registration request
func (db *Database) CreateRegistration(reg *Registration) error {
	if reg.RegistrationID == "" {
		reg.RegistrationID = uuid.New().String()
	}

	query := `
		INSERT INTO registrations (
			registration_id, client_id, public_key, encrypted_passphrase,
			hostname, username, os, client_version
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.conn.Exec(query,
		reg.RegistrationID, reg.ClientID, reg.PublicKey, reg.EncryptedPassphrase,
		reg.Hostname, reg.Username, reg.OS, reg.ClientVersion,
	)
	if err != nil {
		return fmt.Errorf("failed to create registration: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	reg.ID = id

	return nil
}

// GetRegistrationByID retrieves a registration by ID
func (db *Database) GetRegistrationByID(registrationID string) (*Registration, error) {
	query := `
		SELECT id, registration_id, client_id, public_key, encrypted_passphrase,
			status, hostname, username, os, client_version,
			requested_at, processed_at, processed_by, rejection_reason
		FROM registrations WHERE registration_id = ?
	`

	reg := &Registration{}
	err := db.conn.QueryRow(query, registrationID).Scan(
		&reg.ID, &reg.RegistrationID, &reg.ClientID, &reg.PublicKey,
		&reg.EncryptedPassphrase, &reg.Status, &reg.Hostname, &reg.Username,
		&reg.OS, &reg.ClientVersion, &reg.RequestedAt, &reg.ProcessedAt,
		&reg.ProcessedBy, &reg.RejectionReason,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get registration: %w", err)
	}

	return reg, nil
}

// GetRegistrationByClientID retrieves a registration by client ID
func (db *Database) GetRegistrationByClientID(clientID string) (*Registration, error) {
	query := `
		SELECT id, registration_id, client_id, public_key, encrypted_passphrase,
			status, hostname, username, os, client_version,
			requested_at, processed_at, processed_by, rejection_reason
		FROM registrations WHERE client_id = ?
	`

	reg := &Registration{}
	err := db.conn.QueryRow(query, clientID).Scan(
		&reg.ID, &reg.RegistrationID, &reg.ClientID, &reg.PublicKey,
		&reg.EncryptedPassphrase, &reg.Status, &reg.Hostname, &reg.Username,
		&reg.OS, &reg.ClientVersion, &reg.RequestedAt, &reg.ProcessedAt,
		&reg.ProcessedBy, &reg.RejectionReason,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get registration: %w", err)
	}

	return reg, nil
}

// ListPendingRegistrations retrieves all pending registrations
func (db *Database) ListPendingRegistrations() ([]*Registration, error) {
	query := `
		SELECT id, registration_id, client_id, public_key, encrypted_passphrase,
			status, hostname, username, os, client_version,
			requested_at, processed_at, processed_by, rejection_reason
		FROM registrations WHERE status = 'pending'
		ORDER BY requested_at ASC
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending registrations: %w", err)
	}
	defer rows.Close()

	var registrations []*Registration
	for rows.Next() {
		reg := &Registration{}
		err := rows.Scan(
			&reg.ID, &reg.RegistrationID, &reg.ClientID, &reg.PublicKey,
			&reg.EncryptedPassphrase, &reg.Status, &reg.Hostname, &reg.Username,
			&reg.OS, &reg.ClientVersion, &reg.RequestedAt, &reg.ProcessedAt,
			&reg.ProcessedBy, &reg.RejectionReason,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan registration: %w", err)
		}
		registrations = append(registrations, reg)
	}

	return registrations, rows.Err()
}

// ApproveRegistration approves a registration
func (db *Database) ApproveRegistration(registrationID, processedBy string) error {
	_, err := db.conn.Exec(`
		UPDATE registrations 
		SET status = 'approved', processed_at = CURRENT_TIMESTAMP, processed_by = ?
		WHERE registration_id = ?
	`, processedBy, registrationID)
	return err
}

// RejectRegistration rejects a registration
func (db *Database) RejectRegistration(registrationID, processedBy, reason string) error {
	_, err := db.conn.Exec(`
		UPDATE registrations 
		SET status = 'rejected', processed_at = CURRENT_TIMESTAMP, 
			processed_by = ?, rejection_reason = ?
		WHERE registration_id = ?
	`, processedBy, reason, registrationID)
	return err
}

// ============================================================================
// Session Operations
// ============================================================================

// CreateSession creates a new session
func (db *Database) CreateSession(session *Session) error {
	if session.SessionID == "" {
		session.SessionID = uuid.New().String()
	}

	query := `
		INSERT INTO sessions (session_id, user_id, client_id, expires_at)
		VALUES (?, ?, ?, ?)
	`

	result, err := db.conn.Exec(query,
		session.SessionID, session.UserID, session.ClientID, session.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	session.ID = id

	return nil
}

// GetSessionByID retrieves a session by ID
func (db *Database) GetSessionByID(sessionID string) (*Session, error) {
	query := `
		SELECT id, session_id, user_id, client_id, dataset_key_loaded,
			created_at, last_heartbeat, expires_at
		FROM sessions WHERE session_id = ?
	`

	session := &Session{}
	err := db.conn.QueryRow(query, sessionID).Scan(
		&session.ID, &session.SessionID, &session.UserID, &session.ClientID,
		&session.DatasetKeyLoaded, &session.CreatedAt, &session.LastHeartbeat,
		&session.ExpiresAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return session, nil
}

// UpdateSessionHeartbeat updates the last heartbeat time
func (db *Database) UpdateSessionHeartbeat(sessionID string) error {
	_, err := db.conn.Exec(`
		UPDATE sessions SET last_heartbeat = CURRENT_TIMESTAMP WHERE session_id = ?
	`, sessionID)
	return err
}

// SetSessionKeyLoaded sets the dataset key loaded flag
func (db *Database) SetSessionKeyLoaded(sessionID string, loaded bool) error {
	_, err := db.conn.Exec(`
		UPDATE sessions SET dataset_key_loaded = ? WHERE session_id = ?
	`, loaded, sessionID)
	return err
}

// DeleteSession deletes a session
func (db *Database) DeleteSession(sessionID string) error {
	_, err := db.conn.Exec(`DELETE FROM sessions WHERE session_id = ?`, sessionID)
	return err
}

// ============================================================================
// Restore Request Operations
// ============================================================================

// CreateRestoreRequest creates a new restore request
func (db *Database) CreateRestoreRequest(req *RestoreRequest) error {
	if req.RequestID == "" {
		req.RequestID = uuid.New().String()
	}

	query := `
		INSERT INTO restore_requests (
			request_id, user_id, client_id, snapshot_id, restore_path, reason
		) VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := db.conn.Exec(query,
		req.RequestID, req.UserID, req.ClientID, req.SnapshotID,
		req.RestorePath, req.Reason,
	)
	if err != nil {
		return fmt.Errorf("failed to create restore request: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	req.ID = id

	return nil
}

// GetRestoreRequestByID retrieves a restore request by ID
func (db *Database) GetRestoreRequestByID(requestID string) (*RestoreRequest, error) {
	query := `
		SELECT id, request_id, user_id, client_id, snapshot_id, restore_path,
			reason, status, requested_at, processed_at, processed_by,
			completed_at, rejection_reason
		FROM restore_requests WHERE request_id = ?
	`

	req := &RestoreRequest{}
	err := db.conn.QueryRow(query, requestID).Scan(
		&req.ID, &req.RequestID, &req.UserID, &req.ClientID, &req.SnapshotID,
		&req.RestorePath, &req.Reason, &req.Status, &req.RequestedAt,
		&req.ProcessedAt, &req.ProcessedBy, &req.CompletedAt, &req.RejectionReason,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get restore request: %w", err)
	}

	return req, nil
}

// ListPendingRestoreRequests retrieves all pending restore requests
func (db *Database) ListPendingRestoreRequests() ([]*RestoreRequest, error) {
	query := `
		SELECT id, request_id, user_id, client_id, snapshot_id, restore_path,
			reason, status, requested_at, processed_at, processed_by,
			completed_at, rejection_reason
		FROM restore_requests WHERE status = 'pending'
		ORDER BY requested_at ASC
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending restore requests: %w", err)
	}
	defer rows.Close()

	var requests []*RestoreRequest
	for rows.Next() {
		req := &RestoreRequest{}
		err := rows.Scan(
			&req.ID, &req.RequestID, &req.UserID, &req.ClientID, &req.SnapshotID,
			&req.RestorePath, &req.Reason, &req.Status, &req.RequestedAt,
			&req.ProcessedAt, &req.ProcessedBy, &req.CompletedAt, &req.RejectionReason,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan restore request: %w", err)
		}
		requests = append(requests, req)
	}

	return requests, rows.Err()
}

// ApproveRestoreRequest approves a restore request
func (db *Database) ApproveRestoreRequest(requestID, processedBy string) error {
	_, err := db.conn.Exec(`
		UPDATE restore_requests 
		SET status = 'approved', processed_at = CURRENT_TIMESTAMP, processed_by = ?
		WHERE request_id = ?
	`, processedBy, requestID)
	return err
}

// RejectRestoreRequest rejects a restore request
func (db *Database) RejectRestoreRequest(requestID, processedBy, reason string) error {
	_, err := db.conn.Exec(`
		UPDATE restore_requests 
		SET status = 'rejected', processed_at = CURRENT_TIMESTAMP, 
			processed_by = ?, rejection_reason = ?
		WHERE request_id = ?
	`, processedBy, reason, requestID)
	return err
}

// CompleteRestoreRequest marks a restore request as completed
func (db *Database) CompleteRestoreRequest(requestID string) error {
	_, err := db.conn.Exec(`
		UPDATE restore_requests 
		SET status = 'completed', completed_at = CURRENT_TIMESTAMP
		WHERE request_id = ?
	`, requestID)
	return err
}

// ============================================================================
// Snapshot Operations
// ============================================================================

// CreateSnapshot creates a new snapshot record
func (db *Database) CreateSnapshot(snapshot *Snapshot) error {
	if snapshot.SnapshotID == "" {
		snapshot.SnapshotID = uuid.New().String()
	}

	query := `
		INSERT INTO snapshots (snapshot_id, user_id, zfs_snapshot_name, size_bytes)
		VALUES (?, ?, ?, ?)
	`

	result, err := db.conn.Exec(query,
		snapshot.SnapshotID, snapshot.UserID, snapshot.ZFSSnapshotName, snapshot.SizeBytes,
	)
	if err != nil {
		return fmt.Errorf("failed to create snapshot: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	snapshot.ID = id

	return nil
}

// ListSnapshotsByUserID retrieves all snapshots for a user
func (db *Database) ListSnapshotsByUserID(userID string) ([]*Snapshot, error) {
	query := `
		SELECT id, snapshot_id, user_id, zfs_snapshot_name, size_bytes, created_at
		FROM snapshots WHERE user_id = ?
		ORDER BY created_at DESC
	`

	rows, err := db.conn.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list snapshots: %w", err)
	}
	defer rows.Close()

	var snapshots []*Snapshot
	for rows.Next() {
		snapshot := &Snapshot{}
		err := rows.Scan(
			&snapshot.ID, &snapshot.SnapshotID, &snapshot.UserID,
			&snapshot.ZFSSnapshotName, &snapshot.SizeBytes, &snapshot.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan snapshot: %w", err)
		}
		snapshots = append(snapshots, snapshot)
	}

	return snapshots, rows.Err()
}

// DeleteSnapshot deletes a snapshot record
func (db *Database) DeleteSnapshot(snapshotID string) error {
	_, err := db.conn.Exec(`DELETE FROM snapshots WHERE snapshot_id = ?`, snapshotID)
	return err
}

// ============================================================================
// Nonce Operations (for replay protection)
// ============================================================================

// CheckAndStoreNonce checks if a nonce has been used and stores it
func (db *Database) CheckAndStoreNonce(nonce, clientID string, timestamp int64) (bool, error) {
	// Check if nonce exists
	var count int
	err := db.conn.QueryRow(`SELECT COUNT(*) FROM nonces WHERE nonce = ?`, nonce).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check nonce: %w", err)
	}

	if count > 0 {
		return false, nil // Nonce already used
	}

	// Store nonce
	_, err = db.conn.Exec(`
		INSERT INTO nonces (nonce, client_id, timestamp) VALUES (?, ?, ?)
	`, nonce, clientID, timestamp)
	if err != nil {
		return false, fmt.Errorf("failed to store nonce: %w", err)
	}

	return true, nil
}

// ============================================================================
// Audit Log Operations
// ============================================================================

// LogAction logs an action to the audit log
func (db *Database) LogAction(userID, clientID, action, details, ipAddress string) error {
	_, err := db.conn.Exec(`
		INSERT INTO audit_log (user_id, client_id, action, details, ip_address)
		VALUES (?, ?, ?, ?, ?)
	`, userID, clientID, action, details, ipAddress)
	return err
}

// GetAuditLogs retrieves audit logs
func (db *Database) GetAuditLogs(limit int) ([]*AuditLog, error) {
	query := `
		SELECT id, user_id, client_id, action, details, ip_address, created_at
		FROM audit_log
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := db.conn.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs: %w", err)
	}
	defer rows.Close()

	var logs []*AuditLog
	for rows.Next() {
		log := &AuditLog{}
		err := rows.Scan(
			&log.ID, &log.UserID, &log.ClientID, &log.Action,
			&log.Details, &log.IPAddress, &log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, rows.Err()
}
