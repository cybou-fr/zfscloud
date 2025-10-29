package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Database represents the application database
type Database struct {
	conn *sql.DB
}

// New creates a new database connection
func New(dbPath string) (*Database, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys
	if _, err := conn.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	db := &Database{conn: conn}

	// Initialize schema
	if err := db.initSchema(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return db, nil
}

// Close closes the database connection
func (db *Database) Close() error {
	return db.conn.Close()
}

// initSchema creates all necessary tables
func (db *Database) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT UNIQUE NOT NULL,
		client_id TEXT UNIQUE NOT NULL,
		public_key BLOB NOT NULL,
		encrypted_passphrase BLOB NOT NULL,
		zfs_dataset TEXT NOT NULL,
		quota_bytes INTEGER NOT NULL,
		used_bytes INTEGER DEFAULT 0,
		revoked BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		last_seen TIMESTAMP,
		
		-- Metadata
		hostname TEXT,
		username TEXT,
		os TEXT,
		client_version TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_users_client_id ON users(client_id);
	CREATE INDEX IF NOT EXISTS idx_users_user_id ON users(user_id);

	CREATE TABLE IF NOT EXISTS registrations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		registration_id TEXT UNIQUE NOT NULL,
		client_id TEXT UNIQUE NOT NULL,
		public_key BLOB NOT NULL,
		encrypted_passphrase BLOB NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending', -- pending, approved, rejected
		
		-- Metadata
		hostname TEXT,
		username TEXT,
		os TEXT,
		client_version TEXT,
		
		requested_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		processed_at TIMESTAMP,
		processed_by TEXT,
		rejection_reason TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_registrations_status ON registrations(status);
	CREATE INDEX IF NOT EXISTS idx_registrations_client_id ON registrations(client_id);

	CREATE TABLE IF NOT EXISTS sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT UNIQUE NOT NULL,
		user_id TEXT NOT NULL,
		client_id TEXT NOT NULL,
		dataset_key_loaded BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		last_heartbeat TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		expires_at TIMESTAMP NOT NULL,
		
		FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);

	CREATE TABLE IF NOT EXISTS restore_requests (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		request_id TEXT UNIQUE NOT NULL,
		user_id TEXT NOT NULL,
		client_id TEXT NOT NULL,
		snapshot_id TEXT NOT NULL,
		restore_path TEXT,
		reason TEXT,
		status TEXT NOT NULL DEFAULT 'pending', -- pending, approved, rejected, completed
		
		requested_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		processed_at TIMESTAMP,
		processed_by TEXT,
		completed_at TIMESTAMP,
		rejection_reason TEXT,
		
		FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_restore_requests_user_id ON restore_requests(user_id);
	CREATE INDEX IF NOT EXISTS idx_restore_requests_status ON restore_requests(status);

	CREATE TABLE IF NOT EXISTS snapshots (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		snapshot_id TEXT UNIQUE NOT NULL,
		user_id TEXT NOT NULL,
		zfs_snapshot_name TEXT NOT NULL,
		size_bytes INTEGER DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		
		FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_snapshots_user_id ON snapshots(user_id);
	CREATE INDEX IF NOT EXISTS idx_snapshots_created_at ON snapshots(created_at);

	CREATE TABLE IF NOT EXISTS nonces (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		nonce TEXT UNIQUE NOT NULL,
		client_id TEXT NOT NULL,
		timestamp INTEGER NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_nonces_nonce ON nonces(nonce);
	CREATE INDEX IF NOT EXISTS idx_nonces_created_at ON nonces(created_at);

	CREATE TABLE IF NOT EXISTS audit_log (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT,
		client_id TEXT,
		action TEXT NOT NULL,
		details TEXT,
		ip_address TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_audit_log_user_id ON audit_log(user_id);
	CREATE INDEX IF NOT EXISTS idx_audit_log_created_at ON audit_log(created_at);
	`

	_, err := db.conn.Exec(schema)
	return err
}

// CleanupExpiredSessions removes expired sessions
func (db *Database) CleanupExpiredSessions() error {
	_, err := db.conn.Exec(`
		DELETE FROM sessions 
		WHERE expires_at < datetime('now')
	`)
	return err
}

// CleanupOldNonces removes nonces older than 10 minutes
func (db *Database) CleanupOldNonces() error {
	_, err := db.conn.Exec(`
		DELETE FROM nonces 
		WHERE created_at < datetime('now', '-10 minutes')
	`)
	return err
}

// User represents a registered user
type User struct {
	ID                  int64
	UserID              string
	ClientID            string
	PublicKey           []byte
	EncryptedPassphrase []byte
	ZFSDataset          string
	QuotaBytes          int64
	UsedBytes           int64
	Revoked             bool
	CreatedAt           time.Time
	LastSeen            *time.Time
	Hostname            string
	Username            string
	OS                  string
	ClientVersion       string
}

// Registration represents a pending registration
type Registration struct {
	ID                  int64
	RegistrationID      string
	ClientID            string
	PublicKey           []byte
	EncryptedPassphrase []byte
	Status              string
	Hostname            string
	Username            string
	OS                  string
	ClientVersion       string
	RequestedAt         time.Time
	ProcessedAt         *time.Time
	ProcessedBy         string
	RejectionReason     string
}

// Session represents an active client session
type Session struct {
	ID               int64
	SessionID        string
	UserID           string
	ClientID         string
	DatasetKeyLoaded bool
	CreatedAt        time.Time
	LastHeartbeat    time.Time
	ExpiresAt        time.Time
}

// RestoreRequest represents a snapshot restore request
type RestoreRequest struct {
	ID              int64
	RequestID       string
	UserID          string
	ClientID        string
	SnapshotID      string
	RestorePath     string
	Reason          string
	Status          string
	RequestedAt     time.Time
	ProcessedAt     *time.Time
	ProcessedBy     string
	CompletedAt     *time.Time
	RejectionReason string
}

// Snapshot represents a ZFS snapshot
type Snapshot struct {
	ID              int64
	SnapshotID      string
	UserID          string
	ZFSSnapshotName string
	SizeBytes       int64
	CreatedAt       time.Time
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID        int64
	UserID    string
	ClientID  string
	Action    string
	Details   string
	IPAddress string
	CreatedAt time.Time
}
