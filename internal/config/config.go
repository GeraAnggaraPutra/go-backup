package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBType     string
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	DBTables   string
	OutputFile string
	Schedule string
}

func LoadConfig() (cfg Config) {
	_ = godotenv.Load("./.env")

	cfg.DBType = getEnv("DB_TYPE")
	cfg.DBHost = getEnv("DB_HOST")
	cfg.DBPort = getEnvAsInt("DB_PORT")
	cfg.DBUser = getEnv("DB_USER")
	cfg.DBPassword = getEnv("DB_PASSWORD")
	cfg.DBName = getEnv("DB_NAME")
	cfg.DBTables = getEnv("DB_TABLES")
	cfg.OutputFile = getEnv("OUTPUT_FILE")
	cfg.Schedule = getEnv("SCHEDULE")

	return
}

func getEnv(key string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return ""
}

func getEnvAsInt(key string) int {
	if valueStr, exists := os.LookupEnv(key); exists {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}

	return 0
}
