package common

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	ddpgx "gopkg.in/DataDog/dd-trace-go.v1/contrib/jackc/pgx.v5"
)

var (
	db     *pgxpool.Pool
	dbOnce sync.Once
)

func GetDBConnection() *pgxpool.Pool {
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
				Logger.Error("database configuration missing - application cannot start")
				os.Exit(1)
			}

			databaseURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbName)
		}

		// Parse connection string and optimize pool configuration
		config, err := pgxpool.ParseConfig(databaseURL)
		if err != nil {
			Logger.Error("unable to parse database URL - application cannot start", "error", err)
			os.Exit(1)
		}

		// Recommended configuration for pgxpool connecting to pgBouncer
		config.MaxConns = 50
		config.MinConns = 10
		config.MaxConnLifetime = 30 * time.Minute
		config.MaxConnIdleTime = 5 * time.Minute
		config.HealthCheckPeriod = 1 * time.Minute
		config.ConnConfig.ConnectTimeout = 5 * time.Second
		if os.Getenv("DISABLE_PREPARED_STATEMENTS") == "true" {
			config.ConnConfig.StatementCacheCapacity = 0
			config.ConnConfig.PreferSimpleProtocol = true
		}

		// Create connection pool with Datadog instrumentation if enabled
		if os.Getenv("DD_ENABLED") == DDEnabledValue {
			db, err = ddpgx.NewPoolWithConfig(context.Background(), config,
				ddpgx.WithServiceName("decorebator-api"),
				ddpgx.WithTraceQuery(true),
				ddpgx.WithTraceBatch(true),
				ddpgx.WithTracePrepare(true),
				ddpgx.WithTraceConnect(true),
			)
		} else {
			db, err = pgxpool.NewWithConfig(context.Background(), config)
		}
		if err != nil {
			Logger.Error("unable to create connection pool - application cannot start", "error", err)
			os.Exit(1)
		}

		// Ping the database to ensure the connection is still alive
		if err := db.Ping(context.Background()); err != nil {
			Logger.Error("database ping failed - application cannot continue", "error", err)
			os.Exit(1)
		}
	})

	return db
}

func CloseDBConnection() {
	fmt.Fprintf(os.Stdout, "Closing db connection\n")

	if db != nil {
		db.Close()
		db = nil
	}
}
