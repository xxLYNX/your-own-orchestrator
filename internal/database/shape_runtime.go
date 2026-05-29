package database

import (
	"database/sql"
	"fmt"

	"yoo/internal/engine"
	"yoo/internal/models"
)

// ShapeRuntime loads structure, persisted state, and dependency evaluation context.
type ShapeRuntime struct {
	NoteTemplateID int64
	Root           *models.ShapeNode
	States         []*models.ShapeState
	Inputs         map[string]interface{}
	Eval           engine.EvalContext
	blockers       []engine.Blocker
}

// LoadShapeRuntime builds dependency evaluation context for a templated note instance.
func LoadShapeRuntime(db *sql.DB, noteTemplateID int64, template *models.Template, inputs map[string]interface{}) (*ShapeRuntime, error) {
	states, err := ListAllShapeStates(db, noteTemplateID)
	if err != nil {
		return nil, err
	}

	root, err := template.Definition.GetStructure()
	if err != nil {
		return nil, err
	}

	eval, err := buildEvalContext(db, noteTemplateID)
	if err != nil {
		return nil, err
	}

	runtime := &ShapeRuntime{
		NoteTemplateID: noteTemplateID,
		Root:           root,
		States:         states,
		Inputs:         inputs,
		Eval:           eval,
		blockers:       engine.FindBlockers(root, states, inputs, eval),
	}
	return runtime, nil
}

func buildEvalContext(db *sql.DB, noteTemplateID int64) (engine.EvalContext, error) {
	eval := engine.EvalContext{
		RecordCountByStack: map[string]int{},
	}

	records, err := ListTemplateRecords(db, noteTemplateID, nil)
	if err != nil {
		return eval, err
	}
	eval.RecordCountTotal = len(records)
	for _, record := range records {
		eval.RecordCountByStack[record.RepeatStack.JSON()]++
	}

	artifacts, err := ListArtifacts(db, noteTemplateID)
	if err != nil {
		return eval, err
	}
	eval.ArtifactCount = len(artifacts)
	return eval, nil
}

// Refresh reloads shape states and recomputes blockers after a mutation.
func (r *ShapeRuntime) Refresh(db *sql.DB) error {
	if r == nil {
		return nil
	}
	states, err := ListAllShapeStates(db, r.NoteTemplateID)
	if err != nil {
		return err
	}
	eval, err := buildEvalContext(db, r.NoteTemplateID)
	if err != nil {
		return err
	}
	r.States = states
	r.Eval = eval
	r.blockers = engine.FindBlockers(r.Root, states, r.Inputs, eval)
	return nil
}

// IsBlocked reports whether forward progress on an instance should be rejected.
func (r *ShapeRuntime) IsBlocked(state *models.ShapeState) bool {
	if r == nil || state == nil {
		return false
	}
	return engine.IsInstanceBlocked(r.blockers, state.ShapePath, state.RepeatStack)
}

// EnsureUnblocked returns ErrBlocked when dependencies prevent forward progress.
func (r *ShapeRuntime) EnsureUnblocked(state *models.ShapeState) error {
	if r == nil || state == nil {
		return nil
	}
	blockers := engine.BlockersForInstance(r.blockers, state.ShapePath, state.RepeatStack)
	if len(blockers) == 0 {
		return nil
	}
	return &engine.ErrBlocked{
		ShapePath:   state.ShapePath,
		RepeatStack: state.RepeatStack,
		Blockers:    blockers,
	}
}

// BlockedShapePaths returns shape paths with unsatisfied dependencies in the current scope.
func (r *ShapeRuntime) BlockedShapePaths() map[string]bool {
	out := map[string]bool{}
	if r == nil {
		return out
	}
	for _, blocker := range r.blockers {
		key := fmt.Sprintf("%s|%s", blocker.ShapePath, blocker.RepeatStack.JSON())
		out[key] = true
	}
	return out
}

// IsShapeBlocked reports whether a path/stack pair is blocked.
func (r *ShapeRuntime) IsShapeBlocked(shapePath string, stack models.RepeatStack) bool {
	if r == nil {
		return false
	}
	return engine.IsInstanceBlocked(r.blockers, shapePath, stack)
}
