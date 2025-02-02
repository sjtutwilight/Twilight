package config

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type Config struct {
	Port     string
	DBConfig DBConfig
}

type DBConfig struct {
	Host     string
	Port     string
	Database string
	User     string
	Password string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: .env file not found\n")
	}

	return &Config{
		Port: getEnvOrDefault("PORT", "8091"),
		DBConfig: DBConfig{
			Host:     getEnvOrDefault("DB_HOST", "localhost"),
			Port:     getEnvOrDefault("DB_PORT", "5432"),
			Database: getEnvOrDefault("DB_NAME", "twilight"),
			User:     getEnvOrDefault("DB_USER", "twilight"),
			Password: getEnvOrDefault("DB_PASSWORD", "twilight123"),
		},
	}, nil
}

// InitDB initializes database connection
func InitDB(cfg *Config) (*pgxpool.Pool, error) {
	connString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s",
		cfg.DBConfig.Host,
		cfg.DBConfig.Port,
		cfg.DBConfig.User,
		cfg.DBConfig.Password,
		cfg.DBConfig.Database,
	)

	// Create connection pool
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %v", err)
	}

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("unable to ping database: %v", err)
	}

	return pool, nil
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
