package cmd

import (
	"fmt"
	"strconv"

	"yoo/internal/database"
)

func parseNoteIDArg(arg string) (int64, error) {
	id, err := strconv.ParseInt(arg, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid note ID: %w", err)
	}
	return id, nil
}

func parseRecordIndexArg(arg string) (int, error) {
	idx, err := strconv.Atoi(arg)
	if err != nil {
		return 0, fmt.Errorf("invalid record index: %w", err)
	}
	return idx, nil
}

type templatedNoteFn func(db *database.DB, ctx *database.TemplatedNoteContext) error

func runForTemplatedNote(noteID int64, fn templatedNoteFn) error {
	return withDB(func(db *database.DB) error {
		ctx, err := database.LoadTemplatedNoteContext(db.Conn(), noteID)
		if err != nil {
			return err
		}
		return fn(db, ctx)
	})
}

func runForTemplatedNoteArg(args []string, fn templatedNoteFn) error {
	noteID, err := parseNoteIDArg(args[0])
	if err != nil {
		return err
	}
	return runForTemplatedNote(noteID, fn)
}

func setStepCompletion(noteID int64, stepNumber int, completed bool) error {
	return runForTemplatedNote(noteID, func(db *database.DB, ctx *database.TemplatedNoteContext) error {
		_, state, err := loadProcedureState(db, noteID, stepNumber)
		if err != nil {
			return err
		}
		if state.Completed == completed {
			if completed {
				fmt.Printf("Step %d is already complete\n", stepNumber)
			} else {
				fmt.Printf("Step %d is already incomplete\n", stepNumber)
			}
			return nil
		}
		runtime, err := database.LoadShapeRuntime(db.Conn(), ctx.NoteTemplate.ID, ctx.Template, ctx.NoteTemplate.TemplateData.Inputs)
		if err != nil {
			return err
		}
		if err := database.ToggleShapeComplete(db.Conn(), runtime, state, completed); err != nil {
			return err
		}
		icon := "○"
		if completed {
			icon = "✓"
		}
		fmt.Printf("%s Step %d marked %s: %s\n  Note: %s\n", icon, stepNumber, completionLabel(completed), state.Title, ctx.Note.Title)
		return nil
	})
}

func completionLabel(completed bool) string {
	if completed {
		return "complete"
	}
	return "incomplete"
}
