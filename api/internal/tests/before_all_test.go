package tests

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"decorebator.com/internal/common"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// CustomLogger implements the migrate.Logger interface
type CustomLogger struct{}

func (l *CustomLogger) Printf(format string, v ...interface{}) {
	if strings.Contains(format, "Finished") {
		log.Printf(format, v...)
	}
}

func (l *CustomLogger) Verbose() bool {
	return true
}

// TestMain will be called before any tests are run.
func TestMain(m *testing.M) {
	fmt.Println("Running global setup (BeforeAll)")

	url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		common.Env.PostgresUser, common.Env.PostgresPassword,
		common.Env.PostgresHost,
		common.Env.PostgresPort, common.Env.PostgresDB)
	db, err := sql.Open("postgres", url)

	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	// project is copied to folder /app
	migrator, err := migrate.NewWithDatabaseInstance("file:///app/migrations", "postgres", driver)

	// Attach the custom logger
	migrator.Log = &CustomLogger{}

	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	migrator.Up()

	// Run all the tests
	code := m.Run()

	// AfterAll: Cleanup code that runs once after all tests
	afterAll()

	// Exit with the status code returned by m.Run()
	os.Exit(code)
}

func afterAll() {

}
