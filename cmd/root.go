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

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (overrides default search paths)")
	rootCmd.PersistentFlags().String("db", "", "database file path (default is $HOME/.local/share/yoo/yoo.db)")

	if err := viper.BindPFlag("database.path", rootCmd.PersistentFlags().Lookup("db")); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initConfig() {
	if err := config.Setup(config.Options{ConfigFile: cfgFile}); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
