package cmd

import (
	"fmt"
	"os"
	"time"

	"yoo/internal/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "yoo",
	Short: "Your Own Orchestrator - A TUI for schedule and task management",
	Long: `yoo is a terminal-based task and schedule management tool.
It helps you keep track of your daily to-dos, tasks, reminders, and actions
with a clean TUI interface backed by SQLite for local storage.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSchedule(time.Now())
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.yoo.yaml)")
	rootCmd.PersistentFlags().String("db", "", "database file path (default is $HOME/.yoo/yoo.db)")

	// Bind flags to viper
	viper.BindPFlag("database.path", rootCmd.PersistentFlags().Lookup("db"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".yoo" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".yoo")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		// Config file loaded successfully
	}

	// Initialize configuration
	config.Init()
}
