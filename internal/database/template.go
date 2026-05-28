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
	defer rows.Close()

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

	// Create step instances in the database
	for i := range instance.Steps {
		if err := createStepInstance(db, id, &instance.Steps[i]); err != nil {
			return nil, fmt.Errorf("failed to create step instance: %w", err)
		}
	}

	template, err := GetTemplateByID(db, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to load template for shape state: %w", err)
	}
	comp, err := template.Definition.GetComposition()
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

// createStepInstance creates a step instance in the database
func createStepInstance(db *sql.DB, noteTemplateID int64, step *models.StepInstance) error {
	query := `
		INSERT INTO note_steps (note_template_id, step_number, title, description, completed, completed_at, notes, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.Exec(query,
		noteTemplateID,
		step.StepNumber,
		step.Title,
		step.Description,
		step.Completed,
		step.CompletedAt,
		step.Notes,
		step.CreatedAt,
	)
	if err != nil {
		return err
	}

	step.ID, err = result.LastInsertId()
	return err
}

// CompleteStep marks a step as completed
func CompleteStep(db *sql.DB, noteTemplateID int64, stepNumber int) error {
	query := `
		UPDATE note_steps
		SET completed = 1, completed_at = ?
		WHERE note_template_id = ? AND step_number = ?
	`

	now := time.Now()
	_, err := db.Exec(query, now, noteTemplateID, stepNumber)
	return err
}

// GetStepsByNoteTemplate retrieves all steps for a note template
func GetStepsByNoteTemplate(db *sql.DB, noteTemplateID int64) ([]models.StepInstance, error) {
	query := `
		SELECT id, step_number, title, description, completed, completed_at, notes, created_at
		FROM note_steps
		WHERE note_template_id = ?
		ORDER BY step_number ASC
	`

	rows, err := db.Query(query, noteTemplateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var steps []models.StepInstance
	for rows.Next() {
		step := models.StepInstance{}
		err := rows.Scan(
			&step.ID,
			&step.StepNumber,
			&step.Title,
			&step.Description,
			&step.Completed,
			&step.CompletedAt,
			&step.Notes,
			&step.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}

	return steps, rows.Err()
}

// AddArtifact adds an input or output artifact
func AddArtifact(db *sql.DB, artifact *models.Artifact) error {
	query := `
		INSERT INTO artifacts (note_template_id, artifact_type, name, type, value, description, required, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	result, err := db.Exec(query,
		artifact.NoteTemplateID,
		artifact.ArtifactType,
		artifact.Name,
		artifact.Type,
		artifact.Value,
		artifact.Description,
		artifact.Required,
		now,
		now,
	)
	if err != nil {
		return err
	}

	artifact.ID, err = result.LastInsertId()
	artifact.CreatedAt = now
	artifact.UpdatedAt = now
	return err
}

// GetArtifactsByNoteTemplate retrieves all artifacts for a note template
func GetArtifactsByNoteTemplate(db *sql.DB, noteTemplateID int64) ([]*models.Artifact, error) {
	query := `
		SELECT id, note_template_id, artifact_type, name, type, value, description, required, created_at, updated_at
		FROM artifacts
		WHERE note_template_id = ?
		ORDER BY artifact_type, name
	`

	rows, err := db.Query(query, noteTemplateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var artifacts []*models.Artifact
	for rows.Next() {
		artifact := &models.Artifact{}
		err := rows.Scan(
			&artifact.ID,
			&artifact.NoteTemplateID,
			&artifact.ArtifactType,
			&artifact.Name,
			&artifact.Type,
			&artifact.Value,
			&artifact.Description,
			&artifact.Required,
			&artifact.CreatedAt,
			&artifact.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		artifacts = append(artifacts, artifact)
	}

	return artifacts, rows.Err()
}

// AddReference adds a reference to external resource
func AddReference(db *sql.DB, reference *models.Reference) error {
	query := `
		INSERT INTO note_references (note_id, ref_type, ref_value, description, created_at)
		VALUES (?, ?, ?, ?, ?)
	`

	now := time.Now()
	result, err := db.Exec(query,
		reference.NoteID,
		reference.Type,
		reference.Value,
		reference.Description,
		now,
	)
	if err != nil {
		return err
	}

	reference.ID, err = result.LastInsertId()
	reference.CreatedAt = now
	return err
}

// GetReferencesByNote retrieves all references for a note
func GetReferencesByNote(db *sql.DB, noteID int64) ([]*models.Reference, error) {
	query := `
		SELECT id, note_id, ref_type, ref_value, description, created_at
		FROM note_references
		WHERE note_id = ?
		ORDER BY created_at DESC
	`

	rows, err := db.Query(query, noteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var references []*models.Reference
	for rows.Next() {
		ref := &models.Reference{}
		err := rows.Scan(
			&ref.ID,
			&ref.NoteID,
			&ref.Type,
			&ref.Value,
			&ref.Description,
			&ref.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		references = append(references, ref)
	}

	return references, rows.Err()
}

// DeleteReference removes a reference
func DeleteReference(db *sql.DB, id int64) error {
	query := `DELETE FROM note_references WHERE id = ?`
	_, err := db.Exec(query, id)
	return err
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

// IsNoteTemplated checks if a note uses a template
func IsNoteTemplated(db *sql.DB, noteID int64) (bool, error) {
	query := `SELECT COUNT(*) FROM note_templates WHERE note_id = ?`

	var count int
	err := db.QueryRow(query, noteID).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
