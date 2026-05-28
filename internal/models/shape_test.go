package models_test

import (
	"testing"

	"yoo/internal/models"
)

func TestNormalizeRepeatModifier(t *testing.T) {
	root := &models.ShapeNode{
		ID:     "applications",
		Kind:   models.ShapeProcedure,
		Title:  "Applications",
		Repeat: &models.RepeatSpec{Count: "5"},
		Steps: []models.ShapeNode{
			{
				ID:    "child",
				Kind:  models.ShapeChecklist,
				Title: "Child checklist",
				Items: []models.ShapeNode{
					{ID: "a", Kind: models.ShapeAction, Title: "A"},
					{ID: "b", Kind: models.ShapeAction, Title: "B"},
				},
			},
		},
	}

	normalized := models.NormalizeShapeTree(root)
	if normalized.Kind != models.ShapeProcedure {
		t.Fatalf("expected procedure, got %s", normalized.Kind)
	}
	if normalized.Repeat == nil || normalized.Repeat.Count != "5" {
		t.Fatalf("expected repeat count 5, got %#v", normalized.Repeat)
	}
	if len(normalized.Steps) != 1 || len(normalized.Steps[0].Items) != 2 {
		t.Fatalf("expected checklist with 2 items, got %#v", normalized.Steps)
	}
}

func TestNestedRepeatStateInits(t *testing.T) {
	root := &models.ShapeNode{
		ID:     "root",
		Kind:   models.ShapeChecklist,
		Title:  "Parent",
		Repeat: &models.RepeatSpec{Count: "5"},
		Items: []models.ShapeNode{
			{ID: "item-1", Kind: models.ShapeAction, Title: "Item 1"},
			{
				ID:     "child-list",
				Kind:   models.ShapeChecklist,
				Title:  "Child checklist",
				Repeat: &models.RepeatSpec{Count: "4"},
				Items: []models.ShapeNode{
					{ID: "child-a", Kind: models.ShapeAction, Title: "Child A"},
					{ID: "child-b", Kind: models.ShapeAction, Title: "Child B"},
				},
			},
			{ID: "item-3", Kind: models.ShapeAction, Title: "Item 3"},
		},
	}
	root = models.NormalizeShapeTree(root)

	inits := root.CollectShapeStateInits([]string{}, nil, map[string]interface{}{})
	childActionCount := 0
	for _, init := range inits {
		if init.Kind == models.ShapeAction && init.Title == "Child A" {
			childActionCount++
		}
	}
	if childActionCount != 20 {
		t.Fatalf("expected 20 child-a occurrences, got %d", childActionCount)
	}
}

func TestRepeatStackJSONRoundTrip(t *testing.T) {
	stack := models.RepeatStack{
		{ShapeID: "parent", Index: 2},
		{ShapeID: "child", Index: 4},
	}
	raw := stack.JSON()
	parsed := models.ParseRepeatStackJSON(raw)
	if !stack.Equal(parsed) {
		t.Fatalf("stack round trip failed: %#v vs %#v", stack, parsed)
	}
}

func TestSimpleNoteStructure(t *testing.T) {
	root := models.SimpleNoteStructure("Buy milk")
	if root.Kind != models.ShapeAction {
		t.Fatalf("expected action kind, got %s", root.Kind)
	}
	if root.EffectiveRepeatCount(nil) != 1 {
		t.Fatalf("expected implicit repeat 1, got %d", root.EffectiveRepeatCount(nil))
	}
}

func TestProcedureSequentialDependencies(t *testing.T) {
	root := &models.ShapeNode{
		ID:   "root",
		Kind: models.ShapeProcedure,
		Steps: []models.ShapeNode{
			{ID: "a", Kind: models.ShapeAction, Title: "A"},
			{ID: "b", Kind: models.ShapeAction, Title: "B"},
			{ID: "c", Kind: models.ShapeAction, Title: "C"},
		},
	}
	root = models.NormalizeShapeTree(root)
	if len(root.Steps[1].DependsOn) != 1 || root.Steps[1].DependsOn[0].Target.ShapeID != "a" {
		t.Fatalf("expected b depends on a, got %#v", root.Steps[1].DependsOn)
	}
	if len(root.Steps[2].DependsOn) != 1 || root.Steps[2].DependsOn[0].Target.ShapeID != "b" {
		t.Fatalf("expected c depends on b, got %#v", root.Steps[2].DependsOn)
	}
}
