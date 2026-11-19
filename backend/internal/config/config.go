package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
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
	RetentionDays  int
}

func Load() *Config {
	aiTimeout, err := time.ParseDuration(getEnv("AI_SERVICE_TIMEOUT", "5s"))
	if err != nil {
		aiTimeout = 5 * time.Second
	}
	retentionDays, err := strconv.Atoi(getEnv("TELEMETRY_RETENTION_DAYS", "30"))
	if err != nil || retentionDays < 1 || retentionDays > 3650 {
		retentionDays = 30
	}

	return &Config{
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         getEnv("DB_PORT", "5432"),
		DBUser:         getEnv("DB_USER", "admin"),
		DBPassword:     os.Getenv("DB_PASSWORD"),
		DBName:         getEnv("DB_NAME", "industrial_monitoring"),
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		AIServiceURL:   getEnv("AI_SERVICE_URL", "http://localhost:8000"),
		AITimeout:      aiTimeout,
		SimulatorToken: os.Getenv("SIMULATOR_TOKEN"),
		RetentionDays:  retentionDays,
	}
}

func (c *Config) Validate() error {
	if len(c.DBPassword) < 16 {
		return errors.New("DB_PASSWORD must contain at least 16 characters")
	}
	if len(c.JWTSecret) < 32 {
		return errors.New("JWT_SECRET must contain at least 32 characters")
	}
	if len(c.SimulatorToken) < 32 {
		return errors.New("SIMULATOR_TOKEN must contain at least 32 characters")
	}
	return nil
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
