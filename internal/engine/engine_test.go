package engine_test

import (
	"testing"

	"yoo/internal/engine"
	"yoo/internal/models"
)

func TestEvaluateNestedRepeatProgress(t *testing.T) {
	root := models.NormalizeShapeTree(&models.ShapeNode{
		ID:   "root",
		Kind: models.ShapeChecklist,
		Title: "Parent",
		Repeat: &models.RepeatSpec{Count: "2"},
		Items: []models.ShapeNode{
			{
				ID:    "child-list",
				Kind:  models.ShapeChecklist,
				Title: "Child checklist",
				Repeat: &models.RepeatSpec{Count: "2"},
				Items: []models.ShapeNode{
					{ID: "child-a", Kind: models.ShapeAction, Title: "Child A"},
					{ID: "child-b", Kind: models.ShapeAction, Title: "Child B"},
				},
			},
		},
	})

	states := []*models.ShapeState{
		{
			ShapePath:   "root.child-list",
			RepeatStack: models.RepeatStack{{ShapeID: "root", Index: 1}, {ShapeID: "child-list", Index: 1}},
			Kind:        models.ShapeChecklist,
			Data: models.ShapeStateData{
				ItemCompletion: map[string]bool{"child-a": true, "child-b": false},
			},
		},
		{
			ShapePath:   "root.child-list",
			RepeatStack: models.RepeatStack{{ShapeID: "root", Index: 1}, {ShapeID: "child-list", Index: 2}},
			Kind:        models.ShapeChecklist,
			Data: models.ShapeStateData{
				ItemCompletion: map[string]bool{"child-a": false, "child-b": false},
			},
		},
		{
			ShapePath:   "root.child-list",
			RepeatStack: models.RepeatStack{{ShapeID: "root", Index: 2}, {ShapeID: "child-list", Index: 1}},
			Kind:        models.ShapeChecklist,
			Data: models.ShapeStateData{
				ItemCompletion: map[string]bool{"child-a": false, "child-b": false},
			},
		},
		{
			ShapePath:   "root.child-list",
			RepeatStack: models.RepeatStack{{ShapeID: "root", Index: 2}, {ShapeID: "child-list", Index: 2}},
			Kind:        models.ShapeChecklist,
			Data: models.ShapeStateData{
				ItemCompletion: map[string]bool{"child-a": false, "child-b": false},
			},
		},
	}

	progress := engine.EvaluateProgress(root, states, nil, 0)
	if progress.Total != 8 {
		t.Fatalf("expected 8 total units, got %d", progress.Total)
	}
	if progress.Done != 1 {
		t.Fatalf("expected 1 done unit, got %d", progress.Done)
	}
}

func TestDependencyBlocksUntilComplete(t *testing.T) {
	root := models.NormalizeShapeTree(&models.ShapeNode{
		ID:   "root",
		Kind: models.ShapeProcedure,
		Steps: []models.ShapeNode{
			{ID: "a", Kind: models.ShapeAction, Title: "A"},
			{
				ID:   "b",
				Kind: models.ShapeAction,
				Title: "B",
				DependsOn: []models.DependencySpec{{
					Target:      models.TargetRef{ShapeID: "a"},
					Requirement: models.RequirementCompleted,
					Scope:       models.ScopeSameRepeatContext,
				}},
			},
		},
	})

	states := []*models.ShapeState{}
	blockers := engine.FindBlockers(root, states, nil)
	if len(blockers) != 1 {
		t.Fatalf("expected one blocker, got %d", len(blockers))
	}
}
