package database

import (
	"database/sql"
	"fmt"

	"yoo/internal/models"
)

// ListNoteSteps returns all steps for a given note template
func ListNoteSteps(conn *sql.DB, noteTemplateID int64) ([]*models.StepInstance, error) {
	query := `
		SELECT id, step_number, title, description, completed, completed_at, notes, created_at
		FROM note_steps
		WHERE note_template_id = ?
		ORDER BY step_number ASC
	`

	rows, err := conn.Query(query, noteTemplateID)
	if err != nil {
		return nil, fmt.Errorf("failed to query steps: %w", err)
	}
	defer rows.Close()

	var steps []*models.StepInstance
	for rows.Next() {
		step := &models.StepInstance{}
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
			return nil, fmt.Errorf("failed to scan step: %w", err)
		}
		steps = append(steps, step)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating steps: %w", err)
	}

	return steps, nil
}

// GetNoteStep returns a specific step for a given note template
func GetNoteStep(conn *sql.DB, noteTemplateID int64, stepNumber int) (*models.StepInstance, error) {
	query := `
		SELECT id, step_number, title, description, completed, completed_at, notes, created_at
		FROM note_steps
		WHERE note_template_id = ? AND step_number = ?
	`

	step := &models.StepInstance{}
	err := conn.QueryRow(query, noteTemplateID, stepNumber).Scan(
		&step.ID,
		&step.StepNumber,
		&step.Title,
		&step.Description,
		&step.Completed,
		&step.CompletedAt,
		&step.Notes,
		&step.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("step %d not found", stepNumber)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get step: %w", err)
	}

	return step, nil
}

// UncompleteStep marks a step as incomplete
func UncompleteStep(conn *sql.DB, noteTemplateID int64, stepNumber int) error {
	query := `
		UPDATE note_steps
		SET completed = 0, completed_at = NULL
		WHERE note_template_id = ? AND step_number = ?
	`

	result, err := conn.Exec(query, noteTemplateID, stepNumber)
	if err != nil {
		return fmt.Errorf("failed to uncomplete step: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("step %d not found", stepNumber)
	}

	return nil
}

// UpdateStepNotes updates the notes for a specific step
func UpdateStepNotes(conn *sql.DB, noteTemplateID int64, stepNumber int, notes string) error {
	query := `
		UPDATE note_steps
		SET notes = ?
		WHERE note_template_id = ? AND step_number = ?
	`

	result, err := conn.Exec(query, notes, noteTemplateID, stepNumber)
	if err != nil {
		return fmt.Errorf("failed to update step notes: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("step %d not found", stepNumber)
	}

	return nil
}
