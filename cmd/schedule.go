package cmd

import (
	"fmt"
	"time"

	"yoo/internal/config"
	"yoo/internal/database"
	"yoo/internal/tui"

	"github.com/spf13/cobra"
)

var scheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "View your schedule",
	Long:  `Display notes and tasks for a specific date. Defaults to today if no date is specified.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get the date flag
		dateStr, _ := cmd.Flags().GetString("date")

		var targetDate time.Time
		var err error

		if dateStr == "" {
			// Default to today
			targetDate = time.Now()
		} else {
			// Parse the provided date
			targetDate, err = time.Parse("2006-01-02", dateStr)
			if err != nil {
				return fmt.Errorf("invalid date format. Use YYYY-MM-DD: %w", err)
			}
		}

		// Initialize database
		dbPath := config.GetDatabasePath()
		db, err := database.New(dbPath)
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		defer db.Close()

		// Query notes for the target date
		notes, err := database.GetNotesByDate(db.Conn(), targetDate)
		if err != nil {
			return fmt.Errorf("failed to query notes: %w", err)
		}

		// Launch TUI
		if err := tui.ShowSchedule(notes, targetDate); err != nil {
			return fmt.Errorf("failed to display schedule: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(scheduleCmd)

	// Add date flag
	scheduleCmd.Flags().StringP("date", "d", "", "Date to query (YYYY-MM-DD format). Defaults to today.")
}
