package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"yoo/internal/database"
	"yoo/internal/models"
	"yoo/internal/strutil"

	"github.com/spf13/cobra"
)

var stepsCmd = &cobra.Command{
	Use:     "step",
	Aliases: []string{"steps"},
	Short:   "Manage procedure states for templated notes",
	Long:    `List and update top-level procedure shape states for templated notes.`,
}

func loadProcedureState(db *database.DB, noteID int64, stepNumber int) (*database.TemplatedNoteContext, *models.ShapeState, error) {
	ctx, err := database.LoadTemplatedNoteContext(db.Conn(), noteID)
	if err != nil {
		return nil, nil, err
	}
	states, err := database.ListTopLevelProcedureStates(db.Conn(), ctx.NoteTemplate.ID)
	if err != nil {
		return nil, nil, err
	}
	if stepNumber < 1 || stepNumber > len(states) {
		return nil, nil, fmt.Errorf("step %d not found", stepNumber)
	}
	return ctx, states[stepNumber-1], nil
}

var stepListCmd = &cobra.Command{
	Use:   "list <note-id>",
	Short: "List top-level procedure states for a templated note",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runForTemplatedNoteArg(args, func(db *database.DB, ctx *database.TemplatedNoteContext) error {
			states, err := database.ListTopLevelProcedureStates(db.Conn(), ctx.NoteTemplate.ID)
			if err != nil {
				return fmt.Errorf("failed to list procedure states: %w", err)
			}
			if len(states) == 0 {
				fmt.Printf("No procedure states found for note: %s\n", ctx.Note.Title)
				return nil
			}

			fmt.Printf("Procedures for: %s\n\n", ctx.Note.Title)
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			_, _ = fmt.Fprintln(w, "STATUS\tSTEP\tTITLE\tCOMPLETED")
			_, _ = fmt.Fprintln(w, "------\t----\t-----\t---------")

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
				title := strutil.Truncate(state.Title, 60)
				_, _ = fmt.Fprintf(w, "%s\t%d\t%s\t%s\n", status, i+1, title, completedAt)
			}
			_ = w.Flush()
			fmt.Printf("\nProgress: %d/%d complete\n", done, len(states))
			return nil
		})
	},
}

var stepCompleteCmd = &cobra.Command{
	Use:   "complete <note-id> <step-number>",
	Short: "Mark a procedure state complete",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		noteID, err := parseNoteIDArg(args[0])
		if err != nil {
			return err
		}
		stepNumber, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid step number: %w", err)
		}
		return setStepCompletion(noteID, stepNumber, true)
	},
}

var stepUncompleteCmd = &cobra.Command{
	Use:   "uncomplete <note-id> <step-number>",
	Short: "Mark a procedure state incomplete",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		noteID, err := parseNoteIDArg(args[0])
		if err != nil {
			return err
		}
		stepNumber, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid step number: %w", err)
		}
		return setStepCompletion(noteID, stepNumber, false)
	},
}

var stepNoteCmd = &cobra.Command{
	Use:   "note <note-id> <step-number> <text>",
	Short: "Add notes to a procedure state",
	Args:  cobra.MinimumNArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		noteID, err := parseNoteIDArg(args[0])
		if err != nil {
			return err
		}
		stepNumber, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid step number: %w", err)
		}
		noteText := strings.Join(args[2:], " ")

		return runForTemplatedNote(noteID, func(db *database.DB, ctx *database.TemplatedNoteContext) error {
			_, state, err := loadProcedureState(db, noteID, stepNumber)
			if err != nil {
				return err
			}
			if err := database.UpdateShapeNotes(db.Conn(), state, noteText); err != nil {
				return err
			}
			fmt.Printf("✓ Notes saved for step %d on note: %s\n", stepNumber, ctx.Note.Title)
			return nil
		})
	},
}

func init() {
	rootCmd.AddCommand(stepsCmd)
	stepsCmd.AddCommand(stepListCmd)
	stepsCmd.AddCommand(stepCompleteCmd)
	stepsCmd.AddCommand(stepUncompleteCmd)
	stepsCmd.AddCommand(stepNoteCmd)
}
