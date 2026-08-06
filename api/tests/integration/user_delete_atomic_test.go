package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"decorebator.com/internal/repository"
	"decorebator.com/internal/service"
	"decorebator.com/tests/integration/setup"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepositoryDeleteRemovesAccountDataAtomically(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setup.CreateTestDB(t)
	defer db.Close()
	ctx := context.Background()
	require.NoError(t, setup.RunMigrations(db))
	require.NoError(t, setup.CleanTestData(db))

	repo := &repository.UserRepository{Db: db}
	riverClient, err := river.NewClient(riverpgxv5.New(db), &river.Config{})
	require.NoError(t, err)
	jobs := service.NewJobService(riverClient)
	scheduleCleanup := func(ctx context.Context, tx pgx.Tx, prefix string) error {
		return jobs.ScheduleAccountCleanupJob(ctx, prefix, &tx)
	}

	t.Run("deletes the user and non-cascading account data", func(t *testing.T) {
		userID := createUserDeletionFixture(ctx, t, db, "success")

		require.NoError(t, repo.Delete(ctx, userID, scheduleCleanup))
		assertUserDeletionCounts(ctx, t, db, userID, 0, 0, 0, 0)
		assertCleanupJobCount(ctx, t, db, userID, 1)
	})

	t.Run("rolls every deletion back when the user delete fails", func(t *testing.T) {
		userID := createUserDeletionFixture(ctx, t, db, "rollback")
		_, err := db.Exec(ctx, fmt.Sprintf(`
			CREATE OR REPLACE FUNCTION test_reject_user_%d_delete()
			RETURNS trigger AS $$
			BEGIN
				RAISE EXCEPTION 'forced user deletion failure';
			END;
			$$ LANGUAGE plpgsql;
			CREATE TRIGGER test_reject_user_%d_delete_trigger
			BEFORE DELETE ON users
			FOR EACH ROW WHEN (OLD.id = %d)
			EXECUTE FUNCTION test_reject_user_%d_delete();
		`, userID, userID, userID, userID))
		require.NoError(t, err)
		defer func() {
			_, cleanupErr := db.Exec(ctx, fmt.Sprintf(`
				DROP TRIGGER IF EXISTS test_reject_user_%d_delete_trigger ON users;
				DROP FUNCTION IF EXISTS test_reject_user_%d_delete();
			`, userID, userID))
			require.NoError(t, cleanupErr)
		}()

		require.ErrorContains(t, repo.Delete(ctx, userID, scheduleCleanup), "forced user deletion failure")
		assertUserDeletionCounts(ctx, t, db, userID, 1, 1, 1, 1)
		assertCleanupJobCount(ctx, t, db, userID, 0)
	})
}

func assertCleanupJobCount(ctx context.Context, t *testing.T, db *pgxpool.Pool, userID int64, want int) {
	t.Helper()
	var count int
	require.NoError(t, db.QueryRow(ctx, `
		SELECT count(*) FROM river_job
		WHERE kind='account_cleanup' AND args->>'profile_object_prefix'=$1
	`, fmt.Sprintf("users/%d-", userID)).Scan(&count))
	assert.Equal(t, want, count)
}

func createUserDeletionFixture(ctx context.Context, t *testing.T, db *pgxpool.Pool, suffix string) int64 {
	t.Helper()

	var userID, wordlistID, wordID int64
	email := fmt.Sprintf("delete-%s-%d@example.com", suffix, time.Now().UnixNano())
	require.NoError(t, db.QueryRow(ctx, `
		INSERT INTO users (first_name, last_name, email, password_hash)
		VALUES ('Delete', 'Test', $1, 'not-a-real-password')
		RETURNING id
	`, email).Scan(&userID))
	require.NoError(t, db.QueryRow(ctx, `
		INSERT INTO wordlists (name, description, user_id)
		VALUES ('Delete fixture', 'Atomic account deletion fixture', $1)
		RETURNING id
	`, userID).Scan(&wordlistID))
	require.NoError(t, db.QueryRow(ctx, `
		INSERT INTO words (name, user_id, wordlist_id)
		VALUES ('fixture', $1, $2)
		RETURNING id
	`, userID, wordlistID).Scan(&wordID))
	_, err := db.Exec(ctx, `
		INSERT INTO error_reports (user_id, word_id, error_type)
		VALUES ($1, $2, 'missing_image')
	`, userID, wordID)
	require.NoError(t, err)

	return userID
}

func assertUserDeletionCounts(
	ctx context.Context,
	t *testing.T,
	db *pgxpool.Pool,
	userID int64,
	wantUsers, wantWordlists, wantWords, wantReports int,
) {
	t.Helper()

	var users, wordlists, words, reports int
	require.NoError(t, db.QueryRow(ctx, `SELECT count(*) FROM users WHERE id=$1`, userID).Scan(&users))
	require.NoError(t, db.QueryRow(ctx, `SELECT count(*) FROM wordlists WHERE user_id=$1`, userID).Scan(&wordlists))
	require.NoError(t, db.QueryRow(ctx, `SELECT count(*) FROM words WHERE user_id=$1`, userID).Scan(&words))
	require.NoError(t, db.QueryRow(ctx, `SELECT count(*) FROM error_reports WHERE user_id=$1`, userID).Scan(&reports))
	assert.Equal(t, wantUsers, users)
	assert.Equal(t, wantWordlists, wordlists)
	assert.Equal(t, wantWords, words)
	assert.Equal(t, wantReports, reports)
}
