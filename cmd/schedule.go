package cmd

import (
	"fmt"
	"time"

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

		return runSchedule(targetDate)
	},
}

func init() {
	rootCmd.AddCommand(scheduleCmd)

	// Add date flag
	scheduleCmd.Flags().StringP("date", "d", "", "Date to query (YYYY-MM-DD format). Defaults to today.")
}
