package config

import (
	"os"
	"strconv"
)

type Config struct {
	APP_NAME       string
	DATABASE_URL   string
	CLICKHOUSE_URL string
	BIND_ADDRESS   string
	PORT           int
	DEBUG          bool
	JWT_SECRET     string
	DEV_MODE       bool
	SMTP_HOST      string
	SMTP_PORT      int
	SMTP_FROM      string
	SMTP_USERNAME  string
	SMTP_PASSWORD  string
}

var config *Config

func Load() *Config {
	if config == nil {
		config = &Config{
			APP_NAME:       getEnvAsString("APP_NAME", "clutter"),
			DATABASE_URL:   getEnvAsString("DATABASE_URL", "postgres://admin:admin@localhost:5432/mydb?sslmode=disable"),
			CLICKHOUSE_URL: getEnvAsString("CLICKHOUSE_URL", "clickhouse://default:@localhost:9000?database=clutter"),
			BIND_ADDRESS:   getEnvAsString("BIND_ADDRESS", "127.0.0.1"),
			PORT:           getEnvAsInt("PORT", 8081),
			DEBUG:          getEnvAsBool("DEBUG", false),
			JWT_SECRET:     getEnvAsString("JWT_SECRET", "supersecretkey"),
			DEV_MODE:       getEnvAsBool("DEV_MODE", true),
			SMTP_HOST:      getEnvAsString("SMTP_HOST", ""),
			SMTP_PORT:      getEnvAsInt("SMTP_PORT", 587),
			SMTP_FROM:      getEnvAsString("SMTP_FROM", ""),
			SMTP_USERNAME:  getEnvAsString("SMTP_USERNAME", ""),
			SMTP_PASSWORD:  getEnvAsString("SMTP_PASSWORD", ""),
		}
	}
	return config
}

func Get() *Config {
	if config == nil {
		panic("configuration not loaded")
	}
	return config
}

func getEnvAsString(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
