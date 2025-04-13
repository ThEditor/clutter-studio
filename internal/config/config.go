package config

import (
	"os"
	"strconv"
)

type Config struct {
	DATABASE_URL string
	BIND_ADDRESS string
	PORT         int
	DEBUG        bool
}

var config *Config

func Load() *Config {
	if config == nil {
		config = &Config{
			DATABASE_URL: getEnvAsString("DATABASE_URL", "postgres://admin:admin@localhost:5432/mydb?sslmode=disable"),
			BIND_ADDRESS: getEnvAsString("BIND_ADDRESS", "127.0.0.1"),
			PORT:         getEnvAsInt("PORT", 8081),
			DEBUG:        getEnvAsBool("DEBUG", false),
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
