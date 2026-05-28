package database

const schemaSQL = `
CREATE TABLE IF NOT EXISTS notes (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT NOT NULL,
	description TEXT,
	scheduled_at DATETIME NOT NULL,
	status TEXT DEFAULT 'pending',
	priority INTEGER DEFAULT 0,
	is_templated BOOLEAN DEFAULT 0,
	template_progress REAL DEFAULT 0.0,
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

CREATE INDEX IF NOT EXISTS idx_templates_name ON templates(name);
CREATE INDEX IF NOT EXISTS idx_templates_category ON templates(category);

CREATE TRIGGER IF NOT EXISTS update_templates_timestamp
AFTER UPDATE ON templates
BEGIN
	UPDATE templates SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TABLE IF NOT EXISTS note_templates (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	note_id INTEGER NOT NULL,
	template_id INTEGER NOT NULL,
	template_data TEXT NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (note_id) REFERENCES notes(id) ON DELETE CASCADE,
	FOREIGN KEY (template_id) REFERENCES templates(id)
);

CREATE INDEX IF NOT EXISTS idx_note_templates_note ON note_templates(note_id);
CREATE INDEX IF NOT EXISTS idx_note_templates_template ON note_templates(template_id);

CREATE TABLE IF NOT EXISTS note_shape_state (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	note_template_id INTEGER NOT NULL,
	shape_path TEXT NOT NULL,
	repeat_stack_json TEXT NOT NULL DEFAULT '[]',
	kind TEXT NOT NULL,
	title TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT 'not_started',
	completed BOOLEAN DEFAULT 0,
	completed_at DATETIME,
	notes TEXT,
	data TEXT NOT NULL DEFAULT '{}',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (note_template_id) REFERENCES note_templates(id) ON DELETE CASCADE,
	UNIQUE(note_template_id, shape_path, repeat_stack_json)
);

CREATE INDEX IF NOT EXISTS idx_note_shape_state_template ON note_shape_state(note_template_id);
CREATE INDEX IF NOT EXISTS idx_note_shape_state_stack ON note_shape_state(note_template_id, repeat_stack_json);
CREATE INDEX IF NOT EXISTS idx_note_shape_state_path ON note_shape_state(note_template_id, shape_path);

CREATE TRIGGER IF NOT EXISTS update_note_shape_state_timestamp
AFTER UPDATE ON note_shape_state
BEGIN
	UPDATE note_shape_state SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TABLE IF NOT EXISTS template_records (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	note_template_id INTEGER NOT NULL,
	repeat_stack_json TEXT NOT NULL DEFAULT '[]',
	record_index INTEGER NOT NULL,
	data TEXT NOT NULL,
	status TEXT DEFAULT 'draft',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (note_template_id) REFERENCES note_templates(id) ON DELETE CASCADE,
	UNIQUE(note_template_id, repeat_stack_json, record_index)
);

CREATE INDEX IF NOT EXISTS idx_template_records_note_template ON template_records(note_template_id);
CREATE INDEX IF NOT EXISTS idx_template_records_status ON template_records(status);
CREATE INDEX IF NOT EXISTS idx_template_records_index ON template_records(note_template_id, repeat_stack_json, record_index);

CREATE TRIGGER IF NOT EXISTS update_template_records_timestamp
AFTER UPDATE ON template_records
BEGIN
	UPDATE template_records SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

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

CREATE INDEX IF NOT EXISTS idx_artifacts_template ON artifacts(note_template_id);
CREATE INDEX IF NOT EXISTS idx_artifacts_type ON artifacts(artifact_type);

CREATE TRIGGER IF NOT EXISTS update_artifacts_timestamp
AFTER UPDATE ON artifacts
BEGIN
	UPDATE artifacts SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

`
