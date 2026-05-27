package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"yoo/internal/config"

	_ "modernc.org/sqlite"
)

// DB holds the database connection
type DB struct {
	conn *sql.DB
}

// New creates a new database connection and initializes the schema
func New(dbPath string) (*DB, error) {
	// Ensure the directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{conn: conn}

	// Initialize schema
	if err := db.initSchema(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return db, nil
}

// initSchema creates the necessary tables if they don't exist
func (db *DB) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS notes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		description TEXT,
		scheduled_at DATETIME NOT NULL,
		status TEXT DEFAULT 'pending',
		priority INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_notes_scheduled_at ON notes(scheduled_at);
	CREATE INDEX IF NOT EXISTS idx_notes_status ON notes(status);

	CREATE TRIGGER IF NOT EXISTS update_notes_timestamp
	AFTER UPDATE ON notes
	BEGIN
		UPDATE notes SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
	END;
	`

	_, err := db.conn.Exec(schema)
	return err
}

// Close closes the database connection
func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// Conn returns the underlying database connection
func (db *DB) Conn() *sql.DB {
	return db.conn
}

// Init initializes the database using the configured path
func Init() (*sql.DB, error) {
	dbPath := config.GetDatabasePath()
	db, err := New(dbPath)
	if err != nil {
		return nil, err
	}
	return db.Conn(), nil
}

// InitDB initializes the database (alias for Init for backward compatibility)
func InitDB() (*sql.DB, error) {
	return Init()
}
