package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var pool *pgxpool.Pool

// InitDB initializes the database connection pool
func InitDB() error {
	// Build connection string with additional parameters
	connString := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=disable&connect_timeout=10",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	// Log connection attempt
	log.Printf("Attempting to connect to database at %s:%s", os.Getenv("DB_HOST"), os.Getenv("DB_PORT"))

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return fmt.Errorf("error parsing connection string: %v", err)
	}

	// Set connection pool settings
	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute
	config.ConnConfig.ConnectTimeout = 10 * time.Second

	// Create connection pool with extended retry
	maxRetries := 5
	var retryCount int
	var lastErr error

	for retryCount < maxRetries {
		// Create context with timeout for each attempt
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		pool, err = pgxpool.NewWithConfig(ctx, config)
		cancel()

		if err == nil {
			// Test the connection
			if err := pool.Ping(context.Background()); err == nil {
				log.Printf("Successfully connected to database at %s:%s", os.Getenv("DB_HOST"), os.Getenv("DB_PORT"))
				return nil
			} else {
				lastErr = err
				log.Printf("Connection test failed (attempt %d/%d): %v", retryCount+1, maxRetries, err)
			}
		} else {
			lastErr = err
			log.Printf("Failed to create connection pool (attempt %d/%d): %v", retryCount+1, maxRetries, err)
		}

		retryCount++
		if retryCount < maxRetries {
			// Exponential backoff: 2, 4, 8, 16 seconds
			backoff := time.Duration(1<<uint(retryCount)) * time.Second
			log.Printf("Waiting %v before next attempt...", backoff)
			time.Sleep(backoff)
		}
	}

	return fmt.Errorf("failed to connect after %d attempts. Last error: %v", maxRetries, lastErr)
}

// GetDB returns the database connection pool
func GetDB() *pgxpool.Pool {
	if pool == nil {
		log.Fatal("Database connection pool is not initialized")
	}
	return pool
}

// CloseDB closes the database connection pool
func CloseDB() {
	if pool != nil {
		pool.Close()
		log.Println("Database connection pool closed")
	}
}
