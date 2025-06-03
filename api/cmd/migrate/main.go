package main

import (
	"database/sql"
	"fmt"
	"log"

	"embed"

	"decorebator.com/internal/common"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type verboseLogger struct{}

func (*verboseLogger) Printf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}

func (*verboseLogger) Verbose() bool {
	return true
}

func main() {

	// Use database/sql
	db, err := sql.Open("postgres", common.Env.DatabaseUrl)
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	// Create Postgres driver for migrate
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("Failed to create migrate driver: %v", err)
	}

	sourceDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		log.Fatal(err)
	}

	m, err := migrate.NewWithInstance(
		"iofs", sourceDriver,
		"postgres", driver)
	defer m.Close()

	m.Log = &verboseLogger{}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("✅ Migration applied successfully")
}
