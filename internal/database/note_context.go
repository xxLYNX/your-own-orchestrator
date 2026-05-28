package database

import (
	"database/sql"
	"fmt"

	"yoo/internal/models"
)

// TemplatedNoteContext holds loaded note, instance, and template data.
type TemplatedNoteContext struct {
	Note         *Note
	NoteTemplate *models.NoteTemplate
	Template     *models.Template
}

// LoadTemplatedNoteContext loads a note and its template instance.
func LoadTemplatedNoteContext(db *sql.DB, noteID int64) (*TemplatedNoteContext, error) {
	note, err := GetNoteByID(db, noteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get note: %w", err)
	}
	noteTemplate, err := GetNoteTemplate(db, noteID)
	if err != nil {
		return nil, fmt.Errorf("note is not templated: %w", err)
	}
	template, err := GetTemplateByID(db, noteTemplate.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}
	return &TemplatedNoteContext{
		Note:         note,
		NoteTemplate: noteTemplate,
		Template:     template,
	}, nil
}
