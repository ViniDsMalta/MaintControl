package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	ServerPort     string
	JWTSecret      string
	AIServiceURL   string
	AITimeout      time.Duration
	SimulatorToken string
}

func Load() *Config {
	aiTimeout, err := time.ParseDuration(getEnv("AI_SERVICE_TIMEOUT", "5s"))
	if err != nil {
		aiTimeout = 5 * time.Second
	}

	return &Config{
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         getEnv("DB_PORT", "5432"),
		DBUser:         getEnv("DB_USER", "admin"),
		DBPassword:     getEnv("DB_PASSWORD", "admin"),
		DBName:         getEnv("DB_NAME", "industrial_monitoring"),
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		JWTSecret:      getEnv("JWT_SECRET", "change-this-secret-in-production"),
		AIServiceURL:   getEnv("AI_SERVICE_URL", "http://localhost:8000"),
		AITimeout:      aiTimeout,
		SimulatorToken: getEnv("SIMULATOR_TOKEN", "change-this-simulator-token"),
	}
}

func (c *Config) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost,
		c.DBPort,
		c.DBUser,
		c.DBPassword,
		c.DBName,
	)
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
