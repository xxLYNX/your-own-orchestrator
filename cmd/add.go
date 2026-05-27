package cmd

import (
	"fmt"
	"log"
	"strings"
	"time"

	"yoo/internal/config"
	"yoo/internal/database"

	"github.com/spf13/cobra"
)

var (
	addDate     string
	addPriority string
	addTags     []string
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add [note text]",
	Short: "Add a new note to your schedule",
	Long: `Add a new note, task, reminder, or action to your schedule.

By default, the note is added to today's schedule. You can specify a different
date using the --date flag.

Examples:
  yoo add "Complete project documentation"
  yoo add "Team meeting at 2pm" --date 2024-01-15
  yoo add "Review PR" --priority high --tags work,urgent`,
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

		// Create the note
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
}

// isToday checks if the given date is today
func isToday(date time.Time) bool {
	now := time.Now()
	return date.Year() == now.Year() && date.YearDay() == now.YearDay()
}
