package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	DatabasePath string
	DateFormat   string
	TimeFormat   string
}

var cfg *Config

// Init initializes the configuration using Viper
func Init() error {
	// Set default values
	viper.SetDefault("database.path", getDefaultDatabasePath())
	viper.SetDefault("format.date", "2006-01-02")
	viper.SetDefault("format.time", "15:04")

	// Set config name and type
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Add config paths
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.config/yoo")
	viper.AddConfigPath("/etc/yoo")

	// Read environment variables
	viper.SetEnvPrefix("YOO")
	viper.AutomaticEnv()

	// Read config file (if it exists)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found; using defaults
	}

	// Populate config struct
	cfg = &Config{
		DatabasePath: viper.GetString("database.path"),
		DateFormat:   viper.GetString("format.date"),
		TimeFormat:   viper.GetString("format.time"),
	}

	// Ensure database directory exists
	dbDir := filepath.Dir(cfg.DatabasePath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	return nil
}

// Get returns the current configuration
func Get() *Config {
	if cfg == nil {
		// Initialize with defaults if not already initialized
		if err := Init(); err != nil {
			panic(fmt.Sprintf("failed to initialize config: %v", err))
		}
	}
	return cfg
}

// getDefaultDatabasePath returns the default path for the SQLite database
func getDefaultDatabasePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "./yoo.db"
	}
	return filepath.Join(homeDir, ".local", "share", "yoo", "yoo.db")
}

// GetDatabasePath returns the configured database path
func GetDatabasePath() string {
	return Get().DatabasePath
}

// GetDateFormat returns the configured date format
func GetDateFormat() string {
	return Get().DateFormat
}

// GetTimeFormat returns the configured time format
func GetTimeFormat() string {
	return Get().TimeFormat
}
