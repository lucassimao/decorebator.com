package common

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	db        *pgxpool.Pool
	dbOnce    sync.Once
	dbMutex   sync.Mutex
	dbInitErr error
)

// configureConnectionPool optimizes connection pool settings for DigitalOcean App Platform
func configureConnectionPool(config *pgxpool.Config) {
	env := os.Getenv("ENV")

	switch env {
	case ProductionEnv:
		// Optimized for DigitalOcean App Platform + managed PostgreSQL
		config.MaxConns = 10
		config.MinConns = 2
		config.MaxConnLifetime = time.Hour
		config.MaxConnIdleTime = 15 * time.Minute
		config.HealthCheckPeriod = time.Minute
		config.ConnConfig.ConnectTimeout = 10 * time.Second
	case DevelopmentEnv:
		// Development environment with moderate resource usage
		config.MaxConns = 5
		config.MinConns = 1
		config.MaxConnLifetime = 30 * time.Minute
		config.MaxConnIdleTime = 10 * time.Minute
		config.HealthCheckPeriod = 30 * time.Second
		config.ConnConfig.ConnectTimeout = 5 * time.Second
	default: // testing
		// Testing environment with minimal resource usage
		config.MaxConns = 3
		config.MinConns = 1
		config.MaxConnLifetime = 10 * time.Minute
		config.MaxConnIdleTime = 5 * time.Minute
		config.HealthCheckPeriod = 10 * time.Second
		config.ConnConfig.ConnectTimeout = 2 * time.Second
	}
}

func GetDBConnection() (*pgxpool.Pool, error) {
	// Initialize the database connection pool once
	dbOnce.Do(func() {
		databaseURL := os.Getenv("DATABASE_URL")
		if databaseURL == "" {
			// Build URL from individual components if DATABASE_URL not set
			user := os.Getenv("POSTGRES_USER")
			password := os.Getenv("POSTGRES_PASSWORD")
			host := os.Getenv("POSTGRES_HOST")
			port := os.Getenv("POSTGRES_PORT")
			dbName := os.Getenv("POSTGRES_DB")

			if user == "" || password == "" || host == "" || port == "" || dbName == "" {
				dbInitErr = fmt.Errorf("database configuration missing. Set DATABASE_URL or individual POSTGRES_* variables")
				return
			}

			databaseURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbName)
		}

		// Parse connection string and optimize pool configuration
		config, err := pgxpool.ParseConfig(databaseURL)
		if err != nil {
			dbInitErr = fmt.Errorf("unable to parse database URL: %w", err)
			return
		}

		// Apply environment-specific connection pool optimization
		configureConnectionPool(config)

		db, err = pgxpool.NewWithConfig(context.Background(), config)
		if err != nil {
			dbInitErr = fmt.Errorf("unable to create connection pool: %w", err)
			return
		}
	})

	// Return initialization error if it occurred
	if dbInitErr != nil {
		return nil, dbInitErr
	}

	// Ping the database to ensure the connection is still alive
	if err := db.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}
	return db, nil
}

func CloseDBConnection() {
	fmt.Fprintf(os.Stdout, "Closing db connection\n")

	dbMutex.Lock()
	defer dbMutex.Unlock()

	if db != nil {
		db.Close()
		db = nil
	}
}
