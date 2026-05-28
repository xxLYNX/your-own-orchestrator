package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"yoo/internal/models"
)

// BackfillShapeStatesIfMissing initializes shape state rows for notes created before migration 4.
func BackfillShapeStatesIfMissing(db *sql.DB, noteTemplateID int64, template *models.Template, inputs map[string]interface{}) error {
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM note_shape_state WHERE note_template_id = ?`, noteTemplateID).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	comp, err := template.Definition.GetComposition()
	if err != nil {
		return err
	}
	return InitializeShapeStates(db, noteTemplateID, comp, inputs)
}

// InitializeShapeStates materializes runtime rows for every trackable node in a composition tree.
func InitializeShapeStates(db *sql.DB, noteTemplateID int64, composition *models.ShapeNode, inputs map[string]interface{}) error {
	if composition == nil {
		return nil
	}

	inits := composition.CollectShapeStateInits([]string{}, 0, inputs)
	for _, init := range inits {
		data := models.ShapeStateData{}
		if init.Kind == models.ShapeChecklist {
			data.ItemCompletion = make(map[string]bool, len(init.ItemIDs))
			for _, id := range init.ItemIDs {
				data.ItemCompletion[id] = false
			}
		}

		if err := upsertShapeState(db, noteTemplateID, init.Path, init.RepeatIndex, init.Kind, init.Title, data); err != nil {
			return err
		}
	}

	return SyncLegacyNoteStepsFromShapeStates(db, noteTemplateID)
}

// EnsureRepeatScope materializes shape state for a repeat iteration if missing.
func EnsureRepeatScope(db *sql.DB, noteTemplateID int64, composition *models.ShapeNode, repeatNode *models.ShapeNode, repeatIndex int, inputs map[string]interface{}) error {
	if repeatNode == nil || repeatNode.RepeatBody == nil || repeatIndex <= 0 {
		return nil
	}

	path := composition.PathTo(repeatNode.ID)
	if len(path) == 0 {
		path = []string{composition.ID, repeatNode.ID}
	}

	inits := repeatNode.RepeatBody.CollectShapeStateInits(path, repeatIndex, inputs)
	for _, init := range inits {
		existing, err := GetShapeState(db, noteTemplateID, init.Path, init.RepeatIndex)
		if err != nil {
			return err
		}
		if existing != nil {
			continue
		}

		data := models.ShapeStateData{}
		if init.Kind == models.ShapeChecklist {
			data.ItemCompletion = make(map[string]bool, len(init.ItemIDs))
			for _, id := range init.ItemIDs {
				data.ItemCompletion[id] = false
			}
		}
		if err := upsertShapeState(db, noteTemplateID, init.Path, init.RepeatIndex, init.Kind, init.Title, data); err != nil {
			return err
		}
	}
	return nil
}

func upsertShapeState(db *sql.DB, noteTemplateID int64, shapePath string, repeatIndex int, kind, title string, data models.ShapeStateData) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal shape state: %w", err)
	}

	now := time.Now()
	query := `
		INSERT INTO note_shape_state (
			note_template_id, shape_path, repeat_index, kind, title,
			completed, data, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, 0, ?, ?, ?)
		ON CONFLICT(note_template_id, shape_path, repeat_index) DO NOTHING
	`
	_, err = db.Exec(query, noteTemplateID, shapePath, repeatIndex, kind, title, string(dataJSON), now, now)
	if err != nil {
		return fmt.Errorf("failed to insert shape state: %w", err)
	}
	return nil
}

// GetShapeState loads a single shape state row.
func GetShapeState(db *sql.DB, noteTemplateID int64, shapePath string, repeatIndex int) (*models.ShapeState, error) {
	query := `
		SELECT id, note_template_id, shape_path, repeat_index, kind, title,
		       completed, completed_at, notes, data, created_at, updated_at
		FROM note_shape_state
		WHERE note_template_id = ? AND shape_path = ? AND repeat_index = ?
	`

	state := &models.ShapeState{}
	var dataJSON string
	var completedAt, notes sql.NullString
	var createdAt, updatedAt time.Time

	err := db.QueryRow(query, noteTemplateID, shapePath, repeatIndex).Scan(
		&state.ID,
		&state.NoteTemplateID,
		&state.ShapePath,
		&state.RepeatIndex,
		&state.Kind,
		&state.Title,
		&state.Completed,
		&completedAt,
		&notes,
		&dataJSON,
		&createdAt,
		&updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get shape state: %w", err)
	}

	if completedAt.Valid {
		state.CompletedAt = &completedAt.String
	}
	if notes.Valid {
		state.Notes = notes.String
	}
	state.CreatedAt = createdAt.Format(time.RFC3339)
	state.UpdatedAt = updatedAt.Format(time.RFC3339)

	if dataJSON != "" {
		if err := json.Unmarshal([]byte(dataJSON), &state.Data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal shape state data: %w", err)
		}
	}
	return state, nil
}

// ListShapeStates returns shape states for a note template scope.
func ListShapeStates(db *sql.DB, noteTemplateID int64, repeatIndex int) ([]*models.ShapeState, error) {
	var (
		query string
		args  []interface{}
	)
	if repeatIndex < 0 {
		query = `
			SELECT id, note_template_id, shape_path, repeat_index, kind, title,
			       completed, completed_at, notes, data, created_at, updated_at
			FROM note_shape_state
			WHERE note_template_id = ?
			ORDER BY repeat_index ASC, shape_path ASC
		`
		args = []interface{}{noteTemplateID}
	} else {
		query = `
			SELECT id, note_template_id, shape_path, repeat_index, kind, title,
			       completed, completed_at, notes, data, created_at, updated_at
			FROM note_shape_state
			WHERE note_template_id = ? AND repeat_index = ?
			ORDER BY shape_path ASC
		`
		args = []interface{}{noteTemplateID, repeatIndex}
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list shape states: %w", err)
	}
	defer rows.Close()

	return scanShapeStates(rows)
}

// ListTopLevelProcedureStates returns non-repeat-scoped procedure rows for legacy step views.
func ListTopLevelProcedureStates(db *sql.DB, noteTemplateID int64) ([]*models.ShapeState, error) {
	query := `
		SELECT id, note_template_id, shape_path, repeat_index, kind, title,
		       completed, completed_at, notes, data, created_at, updated_at
		FROM note_shape_state
		WHERE note_template_id = ? AND repeat_index = 0 AND kind = ?
		  AND shape_path NOT LIKE '%.apply_one.%' AND shape_path NOT LIKE '%.apply_one'
		ORDER BY id ASC
	`
	rows, err := db.Query(query, noteTemplateID, models.ShapeProcedure)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanShapeStates(rows)
}

func scanShapeStates(rows *sql.Rows) ([]*models.ShapeState, error) {
	var states []*models.ShapeState
	for rows.Next() {
		state := &models.ShapeState{}
		var dataJSON string
		var completedAt, notes sql.NullString
		var createdAt, updatedAt time.Time

		if err := rows.Scan(
			&state.ID,
			&state.NoteTemplateID,
			&state.ShapePath,
			&state.RepeatIndex,
			&state.Kind,
			&state.Title,
			&state.Completed,
			&completedAt,
			&notes,
			&dataJSON,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, err
		}
		if completedAt.Valid {
			state.CompletedAt = &completedAt.String
		}
		if notes.Valid {
			state.Notes = notes.String
		}
		state.CreatedAt = createdAt.Format(time.RFC3339)
		state.UpdatedAt = updatedAt.Format(time.RFC3339)
		if dataJSON != "" {
			if err := json.Unmarshal([]byte(dataJSON), &state.Data); err != nil {
				return nil, err
			}
		}
		states = append(states, state)
	}
	return states, rows.Err()
}

// UpdateShapeState persists a shape state row.
func UpdateShapeState(db *sql.DB, state *models.ShapeState) error {
	dataJSON, err := json.Marshal(state.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal shape state: %w", err)
	}

	now := time.Now()
	query := `
		UPDATE note_shape_state
		SET completed = ?, completed_at = ?, notes = ?, data = ?, updated_at = ?
		WHERE id = ?
	`

	var completedAt interface{}
	if state.Completed {
		completedAt = now
	}

	_, err = db.Exec(query, state.Completed, completedAt, state.Notes, string(dataJSON), now, state.ID)
	if err != nil {
		return fmt.Errorf("failed to update shape state: %w", err)
	}
	state.UpdatedAt = now.Format(time.RFC3339)
	return nil
}

// ToggleChecklistItem updates one checklist item and auto-completes the parent when all items are done.
func ToggleChecklistItem(db *sql.DB, state *models.ShapeState, itemID string, completed bool) error {
	if state.Data.ItemCompletion == nil {
		state.Data.ItemCompletion = map[string]bool{}
	}
	state.Data.ItemCompletion[itemID] = completed
	state.Completed = models.AllChecklistItemsComplete(state)
	if err := UpdateShapeState(db, state); err != nil {
		return err
	}
	if err := syncParentProcedureFromChecklist(db, state); err != nil {
		return err
	}
	return SyncLegacyNoteStepsFromShapeStates(db, state.NoteTemplateID)
}

func syncParentProcedureFromChecklist(db *sql.DB, checklistState *models.ShapeState) error {
	parts := strings.Split(checklistState.ShapePath, ".")
	if len(parts) < 2 {
		return nil
	}
	parentPath := strings.Join(parts[:len(parts)-1], ".")
	parent, err := GetShapeState(db, checklistState.NoteTemplateID, parentPath, checklistState.RepeatIndex)
	if err != nil {
		return err
	}
	if parent == nil || parent.Kind != models.ShapeProcedure {
		return nil
	}
	parent.Completed = checklistState.Completed
	return UpdateShapeState(db, parent)
}

// SyncLegacyNoteStepsFromShapeStates keeps flat note_steps aligned with top-level procedure shape states.
func SyncLegacyNoteStepsFromShapeStates(db *sql.DB, noteTemplateID int64) error {
	states, err := ListTopLevelProcedureStates(db, noteTemplateID)
	if err != nil {
		return err
	}

	for i, state := range states {
		stepNumber := i + 1
		var completedAt interface{}
		if state.Completed && state.CompletedAt != nil {
			completedAt = state.CompletedAt
		} else if state.Completed {
			completedAt = time.Now()
		}

		_, err := db.Exec(`
			UPDATE note_steps
			SET completed = ?, completed_at = ?, title = ?
			WHERE note_template_id = ? AND step_number = ?
		`, state.Completed, completedAt, state.Title, noteTemplateID, stepNumber)
		if err != nil {
			return err
		}
	}
	return nil
}

// ComputeTemplateProgress calculates overall completion for a templated note instance.
func ComputeTemplateProgress(db *sql.DB, noteTemplateID int64, template *models.Template, inputs map[string]interface{}) (float64, error) {
	states, err := ListShapeStates(db, noteTemplateID, -1)
	if err != nil {
		return 0, err
	}

	total := 0
	done := 0
	for _, state := range states {
		switch state.Kind {
		case models.ShapeChecklist:
			for _, completed := range state.Data.ItemCompletion {
				total++
				if completed {
					done++
				}
			}
		case models.ShapeProcedure:
			total++
			if state.Completed {
				done++
			}
		}
	}

	comp, err := template.Definition.GetComposition()
	if err == nil && comp != nil {
		var repeatNode *models.ShapeNode
		var findRepeat func(*models.ShapeNode)
		findRepeat = func(n *models.ShapeNode) {
			if n == nil || repeatNode != nil {
				return
			}
			if n.Kind == models.ShapeRepeat {
				repeatNode = n
				return
			}
			for i := range n.Steps {
				findRepeat(&n.Steps[i])
			}
			if n.RepeatBody != nil {
				findRepeat(n.RepeatBody)
			}
		}
		findRepeat(comp)
		if repeatNode != nil {
			target := repeatNode.ResolveRepeatCount(inputs)
			if target > 0 {
				recordCount, err := CountTemplateRecords(db, noteTemplateID, -1)
				if err != nil {
					return 0, err
				}
				if recordCount > target {
					recordCount = target
				}
				total += target
				done += recordCount
			}
		}
	}

	if total == 0 {
		return 0, nil
	}
	return float64(done) / float64(total), nil
}

// PersistTemplateProgress saves computed progress on the parent note.
func PersistTemplateProgress(db *sql.DB, noteID int64, noteTemplateID int64, template *models.Template, inputs map[string]interface{}) error {
	progress, err := ComputeTemplateProgress(db, noteTemplateID, template, inputs)
	if err != nil {
		return err
	}
	return UpdateTemplateProgress(db, noteID, progress)
}
