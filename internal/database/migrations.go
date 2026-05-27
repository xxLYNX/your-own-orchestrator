package database

import (
	"fmt"
	"sort"
	"time"
)

// Migration represents a database migration
type Migration struct {
	Version     int
	Name        string
	Description string
	Up          string // SQL for applying the migration
	Down        string // SQL for rolling back the migration
}

// MigrationRecord tracks applied migrations
type MigrationRecord struct {
	Version   int       `db:"version"`
	Name      string    `db:"name"`
	AppliedAt time.Time `db:"applied_at"`
}

// migrations holds all available migrations in order
var migrations = []Migration{
	{
		Version:     1,
		Name:        "add_templates",
		Description: "Add template system tables",
		Up: `
-- Template definitions table
CREATE TABLE IF NOT EXISTS templates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    version TEXT NOT NULL,
    description TEXT,
    category TEXT,
    definition TEXT NOT NULL,
    is_builtin BOOLEAN DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Note templates (links notes to templates)
CREATE TABLE IF NOT EXISTS note_templates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    note_id INTEGER NOT NULL,
    template_id INTEGER NOT NULL,
    template_data TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (note_id) REFERENCES notes(id) ON DELETE CASCADE,
    FOREIGN KEY (template_id) REFERENCES templates(id)
);

-- Step tracking for templated notes
CREATE TABLE IF NOT EXISTS note_steps (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    note_template_id INTEGER NOT NULL,
    step_number INTEGER NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    completed BOOLEAN DEFAULT 0,
    completed_at DATETIME,
    notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (note_template_id) REFERENCES note_templates(id) ON DELETE CASCADE
);

-- Artifacts (inputs/outputs)
CREATE TABLE IF NOT EXISTS artifacts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    note_template_id INTEGER NOT NULL,
    artifact_type TEXT NOT NULL,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    value TEXT NOT NULL,
    description TEXT,
    required BOOLEAN DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (note_template_id) REFERENCES note_templates(id) ON DELETE CASCADE
);

-- External references
CREATE TABLE IF NOT EXISTS note_references (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    note_id INTEGER NOT NULL,
    ref_type TEXT NOT NULL,
    ref_value TEXT NOT NULL,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (note_id) REFERENCES notes(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_note_templates_note ON note_templates(note_id);
CREATE INDEX IF NOT EXISTS idx_note_templates_template ON note_templates(template_id);
CREATE INDEX IF NOT EXISTS idx_note_steps_template ON note_steps(note_template_id);
CREATE INDEX IF NOT EXISTS idx_artifacts_template ON artifacts(note_template_id);
CREATE INDEX IF NOT EXISTS idx_artifacts_type ON artifacts(artifact_type);
CREATE INDEX IF NOT EXISTS idx_note_references_note ON note_references(note_id);
CREATE INDEX IF NOT EXISTS idx_templates_name ON templates(name);
CREATE INDEX IF NOT EXISTS idx_templates_category ON templates(category);

-- Add template support columns to notes table
ALTER TABLE notes ADD COLUMN is_templated BOOLEAN DEFAULT 0;
ALTER TABLE notes ADD COLUMN template_progress REAL DEFAULT 0.0;

-- Trigger to update templates timestamp
CREATE TRIGGER IF NOT EXISTS update_templates_timestamp
AFTER UPDATE ON templates
BEGIN
    UPDATE templates SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Trigger to update artifacts timestamp
CREATE TRIGGER IF NOT EXISTS update_artifacts_timestamp
AFTER UPDATE ON artifacts
BEGIN
    UPDATE artifacts SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
`,
		Down: `
-- Drop triggers
DROP TRIGGER IF EXISTS update_artifacts_timestamp;
DROP TRIGGER IF EXISTS update_templates_timestamp;

-- Drop indexes
DROP INDEX IF EXISTS idx_templates_category;
DROP INDEX IF EXISTS idx_templates_name;
DROP INDEX IF EXISTS idx_note_references_note;
DROP INDEX IF EXISTS idx_artifacts_type;
DROP INDEX IF EXISTS idx_artifacts_template;
DROP INDEX IF EXISTS idx_note_steps_template;
DROP INDEX IF EXISTS idx_note_templates_template;
DROP INDEX IF EXISTS idx_note_templates_note;

-- Drop tables (order matters due to foreign keys)
DROP TABLE IF EXISTS note_references;
DROP TABLE IF EXISTS artifacts;
DROP TABLE IF EXISTS note_steps;
DROP TABLE IF EXISTS note_templates;
DROP TABLE IF EXISTS templates;

-- Note: Cannot easily remove columns from notes table in SQLite
-- The is_templated and template_progress columns will remain
-- but will be unused if this migration is rolled back
`,
	},
	{
		Version:     2,
		Name:        "add_template_records",
		Description: "Add template records (log shape) for repeating structured data",
		Up: `
-- Template records table (log-shaped data)
-- This allows templates to define repeating structured records (like job applications, contacts, etc.)
CREATE TABLE IF NOT EXISTS template_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    note_template_id INTEGER NOT NULL,
    record_index INTEGER NOT NULL,
    data TEXT NOT NULL,
    status TEXT DEFAULT 'draft',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (note_template_id) REFERENCES note_templates(id) ON DELETE CASCADE,
    UNIQUE(note_template_id, record_index)
);

-- Indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_template_records_note_template ON template_records(note_template_id);
CREATE INDEX IF NOT EXISTS idx_template_records_status ON template_records(status);
CREATE INDEX IF NOT EXISTS idx_template_records_index ON template_records(note_template_id, record_index);

-- Trigger to update template_records timestamp
CREATE TRIGGER IF NOT EXISTS update_template_records_timestamp
AFTER UPDATE ON template_records
BEGIN
    UPDATE template_records SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
`,
		Down: `
-- Drop trigger
DROP TRIGGER IF EXISTS update_template_records_timestamp;

-- Drop indexes
DROP INDEX IF EXISTS idx_template_records_index;
DROP INDEX IF EXISTS idx_template_records_status;
DROP INDEX IF EXISTS idx_template_records_note_template;

-- Drop table
DROP TABLE IF EXISTS template_records;
`,
	},
}

// initMigrationsTable creates the migrations tracking table
func (db *DB) initMigrationsTable() error {
	schema := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := db.conn.Exec(schema)
	return err
}

// getAppliedMigrations returns a list of applied migration versions
func (db *DB) getAppliedMigrations() (map[int]bool, error) {
	applied := make(map[int]bool)

	rows, err := db.conn.Query("SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

// applyMigration applies a single migration
func (db *DB) applyMigration(migration Migration) error {
	// Start transaction
	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute migration SQL
	if _, err := tx.Exec(migration.Up); err != nil {
		return fmt.Errorf("failed to execute migration %d (%s): %w", migration.Version, migration.Name, err)
	}

	// Record migration
	_, err = tx.Exec(
		"INSERT INTO schema_migrations (version, name, applied_at) VALUES (?, ?, ?)",
		migration.Version,
		migration.Name,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration %d: %w", migration.Version, err)
	}

	return nil
}

// rollbackMigration rolls back a single migration
func (db *DB) rollbackMigration(migration Migration) error {
	// Start transaction
	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute rollback SQL
	if _, err := tx.Exec(migration.Down); err != nil {
		return fmt.Errorf("failed to rollback migration %d (%s): %w", migration.Version, migration.Name, err)
	}

	// Remove migration record
	_, err = tx.Exec("DELETE FROM schema_migrations WHERE version = ?", migration.Version)
	if err != nil {
		return fmt.Errorf("failed to remove migration record %d: %w", migration.Version, err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rollback %d: %w", migration.Version, err)
	}

	return nil
}

// RunMigrations applies all pending migrations
func (db *DB) RunMigrations() error {
	// Initialize migrations table
	if err := db.initMigrationsTable(); err != nil {
		return fmt.Errorf("failed to initialize migrations table: %w", err)
	}

	// Get applied migrations
	applied, err := db.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	// Apply pending migrations
	appliedCount := 0
	for _, migration := range migrations {
		if applied[migration.Version] {
			continue // Already applied
		}

		fmt.Printf("Applying migration %d: %s...\n", migration.Version, migration.Name)
		if err := db.applyMigration(migration); err != nil {
			return fmt.Errorf("migration %d failed: %w", migration.Version, err)
		}
		fmt.Printf("✓ Migration %d applied successfully\n", migration.Version)
		appliedCount++
	}

	if appliedCount == 0 {
		fmt.Println("No pending migrations")
	} else {
		fmt.Printf("Applied %d migration(s)\n", appliedCount)
	}

	return nil
}

// RollbackMigration rolls back the most recent migration
func (db *DB) RollbackMigration() error {
	// Get applied migrations
	applied, err := db.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	if len(applied) == 0 {
		return fmt.Errorf("no migrations to rollback")
	}

	// Find the highest version that's been applied
	var maxVersion int
	for version := range applied {
		if version > maxVersion {
			maxVersion = version
		}
	}

	// Find the migration
	var migration *Migration
	for i := range migrations {
		if migrations[i].Version == maxVersion {
			migration = &migrations[i]
			break
		}
	}

	if migration == nil {
		return fmt.Errorf("migration %d not found", maxVersion)
	}

	fmt.Printf("Rolling back migration %d: %s...\n", migration.Version, migration.Name)
	if err := db.rollbackMigration(*migration); err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}
	fmt.Printf("✓ Migration %d rolled back successfully\n", migration.Version)

	return nil
}

// GetMigrationStatus returns the current migration status
func (db *DB) GetMigrationStatus() ([]MigrationStatus, error) {
	// Initialize migrations table if it doesn't exist
	if err := db.initMigrationsTable(); err != nil {
		return nil, err
	}

	applied, err := db.getAppliedMigrations()
	if err != nil {
		return nil, err
	}

	var statuses []MigrationStatus
	for _, migration := range migrations {
		status := MigrationStatus{
			Version:     migration.Version,
			Name:        migration.Name,
			Description: migration.Description,
			Applied:     applied[migration.Version],
		}

		if status.Applied {
			// Get applied time
			var appliedAt time.Time
			err := db.conn.QueryRow(
				"SELECT applied_at FROM schema_migrations WHERE version = ?",
				migration.Version,
			).Scan(&appliedAt)
			if err == nil {
				status.AppliedAt = &appliedAt
			}
		}

		statuses = append(statuses, status)
	}

	return statuses, nil
}

// MigrationStatus represents the status of a migration
type MigrationStatus struct {
	Version     int
	Name        string
	Description string
	Applied     bool
	AppliedAt   *time.Time
}
