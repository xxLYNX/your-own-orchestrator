package tui

import (
	"database/sql"
	"fmt"
	"time"

	"yoo/internal/database"
	"yoo/internal/models"
)

// ExampleStepsView demonstrates how to use the StepsViewModel
func ExampleStepsView() {
	// Open database connection
	db, err := sql.Open("sqlite", "yoo.db")
	if err != nil {
		fmt.Printf("Failed to open database: %v\n", err)
		return
	}
	defer db.Close()

	// Example: Load a note and its associated template
	noteID := int64(1)
	note, err := database.GetNoteByID(db, noteID)
	if err != nil {
		fmt.Printf("Failed to load note: %v\n", err)
		return
	}

	fmt.Printf("Loading steps for note: %s\n", note.Title)

	// Get the note template (assumes note is templated)
	noteTemplate, err := database.GetNoteTemplate(db, noteID)
	if err != nil {
		fmt.Printf("Failed to load note template: %v\n", err)
		return
	}

	// Get the template definition
	template, err := database.GetTemplateByID(db, noteTemplate.TemplateID)
	if err != nil {
		fmt.Printf("Failed to load template: %v\n", err)
		return
	}

	// Launch the interactive steps view
	if err := ShowSteps(db, noteID, noteTemplate.ID, template); err != nil {
		fmt.Printf("Error in steps view: %v\n", err)
		return
	}

	fmt.Println("Steps view closed successfully")
}

// ExampleCreateAndShowSteps demonstrates creating a templated note and showing its steps
func ExampleCreateAndShowSteps() {
	// Open database connection
	db, err := sql.Open("sqlite", "yoo.db")
	if err != nil {
		fmt.Printf("Failed to open database: %v\n", err)
		return
	}
	defer db.Close()

	// Get a template (e.g., job application template)
	templates, err := database.ListTemplates(db, "")
	if err != nil || len(templates) == 0 {
		fmt.Printf("No templates available: %v\n", err)
		return
	}
	template := templates[0]

	// Create a new note for this template
	note := &database.Note{
		Title:       fmt.Sprintf("Job Application - %s", template.Name),
		Description: "Tracking job application workflow",
		ScheduledAt: time.Now(),
		Status:      "pending",
		Priority:    1,
		IsTemplated: true,
	}

	if err := database.CreateNote(db, note); err != nil {
		fmt.Printf("Failed to create note: %v\n", err)
		return
	}

	// Create template inputs (example values)
	inputs := map[string]interface{}{
		"company_name": "Acme Corp",
		"position":     "Software Engineer",
		"job_url":      "https://example.com/jobs/123",
	}

	// Validate inputs
	if err := template.Definition.ValidateInputs(inputs); err != nil {
		fmt.Printf("Invalid inputs: %v\n", err)
		return
	}

	// Create template instance
	instance := models.NewTemplateInstance(&template.Definition, inputs)

	// Create note template association
	noteTemplate, err := database.AttachTemplateToNote(db, note.ID, template.ID, instance)
	if err != nil {
		fmt.Printf("Failed to create note template: %v\n", err)
		return
	}

	// Launch the interactive steps view
	if err := ShowSteps(db, note.ID, noteTemplate.ID, template); err != nil {
		fmt.Printf("Error in steps view: %v\n", err)
		return
	}

	// After the view closes, check progress
	steps, err := database.ListNoteSteps(db, noteTemplate.ID)
	if err != nil {
		fmt.Printf("Failed to load steps: %v\n", err)
		return
	}

	// Calculate completion
	completed := 0
	for _, step := range steps {
		if step.Completed {
			completed++
		}
	}

	fmt.Printf("Progress: %d/%d steps completed (%.1f%%)\n",
		completed, len(steps), float64(completed)/float64(len(steps))*100)
}

// ExampleProgrammaticStepUpdate demonstrates updating steps programmatically
func ExampleProgrammaticStepUpdate() {
	// Open database connection
	db, err := sql.Open("sqlite", "yoo.db")
	if err != nil {
		fmt.Printf("Failed to open database: %v\n", err)
		return
	}
	defer db.Close()

	noteTemplateID := int64(1) // Example ID

	// Get all steps
	steps, err := database.ListNoteSteps(db, noteTemplateID)
	if err != nil {
		fmt.Printf("Failed to load steps: %v\n", err)
		return
	}

	// Complete the first step
	if len(steps) > 0 {
		firstStep := steps[0]
		if err := database.CompleteStep(db, noteTemplateID, firstStep.StepNumber); err != nil {
			fmt.Printf("Failed to complete step: %v\n", err)
			return
		}
		fmt.Printf("✓ Completed step %d: %s\n", firstStep.StepNumber, firstStep.Title)
	}

	// Add notes to the second step
	if len(steps) > 1 {
		secondStep := steps[1]
		notes := "Remember to highlight experience with Go and distributed systems"
		if err := database.UpdateStepNotes(db, noteTemplateID, secondStep.StepNumber, notes); err != nil {
			fmt.Printf("Failed to add notes: %v\n", err)
			return
		}
		fmt.Printf("📝 Added notes to step %d: %s\n", secondStep.StepNumber, secondStep.Title)
	}

	// Uncomplete a step (if needed)
	if len(steps) > 0 {
		firstStep := steps[0]
		if err := database.UncompleteStep(db, noteTemplateID, firstStep.StepNumber); err != nil {
			fmt.Printf("Failed to uncomplete step: %v\n", err)
			return
		}
		fmt.Printf("○ Uncompleted step %d: %s\n", firstStep.StepNumber, firstStep.Title)
	}
}

// ExampleIntegrateWithSchedule demonstrates integrating steps view with schedule
func ExampleIntegrateWithSchedule() {
	// Open database connection
	db, err := sql.Open("sqlite", "yoo.db")
	if err != nil {
		fmt.Printf("Failed to open database: %v\n", err)
		return
	}
	defer db.Close()

	// Get today's notes
	today := time.Now()
	notes, err := database.GetNotesByDate(db, today)
	if err != nil {
		fmt.Printf("Failed to load notes: %v\n", err)
		return
	}

	// Display schedule first
	fmt.Printf("📅 Schedule for %s\n\n", today.Format("Monday, January 2, 2006"))

	for i, note := range notes {
		status := "○"
		if note.Status == "completed" {
			status = "✓"
		}

		fmt.Printf("%d. %s %s", i+1, status, note.Title)

		if note.IsTemplated {
			// Show progress for templated notes
			fmt.Printf(" [%.0f%% complete]", note.TemplateProgress*100)
		}

		fmt.Println()
	}

	// User selects a templated note
	fmt.Println("\nEnter note number to view steps (or 0 to quit):")
	var selection int
	fmt.Scanf("%d", &selection)

	if selection > 0 && selection <= len(notes) {
		selectedNote := notes[selection-1]

		if selectedNote.IsTemplated {
			// Load template data
			noteTemplate, err := database.GetNoteTemplate(db, selectedNote.ID)
			if err != nil {
				fmt.Printf("Failed to load note template: %v\n", err)
				return
			}

			template, err := database.GetTemplateByID(db, noteTemplate.TemplateID)
			if err != nil {
				fmt.Printf("Failed to load template: %v\n", err)
				return
			}

			// Launch steps view
			if err := ShowSteps(db, selectedNote.ID, noteTemplate.ID, template); err != nil {
				fmt.Printf("Error in steps view: %v\n", err)
				return
			}

			// Recalculate progress after changes
			steps, _ := database.ListNoteSteps(db, noteTemplate.ID)
			completed := 0
			for _, step := range steps {
				if step.Completed {
					completed++
				}
			}
			progress := float64(completed) / float64(len(steps))

			// Update note progress
			selectedNote.TemplateProgress = progress
			if err := database.UpdateNote(db, selectedNote); err != nil {
				fmt.Printf("Failed to update note progress: %v\n", err)
			}
		} else {
			fmt.Println("This note is not templated and doesn't have steps.")
		}
	}
}
