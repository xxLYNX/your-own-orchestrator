package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"yoo/internal/models"
)

// GetNextRecordIndex gets the next available record index for a note template scope.
func GetNextRecordIndex(conn *sql.DB, noteTemplateID int64, stack models.RepeatStack) (int, error) {
	query := `
		SELECT COALESCE(MAX(record_index), 0) + 1
		FROM template_records
		WHERE note_template_id = ? AND repeat_stack_json = ?
	`

	var nextIndex int
	err := conn.QueryRow(query, noteTemplateID, stack.JSON()).Scan(&nextIndex)
	if err != nil {
		return 0, fmt.Errorf("failed to get next record index: %w", err)
	}

	return nextIndex, nil
}

// CreateTemplateRecord inserts a new template record into the database
func CreateTemplateRecord(conn *sql.DB, record *models.TemplateRecord) error {
	dataJSON, err := json.Marshal(record.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal record data: %w", err)
	}

	if record.RepeatStack == nil {
		record.RepeatStack = nil
	}

	if record.RecordIndex == 0 {
		record.RecordIndex, err = GetNextRecordIndex(conn, record.NoteTemplateID, record.RepeatStack)
		if err != nil {
			return err
		}
	}

	if record.Status == "" {
		record.Status = "draft"
	}

	query := `
		INSERT INTO template_records (note_template_id, repeat_stack_json, record_index, data, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	result, err := conn.Exec(query,
		record.NoteTemplateID,
		record.RepeatStack.JSON(),
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

// ListTemplateRecords retrieves records for a note template.
// stack == nil returns all repeat scopes; otherwise filters to that occurrence context.
func ListTemplateRecords(conn *sql.DB, noteTemplateID int64, stack models.RepeatStack) ([]*models.TemplateRecord, error) {
	var (
		query string
		args  []interface{}
	)

	if stack == nil {
		query = `
			SELECT id, note_template_id, repeat_stack_json, record_index, data, status, created_at, updated_at
			FROM template_records
			WHERE note_template_id = ?
			ORDER BY repeat_stack_json ASC, record_index ASC
		`
		args = []interface{}{noteTemplateID}
	} else {
		query = `
			SELECT id, note_template_id, repeat_stack_json, record_index, data, status, created_at, updated_at
			FROM template_records
			WHERE note_template_id = ? AND repeat_stack_json = ?
			ORDER BY record_index ASC
		`
		args = []interface{}{noteTemplateID, stack.JSON()}
	}

	rows, err := conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list template records: %w", err)
	}
	defer rows.Close()

	var records []*models.TemplateRecord
	for rows.Next() {
		record := &models.TemplateRecord{}
		var dataJSON, stackJSON string

		err := rows.Scan(
			&record.ID,
			&record.NoteTemplateID,
			&stackJSON,
			&record.RecordIndex,
			&dataJSON,
			&record.Status,
			&record.CreatedAt,
			&record.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan template record: %w", err)
		}

		record.RepeatStack = models.ParseRepeatStackJSON(stackJSON)

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

// CountTemplateRecords returns how many records exist for a note template scope.
func CountTemplateRecords(conn *sql.DB, noteTemplateID int64, stack models.RepeatStack) (int, error) {
	var query string
	var args []interface{}

	if stack == nil {
		query = `SELECT COUNT(*) FROM template_records WHERE note_template_id = ?`
		args = []interface{}{noteTemplateID}
	} else {
		query = `SELECT COUNT(*) FROM template_records WHERE note_template_id = ? AND repeat_stack_json = ?`
		args = []interface{}{noteTemplateID, stack.JSON()}
	}

	var count int
	if err := conn.QueryRow(query, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count template records: %w", err)
	}
	return count, nil
}

// GetTemplateRecord retrieves a specific record by note template ID, repeat stack, and record index.
func GetTemplateRecord(conn *sql.DB, noteTemplateID int64, stack models.RepeatStack, recordIndex int) (*models.TemplateRecord, error) {
	query := `
		SELECT id, note_template_id, repeat_stack_json, record_index, data, status, created_at, updated_at
		FROM template_records
		WHERE note_template_id = ? AND repeat_stack_json = ? AND record_index = ?
	`

	record := &models.TemplateRecord{}
	var dataJSON, stackJSON string

	err := conn.QueryRow(query, noteTemplateID, stack.JSON(), recordIndex).Scan(
		&record.ID,
		&record.NoteTemplateID,
		&stackJSON,
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

	record.RepeatStack = models.ParseRepeatStackJSON(stackJSON)

	if err := json.Unmarshal([]byte(dataJSON), &record.Data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal record data: %w", err)
	}

	return record, nil
}

// UpdateTemplateRecord updates an existing template record
func UpdateTemplateRecord(conn *sql.DB, record *models.TemplateRecord) error {
	dataJSON, err := json.Marshal(record.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal record data: %w", err)
	}

	query := `
		UPDATE template_records
		SET data = ?, status = ?, updated_at = ?
		WHERE note_template_id = ? AND repeat_stack_json = ? AND record_index = ?
	`

	record.UpdatedAt = time.Now()

	result, err := conn.Exec(query,
		string(dataJSON),
		record.Status,
		record.UpdatedAt,
		record.NoteTemplateID,
		record.RepeatStack.JSON(),
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

// DeleteTemplateRecord deletes a template record by scope keys.
func DeleteTemplateRecord(conn *sql.DB, noteTemplateID int64, stack models.RepeatStack, recordIndex int) error {
	query := `DELETE FROM template_records WHERE note_template_id = ? AND repeat_stack_json = ? AND record_index = ?`

	result, err := conn.Exec(query, noteTemplateID, stack.JSON(), recordIndex)
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
func GetNote(conn *sql.DB, noteID int64) (*Note, error) {
	return GetNoteByID(conn, noteID)
}
