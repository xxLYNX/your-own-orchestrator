package cmd

import (
	"fmt"

	"yoo/internal/config"
	"yoo/internal/database"
)

// withDB opens the configured database and runs fn, ensuring the connection is closed.
func withDB(fn func(db *database.DB) error) error {
	dbPath := config.GetDatabasePath()
	db, err := database.New(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()
	return fn(db)
}
