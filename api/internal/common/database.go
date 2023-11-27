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
		url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
			os.Getenv("POSTGRES_USER"),
			os.Getenv("POSTGRES_PASSWORD"),
			os.Getenv("POSTGRES_HOST"),
			os.Getenv("POSTGRES_PORT"),
			os.Getenv("POSTGRES_DB"))

		var err error
		db, err = pgxpool.New(context.Background(), url)
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
