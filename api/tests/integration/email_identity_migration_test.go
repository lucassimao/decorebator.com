package integration

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func TestCanonicalEmailIdentityMigrationDrill(t *testing.T) {
	ctx := context.Background()
	baseURL, err := url.Parse(os.Getenv("TEST_DATABASE_URL"))
	require.NoError(t, err)
	require.NotEmpty(t, baseURL.Host)

	adminURL := *baseURL
	adminURL.Path = "/postgres"
	adminPool, err := pgxpool.New(ctx, adminURL.String())
	require.NoError(t, err)
	defer adminPool.Close()

	databaseName := fmt.Sprintf("email_identity_drill_%d", time.Now().UnixNano())
	quotedDatabase := pgx.Identifier{databaseName}.Sanitize()
	_, err = adminPool.Exec(ctx, "CREATE DATABASE "+quotedDatabase)
	require.NoError(t, err)

	drillURL := *baseURL
	drillURL.Path = "/" + databaseName
	var drillPool *pgxpool.Pool
	var migrationRunner *migrate.Migrate
	defer func() {
		if drillPool != nil {
			drillPool.Close()
		}
		if migrationRunner != nil {
			_, _ = migrationRunner.Close()
		}
		_, _ = adminPool.Exec(ctx, `SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname=$1`, databaseName)
		if _, dropErr := adminPool.Exec(ctx, "DROP DATABASE IF EXISTS "+quotedDatabase); dropErr != nil {
			t.Errorf("drop migration drill database: %v", dropErr)
		}
	}()

	migrationsPath, err := filepath.Abs("../../cmd/migrate/migrations")
	require.NoError(t, err)
	migrationRunner, err = migrate.New("file://"+migrationsPath, drillURL.String())
	require.NoError(t, err)
	require.NoError(t, migrationRunner.Migrate(77))
	reopenMigrationRunner := func() {
		t.Helper()
		_, _ = migrationRunner.Close()
		migrationRunner, err = migrate.New("file://"+migrationsPath, drillURL.String())
		require.NoError(t, err)
	}

	drillPool, err = pgxpool.New(ctx, drillURL.String())
	require.NoError(t, err)

	insertDrillUser := func(email string) {
		t.Helper()
		_, insertErr := drillPool.Exec(ctx, `
			INSERT INTO users (first_name, last_name, email, password_hash)
			VALUES ('Migration', 'Drill', $1, 'not-a-real-password-hash')
		`, email)
		require.NoError(t, insertErr)
	}
	resetAfterFailedMigration := func() {
		t.Helper()
		require.True(t, migrationIndexExists(ctx, t, drillPool, "users_email_unique_lower"))
		require.False(t, migrationIndexExists(ctx, t, drillPool, "users_email_unique_canonical"))
		reopenMigrationRunner()
		require.NoError(t, migrationRunner.Force(77))
		_, deleteErr := drillPool.Exec(ctx, `DELETE FROM users`)
		require.NoError(t, deleteErr)
	}

	insertDrillUser("Case@example.com")
	insertDrillUser(" case@example.com ")
	require.Error(t, migrationRunner.Steps(1))
	resetAfterFailedMigration()

	insertDrillUser("usér@example.com")
	require.Error(t, migrationRunner.Steps(1))
	resetAfterFailedMigration()

	insertDrillUser("Name <user@example.com>")
	require.Error(t, migrationRunner.Steps(1))
	resetAfterFailedMigration()

	insertDrillUser("Legacy.I@Example.COM")
	require.NoError(t, migrationRunner.Steps(1))
	require.False(t, migrationIndexExists(ctx, t, drillPool, "users_email_unique_lower"))
	require.True(t, migrationIndexExists(ctx, t, drillPool, "users_email_unique_canonical"))

	var valid, ready bool
	require.NoError(t, drillPool.QueryRow(ctx, `
		SELECT indisvalid, indisready
		FROM pg_index
		WHERE indexrelid='users_email_unique_canonical'::regclass
	`).Scan(&valid, &ready))
	require.True(t, valid)
	require.True(t, ready)

	_, err = drillPool.Exec(ctx, `
		INSERT INTO users (first_name, last_name, email, password_hash)
		VALUES ('Duplicate', 'Drill', ' legacy.i@example.com ', 'not-a-real-password-hash')
	`)
	var postgresError *pgconn.PgError
	require.ErrorAs(t, err, &postgresError)
	require.Equal(t, "23505", postgresError.Code)
	require.Equal(t, "users_email_unique_canonical", postgresError.ConstraintName)

	require.NoError(t, migrationRunner.Steps(-1))
	require.True(t, migrationIndexExists(ctx, t, drillPool, "users_email_unique_lower"))
	require.False(t, migrationIndexExists(ctx, t, drillPool, "users_email_unique_canonical"))
}

func migrationIndexExists(ctx context.Context, t *testing.T, pool *pgxpool.Pool, indexName string) bool {
	t.Helper()
	var exists bool
	require.NoError(t, pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM pg_indexes
			WHERE schemaname='public' AND indexname=$1
		)
	`, indexName).Scan(&exists))
	return exists
}
