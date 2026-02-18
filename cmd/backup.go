package cmd

import (
	"context"
	"database/sql"
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

func exitWithError(msg string, err error) {
	if err != nil {
		fmt.Printf("\n%s[Error] %s: %v%s\n", constant.ColorRed, msg, err, constant.ColorReset)
	} else {
		fmt.Printf("\n%s[Error] %s%s\n", constant.ColorRed, msg, constant.ColorReset)
	}

	os.Exit(1)
}

var rootCmd = &cobra.Command{
	Use: "go-backup",
}

var backupCmd = &cobra.Command{
	Use:           "backup",
	Short:         "Database backup",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("%s[System] Database Backup Tool Starting...%s\n", constant.ColorCyan, constant.ColorReset)
		fmt.Printf("%s[System] Timezone: %s%s\n", constant.ColorCyan, time.Now().Format("2006-01-02 15:04:05 MST"), constant.ColorReset)

		configSource := "Manual Input"
		if _, err := os.Stat(".env"); err == nil {
			prompt := &survey.Select{
				Message: "Select configuration source:",
				Options: []string{".env file", "Manual Input"},
				Default: ".env file",
			}

			survey.AskOne(prompt, &configSource)
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

		return backup.Run(cfg)
	},
}

func Execute() error {
	rootCmd.AddCommand(backupCmd)
	return rootCmd.Execute()
}

// Helper Functions
func getConfigFromEnv() (backup.Config, error) {
	fmt.Printf("%s[System] Validating connection and tables from .env...%s\n", constant.ColorCyan, constant.ColorReset)

	dbConfig := config.LoadConfig()
	driver := getDatabaseDriver(dbConfig.DBType)

	conn, err := connectToDatabase(driver, dbConfig.DBHost, dbConfig.DBPort, dbConfig.DBUser, dbConfig.DBPassword, dbConfig.DBName)
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

	finalTables, err := filterTables(existingTables, dbConfig.DBTables)
	if err != nil {
		exitWithError("Table validation failed", err)
	}

	output := getOutputFile(dbConfig.OutputFile)

	fmt.Printf("%s[System] Validation successful. Starting backup...%s\n", constant.ColorGreen, constant.ColorReset)

	return backup.Config{
		DBType:   dbConfig.DBType,
		Host:     dbConfig.DBHost,
		Port:     dbConfig.DBPort,
		Username: dbConfig.DBUser,
		Password: dbConfig.DBPassword,
		DBName:   dbConfig.DBName,
		Output:   output,
		Tables:   finalTables,
		Schedule: dbConfig.Schedule,
	}, nil
}

func getConfigFromInput() (backup.Config, error) {
	var (
		dbType   string
		host     string
		port     int
		username string
		password string
		dbName   string
	)

	survey.AskOne(&survey.Select{
		Message: "Select database type:",
		Options: []string{"mysql", "postgres"},
		Default: "mysql",
	}, &dbType)

	survey.AskOne(&survey.Input{
		Message: "Host:",
		Default: "127.0.0.1",
	}, &host)

	defPort := "3306"
	if dbType == "postgres" {
		defPort = "5432"
	}
	var portStr string
	survey.AskOne(&survey.Input{
		Message: "Port:",
		Default: defPort,
	}, &portStr)
	port, _ = strconv.Atoi(portStr)

	survey.AskOne(&survey.Input{Message: "Username:"}, &username)
	survey.AskOne(&survey.Password{Message: "Password:"}, &password)
	survey.AskOne(&survey.Input{Message: "Database name:"}, &dbName)

	driver := getDatabaseDriver(dbType)
	conn, err := connectToDatabase(driver, host, port, username, password, dbName)
	if err != nil {
		exitWithError("Failed to connect", err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	allTables, err := driver.ListTables(ctx, conn)
	if err != nil {
		exitWithError("Failed to get tables", err)
	}

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
		Default: "Now (Once)",
	}, &runMode)

	cronSchedule := ""
	if runMode == "Scheduled (Cron)" {
		survey.AskOne(&survey.Input{
			Message: "Cron Expression:",
			Default: "* * * * *",
		}, &cronSchedule)
	}

	var outputInput string
	defaultOutput := time.Now().Format("20060102_150405") + "_backup.zip"
	survey.AskOne(&survey.Input{
		Message: "Output file:",
		Default: defaultOutput,
	}, &outputInput)

	output := getOutputFile(outputInput)

	return backup.Config{
		DBType:   dbType,
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		DBName:   dbName,
		Output:   output,
		Tables:   selectedTables,
		Schedule: cronSchedule,
	}, nil
}

func getDatabaseDriver(dbType string) database.Database {
	if strings.ToLower(dbType) == "mysql" {
		return &database.MySQL{}
	}

	return &database.Postgres{}
}

func connectToDatabase(driver database.Database, host string, port int, user, password, dbName string) (*sql.DB, error) {
	return driver.Connect(host, port, user, password, dbName)
}

func filterTables(existingTables []string, envTablesStr string) ([]string, error) {
	if envTablesStr == "" {
		if len(existingTables) == 0 {
			return nil, errors.New("no tables found in database")
		}

		return existingTables, nil
	}

	envTables := strings.Split(envTablesStr, ",")
	tableMap := make(map[string]bool)
	for _, t := range existingTables {
		tableMap[t] = true
	}

	var finalTables []string
	for _, et := range envTables {
		trimmed := strings.TrimSpace(et)
		if trimmed == "" {
			continue
		}

		if !tableMap[trimmed] {
			return nil, fmt.Errorf("table '%s' from .env not found in database", trimmed)
		}

		finalTables = append(finalTables, trimmed)
	}

	if len(finalTables) == 0 {
		return nil, errors.New("no tables to backup")
	}

	return finalTables, nil
}

func getOutputFile(input string) string {
	output := strings.TrimSpace(input)
	defaultOutput := time.Now().Format("20060102_150405") + "_backup.zip"

	if output == "" {
		return defaultOutput
	}

	if !strings.HasSuffix(output, ".zip") {
		return output + ".zip"
	}

	if !strings.HasPrefix(output, "backups/") {
		output = "backups/" + output
	}

	return output
}
