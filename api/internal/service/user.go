package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	repo "decorebator.com/internal/repository"

	"github.com/dgrijalva/jwt-go"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type User = model.User

// UserService handles user-related operations with dependency injection
type UserService struct {
	userRepository      *repo.UserRepository
	subscriptionService *SubscriptionService
	effectiveAccess     *EffectiveAccessService
	jobService          JobService
	deleteProfileObject func(context.Context, string, string) error
}

func (s *UserService) SetEffectiveAccess(service *EffectiveAccessService) {
	s.effectiveAccess = service
}

// NewUserService creates a new UserService with injected dependencies
func NewUserService(db *pgxpool.Pool, subscriptionService *SubscriptionService, jobs ...JobService) *UserService {
	service := &UserService{
		userRepository:      &repo.UserRepository{Db: db},
		subscriptionService: subscriptionService,
		deleteProfileObject: common.MinIODeleteObject,
	}
	if len(jobs) > 0 {
		service.jobService = jobs[0]
	}
	return service
}

const AuthTokenDuration = (24 * time.Hour) * 365 // 1 year

// JWT configuration cached for performance
var (
	jwtKey     []byte
	jwtEnv     string
	jwtOnce    sync.Once
	jwtInitErr error
)

// jwt.StandardClaims is an embedded type to provide expiry time, issued at time, etc.
type Claims struct {
	Email            string                 `json:"email"`
	Environment      string                 `json:"environment"`
	SubscriptionPlan model.SubscriptionPlan `json:"subscriptionPlan"`
	jwt.StandardClaims
}

// initJWTConfig initializes JWT configuration once for performance
func initJWTConfig() {
	jwtKeyStr := os.Getenv("JWT_KEY")
	if jwtKeyStr == "" {
		jwtInitErr = errors.New("JWT_KEY environment variable is required")
		return
	}
	jwtKey = []byte(jwtKeyStr)
	jwtEnv = os.Getenv("ENV")
}

func GenerateJWT(user User) (string, error) {
	// Initialize JWT configuration once
	jwtOnce.Do(initJWTConfig)
	if jwtInitErr != nil {
		return "", jwtInitErr
	}

	claims := &Claims{
		Email:            user.Email,
		Environment:      jwtEnv, // Use cached environment value
		SubscriptionPlan: user.SubscriptionPlan,
		StandardClaims: jwt.StandardClaims{
			Issuer:    "Decorebator",
			ExpiresAt: time.Now().Add(AuthTokenDuration).Unix(), // Token is valid for 1 year
			Subject:   fmt.Sprint(user.ID),
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey) // Use cached JWT key

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *UserService) SaveUser(ctx context.Context, firstName, lastName, password, email string, country *string, preferredLanguage *string) (*User, error) {
	// Validate required parameters
	if firstName == "" || lastName == "" || password == "" || email == "" {
		return nil, common.BusinessError{Message: "firstName, lastName, password, and email are required"}
	}
	canonicalEmail, err := common.NormalizeEmail(email)
	if err != nil {
		return nil, common.BusinessError{Message: "Invalid email address"}
	}

	user, err := s.userRepository.Save(ctx, firstName, lastName, password, canonicalEmail, country, preferredLanguage)
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

func (s *UserService) UpdatePassword(ctx context.Context, userID int64, password string) error {
	err := s.userRepository.UpdatePassword(ctx, userID, password)
	if err != nil {
		common.Logger.Error("failed to save new user", "error", err)
		return errors.New("could not update the password")
	}
	return nil
}

func (s *UserService) LoginUser(ctx context.Context, email, password string) (string, error) {
	startTime := time.Now()
	canonicalEmail, err := common.NormalizeEmail(email)
	if err != nil {
		return "", errors.New("invalid combination of email and/or password")
	}

	args := repo.FindUserArgs{
		Email: &canonicalEmail,
	}

	// Measure database query time
	dbStart := time.Now()
	results, err := s.userRepository.Find(ctx, args)
	dbDuration := time.Since(dbStart)

	if err != nil {
		// Check if error is due to context timeout/cancellation
		if errors.Is(err, context.DeadlineExceeded) {
			common.Logger.Error("login request timed out",
				"db_query_ms", dbDuration.Milliseconds())
			return "", err
		}
		if errors.Is(err, context.Canceled) {
			common.Logger.Error("login request was canceled",
				"db_query_ms", dbDuration.Milliseconds())
			return "", err
		}
		common.Logger.Error("failed to login user",
			"error", err,
			"db_query_ms", dbDuration.Milliseconds())
		return "", errors.New("could not process your request. Try again later")
	}

	if len(results) != 1 {
		return "", errors.New("invalid combination of email and/or password")
	}

	user := results[0]

	// Measure bcrypt comparison time
	bcryptStart := time.Now()
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	bcryptDuration := time.Since(bcryptStart)

	totalDuration := time.Since(startTime)

	// Log performance metrics for successful login attempts
	if err == nil {
		common.Logger.Info("login performance metrics",
			"user_id", user.ID,
			"db_query_ms", dbDuration.Milliseconds(),
			"bcrypt_ms", bcryptDuration.Milliseconds(),
			"total_ms", totalDuration.Milliseconds(),
			"status", "success")
		return GenerateJWT(user)
	}

	// Log performance even for failed password attempts
	common.Logger.Info("login performance metrics",
		"user_id", user.ID,
		"db_query_ms", dbDuration.Milliseconds(),
		"bcrypt_ms", bcryptDuration.Milliseconds(),
		"total_ms", totalDuration.Milliseconds(),
		"status", "failed_password")
	return "", errors.New("invalid combination of email and/or password")
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

func (s *UserService) CompensateProfileUpload(ctx context.Context, objectName string) error {
	if s.deleteProfileObject == nil {
		return errors.New("profile object deleter is unavailable")
	}
	if err := s.deleteProfileObject(ctx, profilePictureBucket, objectName); err == nil {
		return nil
	}
	if s.jobService == nil {
		return errors.New("account cleanup scheduler is unavailable")
	}
	return s.jobService.ScheduleProfileObjectCleanupJob(ctx, objectName)
}

func (s *UserService) UpdateProfile(ctx context.Context, userID int64, firstName, lastName, country, preferredLanguage, profilePictureURL, password *string, dateOfBirth *time.Time, notificationsEnabled *bool) (*User, error) {
	// Validate required fields
	if firstName != nil && strings.TrimSpace(*firstName) == "" {
		return nil, common.BusinessError{Message: "First name is required"}
	}

	if lastName != nil && strings.TrimSpace(*lastName) == "" {
		return nil, common.BusinessError{Message: "Last name is required"}
	}

	if password != nil && strings.TrimSpace(*password) == "" {
		return nil, common.BusinessError{Message: "Password is required"}
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

	user, err := s.userRepository.UpdateUserProfile(ctx, args)
	if err != nil {
		common.Logger.Error("failed to update user profile", "error", err, "userID", userID)
		switch err.(type) {
		case common.BusinessError:
			return nil, err
		default:
			return nil, errors.New("could not update user profile")
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
