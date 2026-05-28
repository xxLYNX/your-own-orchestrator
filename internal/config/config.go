package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application.
type Config struct {
	DatabasePath string
	DateFormat   string
	TimeFormat   string
}

// Options configures the one-time Viper setup performed at startup.
type Options struct {
	ConfigFile string
}

var cfg *Config

// Setup initializes Viper defaults, loads the first available config file, and
// populates the global Config. It should be called once from cmd/root initConfig.
func Setup(opts Options) error {
	viper.SetDefault("database.path", defaultDatabasePath())
	viper.SetDefault("format.date", "2006-01-02")
	viper.SetDefault("format.time", "15:04")

	viper.SetEnvPrefix("YOO")
	viper.AutomaticEnv()

	if err := loadConfigFile(opts.ConfigFile); err != nil {
		return err
	}

	cfg = &Config{
		DatabasePath: viper.GetString("database.path"),
		DateFormat:   viper.GetString("format.date"),
		TimeFormat:   viper.GetString("format.time"),
	}

	dbDir := filepath.Dir(cfg.DatabasePath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	return nil
}

func loadConfigFile(explicit string) error {
	if explicit != "" {
		viper.SetConfigFile(explicit)
		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return fmt.Errorf("error reading config file: %w", err)
			}
		}
		return nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	candidates := []string{
		filepath.Join(home, ".yoo.yaml"),
		filepath.Join(home, ".config", "yoo", "config.yaml"),
		"config.yaml",
		filepath.Join("/etc/yoo", "config.yaml"),
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err != nil {
			continue
		}
		viper.SetConfigFile(path)
		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return fmt.Errorf("error reading config file %s: %w", path, err)
			}
		}
		return nil
	}

	return nil
}

// Get returns the current configuration.
func Get() *Config {
	if cfg == nil {
		if err := Setup(Options{}); err != nil {
			panic(fmt.Sprintf("failed to initialize config: %v", err))
		}
	}
	return cfg
}

func defaultDatabasePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "./yoo.db"
	}
	return filepath.Join(homeDir, ".local", "share", "yoo", "yoo.db")
}

// GetDatabasePath returns the configured database path.
func GetDatabasePath() string {
	return Get().DatabasePath
}

// GetDateFormat returns the configured date format.
func GetDateFormat() string {
	return Get().DateFormat
}

// GetTimeFormat returns the configured time format.
func GetTimeFormat() string {
	return Get().TimeFormat
}
