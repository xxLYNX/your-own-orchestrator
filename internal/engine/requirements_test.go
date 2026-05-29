package engine_test

import (
	"testing"

	"yoo/internal/engine"
	"yoo/internal/models"
)

func TestHasRecordSameRepeatContext(t *testing.T) {
	root := models.NormalizeShapeTree(&models.ShapeNode{
		ID:             "root",
		Kind:           models.ShapeProcedure,
		DependencyMode: models.ProcedureManual,
		Steps: []models.ShapeNode{
			{
				ID:     "applications",
				Kind:   models.ShapeProcedure,
				Repeat: &models.RepeatSpec{Count: "target_count"},
				Steps: []models.ShapeNode{
					{ID: "application-log", Kind: models.ShapeLog, RecordSchema: &models.RecordSchema{
						Fields: []models.RecordField{{Name: "company", Type: "text", Required: true}},
					}},
					{
						ID:   "confirm-logged",
						Kind: models.ShapeAction,
						DependsOn: []models.DependencySpec{{
							Target:      models.TargetRef{ShapeID: "application-log"},
							Requirement: models.RequirementHasRecord,
							Scope:       models.ScopeSameRepeatContext,
						}},
					},
				},
			},
		},
	})

	inputs := map[string]interface{}{"target_count": 10}
	evalEmpty := engine.EvalContext{RecordCountByStack: map[string]int{}}
	blockers := engine.FindBlockers(root, nil, inputs, evalEmpty)
	if len(blockers) != 10 {
		t.Fatalf("expected has_record blocker for each iteration, got %d", len(blockers))
	}

	stack := models.RepeatStack{{ShapeID: "applications", Index: 3}}
	evalWithRecord := engine.EvalContext{RecordCountByStack: map[string]int{
		stack.JSON(): 1,
	}}
	blockers = engine.FindBlockers(root, nil, inputs, evalWithRecord)
	if len(blockers) != 9 {
		t.Fatalf("expected 9 blockers when only iteration 3 is logged, got %d", len(blockers))
	}
}

func TestHasRecordAllOccurrences(t *testing.T) {
	root := models.NormalizeShapeTree(&models.ShapeNode{
		ID:             "root",
		Kind:           models.ShapeProcedure,
		DependencyMode: models.ProcedureManual,
		Steps: []models.ShapeNode{
			{
				ID:     "applications",
				Kind:   models.ShapeProcedure,
				Repeat: &models.RepeatSpec{Count: "target_count"},
				Steps: []models.ShapeNode{
					{ID: "application-log", Kind: models.ShapeLog, RecordSchema: &models.RecordSchema{
						Fields: []models.RecordField{{Name: "company", Type: "text", Required: true}},
					}},
				},
			},
			{
				ID:   "wrapup",
				Kind: models.ShapeAction,
				DependsOn: []models.DependencySpec{{
					Target:      models.TargetRef{ShapeID: "application-log"},
					Requirement: models.RequirementHasRecord,
					Scope:       models.ScopeAllOccurrences,
				}},
			},
		},
	})

	inputs := map[string]interface{}{"target_count": 3}
	evalPartial := engine.EvalContext{RecordCountByStack: map[string]int{
		`[{"shape_id":"applications","index":1}]`: 1,
		`[{"shape_id":"applications","index":2}]`: 1,
	}}
	blockers := engine.FindBlockers(root, nil, inputs, evalPartial)
	if len(blockers) != 1 {
		t.Fatalf("expected blocker until all 3 iterations logged, got %d", len(blockers))
	}

	evalAll := engine.EvalContext{RecordCountByStack: map[string]int{
		`[{"shape_id":"applications","index":1}]`: 1,
		`[{"shape_id":"applications","index":2}]`: 1,
		`[{"shape_id":"applications","index":3}]`: 1,
	}}
	blockers = engine.FindBlockers(root, nil, inputs, evalAll)
	if len(blockers) != 0 {
		t.Fatalf("expected all_occurrences satisfied with 3 records, got %d blockers", len(blockers))
	}
}

func TestIsInstanceBlocked(t *testing.T) {
	blockers := []engine.Blocker{{
		ShapePath:   "root.b",
		RepeatStack: nil,
		Dependency: models.DependencySpec{
			Target:      models.TargetRef{ShapeID: "a"},
			Requirement: models.RequirementCompleted,
		},
	}}
	if !engine.IsInstanceBlocked(blockers, "root.b", nil) {
		t.Fatal("expected instance blocked")
	}
	if engine.IsInstanceBlocked(blockers, "root.a", nil) {
		t.Fatal("expected instance not blocked")
	}
}
