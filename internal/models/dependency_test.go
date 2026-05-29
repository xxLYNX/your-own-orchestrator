package models_test

import (
	"testing"

	"yoo/internal/models"
)

func TestInvalidDependencyRequirementRejected(t *testing.T) {
	root := &models.ShapeNode{
		ID:   "root",
		Kind: models.ShapeProcedure,
		Steps: []models.ShapeNode{
			{
				ID:   "b",
				Kind: models.ShapeAction,
				DependsOn: []models.DependencySpec{{
					Target:      models.TargetRef{ShapeID: "a"},
					Requirement: "approved",
				}},
			},
		},
	}
	def := &models.TemplateDefinition{Structure: root}
	if err := models.NormalizeTemplateDefinition(def); err == nil {
		t.Fatal("expected validation error for removed approved requirement")
	}
}
