package database

import (
	"database/sql"
	"time"
)

// Note represents a task, reminder, or action item in the schedule
type Note struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	ScheduledAt time.Time `json:"scheduled_at"`
	Status      string    `json:"status"` // "pending", "completed", "cancelled"
	Priority    int       `json:"priority"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateNote inserts a new note into the database
func CreateNote(db *sql.DB, note *Note) error {
	query := `
		INSERT INTO notes (title, description, scheduled_at, status, priority, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	if note.CreatedAt.IsZero() {
		note.CreatedAt = now
	}
	if note.UpdatedAt.IsZero() {
		note.UpdatedAt = now
	}
	if note.Status == "" {
		note.Status = "pending"
	}

	result, err := db.Exec(query,
		note.Title,
		note.Description,
		note.ScheduledAt,
		note.Status,
		note.Priority,
		note.CreatedAt,
		note.UpdatedAt,
	)
	if err != nil {
		return err
	}

	note.ID, err = result.LastInsertId()
	return err
}

// GetNoteByID retrieves a note by its ID
func GetNoteByID(db *sql.DB, id int64) (*Note, error) {
	query := `
		SELECT id, title, description, scheduled_at, status, priority, created_at, updated_at
		FROM notes
		WHERE id = ?
	`

	note := &Note{}
	err := db.QueryRow(query, id).Scan(
		&note.ID,
		&note.Title,
		&note.Description,
		&note.ScheduledAt,
		&note.Status,
		&note.Priority,
		&note.CreatedAt,
		&note.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return note, nil
}

// GetNotesByDate retrieves all notes scheduled for a specific date
func GetNotesByDate(db *sql.DB, date time.Time) ([]*Note, error) {
	// Normalize to start and end of day
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	query := `
		SELECT id, title, description, scheduled_at, status, priority, created_at, updated_at
		FROM notes
		WHERE scheduled_at >= ? AND scheduled_at < ?
		ORDER BY scheduled_at ASC, priority DESC
	`

	rows, err := db.Query(query, startOfDay, endOfDay)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []*Note
	for rows.Next() {
		note := &Note{}
		err := rows.Scan(
			&note.ID,
			&note.Title,
			&note.Description,
			&note.ScheduledAt,
			&note.Status,
			&note.Priority,
			&note.CreatedAt,
			&note.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}

	return notes, rows.Err()
}

// GetNotesByDateRange retrieves all notes within a date range
func GetNotesByDateRange(db *sql.DB, startDate, endDate time.Time) ([]*Note, error) {
	query := `
		SELECT id, title, description, scheduled_at, status, priority, created_at, updated_at
		FROM notes
		WHERE scheduled_at >= ? AND scheduled_at < ?
		ORDER BY scheduled_at ASC, priority DESC
	`

	rows, err := db.Query(query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []*Note
	for rows.Next() {
		note := &Note{}
		err := rows.Scan(
			&note.ID,
			&note.Title,
			&note.Description,
			&note.ScheduledAt,
			&note.Status,
			&note.Priority,
			&note.CreatedAt,
			&note.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}

	return notes, rows.Err()
}

// UpdateNote updates an existing note
func UpdateNote(db *sql.DB, note *Note) error {
	query := `
		UPDATE notes
		SET title = ?, description = ?, scheduled_at = ?, status = ?, priority = ?, updated_at = ?
		WHERE id = ?
	`

	note.UpdatedAt = time.Now()

	_, err := db.Exec(query,
		note.Title,
		note.Description,
		note.ScheduledAt,
		note.Status,
		note.Priority,
		note.UpdatedAt,
		note.ID,
	)

	return err
}

// DeleteNote deletes a note by its ID
func DeleteNote(db *sql.DB, id int64) error {
	query := `DELETE FROM notes WHERE id = ?`
	_, err := db.Exec(query, id)
	return err
}

// MarkNoteCompleted marks a note as completed
func MarkNoteCompleted(db *sql.DB, id int64) error {
	query := `
		UPDATE notes
		SET status = 'completed', updated_at = ?
		WHERE id = ?
	`

	_, err := db.Exec(query, time.Now(), id)
	return err
}

// GetAllNotes retrieves all notes
func GetAllNotes(db *sql.DB) ([]*Note, error) {
	query := `
		SELECT id, title, description, scheduled_at, status, priority, created_at, updated_at
		FROM notes
		ORDER BY scheduled_at DESC, priority DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []*Note
	for rows.Next() {
		note := &Note{}
		err := rows.Scan(
			&note.ID,
			&note.Title,
			&note.Description,
			&note.ScheduledAt,
			&note.Status,
			&note.Priority,
			&note.CreatedAt,
			&note.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}

	return notes, rows.Err()
}
