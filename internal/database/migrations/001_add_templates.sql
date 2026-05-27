-- Migration: 001_add_templates
-- Description: Add template system tables and modify notes table for template support
-- Created: 2026-05-27

-- ============================================================================
-- UP Migration: Create template system
-- ============================================================================

-- Templates table: Store template definitions
CREATE TABLE IF NOT EXISTS templates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    version TEXT NOT NULL,
    description TEXT,
    category TEXT,
    definition TEXT NOT NULL,  -- JSON serialized TemplateDefinition
    is_builtin BOOLEAN DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Note templates table: Link notes to templates with instance data
CREATE TABLE IF NOT EXISTS note_templates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    note_id INTEGER NOT NULL,
    template_id INTEGER NOT NULL,
    template_data TEXT NOT NULL,  -- JSON serialized TemplateInstance
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (note_id) REFERENCES notes(id) ON DELETE CASCADE,
    FOREIGN KEY (template_id) REFERENCES templates(id) ON DELETE RESTRICT
);

-- Note steps table: Track step completion for templated notes
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

-- Artifacts table: Store inputs and outputs for templated notes
CREATE TABLE IF NOT EXISTS artifacts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    note_template_id INTEGER NOT NULL,
    artifact_type TEXT NOT NULL CHECK(artifact_type IN ('input', 'output')),
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK(type IN ('file', 'url', 'text', 'folder')),
    value TEXT NOT NULL,
    description TEXT,
    required BOOLEAN DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (note_template_id) REFERENCES note_templates(id) ON DELETE CASCADE
);

-- Note references table: External references for any note
CREATE TABLE IF NOT EXISTS note_references (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    note_id INTEGER NOT NULL,
    ref_type TEXT NOT NULL CHECK(ref_type IN ('file', 'url', 'command', 'note')),
    ref_value TEXT NOT NULL,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (note_id) REFERENCES notes(id) ON DELETE CASCADE
);

-- Modify notes table to support templates
ALTER TABLE notes ADD COLUMN is_templated BOOLEAN DEFAULT 0;
ALTER TABLE notes ADD COLUMN template_progress REAL DEFAULT 0.0;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_templates_name ON templates(name);
CREATE INDEX IF NOT EXISTS idx_templates_category ON templates(category);
CREATE INDEX IF NOT EXISTS idx_templates_builtin ON templates(is_builtin);

CREATE INDEX IF NOT EXISTS idx_note_templates_note ON note_templates(note_id);
CREATE INDEX IF NOT EXISTS idx_note_templates_template ON note_templates(template_id);

CREATE INDEX IF NOT EXISTS idx_note_steps_template ON note_steps(note_template_id);
CREATE INDEX IF NOT EXISTS idx_note_steps_number ON note_steps(note_template_id, step_number);
CREATE INDEX IF NOT EXISTS idx_note_steps_completed ON note_steps(completed);

CREATE INDEX IF NOT EXISTS idx_artifacts_template ON artifacts(note_template_id);
CREATE INDEX IF NOT EXISTS idx_artifacts_type ON artifacts(artifact_type);
CREATE INDEX IF NOT EXISTS idx_artifacts_name ON artifacts(note_template_id, name);

CREATE INDEX IF NOT EXISTS idx_note_references_note ON note_references(note_id);
CREATE INDEX IF NOT EXISTS idx_note_references_type ON note_references(ref_type);

-- Create triggers for automatic timestamp updates
CREATE TRIGGER IF NOT EXISTS update_templates_timestamp
AFTER UPDATE ON templates
BEGIN
    UPDATE templates SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_artifacts_timestamp
AFTER UPDATE ON artifacts
BEGIN
    UPDATE artifacts SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Create trigger to update note template_progress when steps are completed
CREATE TRIGGER IF NOT EXISTS update_note_progress_on_step_complete
AFTER UPDATE OF completed ON note_steps
WHEN NEW.completed = 1
BEGIN
    UPDATE notes
    SET template_progress = (
        SELECT CAST(SUM(CASE WHEN completed = 1 THEN 1 ELSE 0 END) AS REAL) / COUNT(*)
        FROM note_steps
        WHERE note_template_id IN (
            SELECT id FROM note_templates WHERE note_id = (
                SELECT note_id FROM note_templates WHERE id = NEW.note_template_id
            )
        )
    )
    WHERE id IN (
        SELECT note_id FROM note_templates WHERE id = NEW.note_template_id
    );
END;

-- ============================================================================
-- DOWN Migration: Rollback template system
-- ============================================================================

-- Note: To rollback this migration, uncomment and run the following:

-- DROP TRIGGER IF EXISTS update_note_progress_on_step_complete;
-- DROP TRIGGER IF EXISTS update_artifacts_timestamp;
-- DROP TRIGGER IF EXISTS update_templates_timestamp;
--
-- DROP INDEX IF EXISTS idx_note_references_type;
-- DROP INDEX IF EXISTS idx_note_references_note;
-- DROP INDEX IF EXISTS idx_artifacts_name;
-- DROP INDEX IF EXISTS idx_artifacts_type;
-- DROP INDEX IF EXISTS idx_artifacts_template;
-- DROP INDEX IF EXISTS idx_note_steps_completed;
-- DROP INDEX IF EXISTS idx_note_steps_number;
-- DROP INDEX IF EXISTS idx_note_steps_template;
-- DROP INDEX IF EXISTS idx_note_templates_template;
-- DROP INDEX IF EXISTS idx_note_templates_note;
-- DROP INDEX IF EXISTS idx_templates_builtin;
-- DROP INDEX IF EXISTS idx_templates_category;
-- DROP INDEX IF EXISTS idx_templates_name;
--
-- DROP TABLE IF EXISTS note_references;
-- DROP TABLE IF EXISTS artifacts;
-- DROP TABLE IF EXISTS note_steps;
-- DROP TABLE IF EXISTS note_templates;
-- DROP TABLE IF EXISTS templates;
--
-- ALTER TABLE notes DROP COLUMN template_progress;
-- ALTER TABLE notes DROP COLUMN is_templated;

-- ============================================================================
-- Migration Complete
-- ============================================================================
