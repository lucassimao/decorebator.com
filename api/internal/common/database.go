package common

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	db      *pgxpool.Pool
	dbOnce  sync.Once
	dbMutex sync.Mutex
)

func GetDBConnection() (*pgxpool.Pool, error) {
	// Initialize the database connection pool once
	dbOnce.Do(func() {
		var err error
		databaseURL := os.Getenv("DATABASE_URL")
		if databaseURL == "" {
			// Build URL from individual components if DATABASE_URL not set
			user := os.Getenv("POSTGRES_USER")
			password := os.Getenv("POSTGRES_PASSWORD")
			host := os.Getenv("POSTGRES_HOST")
			port := os.Getenv("POSTGRES_PORT")
			dbName := os.Getenv("POSTGRES_DB")

			if user == "" || password == "" || host == "" || port == "" || dbName == "" {
				fmt.Fprintf(os.Stderr, "Database configuration missing. Set DATABASE_URL or individual POSTGRES_* variables\n")
				os.Exit(1)
			}

			databaseURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbName)
		}

		db, err = pgxpool.New(context.Background(), databaseURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
			os.Exit(1)
		}
	})

	// Ping the database to ensure the connection is still alive
	if err := db.Ping(context.Background()); err != nil {
		return nil, err
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
