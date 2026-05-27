package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"yoo/internal/config"
	"yoo/internal/database"

	"github.com/spf13/cobra"
)

// stepsCmd represents the steps command
var stepsCmd = &cobra.Command{
	Use:     "step",
	Aliases: []string{"steps"},
	Short:   "Manage steps for templated notes",
	Long: `Manage procedural steps for templated notes.

Track progress through multi-step workflows by marking steps complete,
adding notes, and monitoring overall completion status.`,
}

// stepListCmd lists all steps for a note
var stepListCmd = &cobra.Command{
	Use:   "list <note-id>",
	Short: "List all steps for a templated note",
	Long: `List all steps for a templated note with completion status.

Shows each step with its number, title, and completion indicator.

Examples:
  yoo step list 42
  yoo steps list 42`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		noteID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid note ID: %w", err)
		}

		// Initialize database
		dbPath := config.GetDatabasePath()
		db, err := database.New(dbPath)
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		defer db.Close()

		// Get the note
		note, err := database.GetNote(db.Conn(), noteID)
		if err != nil {
			return fmt.Errorf("failed to get note: %w", err)
		}

		// Get the template link
		noteTemplate, err := database.GetNoteTemplate(db.Conn(), noteID)
		if err != nil {
			return fmt.Errorf("note is not templated: %w", err)
		}

		// Get all steps
		steps, err := database.ListNoteSteps(db.Conn(), noteTemplate.ID)
		if err != nil {
			return fmt.Errorf("failed to list steps: %w", err)
		}

		if len(steps) == 0 {
			fmt.Printf("No steps found for note: %s\n", note.Title)
			return nil
		}

		// Display steps
		fmt.Printf("Steps for: %s\n\n", note.Title)

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "STATUS\tSTEP\tTITLE\tCOMPLETED")
		fmt.Fprintln(w, "------\t----\t-----\t---------")

		completedCount := 0
		for _, step := range steps {
			status := "○"
			completedAt := ""
			if step.Completed {
				status = "✓"
				completedCount++
				if step.CompletedAt != nil {
					completedAt = step.CompletedAt.Format("2006-01-02")
				}
			}

			title := step.Title
			if len(title) > 60 {
				title = title[:57] + "..."
			}

			fmt.Fprintf(w, "%s\t%d\t%s\t%s\n",
				status,
				step.StepNumber,
				title,
				completedAt,
			)
		}
		w.Flush()

		// Display progress summary
		percentage := float64(completedCount) / float64(len(steps)) * 100
		fmt.Printf("\nProgress: %d/%d steps completed (%.0f%%)\n", completedCount, len(steps), percentage)

		return nil
	},
}

// stepShowCmd shows details of a specific step
var stepShowCmd = &cobra.Command{
	Use:   "show <note-id> <step-number>",
	Short: "Show details of a specific step",
	Long: `Display detailed information about a specific step including checklist items.

Shows the step title, description, completion status, notes, and any checklist items
defined in the template.

Examples:
  yoo step show 42 1
  yoo step show 42 3`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		noteID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid note ID: %w", err)
		}

		stepNumber, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid step number: %w", err)
		}

		// Initialize database
		dbPath := config.GetDatabasePath()
		db, err := database.New(dbPath)
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		defer db.Close()

		// Get the note
		note, err := database.GetNote(db.Conn(), noteID)
		if err != nil {
			return fmt.Errorf("failed to get note: %w", err)
		}

		// Get the template link
		noteTemplate, err := database.GetNoteTemplate(db.Conn(), noteID)
		if err != nil {
			return fmt.Errorf("note is not templated: %w", err)
		}

		// Get the template definition
		template, err := database.GetTemplateByID(db.Conn(), noteTemplate.TemplateID)
		if err != nil {
			return fmt.Errorf("failed to get template: %w", err)
		}

		// Get the step instance
		step, err := database.GetNoteStep(db.Conn(), noteTemplate.ID, stepNumber)
		if err != nil {
			return fmt.Errorf("failed to get step: %w", err)
		}

		// Display step details
		fmt.Printf("Note: %s\n", note.Title)
		fmt.Printf("Step %d: %s\n\n", step.StepNumber, step.Title)

		// Status
		if step.Completed {
			fmt.Println("Status: ✓ Complete")
			if step.CompletedAt != nil {
				fmt.Printf("Completed: %s\n", step.CompletedAt.Format("2006-01-02 15:04:05"))
			}
		} else {
			fmt.Println("Status: ○ Incomplete")
		}
		fmt.Println()

		// Description
		if step.Description != "" {
			fmt.Println("Description:")
			fmt.Println(step.Description)
			fmt.Println()
		}

		// Get checklist from template definition
		var templateStep *struct {
			Checklist     []string
			EstimatedTime string
		}

		for _, ts := range template.Definition.Steps {
			if ts.ID == stepNumber {
				templateStep = &struct {
					Checklist     []string
					EstimatedTime string
				}{
					Checklist:     ts.Checklist,
					EstimatedTime: ts.EstimatedTime,
				}
				break
			}
		}

		// Display checklist
		if templateStep != nil && len(templateStep.Checklist) > 0 {
			fmt.Println("Checklist:")
			for _, item := range templateStep.Checklist {
				fmt.Printf("  ☐ %s\n", item)
			}
			fmt.Println()
		}

		// Estimated time
		if templateStep != nil && templateStep.EstimatedTime != "" {
			fmt.Printf("Estimated Time: %s\n\n", templateStep.EstimatedTime)
		}

		// Notes
		if step.Notes != "" {
			fmt.Println("Notes:")
			fmt.Println(step.Notes)
			fmt.Println()
		}

		return nil
	},
}

// stepCompleteCmd marks a step as complete
var stepCompleteCmd = &cobra.Command{
	Use:   "complete <note-id> <step-number>",
	Short: "Mark a step as complete",
	Long: `Mark a step as complete and record the completion timestamp.

Examples:
  yoo step complete 42 1
  yoo step complete 42 3`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		noteID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid note ID: %w", err)
		}

		stepNumber, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid step number: %w", err)
		}

		// Initialize database
		dbPath := config.GetDatabasePath()
		db, err := database.New(dbPath)
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		defer db.Close()

		// Get the note
		note, err := database.GetNote(db.Conn(), noteID)
		if err != nil {
			return fmt.Errorf("failed to get note: %w", err)
		}

		// Get the template link
		noteTemplate, err := database.GetNoteTemplate(db.Conn(), noteID)
		if err != nil {
			return fmt.Errorf("note is not templated: %w", err)
		}

		// Get the step to verify it exists and get its title
		step, err := database.GetNoteStep(db.Conn(), noteTemplate.ID, stepNumber)
		if err != nil {
			return fmt.Errorf("failed to get step: %w", err)
		}

		// Check if already completed
		if step.Completed {
			fmt.Printf("Step %d is already marked as complete\n", stepNumber)
			return nil
		}

		// Mark as complete
		if err := database.CompleteStep(db.Conn(), noteTemplate.ID, stepNumber); err != nil {
			return fmt.Errorf("failed to complete step: %w", err)
		}

		fmt.Printf("✓ Step %d marked as complete: %s\n", stepNumber, step.Title)
		fmt.Printf("  Note: %s\n", note.Title)

		// Show progress
		steps, err := database.ListNoteSteps(db.Conn(), noteTemplate.ID)
		if err == nil {
			completedCount := 0
			for _, s := range steps {
				if s.Completed {
					completedCount++
				}
			}
			percentage := float64(completedCount) / float64(len(steps)) * 100
			fmt.Printf("\nProgress: %d/%d steps completed (%.0f%%)\n", completedCount, len(steps), percentage)
		}

		return nil
	},
}

// stepUncompleteCmd marks a step as incomplete
var stepUncompleteCmd = &cobra.Command{
	Use:   "uncomplete <note-id> <step-number>",
	Short: "Mark a step as incomplete",
	Long: `Mark a step as incomplete and clear the completion timestamp.

Use this to revert a step back to incomplete status.

Examples:
  yoo step uncomplete 42 1
  yoo step uncomplete 42 3`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		noteID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid note ID: %w", err)
		}

		stepNumber, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid step number: %w", err)
		}

		// Initialize database
		dbPath := config.GetDatabasePath()
		db, err := database.New(dbPath)
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		defer db.Close()

		// Get the note
		note, err := database.GetNote(db.Conn(), noteID)
		if err != nil {
			return fmt.Errorf("failed to get note: %w", err)
		}

		// Get the template link
		noteTemplate, err := database.GetNoteTemplate(db.Conn(), noteID)
		if err != nil {
			return fmt.Errorf("note is not templated: %w", err)
		}

		// Get the step to verify it exists and get its title
		step, err := database.GetNoteStep(db.Conn(), noteTemplate.ID, stepNumber)
		if err != nil {
			return fmt.Errorf("failed to get step: %w", err)
		}

		// Check if already incomplete
		if !step.Completed {
			fmt.Printf("Step %d is already marked as incomplete\n", stepNumber)
			return nil
		}

		// Mark as incomplete
		if err := database.UncompleteStep(db.Conn(), noteTemplate.ID, stepNumber); err != nil {
			return fmt.Errorf("failed to uncomplete step: %w", err)
		}

		fmt.Printf("○ Step %d marked as incomplete: %s\n", stepNumber, step.Title)
		fmt.Printf("  Note: %s\n", note.Title)

		return nil
	},
}

// stepNoteCmd adds notes to a step
var stepNoteCmd = &cobra.Command{
	Use:   "note <note-id> <step-number> <text>",
	Short: "Add notes to a step",
	Long: `Add or update notes for a specific step.

Notes can contain observations, progress updates, blockers, or any other
relevant information about the step.

Examples:
  yoo step note 42 1 "Completed initial research phase"
  yoo step note 42 3 "Blocked: waiting for approval"`,
	Args: cobra.MinimumNArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		noteID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid note ID: %w", err)
		}

		stepNumber, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid step number: %w", err)
		}

		// Join remaining args as the note text
		noteText := strings.Join(args[2:], " ")

		// Initialize database
		dbPath := config.GetDatabasePath()
		db, err := database.New(dbPath)
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		defer db.Close()

		// Get the note
		note, err := database.GetNote(db.Conn(), noteID)
		if err != nil {
			return fmt.Errorf("failed to get note: %w", err)
		}

		// Get the template link
		noteTemplate, err := database.GetNoteTemplate(db.Conn(), noteID)
		if err != nil {
			return fmt.Errorf("note is not templated: %w", err)
		}

		// Get the step to verify it exists and get its title
		step, err := database.GetNoteStep(db.Conn(), noteTemplate.ID, stepNumber)
		if err != nil {
			return fmt.Errorf("failed to get step: %w", err)
		}

		// Update notes
		if err := database.UpdateStepNotes(db.Conn(), noteTemplate.ID, stepNumber, noteText); err != nil {
			return fmt.Errorf("failed to update step notes: %w", err)
		}

		fmt.Printf("✓ Notes added to step %d: %s\n", stepNumber, step.Title)
		fmt.Printf("  Note: %s\n", note.Title)
		fmt.Printf("  Text: %s\n", noteText)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(stepsCmd)

	// Add subcommands
	stepsCmd.AddCommand(stepListCmd)
	stepsCmd.AddCommand(stepShowCmd)
	stepsCmd.AddCommand(stepCompleteCmd)
	stepsCmd.AddCommand(stepUncompleteCmd)
	stepsCmd.AddCommand(stepNoteCmd)
}
