package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	repo "decorebator.com/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type User = model.User

var ErrInvalidLoginCredentials = errors.New("invalid combination of email and/or password")

var ErrAuthHardeningWritesDisabled = errors.New("AUTH-3 writes are disabled for rollback")

// UserService handles user-related operations with dependency injection
type UserService struct {
	db                      *pgxpool.Pool
	userRepository          *repo.UserRepository
	subscriptionService     *SubscriptionService
	effectiveAccess         *EffectiveAccessService
	authSessions            *AuthSessionService
	jobService              JobService
	dummyPasswordHash       []byte
	passwordComparisons     atomic.Uint64
	pendingLookups          atomic.Uint64
	deleteProfileObject     func(context.Context, string, string, string) error
	profileObjectReferenced func(context.Context, int64, string) (bool, error)
}

func (s *UserService) SetEffectiveAccess(service *EffectiveAccessService) {
	s.effectiveAccess = service
}

// NewUserService creates a new UserService with injected dependencies
func NewUserService(
	db *pgxpool.Pool,
	subscriptionService *SubscriptionService,
	authSessions *AuthSessionService,
	jobs ...JobService,
) (*UserService, error) {
	dummyPasswordHash, err := bcrypt.GenerateFromPassword(
		[]byte("decorebator-auth-timing-sentinel"), common.GetBcryptCost(),
	)
	if err != nil {
		return nil, fmt.Errorf("initialize login timing hash: %w", err)
	}
	service := &UserService{
		db:                  db,
		userRepository:      &repo.UserRepository{Db: db},
		subscriptionService: subscriptionService,
		authSessions:        authSessions,
		dummyPasswordHash:   dummyPasswordHash,
		deleteProfileObject: common.MinIODeleteObjectVersion,
		profileObjectReferenced: func(ctx context.Context, userID int64, objectURL string) (bool, error) {
			var referenced bool
			err := db.QueryRow(ctx, `SELECT EXISTS(
				SELECT 1 FROM users WHERE id=$1 AND profile_picture_url=$2
			)`, userID, objectURL).Scan(&referenced)
			return referenced, err
		},
	}
	if len(jobs) > 0 {
		service.jobService = jobs[0]
	}
	return service, nil
}

func (s *UserService) CreateSession(ctx context.Context, userID int64) (SessionCredentials, error) {
	if s.authSessions == nil {
		return SessionCredentials{}, errors.New("auth-session service is unavailable")
	}
	return s.authSessions.Create(ctx, userID)
}

func (s *UserService) SaveUser(
	ctx context.Context,
	firstName, lastName, password, email string,
	country *string,
	preferredLanguage *string,
	transactions ...pgx.Tx,
) (*User, error) {
	// Validate required parameters
	if firstName == "" || lastName == "" || password == "" || email == "" {
		return nil, common.BusinessError{Message: "firstName, lastName, password, and email are required"}
	}
	if err := common.ValidatePassword(password); err != nil {
		return nil, common.BusinessError{Message: err.Error()}
	}
	canonicalEmail, err := common.NormalizeEmail(email)
	if err != nil {
		return nil, common.BusinessError{Message: "Invalid email address"}
	}

	user, err := s.userRepository.Save(ctx, firstName, lastName, password, canonicalEmail, country, preferredLanguage, transactions...)
	if err != nil {
		common.Logger.Error("failed to save new user", "error", err)
		switch err.(type) {
		case common.BusinessError:
			return nil, err
		default:
			return nil, errors.New("could not save new user")
		}
	}
	return user, nil
}

func (s *UserService) SaveUserAndSession(
	ctx context.Context,
	firstName, lastName, password, email string,
	country *string,
	preferredLanguage *string,
) (*User, SessionCredentials, error) {
	var user *User
	var credentials SessionCredentials
	err := pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
		var saveErr error
		user, saveErr = s.SaveUser(
			ctx, firstName, lastName, password, email, country, preferredLanguage, tx,
		)
		if saveErr != nil {
			return saveErr
		}
		credentials, saveErr = s.authSessions.CreateTx(ctx, tx, user.ID)
		return saveErr
	})
	if err != nil {
		return nil, SessionCredentials{}, err
	}
	return user, credentials, nil
}

// RegisterPendingUser creates an account that cannot authenticate until the
// mailbox owner follows the generic account-security email. Duplicate emails
// intentionally produce the same successful outcome as new registrations.
func (s *UserService) RegisterPendingUser(
	ctx context.Context,
	firstName, lastName, password, email string,
	country *string,
	preferredLanguage *string,
) error {
	if s.jobService == nil {
		return errors.New("account security job service is unavailable")
	}
	if err := common.ValidatePassword(password); err != nil {
		return common.BusinessError{Message: err.Error()}
	}
	pendingPassword, err := newPendingAccountPassword()
	if err != nil {
		return err
	}
	err = pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
		if stateErr := requireAuthHardeningWrites(ctx, tx); stateErr != nil {
			return stateErr
		}
		user, saveErr := s.SaveUser(
			ctx, firstName, lastName, pendingPassword, email, country, preferredLanguage, tx,
		)
		if saveErr != nil {
			return saveErr
		}
		_, saveErr = tx.Exec(ctx, `
			INSERT INTO pending_email_verifications (user_id) VALUES ($1)
		`, user.ID)
		if saveErr != nil {
			return saveErr
		}
		return s.jobService.ScheduleResetPasswordEmailJob(ctx, email, tx)
	})
	if err == nil {
		return nil
	}
	var businessError common.BusinessError
	if errors.As(err, &businessError) && businessError.Error() == "Email already exists." {
		return pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
			if stateErr := requireAuthHardeningWrites(ctx, tx); stateErr != nil {
				return stateErr
			}
			return s.jobService.ScheduleResetPasswordEmailJob(ctx, email, tx)
		})
	}
	return err
}

func requireAuthHardeningWrites(ctx context.Context, tx pgx.Tx) error {
	var writesEnabled bool
	if err := tx.QueryRow(ctx, `
		SELECT writes_enabled
		FROM auth_hardening_rollout_state
		WHERE singleton=TRUE
		FOR SHARE
	`).Scan(&writesEnabled); err != nil {
		return fmt.Errorf("read auth-hardening rollout state: %w", err)
	}
	if !writesEnabled {
		return ErrAuthHardeningWritesDisabled
	}
	return nil
}

func newPendingAccountPassword() (string, error) {
	var random [32]byte
	if _, err := rand.Read(random[:]); err != nil {
		return "", fmt.Errorf("generate pending-account password: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(random[:]), nil
}

func (s *UserService) isEmailVerificationPending(ctx context.Context, userID int64, transactions ...pgx.Tx) (bool, error) {
	s.pendingLookups.Add(1)
	var queryer interface {
		QueryRow(context.Context, string, ...any) pgx.Row
	} = s.db
	if len(transactions) > 0 {
		queryer = transactions[0]
	}
	var pending bool
	err := queryer.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM pending_email_verifications WHERE user_id=$1)
	`, userID).Scan(&pending)
	return pending, err
}

func (s *UserService) UpdatePassword(ctx context.Context, userID int64, password string) error {
	if err := common.ValidatePassword(password); err != nil {
		return common.BusinessError{Message: err.Error()}
	}
	err := s.userRepository.UpdatePassword(ctx, userID, password)
	if err != nil {
		common.Logger.Error("failed to save new user", "error", err)
		return errors.New("could not update the password")
	}
	return nil
}

func (s *UserService) LoginUser(ctx context.Context, email, password string) (SessionCredentials, error) {
	startTime := time.Now()
	canonicalEmail, err := common.NormalizeEmail(email)
	if err != nil {
		s.passwordComparisons.Add(1)
		_ = bcrypt.CompareHashAndPassword(s.dummyPasswordHash, []byte(password))
		if _, pendingErr := s.isEmailVerificationPending(ctx, 0); pendingErr != nil {
			return SessionCredentials{}, errors.New("could not process your request. Try again later")
		}
		return SessionCredentials{}, ErrInvalidLoginCredentials
	}

	dbStart := time.Now()
	var user User
	var credentials SessionCredentials
	var passwordMatches, pending bool
	var bcryptDuration time.Duration
	err = pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
		lockedUser, findErr := s.userRepository.FindLoginUserForUpdate(ctx, tx, canonicalEmail)
		if findErr != nil && !errors.Is(findErr, pgx.ErrNoRows) {
			return findErr
		}
		if findErr == nil {
			user = lockedUser
		}

		users := []User(nil)
		if user.ID > 0 {
			users = []User{user}
		}
		bcryptStart := time.Now()
		_, passwordMatches = compareLoginPassword(
			users,
			password,
			s.dummyPasswordHash,
			func(hash, supplied []byte) error {
				s.passwordComparisons.Add(1)
				return bcrypt.CompareHashAndPassword(hash, supplied)
			},
		)
		bcryptDuration = time.Since(bcryptStart)

		pendingLookupID := int64(0)
		if user.ID > 0 {
			pendingLookupID = user.ID
		}
		pending, findErr = s.isEmailVerificationPending(ctx, pendingLookupID, tx)
		if findErr != nil {
			return findErr
		}
		if !passwordMatches || pending {
			return nil
		}
		if s.authSessions == nil {
			return errors.New("auth-session service is unavailable")
		}
		var createErr error
		credentials, createErr = s.authSessions.CreateTx(ctx, tx, user.ID)
		return createErr
	})
	dbDuration := time.Since(dbStart) - bcryptDuration

	if err != nil {
		// Check if error is due to context timeout/cancellation
		if errors.Is(err, context.DeadlineExceeded) {
			common.Logger.Error("login request timed out",
				"db_query_ms", dbDuration.Milliseconds())
			return SessionCredentials{}, err
		}
		if errors.Is(err, context.Canceled) {
			common.Logger.Error("login request was canceled",
				"db_query_ms", dbDuration.Milliseconds())
			return SessionCredentials{}, err
		}
		common.Logger.Error("failed to login user",
			"error", err,
			"db_query_ms", dbDuration.Milliseconds())
		return SessionCredentials{}, errors.New("could not process your request. Try again later")
	}

	totalDuration := time.Since(startTime)

	// Log performance metrics for successful login attempts
	if passwordMatches {
		if pending {
			return SessionCredentials{}, ErrInvalidLoginCredentials
		}
		common.Logger.Info("login performance metrics",
			"user_id", user.ID,
			"db_query_ms", dbDuration.Milliseconds(),
			"bcrypt_ms", bcryptDuration.Milliseconds(),
			"total_ms", totalDuration.Milliseconds(),
			"status", "success")
		return credentials, nil
	}

	// Log performance even for failed password attempts
	logFields := []any{
		"db_query_ms", dbDuration.Milliseconds(),
		"bcrypt_ms", bcryptDuration.Milliseconds(),
		"total_ms", totalDuration.Milliseconds(),
		"status", "failed_password",
	}
	if user.ID > 0 {
		logFields = append(logFields, "user_id", user.ID)
	}
	common.Logger.Info("login performance metrics", logFields...)
	return SessionCredentials{}, ErrInvalidLoginCredentials
}

// AuthPasswordComparisonCount supports authentication observability and proves
// that every login path performs exactly one password-hash comparison.
func (s *UserService) AuthPasswordComparisonCount() uint64 {
	return s.passwordComparisons.Load()
}

func (s *UserService) AuthPendingLookupCount() uint64 {
	return s.pendingLookups.Load()
}

func (s *UserService) ResetPasswordAndVerifyEmail(
	ctx context.Context,
	userID int64,
	tokenID string,
	password string,
) error {
	if err := common.ValidatePassword(password); err != nil {
		return common.BusinessError{Message: err.Error()}
	}
	if tokenID == "" {
		return common.BusinessError{Message: "token_invalid"}
	}
	tokenHash := sha256.Sum256([]byte(tokenID))
	return pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
		if stateErr := requireAuthHardeningWrites(ctx, tx); stateErr != nil {
			return stateErr
		}
		if lockErr := lockAuthUser(ctx, tx, userID); lockErr != nil {
			return lockErr
		}
		var consumed bool
		err := tx.QueryRow(ctx, `
			UPDATE password_reset_tokens
			SET consumed_at=NOW()
			WHERE token_hash=$1 AND user_id=$2 AND consumed_at IS NULL AND expires_at > NOW()
			RETURNING true
		`, tokenHash[:], userID).Scan(&consumed)
		if errors.Is(err, pgx.ErrNoRows) {
			return common.BusinessError{Message: "token_invalid"}
		}
		if err != nil || !consumed {
			return err
		}
		err = s.userRepository.UpdatePassword(ctx, userID, password, tx)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `DELETE FROM pending_email_verifications WHERE user_id=$1`, userID)
		return err
	})
}

func (s *UserService) ResetLegacyPassword(
	ctx context.Context,
	userID int64,
	encryptedToken string,
	password string,
) error {
	if err := common.ValidatePassword(password); err != nil {
		return common.BusinessError{Message: err.Error()}
	}
	if userID <= 0 || encryptedToken == "" {
		return common.BusinessError{Message: "token_invalid"}
	}
	tokenHash := sha256.Sum256([]byte(encryptedToken))
	return pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
		if stateErr := requireAuthHardeningWrites(ctx, tx); stateErr != nil {
			return stateErr
		}
		if lockErr := lockAuthUser(ctx, tx, userID); lockErr != nil {
			return lockErr
		}
		var consumed bool
		err := tx.QueryRow(ctx, `
			INSERT INTO legacy_password_reset_consumptions (token_hash,user_id)
			VALUES ($1,$2)
			ON CONFLICT (token_hash) DO NOTHING
			RETURNING true
		`, tokenHash[:], userID).Scan(&consumed)
		if errors.Is(err, pgx.ErrNoRows) {
			return common.BusinessError{Message: "token_invalid"}
		}
		if err != nil || !consumed {
			return err
		}
		if err = s.userRepository.UpdatePassword(ctx, userID, password, tx); err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `DELETE FROM pending_email_verifications WHERE user_id=$1`, userID)
		return err
	})
}

func lockAuthUser(ctx context.Context, tx pgx.Tx, userID int64) error {
	var lockedUserID int64
	if err := tx.QueryRow(ctx, `SELECT id FROM users WHERE id=$1 FOR UPDATE`, userID).Scan(&lockedUserID); err != nil {
		return err
	}
	return nil
}

func compareLoginPassword(
	users []User,
	password string,
	dummyHash []byte,
	compare func([]byte, []byte) error,
) (User, bool) {
	user := User{}
	hash := dummyHash
	if len(users) == 1 {
		user = users[0]
		hash = []byte(user.PasswordHash)
	}
	compareErr := compare(hash, []byte(password))
	return user, len(users) == 1 && compareErr == nil
}

func (s *UserService) VerifyPassword(ctx context.Context, email, password string) error {
	canonicalEmail, err := common.NormalizeEmail(email)
	if err != nil {
		return errors.New("invalid combination of email and/or password")
	}
	users, err := s.userRepository.Find(ctx, repo.FindUserArgs{Email: &canonicalEmail})
	if err != nil || len(users) != 1 {
		return errors.New("invalid combination of email and/or password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(users[0].PasswordHash), []byte(password)); err != nil {
		return errors.New("invalid combination of email and/or password")
	}
	return nil
}

func (s *UserService) GetProfile(ctx context.Context, userID int64) (*User, bool, error) {
	users, err := s.userRepository.Find(ctx, repo.FindUserArgs{
		ID: &userID,
	})
	if err != nil {
		return nil, false, err
	}

	if len(users) == 0 {
		return nil, false, errors.New("user not found")
	}

	user := &users[0]
	planChanged := false

	// Check if user needs plan downgrade due to expired grace period
	// (checkAndDowngradeExpiredSubscription now handles the premium check internally)
	downgraded, err := s.checkAndDowngradeExpiredSubscription(ctx, userID, user)
	if err != nil {
		// Log error but don't fail the request - graceful degradation
		common.Logger.Error("failed to check subscription grace period", "userId", userID, "error", err)
	} else {
		planChanged = downgraded
	}
	if s.effectiveAccess != nil {
		effectivePlan, accessErr := s.effectiveAccess.Plan(ctx, userID, user.SubscriptionPlan)
		if accessErr != nil {
			return nil, false, accessErr
		}
		user.SubscriptionPlan = effectivePlan
	}

	return user, planChanged, nil
}

func (s *UserService) checkAndDowngradeExpiredSubscription(ctx context.Context, userID int64, user *User) (bool, error) {
	// Only check for downgrade if user currently has a premium plan
	if user.SubscriptionPlan == model.PlanFree {
		return false, nil // Already free, no downgrade needed
	}

	// Check if user has active subscription (includes grace period)
	activeSub, err := s.subscriptionService.GetActiveSubscriptionForUser(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get subscription: %w", err)
	}

	// If no active subscription found, user is beyond grace period
	if activeSub == nil {
		// Downgrade plan to free
		if err := s.userRepository.UpdateSubscriptionPlan(ctx, userID, model.PlanFree); err != nil {
			return false, fmt.Errorf("failed to downgrade subscription plan: %w", err)
		}

		// Update the user object to reflect the change
		originalPlan := user.SubscriptionPlan
		user.SubscriptionPlan = model.PlanFree

		common.Logger.Info("downgraded user subscription plan due to expired grace period",
			"userId", userID, "previousPlan", originalPlan, "newPlan", model.PlanFree)

		return true, nil
	}

	return false, nil
}

func (s *UserService) Delete(ctx context.Context, userID int64) error {
	if s.jobService == nil {
		return errors.New("account cleanup scheduler is unavailable")
	}
	return s.userRepository.Delete(ctx, userID, func(ctx context.Context, tx pgx.Tx, prefix string) error {
		return s.jobService.ScheduleAccountCleanupJob(ctx, prefix, &tx)
	})
}

func (s *UserService) Exists(ctx context.Context, userID int64) (bool, error) {
	return s.userRepository.Exists(ctx, userID)
}

func (s *UserService) CompensateProfileUpload(ctx context.Context, userID int64, receipt common.UploadReceipt) error {
	if s.profileObjectReferenced == nil {
		return errors.New("profile object reference checker is unavailable")
	}
	referenced, referenceErr := s.profileObjectReferenced(ctx, userID, receipt.URL)
	if referenceErr == nil && referenced {
		return nil
	}
	if referenceErr == nil && s.deleteProfileObject != nil {
		if err := s.deleteProfileObject(ctx, receipt.Bucket, receipt.ObjectName, receipt.VersionID); err == nil {
			return nil
		}
	}
	return s.ScheduleProfileUploadReconciliation(ctx, userID, receipt)
}

func (s *UserService) ScheduleProfileUploadReconciliation(ctx context.Context, userID int64, receipt common.UploadReceipt) error {
	if s.jobService == nil {
		return errors.New("account cleanup scheduler is unavailable")
	}
	_, err := s.jobService.ScheduleProfileUploadReconciliationJob(ctx, ProfileUploadReconciliationArgs{
		UserID: userID, Bucket: receipt.Bucket, ObjectName: receipt.ObjectName,
		ObjectURL: receipt.URL, ObjectVersionID: receipt.VersionID,
	})
	return err
}

func (s *UserService) ScheduleUncertainProfileUploadCleanup(ctx context.Context, userID int64, bucket, objectName, objectURL string) (int64, error) {
	if s.jobService == nil {
		return 0, errors.New("profile upload reconciliation scheduler is unavailable")
	}
	return s.jobService.ScheduleProfileUploadReconciliationJob(ctx, ProfileUploadReconciliationArgs{
		UserID: userID, Bucket: bucket, ObjectName: objectName, ObjectURL: objectURL, DeleteAllVersions: true,
	})
}

func ValidateProfileUpdate(firstName, lastName, password *string) error {
	if firstName != nil && strings.TrimSpace(*firstName) == "" {
		return common.BusinessError{Message: "First name is required"}
	}
	if firstName != nil && (len(*firstName) > 400 || utf8.RuneCountInString(*firstName) > 100) {
		return common.BusinessError{Message: "First name is too long"}
	}
	if lastName != nil && strings.TrimSpace(*lastName) == "" {
		return common.BusinessError{Message: "Last name is required"}
	}
	if lastName != nil && (len(*lastName) > 400 || utf8.RuneCountInString(*lastName) > 100) {
		return common.BusinessError{Message: "Last name is too long"}
	}
	if password != nil {
		if err := common.ValidatePassword(*password); err != nil {
			return common.BusinessError{Message: err.Error()}
		}
	}
	return nil
}

func (s *UserService) UpdateProfile(ctx context.Context, userID int64, firstName, lastName, country, preferredLanguage, profilePictureURL, password *string, dateOfBirth *time.Time, notificationsEnabled *bool) (*User, error) {
	return s.updateProfile(ctx, userID, firstName, lastName, country, preferredLanguage, profilePictureURL, password, dateOfBirth, notificationsEnabled, 0, common.UploadReceipt{})
}

func (s *UserService) UpdateProfileWithUploadIntent(ctx context.Context, userID int64, firstName, lastName, country, preferredLanguage, profilePictureURL, password *string, dateOfBirth *time.Time, notificationsEnabled *bool, cleanupIntentID int64, receipt common.UploadReceipt) (*User, error) {
	if cleanupIntentID <= 0 || receipt.ObjectName == "" || profilePictureURL == nil || *profilePictureURL != receipt.URL {
		return nil, errors.New("invalid profile upload persistence intent")
	}
	return s.updateProfile(ctx, userID, firstName, lastName, country, preferredLanguage, profilePictureURL, password, dateOfBirth, notificationsEnabled, cleanupIntentID, receipt)
}

func (s *UserService) updateProfile(ctx context.Context, userID int64, firstName, lastName, country, preferredLanguage, profilePictureURL, password *string, dateOfBirth *time.Time, notificationsEnabled *bool, cleanupIntentID int64, receipt common.UploadReceipt) (*User, error) {
	if err := ValidateProfileUpdate(firstName, lastName, password); err != nil {
		return nil, err
	}

	args := repo.UpdateUserProfileArgs{
		ID:                   userID,
		FirstName:            firstName,
		LastName:             lastName,
		Country:              country,
		DateOfBirth:          dateOfBirth,
		PreferredLanguage:    preferredLanguage,
		ProfilePictureURL:    profilePictureURL,
		Password:             password,
		NotificationsEnabled: notificationsEnabled,
	}

	var user *User
	var err error
	if cleanupIntentID > 0 {
		if s.jobService == nil {
			return nil, errors.New("profile upload reconciliation scheduler is unavailable")
		}
		err = pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
			if finalizeErr := s.jobService.FinalizeProfileUploadReconciliationJob(ctx, cleanupIntentID, ProfileUploadReconciliationArgs{
				UserID: userID, Bucket: receipt.Bucket, ObjectName: receipt.ObjectName,
				ObjectURL: receipt.URL, ObjectVersionID: receipt.VersionID,
			}, tx); finalizeErr != nil {
				return finalizeErr
			}
			args.Transaction = tx
			user, err = s.userRepository.UpdateUserProfile(ctx, args)
			return err
		})
	} else {
		user, err = s.userRepository.UpdateUserProfile(ctx, args)
	}
	if err != nil {
		common.Logger.Error("failed to update user profile", "error", err, "userID", userID)
		switch err.(type) {
		case common.BusinessError:
			return nil, err
		default:
			return nil, fmt.Errorf("could not update user profile: %w", err)
		}
	}

	return user, nil
}

// ValidateUserEligibilityForWorkers checks if a user (especially free tier) is eligible for worker processing
// Free users are limited to:
// - 1 wordlist only
// - Maximum 10 words total
func (s *UserService) ValidateUserEligibilityForWorkers(ctx context.Context, userID int64) error {
	// Get user information including subscription plan
	var subscriptionPlan model.SubscriptionPlan
	err := s.userRepository.Db.QueryRow(ctx,
		"SELECT subscription_plan FROM users WHERE id = $1",
		userID).Scan(&subscriptionPlan)

	if err != nil {
		return fmt.Errorf("failed to get user subscription plan: %w", err)
	}

	// Premium users (monthly/annual) have no restrictions
	if subscriptionPlan == model.PlanMonthly || subscriptionPlan == model.PlanAnnual {
		return nil
	}
	if s.effectiveAccess != nil {
		hasPremium, accessErr := s.effectiveAccess.HasPremium(ctx, userID)
		if accessErr != nil {
			return accessErr
		}
		if hasPremium {
			return nil
		}
	}

	// For free users, check wordlist and word count limits
	var wordlistCount int
	var totalWordCount int

	// Count wordlists
	err = s.userRepository.Db.QueryRow(ctx,
		"SELECT COUNT(*) FROM wordlists WHERE user_id = $1",
		userID).Scan(&wordlistCount)

	if err != nil {
		return fmt.Errorf("failed to count wordlists: %w", err)
	}

	// Count total words across all wordlists
	err = s.userRepository.Db.QueryRow(ctx,
		"SELECT COUNT(*) FROM words WHERE user_id = $1 AND learned = false",
		userID).Scan(&totalWordCount)

	if err != nil {
		return fmt.Errorf("failed to count words: %w", err)
	}

	// Validate limits
	if wordlistCount > model.FreeWordlistLimit {
		return common.BusinessError{
			Message: fmt.Sprintf("Free users are limited to %d wordlist. Please upgrade to add more content.", model.FreeWordlistLimit),
		}
	}

	if totalWordCount > model.FreeWordsPerList {
		return common.BusinessError{
			Message: fmt.Sprintf("Free users are limited to %d words total. Please upgrade to add more words.", model.FreeWordsPerList),
		}
	}

	return nil
}
