package database_test

import (
	"testing"

	"yoo/internal/database"
)

func TestAverageProgress(t *testing.T) {
	if got := database.AverageProgress(nil); got != 0 {
		t.Fatalf("expected 0 for empty, got %v", got)
	}
	got := database.AverageProgress([]float64{0, 0.5, 1})
	if got != 0.5 {
		t.Fatalf("expected 0.5, got %v", got)
	}
}
