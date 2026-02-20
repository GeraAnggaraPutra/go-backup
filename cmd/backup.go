package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"github.com/GeraAnggaraPutra/go-backup/internal/backup"
	"github.com/GeraAnggaraPutra/go-backup/internal/config"
	"github.com/GeraAnggaraPutra/go-backup/internal/constant"
	"github.com/GeraAnggaraPutra/go-backup/internal/database"
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
		cfg, err = getConfigFromEnv()
	} else {
		cfg, err = getConfigFromInput()
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

func getConfigFromEnv() (backup.Config, error) {
	fmt.Printf("%s[System] Validating connection and tables from .env...%s\n", constant.ColorCyan, constant.ColorReset)

	envConfig := config.LoadConfig()
	driver := getDatabaseDriver(envConfig.DBType)

	conn, err := driver.Connect(envConfig.DBHost, envConfig.DBPort, envConfig.DBUser, envConfig.DBPassword, envConfig.DBName)
	if err != nil {
		exitWithError("Connection failed using .env credentials", err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	existingTables, err := driver.ListTables(ctx, conn)
	if err != nil {
		exitWithError("Failed to fetch table list", err)
	}

	finalTables, err := filterTables(existingTables, envConfig.DBTables)
	if err != nil {
		exitWithError("Table validation failed", err)
	}

	fmt.Printf("%s[System] Validation successful. Starting backup...%s\n", constant.ColorGreen, constant.ColorReset)

	return backup.Config{
		DBType: envConfig.DBType, Host: envConfig.DBHost, Port: envConfig.DBPort,
		Username: envConfig.DBUser, Password: envConfig.DBPassword, DBName: envConfig.DBName,
		Output: getOutputFile(envConfig.OutputFile), Tables: finalTables,
		Schedule: envConfig.Schedule, ENVConfig: envConfig,
	}, nil
}

func getConfigFromInput() (backup.Config, error) {
	var dbType, host, portStr, user, pass, dbName string
	var envCfg config.Config

	survey.AskOne(&survey.Select{
		Message: "Select database type:",
		Options: []string{"mysql", "postgres"},
	}, &dbType)
	survey.AskOne(&survey.Input{
		Message: "Host:",
		Default: "127.0.0.1",
	}, &host)

	defPort := "3306"
	if dbType == "postgres" {
		defPort = "5432"
	}

	survey.AskOne(&survey.Input{
		Message: "Port:",
		Default: defPort,
	}, &portStr)
	port, _ := strconv.Atoi(portStr)

	survey.AskOne(&survey.Input{Message: "Username:"}, &user)
	survey.AskOne(&survey.Password{Message: "Password:"}, &pass)
	survey.AskOne(&survey.Input{Message: "Database name:"}, &dbName)

	driver := getDatabaseDriver(dbType)
	conn, err := driver.Connect(host, port, user, pass, dbName)
	if err != nil {
		exitWithError("Failed to connect", err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	allTables, _ := driver.ListTables(ctx, conn)
	var selectedTables []string
	survey.AskOne(&survey.MultiSelect{
		Message: "Select tables:",
		Options: allTables,
	}, &selectedTables)

	if len(selectedTables) == 0 {
		return backup.Config{}, errors.New("no table selected")
	}

	var runMode string
	survey.AskOne(&survey.Select{
		Message: "Choose run mode:",
		Options: []string{"Now (Once)", "Scheduled (Cron)"},
	}, &runMode)

	cronSchedule := ""
	if runMode == "Scheduled (Cron)" {
		survey.AskOne(&survey.Input{
			Message: "Cron Expression:",
			Default: "* * * * *",
		}, &cronSchedule)
	}

	var outputInput string
	survey.AskOne(&survey.Input{
		Message: "Output file:",
		Default: time.Now().Format("20060102_150405") + "_backup.zip",
	}, &outputInput)

	askTelegramPrompt(&envCfg)
	askGCSPrompt(&envCfg)

	return backup.Config{
		DBType: dbType, Host: host, Port: port, Username: user, Password: pass, DBName: dbName,
		Output: getOutputFile(outputInput), Tables: selectedTables,
		Schedule: cronSchedule, ENVConfig: envCfg,
	}, nil
}

func askTelegramPrompt(cfg *config.Config) {
	send := false
	survey.AskOne(&survey.Confirm{Message: "Send notification to Telegram?", Default: false}, &send)
	if send {
		survey.AskOne(&survey.Input{Message: "Telegram Token:"}, &cfg.TelegramToken)
		survey.AskOne(&survey.Input{Message: "Telegram Chat ID:"}, &cfg.TelegramChatID)
	}
}

func askGCSPrompt(cfg *config.Config) {
	send := false
	survey.AskOne(&survey.Confirm{Message: "Upload backup to Google Cloud Storage (GCS)?", Default: false}, &send)
	if send {
		cfg.GCSEnabled = true
		survey.AskOne(&survey.Input{Message: "GCS Bucket Name:", Default: os.Getenv("GCS_BUCKET_NAME")}, &cfg.GCSBucketName)
		survey.AskOne(&survey.Input{Message: "GCS Service Account File:", Default: "credentials-gcs.json"}, &cfg.GCSServiceAccountFile)
	}
}

func printHeader() {
	fmt.Printf("%s[System] Database Backup Tool Starting...%s\n", constant.ColorCyan, constant.ColorReset)
	fmt.Printf("%s[System] Timezone: %s%s\n", constant.ColorCyan, time.Now().Format("2006-01-02 15:04:05 MST"), constant.ColorReset)
}

func getDatabaseDriver(dbType string) database.Database {
	if strings.ToLower(dbType) == "mysql" {
		return &database.MySQL{}
	}

	return &database.Postgres{}
}

func filterTables(existingTables []string, envTablesStr string) ([]string, error) {
	if envTablesStr == "" {
		if len(existingTables) == 0 {
			return nil, errors.New("no tables found")
		}
		return existingTables, nil
	}

	tableMap := make(map[string]bool)
	for _, t := range existingTables {
		tableMap[t] = true
	}

	var final []string
	for _, et := range strings.Split(envTablesStr, ",") {
		trimmed := strings.TrimSpace(et)
		if trimmed == "" {
			continue
		}
		if !tableMap[trimmed] {
			return nil, fmt.Errorf("table '%s' not found", trimmed)
		}
		final = append(final, trimmed)
	}

	return final, nil
}

func getOutputFile(input string) string {
	output := strings.TrimSpace(input)
	if output == "" {
		output = time.Now().Format("20060102_150405") + "_backup.zip"
	}

	if !strings.HasSuffix(strings.ToLower(output), ".zip") {
		output += ".zip"
	}

	if !strings.HasPrefix(output, "backups/") {
		output = "backups/" + output
	}

	_ = os.MkdirAll("backups", 0755)
	return output
}

func exitWithError(msg string, err error) {
	fmt.Printf("\n%s[Error] %s", constant.ColorRed, msg)
	if err != nil {
		fmt.Printf(": %v", err)
	}

	fmt.Printf("%s\n", constant.ColorReset)
	os.Exit(1)
}
