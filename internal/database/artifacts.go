package database

import (
	"database/sql"
	"fmt"
	"time"

	"yoo/internal/models"
)

// ListArtifacts retrieves all artifacts for a note template
func ListArtifacts(conn *sql.DB, noteTemplateID int64) ([]*models.Artifact, error) {
	query := `
		SELECT id, note_template_id, artifact_type, name, type, value, description, required, created_at, updated_at
		FROM artifacts
		WHERE note_template_id = ?
		ORDER BY artifact_type, name
	`

	rows, err := conn.Query(query, noteTemplateID)
	if err != nil {
		return nil, fmt.Errorf("failed to query artifacts: %w", err)
	}
	defer rows.Close()

	artifacts := []*models.Artifact{}
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
			return nil, fmt.Errorf("failed to scan artifact: %w", err)
		}
		artifacts = append(artifacts, artifact)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating artifacts: %w", err)
	}

	return artifacts, nil
}

// GetArtifact retrieves a specific artifact by note template ID and name
func GetArtifact(conn *sql.DB, noteTemplateID int64, name string) (*models.Artifact, error) {
	query := `
		SELECT id, note_template_id, artifact_type, name, type, value, description, required, created_at, updated_at
		FROM artifacts
		WHERE note_template_id = ? AND name = ?
	`

	artifact := &models.Artifact{}
	err := conn.QueryRow(query, noteTemplateID, name).Scan(
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
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("artifact '%s' not found", name)
		}
		return nil, fmt.Errorf("failed to get artifact: %w", err)
	}

	return artifact, nil
}

// CreateArtifact inserts a new artifact into the database
func CreateArtifact(conn *sql.DB, artifact *models.Artifact) error {
	query := `
		INSERT INTO artifacts (note_template_id, artifact_type, name, type, value, description, required, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	result, err := conn.Exec(query,
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
		return fmt.Errorf("failed to create artifact: %w", err)
	}

	artifact.ID, err = result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get artifact ID: %w", err)
	}

	artifact.CreatedAt = now
	artifact.UpdatedAt = now
	return nil
}

// DeleteArtifact removes an artifact from the database
func DeleteArtifact(conn *sql.DB, noteTemplateID int64, name string) error {
	query := `
		DELETE FROM artifacts
		WHERE note_template_id = ? AND name = ?
	`

	result, err := conn.Exec(query, noteTemplateID, name)
	if err != nil {
		return fmt.Errorf("failed to delete artifact: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("artifact '%s' not found", name)
	}

	return nil
}
