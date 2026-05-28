package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"yoo/internal/models"
)

// CreateTemplate inserts a new template into the database
func CreateTemplate(db *sql.DB, template *models.Template) error {
	// Marshal definition to JSON
	definitionJSON, err := json.Marshal(template.Definition)
	if err != nil {
		return fmt.Errorf("failed to marshal template definition: %w", err)
	}

	// Validate template definition
	if err := template.Definition.Validate(); err != nil {
		return fmt.Errorf("template validation failed: %w", err)
	}

	query := `
		INSERT INTO templates (name, version, description, category, definition, is_builtin, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	result, err := db.Exec(query,
		template.Name,
		template.Version,
		template.Description,
		template.Category,
		string(definitionJSON),
		template.IsBuiltin,
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}

	template.ID, err = result.LastInsertId()
	template.CreatedAt = now
	template.UpdatedAt = now
	return err
}

// GetTemplateByID retrieves a template by its ID
func GetTemplateByID(db *sql.DB, id int64) (*models.Template, error) {
	query := `
		SELECT id, name, version, description, category, definition, is_builtin, created_at, updated_at
		FROM templates
		WHERE id = ?
	`

	template := &models.Template{}
	var definitionJSON string

	err := db.QueryRow(query, id).Scan(
		&template.ID,
		&template.Name,
		&template.Version,
		&template.Description,
		&template.Category,
		&definitionJSON,
		&template.IsBuiltin,
		&template.CreatedAt,
		&template.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Unmarshal definition
	if err := json.Unmarshal([]byte(definitionJSON), &template.Definition); err != nil {
		return nil, fmt.Errorf("failed to unmarshal template definition: %w", err)
	}

	return template, nil
}

// GetTemplateByName retrieves a template by its name
func GetTemplateByName(db *sql.DB, name string) (*models.Template, error) {
	query := `
		SELECT id, name, version, description, category, definition, is_builtin, created_at, updated_at
		FROM templates
		WHERE name = ?
	`

	template := &models.Template{}
	var definitionJSON string

	err := db.QueryRow(query, name).Scan(
		&template.ID,
		&template.Name,
		&template.Version,
		&template.Description,
		&template.Category,
		&definitionJSON,
		&template.IsBuiltin,
		&template.CreatedAt,
		&template.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Unmarshal definition
	if err := json.Unmarshal([]byte(definitionJSON), &template.Definition); err != nil {
		return nil, fmt.Errorf("failed to unmarshal template definition: %w", err)
	}

	return template, nil
}

// ListTemplates retrieves all templates, optionally filtered by category
func ListTemplates(db *sql.DB, category string) ([]*models.Template, error) {
	var query string
	var args []interface{}

	if category != "" {
		query = `
			SELECT id, name, version, description, category, definition, is_builtin, created_at, updated_at
			FROM templates
			WHERE category = ?
			ORDER BY is_builtin DESC, name ASC
		`
		args = append(args, category)
	} else {
		query = `
			SELECT id, name, version, description, category, definition, is_builtin, created_at, updated_at
			FROM templates
			ORDER BY is_builtin DESC, name ASC
		`
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var templates []*models.Template
	for rows.Next() {
		template := &models.Template{}
		var definitionJSON string

		err := rows.Scan(
			&template.ID,
			&template.Name,
			&template.Version,
			&template.Description,
			&template.Category,
			&definitionJSON,
			&template.IsBuiltin,
			&template.CreatedAt,
			&template.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Unmarshal definition
		if err := json.Unmarshal([]byte(definitionJSON), &template.Definition); err != nil {
			return nil, fmt.Errorf("failed to unmarshal template definition: %w", err)
		}

		templates = append(templates, template)
	}

	return templates, rows.Err()
}

// UpdateTemplate updates an existing template
func UpdateTemplate(db *sql.DB, template *models.Template) error {
	// Validate template definition
	if err := template.Definition.Validate(); err != nil {
		return fmt.Errorf("template validation failed: %w", err)
	}

	// Marshal definition to JSON
	definitionJSON, err := json.Marshal(template.Definition)
	if err != nil {
		return fmt.Errorf("failed to marshal template definition: %w", err)
	}

	query := `
		UPDATE templates
		SET name = ?, version = ?, description = ?, category = ?, definition = ?, updated_at = ?
		WHERE id = ?
	`

	template.UpdatedAt = time.Now()

	_, err = db.Exec(query,
		template.Name,
		template.Version,
		template.Description,
		template.Category,
		string(definitionJSON),
		template.UpdatedAt,
		template.ID,
	)

	return err
}

// DeleteTemplate deletes a template by ID
func DeleteTemplate(db *sql.DB, id int64) error {
	query := `DELETE FROM templates WHERE id = ?`
	_, err := db.Exec(query, id)
	return err
}

// AttachTemplateToNote creates a note_template association
func AttachTemplateToNote(db *sql.DB, noteID int64, templateID int64, instance *models.TemplateInstance) (*models.NoteTemplate, error) {
	// Marshal instance to JSON
	instanceJSON, err := json.Marshal(instance)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal template instance: %w", err)
	}

	query := `
		INSERT INTO note_templates (note_id, template_id, template_data, created_at)
		VALUES (?, ?, ?, ?)
	`

	now := time.Now()
	result, err := db.Exec(query, noteID, templateID, string(instanceJSON), now)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	noteTemplate := &models.NoteTemplate{
		ID:           id,
		NoteID:       noteID,
		TemplateID:   templateID,
		TemplateData: *instance,
		CreatedAt:    now,
	}

	template, err := GetTemplateByID(db, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to load template for shape state: %w", err)
	}
	comp, err := template.Definition.GetStructure()
	if err != nil {
		return nil, err
	}
	if err := InitializeShapeStates(db, id, comp, instance.Inputs); err != nil {
		return nil, fmt.Errorf("failed to initialize shape states: %w", err)
	}

	return noteTemplate, nil
}

// GetNoteTemplate retrieves the template association for a note
func GetNoteTemplate(db *sql.DB, noteID int64) (*models.NoteTemplate, error) {
	query := `
		SELECT id, note_id, template_id, template_data, created_at
		FROM note_templates
		WHERE note_id = ?
	`

	noteTemplate := &models.NoteTemplate{}
	var instanceJSON string

	err := db.QueryRow(query, noteID).Scan(
		&noteTemplate.ID,
		&noteTemplate.NoteID,
		&noteTemplate.TemplateID,
		&instanceJSON,
		&noteTemplate.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Unmarshal instance
	if err := json.Unmarshal([]byte(instanceJSON), &noteTemplate.TemplateData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal template instance: %w", err)
	}

	return noteTemplate, nil
}

// UpdateTemplateProgress updates the progress percentage for a note
func UpdateTemplateProgress(db *sql.DB, noteID int64, progress float64) error {
	query := `
		UPDATE notes
		SET template_progress = ?
		WHERE id = ?
	`

	_, err := db.Exec(query, progress, noteID)
	return err
}
