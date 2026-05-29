package database

import (
	"database/sql"
	"math"
)

// RefreshNoteProgress recomputes and caches progress for one schedule note.
// Templated notes use shape completion units (checklist items, stages, log targets).
// Plain notes are 0% or 100% from note status.
func RefreshNoteProgress(db *sql.DB, note *Note) (float64, error) {
	if note == nil {
		return 0, nil
	}

	var frac float64
	var err error

	switch {
	case note.IsTemplated:
		ctx, loadErr := LoadTemplatedNoteContext(db, note.ID)
		if loadErr != nil {
			return 0, loadErr
		}
		frac, err = ComputeTemplateProgress(db, ctx.NoteTemplate.ID, ctx.Template, ctx.NoteTemplate.TemplateData.Inputs)
	default:
		if note.Status == "completed" {
			frac = 1
		}
	}

	if err != nil {
		return 0, err
	}

	note.TemplateProgress = frac
	if updateErr := UpdateTemplateProgress(db, note.ID, frac); updateErr != nil {
		return frac, updateErr
	}
	return frac, nil
}

// RefreshScheduleNotes refreshes every note and returns the day's equal-weight average.
func RefreshScheduleNotes(db *sql.DB, notes []*Note) (float64, error) {
	if len(notes) == 0 {
		return 0, nil
	}

	var sum float64
	for _, note := range notes {
		frac, err := RefreshNoteProgress(db, note)
		if err != nil {
			return 0, err
		}
		sum += frac
	}
	return sum / float64(len(notes)), nil
}

// AverageProgress returns the mean of fractions in [0,1], clamping the result.
func AverageProgress(fractions []float64) float64 {
	if len(fractions) == 0 {
		return 0
	}
	var sum float64
	for _, f := range fractions {
		sum += f
	}
	avg := sum / float64(len(fractions))
	return math.Min(1, math.Max(0, avg))
}
