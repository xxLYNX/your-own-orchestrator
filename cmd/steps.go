package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"yoo/internal/config"
	"yoo/internal/database"
	"yoo/internal/models"

	"github.com/spf13/cobra"
)

var stepsCmd = &cobra.Command{
	Use:     "step",
	Aliases: []string{"steps"},
	Short:   "Manage procedure states for templated notes",
	Long:    `List and update top-level procedure shape states for templated notes.`,
}

func loadProcedureState(db *database.DB, noteID int64, stepNumber int) (*database.Note, *models.NoteTemplate, *models.ShapeState, error) {
	note, err := database.GetNote(db.Conn(), noteID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get note: %w", err)
	}
	noteTemplate, err := database.GetNoteTemplate(db.Conn(), noteID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("note is not templated: %w", err)
	}
	states, err := database.ListTopLevelProcedureStates(db.Conn(), noteTemplate.ID)
	if err != nil {
		return nil, nil, nil, err
	}
	if stepNumber < 1 || stepNumber > len(states) {
		return nil, nil, nil, fmt.Errorf("step %d not found", stepNumber)
	}
	return note, noteTemplate, states[stepNumber-1], nil
}

var stepListCmd = &cobra.Command{
	Use:   "list <note-id>",
	Short: "List top-level procedure states for a templated note",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		noteID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid note ID: %w", err)
		}

		dbPath := config.GetDatabasePath()
		db, err := database.New(dbPath)
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		defer db.Close()

		note, err := database.GetNote(db.Conn(), noteID)
		if err != nil {
			return fmt.Errorf("failed to get note: %w", err)
		}
		noteTemplate, err := database.GetNoteTemplate(db.Conn(), noteID)
		if err != nil {
			return fmt.Errorf("note is not templated: %w", err)
		}
		states, err := database.ListTopLevelProcedureStates(db.Conn(), noteTemplate.ID)
		if err != nil {
			return fmt.Errorf("failed to list procedure states: %w", err)
		}
		if len(states) == 0 {
			fmt.Printf("No procedure states found for note: %s\n", note.Title)
			return nil
		}

		fmt.Printf("Procedures for: %s\n\n", note.Title)
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "STATUS\tSTEP\tTITLE\tCOMPLETED")
		fmt.Fprintln(w, "------\t----\t-----\t---------")

		done := 0
		for i, state := range states {
			status := "○"
			completedAt := ""
			if state.Completed {
				status = "✓"
				done++
				if state.CompletedAt != nil {
					completedAt = *state.CompletedAt
				}
			}
			title := state.Title
			if len(title) > 60 {
				title = title[:57] + "..."
			}
			fmt.Fprintf(w, "%s\t%d\t%s\t%s\n", status, i+1, title, completedAt)
		}
		w.Flush()
		fmt.Printf("\nProgress: %d/%d complete\n", done, len(states))
		return nil
	},
}

var stepCompleteCmd = &cobra.Command{
	Use:   "complete <note-id> <step-number>",
	Short: "Mark a procedure state complete",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		noteID, _ := strconv.ParseInt(args[0], 10, 64)
		stepNumber, _ := strconv.Atoi(args[1])

		dbPath := config.GetDatabasePath()
		db, err := database.New(dbPath)
		if err != nil {
			return err
		}
		defer db.Close()

		note, _, state, err := loadProcedureState(db, noteID, stepNumber)
		if err != nil {
			return err
		}
		if state.Completed {
			fmt.Printf("Step %d is already complete\n", stepNumber)
			return nil
		}
		if err := database.ToggleShapeComplete(db.Conn(), state, true); err != nil {
			return err
		}
		fmt.Printf("✓ Step %d marked complete: %s\n  Note: %s\n", stepNumber, state.Title, note.Title)
		return nil
	},
}

var stepUncompleteCmd = &cobra.Command{
	Use:   "uncomplete <note-id> <step-number>",
	Short: "Mark a procedure state incomplete",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		noteID, _ := strconv.ParseInt(args[0], 10, 64)
		stepNumber, _ := strconv.Atoi(args[1])

		dbPath := config.GetDatabasePath()
		db, err := database.New(dbPath)
		if err != nil {
			return err
		}
		defer db.Close()

		note, _, state, err := loadProcedureState(db, noteID, stepNumber)
		if err != nil {
			return err
		}
		if !state.Completed {
			fmt.Printf("Step %d is already incomplete\n", stepNumber)
			return nil
		}
		if err := database.ToggleShapeComplete(db.Conn(), state, false); err != nil {
			return err
		}
		fmt.Printf("○ Step %d marked incomplete: %s\n  Note: %s\n", stepNumber, state.Title, note.Title)
		return nil
	},
}

var stepNoteCmd = &cobra.Command{
	Use:   "note <note-id> <step-number> <text>",
	Short: "Add notes to a procedure state",
	Args:  cobra.MinimumNArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		noteID, _ := strconv.ParseInt(args[0], 10, 64)
		stepNumber, _ := strconv.Atoi(args[1])
		noteText := strings.Join(args[2:], " ")

		dbPath := config.GetDatabasePath()
		db, err := database.New(dbPath)
		if err != nil {
			return err
		}
		defer db.Close()

		note, _, state, err := loadProcedureState(db, noteID, stepNumber)
		if err != nil {
			return err
		}
		if err := database.UpdateShapeNotes(db.Conn(), state, noteText); err != nil {
			return err
		}
		fmt.Printf("✓ Notes saved for step %d on note: %s\n", stepNumber, note.Title)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(stepsCmd)
	stepsCmd.AddCommand(stepListCmd)
	stepsCmd.AddCommand(stepCompleteCmd)
	stepsCmd.AddCommand(stepUncompleteCmd)
	stepsCmd.AddCommand(stepNoteCmd)
}
