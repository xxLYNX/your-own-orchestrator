package cmd

import (
	"fmt"
	"time"

	"yoo/internal/database"
	"yoo/internal/tui"
)

func runSchedule(targetDate time.Time) error {
	return withDB(func(db *database.DB) error {
		notes, err := database.GetNotesByDate(db.Conn(), targetDate)
		if err != nil {
			return fmt.Errorf("failed to query notes: %w", err)
		}
		if err := tui.ShowSchedule(db.Conn(), notes, targetDate); err != nil {
			return fmt.Errorf("failed to display schedule: %w", err)
		}
		return nil
	})
}
