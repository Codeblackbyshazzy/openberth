// Package store owns all persistence for the server. Every table has a
// corresponding file (users.go, deployments.go, secrets.go, sessions.go,
// oauth.go, login_codes.go, bandwidth.go, settings.go). This file holds
// only the connection, schema migrations, and lifecycle helpers.
package store

import (
	"database/sql"
	"os"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)")
	if err != nil {
		return nil, err
	}
	s := &Store{db: db}
	if err := s.ensureSchema(); err != nil {
		return nil, err
	}
	return s, nil
}

// ensureSchema creates all tables and runs idempotent migrations.
// Called on initial open and after restore to ensure the schema is up to date.
func (s *Store) ensureSchema() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			api_key TEXT UNIQUE NOT NULL,
			role TEXT DEFAULT 'user',
			max_deployments INTEGER DEFAULT 10,
			default_ttl_hours INTEGER DEFAULT 72,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS deployments (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			name TEXT NOT NULL,
			subdomain TEXT UNIQUE NOT NULL,
			framework TEXT,
			container_id TEXT,
			port INTEGER,
			status TEXT DEFAULT 'building',
			ttl_hours INTEGER,
			env_json TEXT DEFAULT '{}',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			expires_at DATETIME,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);
		CREATE TABLE IF NOT EXISTS oauth_clients (
			client_id TEXT PRIMARY KEY,
			client_secret TEXT DEFAULT '',
			redirect_uris TEXT DEFAULT '[]',
			client_name TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS oauth_codes (
			code TEXT PRIMARY KEY,
			client_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			redirect_uri TEXT NOT NULL,
			code_challenge TEXT NOT NULL,
			expires_at DATETIME NOT NULL,
			used INTEGER DEFAULT 0
		);
		CREATE TABLE IF NOT EXISTS oauth_tokens (
			token TEXT PRIMARY KEY,
			token_type TEXT NOT NULL,
			client_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			expires_at DATETIME NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			token TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			expires_at DATETIME NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);
		CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS login_codes (
			code TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			callback_url TEXT NOT NULL,
			expires_at DATETIME NOT NULL,
			used INTEGER DEFAULT 0
		);
	`)
	if err != nil {
		return err
	}

	// Migrations (idempotent — fails silently if columns already exist)
	for _, stmt := range []string{
		"ALTER TABLE deployments ADD COLUMN access_mode TEXT DEFAULT 'public'",
		"ALTER TABLE deployments ADD COLUMN access_user TEXT DEFAULT ''",
		"ALTER TABLE deployments ADD COLUMN access_hash TEXT DEFAULT ''",
		"ALTER TABLE deployments ADD COLUMN title TEXT DEFAULT ''",
		"ALTER TABLE deployments ADD COLUMN description TEXT DEFAULT ''",
		"ALTER TABLE users ADD COLUMN display_name TEXT DEFAULT ''",
		"ALTER TABLE users ADD COLUMN password_hash TEXT DEFAULT ''",
		"ALTER TABLE deployments ADD COLUMN mode TEXT DEFAULT 'deploy'",
		"ALTER TABLE deployments ADD COLUMN network_quota TEXT DEFAULT ''",
		"ALTER TABLE deployments ADD COLUMN access_users TEXT DEFAULT ''",
		"ALTER TABLE deployments ADD COLUMN locked INTEGER DEFAULT 0",
		"ALTER TABLE deployments ADD COLUMN memory TEXT DEFAULT ''",
		"ALTER TABLE deployments ADD COLUMN cpus TEXT DEFAULT ''",
	} {
		s.db.Exec(stmt)
	}

	s.db.Exec(`CREATE TABLE IF NOT EXISTS bandwidth_usage (
		deployment_id TEXT NOT NULL,
		period_start TEXT NOT NULL,
		bytes_out INTEGER DEFAULT 0,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (deployment_id, period_start)
	)`)

	s.db.Exec(`CREATE TABLE IF NOT EXISTS secrets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT REFERENCES users(id),
		scope TEXT DEFAULT 'user',
		name TEXT NOT NULL,
		description TEXT DEFAULT '',
		encrypted_dek BLOB NOT NULL,
		dek_nonce BLOB NOT NULL,
		ciphertext BLOB NOT NULL,
		value_nonce BLOB NOT NULL,
		created_by TEXT REFERENCES users(id),
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(user_id, name)
	)`)
	s.db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_secrets_global_name ON secrets(name) WHERE user_id IS NULL`)

	// Migration: add secrets_json to deployments
	s.db.Exec("ALTER TABLE deployments ADD COLUMN secrets_json TEXT DEFAULT '[]'")

	// Migration: add created_by to secrets (pre-existing rows stay NULL → admin-only edit).
	s.db.Exec("ALTER TABLE secrets ADD COLUMN created_by TEXT REFERENCES users(id)")

	return nil
}

func (s *Store) Close() {
	s.db.Close()
}

// Checkpoint flushes the WAL into the main database file.
// Call this before reading the DB file for backup.
func (s *Store) Checkpoint() error {
	_, err := s.db.Exec("PRAGMA wal_checkpoint(TRUNCATE)")
	return err
}

// Reopen closes the current database connection and opens a new one.
// Used after restoring a backup to pick up the new database file.
// Removes stale WAL/SHM files to prevent corruption when the DB file
// has been replaced, and runs schema migrations for compatibility.
func (s *Store) Reopen(dbPath string) error {
	s.db.Close()

	// Remove stale WAL/SHM files left from the previous DB.
	// After a restore, these belong to the OLD database and would corrupt
	// the newly restored one if SQLite tried to replay them.
	os.Remove(dbPath + "-wal")
	os.Remove(dbPath + "-shm")

	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)")
	if err != nil {
		return err
	}
	s.db = db
	return s.ensureSchema()
}
