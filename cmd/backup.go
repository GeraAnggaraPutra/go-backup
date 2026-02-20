package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"github.com/GeraAnggaraPutra/go-backup/cmd/input"
	"github.com/GeraAnggaraPutra/go-backup/internal/backup"
	"github.com/GeraAnggaraPutra/go-backup/internal/constant"
)

var rootCmd = &cobra.Command{
	Use: "go-backup",
}

var backupCmd = &cobra.Command{
	Use:           "backup",
	Short:         "Database backup",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          runBackup,
}

func Execute() error {
	rootCmd.AddCommand(backupCmd)
	return rootCmd.Execute()
}

func runBackup(cmd *cobra.Command, args []string) error {
	printHeader()

	configSource := ".env file"
	if _, err := os.Stat(".env"); err == nil {
		survey.AskOne(&survey.Select{
			Message: "Select configuration source:",
			Options: []string{".env file", "Manual Input"},
			Default: ".env file",
		}, &configSource)
	} else {
		configSource = "Manual Input"
	}

	var cfg backup.Config
	var err error

	if configSource == ".env file" {
		cfg, err = input.GetConfigFromEnv()
	} else {
		cfg, err = input.GetConfigFromInput()
	}

	if err != nil {
		return err
	}

	if err := backup.Run(cfg); err != nil {
		if errors.Is(err, context.Canceled) {
			fmt.Printf("%s[System] Shutdown complete. Goodbye!%s\n", constant.ColorRed, constant.ColorReset)
			return nil
		}

		return err
	}

	return nil
}

func printHeader() {
	fmt.Printf("%s[System] Database Backup Tool Starting...%s\n", constant.ColorCyan, constant.ColorReset)
	fmt.Printf("%s[System] Timezone: %s%s\n", constant.ColorCyan, time.Now().Format("2006-01-02 15:04:05 MST"), constant.ColorReset)
}
