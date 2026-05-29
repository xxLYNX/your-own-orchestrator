package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"yoo/internal/database"
	"yoo/internal/models"

	"github.com/spf13/cobra"
)

var examplesCmd = &cobra.Command{
	Use:   "examples",
	Short: "Example notes for trying the app",
}

var examplesSeedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Add sample notes to today's schedule",
	Long: `Create a handful of example notes that show how yoo works without heavy setup.

Includes simple checklists (grocery, weekly review, home task) plus one plain reminder.
Re-running is safe: existing notes with the same titles are skipped.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return withDB(func(db *database.DB) error {
			if err := loadAllBuiltinTemplates(db); err != nil {
				return err
			}
			created, skipped, err := seedExampleNotes(db)
			if err != nil {
				return err
			}
			fmt.Printf("\n✓ Example notes ready (%d created, %d already existed)\n", created, skipped)
			fmt.Println("\nTry:")
			fmt.Println("  yoo schedule              # see today's notes")
			fmt.Println("  yoo step list <id>        # stages for a templated note")
			fmt.Println("  yoo                         # TUI — open a note and tab to its checklist")
			return nil
		})
	},
}

func init() {
	rootCmd.AddCommand(examplesCmd)
	examplesCmd.AddCommand(examplesSeedCmd)
}

func loadAllBuiltinTemplates(db *database.DB) error {
	templatesDir := "templates"
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		return fmt.Errorf("templates directory not found (run from repo root)")
	}
	patterns := []string{
		filepath.Join(templatesDir, "*.yaml"),
		filepath.Join(templatesDir, "generics", "*.yaml"),
	}
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return err
		}
		for _, file := range matches {
			data, err := os.ReadFile(file)
			if err != nil {
				continue
			}
			template, err := models.ParseTemplateYAML(data, true)
			if err != nil {
				continue
			}
			existing, _ := database.GetTemplateByName(db.Conn(), template.Name)
			if existing != nil {
				template.ID = existing.ID
				template.CreatedAt = existing.CreatedAt
				_ = database.UpdateTemplate(db.Conn(), template)
				continue
			}
			if err := database.CreateTemplate(db.Conn(), template); err != nil {
				return err
			}
		}
	}
	return nil
}

type seedNote struct {
	title    string
	template string
	inputs   []string
}

func seedExampleNotes(db *database.DB) (created, skipped int, err error) {
	today := time.Now()
	examples := []seedNote{
		{title: "Grocery run", template: "Grocery Run"},
		{title: "Friday weekly review", template: "Weekly Review"},
		{title: "Fix kitchen faucet", template: "Home Task"},
	}

	for _, ex := range examples {
		exists, err := noteTitleExistsToday(db, ex.title, today)
		if err != nil {
			return created, skipped, err
		}
		if exists {
			skipped++
			continue
		}
		if err := createTemplatedNote(db, ex.title, today, 1, ex.template, ex.inputs); err != nil {
			return created, skipped, err
		}
		created++
	}

	plainTitle := "Call dentist"
	exists, err := noteTitleExistsToday(db, plainTitle, today)
	if err != nil {
		return created, skipped, err
	}
	if !exists {
		note := &database.Note{
			Title:       plainTitle,
			Description: "Example plain reminder — no template",
			ScheduledAt: today,
			Status:      "pending",
			Priority:    2,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if err := database.CreateNote(db.Conn(), note); err != nil {
			return created, skipped, err
		}
		created++
	} else {
		skipped++
	}

	return created, skipped, nil
}

func noteTitleExistsToday(db *database.DB, title string, day time.Time) (bool, error) {
	notes, err := database.GetNotesByDate(db.Conn(), day)
	if err != nil {
		return false, err
	}
	for _, note := range notes {
		if note.Title == title {
			return true, nil
		}
	}
	return false, nil
}
