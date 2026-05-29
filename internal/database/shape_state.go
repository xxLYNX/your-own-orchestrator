package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"yoo/internal/engine"
	"yoo/internal/models"
)

const shapeStateSelectColumns = `
	id, note_template_id, shape_path, repeat_stack_json, kind, title, status,
	completed, completed_at, notes, data, created_at, updated_at`

// EnsureShapeStates initializes shape state rows when missing.
func EnsureShapeStates(db *sql.DB, noteTemplateID int64, template *models.Template, inputs map[string]interface{}) error {
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM note_shape_state WHERE note_template_id = ?`, noteTemplateID).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	comp, err := template.Definition.GetStructure()
	if err != nil {
		return err
	}
	return InitializeShapeStates(db, noteTemplateID, comp, inputs)
}

// InitializeShapeStates materializes runtime rows for every trackable node in a structure tree.
func InitializeShapeStates(db *sql.DB, noteTemplateID int64, structure *models.ShapeNode, inputs map[string]interface{}) error {
	if structure == nil {
		return nil
	}
	inits := structure.CollectShapeStateInits([]string{}, nil, inputs)
	return insertMissingShapeStateInits(db, noteTemplateID, inits)
}

// EnsureRepeatScope materializes shape state for a repeat occurrence subtree if missing.
func EnsureRepeatScope(db *sql.DB, noteTemplateID int64, structure *models.ShapeNode, node *models.ShapeNode, stack models.RepeatStack, inputs map[string]interface{}) error {
	if node == nil || len(stack) == 0 {
		return nil
	}

	path := structure.PathTo(node.ID)
	if len(path) == 0 {
		path = []string{structure.ID, node.ID}
	}

	inits := node.CollectShapeStateInits(path[:len(path)-1], stack.ParentContext(), inputs)
	var scoped []models.ShapeStateInit
	for _, init := range inits {
		if stack.Equal(init.RepeatStack) || stackHasPrefix(stack, init.RepeatStack) {
			scoped = append(scoped, init)
		}
	}
	return insertMissingShapeStateInits(db, noteTemplateID, scoped)
}

func insertMissingShapeStateInits(db *sql.DB, noteTemplateID int64, inits []models.ShapeStateInit) error {
	for _, init := range inits {
		existing, err := GetShapeState(db, noteTemplateID, init.Path, init.RepeatStack)
		if err != nil {
			return err
		}
		if existing != nil {
			continue
		}
		data := models.NewShapeStateData(init)
		if err := upsertShapeState(db, noteTemplateID, init.Path, init.RepeatStack, init.Kind, init.Title, data); err != nil {
			return err
		}
	}
	return nil
}

func stackHasPrefix(full, prefix models.RepeatStack) bool {
	if len(prefix) > len(full) {
		return false
	}
	for i := range prefix {
		if full[i].ShapeID != prefix[i].ShapeID || full[i].Index != prefix[i].Index {
			return false
		}
	}
	return true
}

func upsertShapeState(db *sql.DB, noteTemplateID int64, shapePath string, stack models.RepeatStack, kind, title string, data models.ShapeStateData) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal shape state: %w", err)
	}

	now := time.Now()
	query := `
		INSERT INTO note_shape_state (
			note_template_id, shape_path, repeat_stack_json, kind, title, status,
			completed, data, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, 0, ?, ?, ?)
		ON CONFLICT(note_template_id, shape_path, repeat_stack_json) DO NOTHING
	`
	_, err = db.Exec(query, noteTemplateID, shapePath, stack.JSON(), kind, title, models.StatusNotStarted, string(dataJSON), now, now)
	if err != nil {
		return fmt.Errorf("failed to insert shape state: %w", err)
	}
	return nil
}

// GetShapeState loads a single shape state row.
func GetShapeState(db *sql.DB, noteTemplateID int64, shapePath string, stack models.RepeatStack) (*models.ShapeState, error) {
	query := `
		SELECT` + shapeStateSelectColumns + `
		FROM note_shape_state
		WHERE note_template_id = ? AND shape_path = ? AND repeat_stack_json = ?
	`

	row := db.QueryRow(query, noteTemplateID, shapePath, stack.JSON())
	state, err := scanShapeStateRow(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return state, err
}

type shapeStateScanner interface {
	Scan(dest ...interface{}) error
}

func scanShapeStateRow(row shapeStateScanner) (*models.ShapeState, error) {
	state := &models.ShapeState{}
	var dataJSON, stackJSON, status string
	var completedAt, notes sql.NullString
	var createdAt, updatedAt time.Time

	err := row.Scan(
		&state.ID,
		&state.NoteTemplateID,
		&state.ShapePath,
		&stackJSON,
		&state.Kind,
		&state.Title,
		&status,
		&state.Completed,
		&completedAt,
		&notes,
		&dataJSON,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	state.RepeatStack = models.ParseRepeatStackJSON(stackJSON)
	state.Status = models.InstanceStatus(status)
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
// When stack is nil, all scopes are returned. Otherwise filters to that repeat context.
func ListShapeStates(db *sql.DB, noteTemplateID int64, stack models.RepeatStack) ([]*models.ShapeState, error) {
	var (
		query string
		args  []interface{}
	)
	if stack == nil {
		query = `
			SELECT` + shapeStateSelectColumns + `
			FROM note_shape_state
			WHERE note_template_id = ?
			ORDER BY repeat_stack_json ASC, shape_path ASC
		`
		args = []interface{}{noteTemplateID}
	} else {
		query = `
			SELECT` + shapeStateSelectColumns + `
			FROM note_shape_state
			WHERE note_template_id = ? AND repeat_stack_json = ?
			ORDER BY shape_path ASC
		`
		args = []interface{}{noteTemplateID, stack.JSON()}
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list shape states: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return scanShapeStates(rows)
}

// ListAllShapeStates returns every shape state row for a note template.
func ListAllShapeStates(db *sql.DB, noteTemplateID int64) ([]*models.ShapeState, error) {
	return ListShapeStates(db, noteTemplateID, nil)
}

// ListTopLevelProcedureStates returns top-level procedure rows for the steps panel.
func ListTopLevelProcedureStates(db *sql.DB, noteTemplateID int64) ([]*models.ShapeState, error) {
	query := `
		SELECT` + shapeStateSelectColumns + `
		FROM note_shape_state
		WHERE note_template_id = ? AND repeat_stack_json = '[]' AND kind = ?
		  AND shape_path NOT LIKE '%.apply_one.%' AND shape_path NOT LIKE '%.apply_one'
		ORDER BY id ASC
	`
	rows, err := db.Query(query, noteTemplateID, models.ShapeProcedure)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return scanShapeStates(rows)
}

func scanShapeStates(rows *sql.Rows) ([]*models.ShapeState, error) {
	var states []*models.ShapeState
	for rows.Next() {
		state, err := scanShapeStateRow(rows)
		if err != nil {
			return nil, err
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
		SET completed = ?, completed_at = ?, notes = ?, data = ?, status = ?, updated_at = ?
		WHERE id = ?
	`

	var completedAt interface{}
	if state.Completed {
		completedAt = now
		if state.Status == "" || state.Status == models.StatusNotStarted {
			state.Status = models.StatusComplete
		}
	}

	_, err = db.Exec(query, state.Completed, completedAt, state.Notes, string(dataJSON), state.Status, now, state.ID)
	if err != nil {
		return fmt.Errorf("failed to update shape state: %w", err)
	}
	state.UpdatedAt = now.Format(time.RFC3339)
	return nil
}

// ToggleChecklistItem updates one checklist item and auto-completes the parent when all items are done.
func ToggleChecklistItem(db *sql.DB, runtime *ShapeRuntime, state *models.ShapeState, itemID string, completed bool) error {
	if completed {
		if runtime != nil {
			if err := runtime.EnsureUnblocked(state); err != nil {
				return err
			}
		}
	}
	if state.Data.ItemCompletion == nil {
		state.Data.ItemCompletion = map[string]bool{}
	}
	state.Data.ItemCompletion[itemID] = completed
	state.Completed = models.AllChecklistItemsComplete(state)
	if state.Completed {
		state.Status = models.StatusComplete
	} else if state.Status == models.StatusComplete {
		state.Status = models.StatusInProgress
	} else if state.Status == models.StatusNotStarted {
		state.Status = models.StatusInProgress
	}
	if err := UpdateShapeState(db, state); err != nil {
		return err
	}
	if err := syncParentProcedureFromChecklist(db, state); err != nil {
		return err
	}
	if runtime != nil {
		return runtime.Refresh(db)
	}
	return nil
}

func syncParentProcedureFromChecklist(db *sql.DB, checklistState *models.ShapeState) error {
	parts := strings.Split(checklistState.ShapePath, ".")
	if len(parts) < 2 {
		return nil
	}
	parentPath := strings.Join(parts[:len(parts)-1], ".")
	parent, err := GetShapeState(db, checklistState.NoteTemplateID, parentPath, checklistState.RepeatStack)
	if err != nil {
		return err
	}
	if parent == nil || parent.Kind != models.ShapeProcedure {
		return nil
	}
	parent.Completed = checklistState.Completed
	if parent.Completed {
		parent.Status = models.StatusComplete
	}
	return UpdateShapeState(db, parent)
}

// ToggleShapeComplete sets completion on a procedure or action shape state.
func ToggleShapeComplete(db *sql.DB, runtime *ShapeRuntime, state *models.ShapeState, completed bool) error {
	if completed {
		if runtime != nil {
			if err := runtime.EnsureUnblocked(state); err != nil {
				return err
			}
		}
	}
	state.Completed = completed
	if completed {
		state.Status = models.StatusComplete
	} else {
		state.Status = models.StatusInProgress
	}
	if err := UpdateShapeState(db, state); err != nil {
		return err
	}
	if runtime != nil {
		return runtime.Refresh(db)
	}
	return nil
}

// UpdateShapeNotes persists notes on a shape state row.
func UpdateShapeNotes(db *sql.DB, state *models.ShapeState, notes string) error {
	state.Notes = notes
	return UpdateShapeState(db, state)
}

// ListProcedureStatesForScope returns procedure shape states under the current navigation scope.
func ListProcedureStatesForScope(db *sql.DB, noteTemplateID int64, stack models.RepeatStack, scopePath []string) ([]*models.ShapeState, error) {
	states, err := ListShapeStates(db, noteTemplateID, stack)
	if err != nil {
		return nil, err
	}
	if len(scopePath) == 0 {
		return ListTopLevelProcedureStates(db, noteTemplateID)
	}

	prefix := models.ShapePath(scopePath) + "."
	var scoped []*models.ShapeState
	for _, state := range states {
		if state.Kind != models.ShapeProcedure {
			continue
		}
		if strings.HasPrefix(state.ShapePath, prefix) && strings.Count(state.ShapePath[len(prefix):], ".") == 0 {
			scoped = append(scoped, state)
		}
	}
	return scoped, nil
}

// ComputeTemplateProgress calculates overall completion for a templated note instance.
func ComputeTemplateProgress(db *sql.DB, noteTemplateID int64, template *models.Template, inputs map[string]interface{}) (float64, error) {
	states, err := ListAllShapeStates(db, noteTemplateID)
	if err != nil {
		return 0, err
	}

	comp, err := template.Definition.GetStructure()
	if err != nil || comp == nil {
		return 0, err
	}

	recordCount, err := CountTemplateRecords(db, noteTemplateID, nil)
	if err != nil {
		return 0, err
	}

	progress := engine.EvaluateProgress(comp, states, inputs, recordCount)
	return progress.Fraction(), nil
}
