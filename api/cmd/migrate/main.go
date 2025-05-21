package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

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

	sslMode := common.Env.PostgresSSLMode
	dbUser := common.Env.PostgresUser
	dbPassword := common.Env.PostgresPassword
	dbName := common.Env.PostgresDB
	dbHost := common.Env.PostgresHost
	dbPort, err := strconv.Atoi(common.Env.PostgresPort)
	if err != nil {
		log.Fatalf("Invalid DB_PORT: %v", err)
	}

	// e.g., postgres://user:pass@host:5432/dbname
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", dbUser, dbPassword,
		dbHost, dbPort, dbName, sslMode)

	// Use database/sql
	db, err := sql.Open("postgres", dbURL)
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
