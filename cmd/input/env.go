package input

import (
	"context"
	"fmt"
	"time"

	"github.com/GeraAnggaraPutra/go-backup/internal/backup"
	"github.com/GeraAnggaraPutra/go-backup/internal/config"
	"github.com/GeraAnggaraPutra/go-backup/internal/constant"
)

func GetConfigFromEnv() (backup.Config, error) {
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
