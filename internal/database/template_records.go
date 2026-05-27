package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"yoo/internal/models"
)

// GetNextRecordIndex gets the next available record index for a note template
func GetNextRecordIndex(conn *sql.DB, noteTemplateID int64) (int, error) {
	query := `
		SELECT COALESCE(MAX(record_index), 0) + 1
		FROM template_records
		WHERE note_template_id = ?
	`

	var nextIndex int
	err := conn.QueryRow(query, noteTemplateID).Scan(&nextIndex)
	if err != nil {
		return 0, fmt.Errorf("failed to get next record index: %w", err)
	}

	return nextIndex, nil
}

// CreateTemplateRecord inserts a new template record into the database
func CreateTemplateRecord(conn *sql.DB, record *models.TemplateRecord) error {
	// Marshal data to JSON
	dataJSON, err := json.Marshal(record.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal record data: %w", err)
	}

	// If record index is not set, get the next available one
	if record.RecordIndex == 0 {
		record.RecordIndex, err = GetNextRecordIndex(conn, record.NoteTemplateID)
		if err != nil {
			return err
		}
	}

	// Set default status if not provided
	if record.Status == "" {
		record.Status = "draft"
	}

	query := `
		INSERT INTO template_records (note_template_id, record_index, data, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	result, err := conn.Exec(query,
		record.NoteTemplateID,
		record.RecordIndex,
		string(dataJSON),
		record.Status,
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to create template record: %w", err)
	}

	record.ID, err = result.LastInsertId()
	if err != nil {
		return err
	}

	record.CreatedAt = now
	record.UpdatedAt = now

	return nil
}

// ListTemplateRecords retrieves all records for a note template
func ListTemplateRecords(conn *sql.DB, noteTemplateID int64) ([]*models.TemplateRecord, error) {
	query := `
		SELECT id, note_template_id, record_index, data, status, created_at, updated_at
		FROM template_records
		WHERE note_template_id = ?
		ORDER BY record_index ASC
	`

	rows, err := conn.Query(query, noteTemplateID)
	if err != nil {
		return nil, fmt.Errorf("failed to list template records: %w", err)
	}
	defer rows.Close()

	var records []*models.TemplateRecord
	for rows.Next() {
		record := &models.TemplateRecord{}
		var dataJSON string

		err := rows.Scan(
			&record.ID,
			&record.NoteTemplateID,
			&record.RecordIndex,
			&dataJSON,
			&record.Status,
			&record.CreatedAt,
			&record.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan template record: %w", err)
		}

		// Unmarshal data
		if err := json.Unmarshal([]byte(dataJSON), &record.Data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal record data: %w", err)
		}

		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

// GetTemplateRecord retrieves a specific record by note template ID and record index
func GetTemplateRecord(conn *sql.DB, noteTemplateID int64, recordIndex int) (*models.TemplateRecord, error) {
	query := `
		SELECT id, note_template_id, record_index, data, status, created_at, updated_at
		FROM template_records
		WHERE note_template_id = ? AND record_index = ?
	`

	record := &models.TemplateRecord{}
	var dataJSON string

	err := conn.QueryRow(query, noteTemplateID, recordIndex).Scan(
		&record.ID,
		&record.NoteTemplateID,
		&record.RecordIndex,
		&dataJSON,
		&record.Status,
		&record.CreatedAt,
		&record.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("template record not found")
		}
		return nil, fmt.Errorf("failed to get template record: %w", err)
	}

	// Unmarshal data
	if err := json.Unmarshal([]byte(dataJSON), &record.Data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal record data: %w", err)
	}

	return record, nil
}

// UpdateTemplateRecord updates an existing template record
func UpdateTemplateRecord(conn *sql.DB, record *models.TemplateRecord) error {
	// Marshal data to JSON
	dataJSON, err := json.Marshal(record.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal record data: %w", err)
	}

	query := `
		UPDATE template_records
		SET data = ?, status = ?, updated_at = ?
		WHERE note_template_id = ? AND record_index = ?
	`

	record.UpdatedAt = time.Now()

	result, err := conn.Exec(query,
		string(dataJSON),
		record.Status,
		record.UpdatedAt,
		record.NoteTemplateID,
		record.RecordIndex,
	)
	if err != nil {
		return fmt.Errorf("failed to update template record: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("template record not found")
	}

	return nil
}

// DeleteTemplateRecord deletes a template record by note template ID and record index
func DeleteTemplateRecord(conn *sql.DB, noteTemplateID int64, recordIndex int) error {
	query := `DELETE FROM template_records WHERE note_template_id = ? AND record_index = ?`

	result, err := conn.Exec(query, noteTemplateID, recordIndex)
	if err != nil {
		return fmt.Errorf("failed to delete template record: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("template record not found")
	}

	return nil
}

// GetNote retrieves a note by its ID
// This is a convenience wrapper around GetNoteByID for consistency
func GetNote(conn *sql.DB, noteID int64) (*Note, error) {
	return GetNoteByID(conn, noteID)
}
