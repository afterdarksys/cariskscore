package database

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
	mu   sync.RWMutex
}

// New opens or creates a new SQLite database
func New(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		conn.Close()
		return nil, err
	}

	return db, nil
}

// migrate creates necessary tables if they don't exist
func (db *DB) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS certificate_authorities (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		common_name TEXT,
		organization TEXT,
		country TEXT,
		certificate_sha256 TEXT,
		source TEXT,
		trusted_by_mozilla BOOLEAN DEFAULT 0,
		trusted_by_microsoft BOOLEAN DEFAULT 0,
		trusted_by_chrome BOOLEAN DEFAULT 0,
		ev_capable BOOLEAN DEFAULT 0,
		indexed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS risk_factors (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ca_id INTEGER NOT NULL,
		risk_type TEXT NOT NULL,
		severity TEXT NOT NULL,
		description TEXT,
		source TEXT,
		detected_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		resolved_at TIMESTAMP,
		FOREIGN KEY (ca_id) REFERENCES certificate_authorities(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS ca_scores (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ca_id INTEGER NOT NULL UNIQUE,
		security_score REAL DEFAULT 0,
		audit_score REAL DEFAULT 0,
		incident_score REAL DEFAULT 0,
		compliance_score REAL DEFAULT 0,
		overall_score REAL DEFAULT 0,
		rank INTEGER,
		computed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (ca_id) REFERENCES certificate_authorities(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS ct_log_entries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		serial_number TEXT NOT NULL,
		issuer_ca_name TEXT,
		common_name TEXT,
		not_before TIMESTAMP,
		not_after TIMESTAMP,
		logged_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		certificate_pem TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_cas_name ON certificate_authorities(name);
	CREATE INDEX IF NOT EXISTS idx_risk_ca_id ON risk_factors(ca_id);
	CREATE INDEX IF NOT EXISTS idx_scores_ca_id ON ca_scores(ca_id);
	CREATE INDEX IF NOT EXISTS idx_ct_issuer ON ct_log_entries(issuer_ca_name);
	`

	_, err := db.conn.Exec(schema)
	return err
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// Conn returns the underlying database connection for queries
func (db *DB) Conn() *sql.DB {
	return db.conn
}
