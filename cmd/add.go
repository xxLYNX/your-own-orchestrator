package cmd

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"yoo/internal/config"
	"yoo/internal/database"
	"yoo/internal/models"

	"github.com/spf13/cobra"
)

var (
	addDate     string
	addPriority string
	addTags     []string
	addTemplate string
	addInputs   []string
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add [note text]",
	Short: "Add a new note to your schedule",
	Long: `Add a new note, task, reminder, or action to your schedule.

By default, the note is added to today's schedule. You can specify a different
date using the --date flag.

Templated Notes:
  Use --template to create a note from a template with structured workflow.
  Provide inputs using --input flags.

Examples:
  yoo add "Complete project documentation"
  yoo add "Team meeting at 2pm" --date 2024-01-15
  yoo add "Review PR" --priority high --tags work,urgent
  yoo add "Apply to 10 tech jobs" --template "Job Applications" --input target_count=10 --input resume=~/resume.pdf`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Join all arguments to form the note text
		noteText := strings.Join(args, " ")

		// Parse the date if provided, otherwise use today
		var scheduleDate time.Time
		var err error

		if addDate != "" {
			scheduleDate, err = time.Parse("2006-01-02", addDate)
			if err != nil {
				log.Fatalf("Invalid date format. Use YYYY-MM-DD: %v", err)
			}
		} else {
			scheduleDate = time.Now()
		}

		// Initialize database
		db, err := database.New(config.GetDatabasePath())
		if err != nil {
			log.Fatalf("Failed to initialize database: %v", err)
		}
		defer db.Close()

		// Parse priority
		priority := 0
		switch strings.ToLower(addPriority) {
		case "low":
			priority = 1
		case "medium":
			priority = 2
		case "high":
			priority = 3
		}

		// Check if this is a templated note
		if addTemplate != "" {
			createTemplatedNote(db, noteText, scheduleDate, priority, addTemplate, addInputs)
			return
		}

		// Create a simple note
		note := &database.Note{
			Title:       noteText,
			Description: "",
			ScheduledAt: scheduleDate,
			Status:      "pending",
			Priority:    priority,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Save to database
		if err := database.CreateNote(db.Conn(), note); err != nil {
			log.Fatalf("Failed to add note: %v", err)
		}

		// Success message
		dateStr := scheduleDate.Format("2006-01-02")
		if isToday(scheduleDate) {
			fmt.Printf("✓ Added note for today: %s\n", noteText)
		} else {
			fmt.Printf("✓ Added note for %s: %s\n", dateStr, noteText)
		}

		if addPriority != "" {
			fmt.Printf("  Priority: %s\n", addPriority)
		}
		if len(addTags) > 0 {
			fmt.Printf("  Tags: %s\n", strings.Join(addTags, ", "))
		}
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	// Flags for the add command
	addCmd.Flags().StringVarP(&addDate, "date", "d", "", "Date for the note (YYYY-MM-DD format, defaults to today)")
	addCmd.Flags().StringVarP(&addPriority, "priority", "p", "", "Priority level (low, medium, high)")
	addCmd.Flags().StringSliceVarP(&addTags, "tags", "t", []string{}, "Tags for categorization (comma-separated)")
	addCmd.Flags().StringVar(&addTemplate, "template", "", "Template to use for this note")
	addCmd.Flags().StringArrayVar(&addInputs, "input", []string{}, "Template input (name=value)")
}

// isToday checks if the given date is today
func isToday(date time.Time) bool {
	now := time.Now()
	return date.Year() == now.Year() && date.YearDay() == now.YearDay()
}

// createTemplatedNote creates a note from a template
func createTemplatedNote(db *database.DB, title string, scheduleDate time.Time, priority int, templateName string, inputs []string) {
	// Get the template
	template, err := database.GetTemplateByName(db.Conn(), templateName)
	if err != nil {
		log.Fatalf("Failed to find template '%s': %v", templateName, err)
	}

	// Parse inputs
	inputMap := make(map[string]interface{})
	for _, input := range inputs {
		parts := strings.SplitN(input, "=", 2)
		if len(parts) != 2 {
			log.Fatalf("Invalid input format '%s', expected name=value", input)
		}
		name := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		inputMap[name] = value
	}

	// Convert input types based on template definition
	for name, value := range inputMap {
		// Find the input definition
		var inputDef *models.TemplateInput
		for i := range template.Definition.Inputs {
			if template.Definition.Inputs[i].Name == name {
				inputDef = &template.Definition.Inputs[i]
				break
			}
		}

		if inputDef != nil {
			strVal, ok := value.(string)
			if ok {
				switch inputDef.Type {
				case "integer":
					if intVal, err := strconv.Atoi(strVal); err == nil {
						inputMap[name] = intVal
					}
				case "boolean":
					if boolVal, err := strconv.ParseBool(strVal); err == nil {
						inputMap[name] = boolVal
					}
				}
			}
		}
	}

	// Validate inputs
	if err := template.Definition.ValidateInputs(inputMap); err != nil {
		log.Fatalf("Input validation failed: %v", err)
	}

	// Create the note
	note := &database.Note{
		Title:       title,
		Description: fmt.Sprintf("Created from template: %s", template.Name),
		ScheduledAt: scheduleDate,
		Status:      "pending",
		Priority:    priority,
		IsTemplated: true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := database.CreateNote(db.Conn(), note); err != nil {
		log.Fatalf("Failed to create note: %v", err)
	}

	// Create template instance
	instance := models.NewTemplateInstance(&template.Definition, inputMap)

	// Create note_template link and initialize steps
	_, err = database.AttachTemplateToNote(db.Conn(), note.ID, template.ID, instance)
	if err != nil {
		log.Fatalf("Failed to link template: %v", err)
	}

	// Success message
	fmt.Printf("✓ Created templated note: %s\n", title)
	fmt.Printf("  Template: %s (v%s)\n", template.Name, template.Version)
	fmt.Printf("  Note ID: %d\n", note.ID)
	fmt.Printf("  Shapes: %s\n", strings.Join(template.Definition.ActiveShapes(), ", "))
	if template.Definition.HasProcedureShape() {
		fmt.Printf("  Manage with: yoo step list %d\n", note.ID)
	}
	if template.Definition.RecordSchema != nil {
		fmt.Printf("  Record schema: %d fields\n", len(template.Definition.RecordSchema.Fields))
		fmt.Printf("  Add records with: yoo records add %d\n", note.ID)
	}
	if len(template.Definition.Outputs) > 0 {
		fmt.Printf("  Outputs required: %d\n", len(template.Definition.GetRequiredOutputs()))
	}
}
