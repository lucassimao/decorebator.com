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
	"decorebator.com/internal/repository"
	repo "decorebator.com/internal/repository"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

type User = model.User

var userRepository *repo.UserRepository
var wordlistRepository *repo.WordlistRepository

const AUTH_TOKEN_DURATION = (24 * time.Hour) * 365 // 1 year

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
			ExpiresAt: time.Now().Add(AUTH_TOKEN_DURATION).Unix(), // Token is valid for 1 year
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

func init() {
	db := common.GetDBConnection()
	userRepository = &repo.UserRepository{Db: db}
	wordlistRepository = &repo.WordlistRepository{Db: db}
}

func SaveUser(firstName, lastName, password, email string, country *string) (*User, error) {
	// Validate required parameters
	if firstName == "" || lastName == "" || password == "" || email == "" {
		return nil, common.BusinessError{Message: "firstName, lastName, password, and email are required"}
	}

	user, err := userRepository.Save(firstName, lastName, password, email, country)
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

func UpdatePassword(userId int64, password string) error {
	err := userRepository.UpdatePassword(userId, password)
	if err != nil {
		common.Logger.Error("failed to save new user", "error", err)
		return errors.New("could not update the password")
	}
	return nil
}

func LoginUser(email, password string) (string, error) {
	lowerCaseEmail := strings.ToLower(email)

	args := repo.FindUserArgs{
		Email: &lowerCaseEmail,
	}
	results, err := userRepository.Find(context.Background(), args)
	if err != nil {
		common.Logger.Error("failed to login user", "error", err)
		return "", errors.New("could not process your request. Try again later")
	}

	if len(results) != 1 {
		return "", errors.New("invalid combination of email and/or password")
	}

	user := results[0]

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err == nil {
		return GenerateJWT(user)
	} else {
		return "", errors.New("invalid combination of email and/or password")
	}

}

func GetProfile(userID int64) (*User, bool, error) {
	users, err := userRepository.Find(context.Background(), repository.FindUserArgs{
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
	downgraded, err := checkAndDowngradeExpiredSubscription(userID, user)
	if err != nil {
		// Log error but don't fail the request - graceful degradation
		common.Logger.Error("failed to check subscription grace period", "userId", userID, "error", err)
	} else {
		planChanged = downgraded
	}

	return user, planChanged, nil
}

func checkAndDowngradeExpiredSubscription(userID int64, user *User) (bool, error) {
	// Only check for downgrade if user currently has a premium plan
	if user.SubscriptionPlan == model.PlanFree {
		return false, nil // Already free, no downgrade needed
	}

	// Get subscription repository
	subRepo := repository.NewSubscriptionRepository(userRepository.Db)

	// Check if user has active subscription (includes grace period)
	activeSub, err := subRepo.GetActiveSubscriptionForUser(context.Background(), userID)
	if err != nil {
		return false, fmt.Errorf("failed to get subscription: %w", err)
	}

	// If no active subscription found, user is beyond grace period
	if activeSub == nil {
		// Downgrade plan to free
		if err := userRepository.UpdateSubscriptionPlan(context.Background(), userID, model.PlanFree); err != nil {
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

func Delete(userID int64) error {
	if _, deleteReportsErr := DeleteUserErrorReports(userID); deleteReportsErr != nil {
		common.Logger.Error("failed to delete user error reports", "userId", userID, "error", deleteReportsErr)
	}
	if _, deleteWordlistsErr := wordlistRepository.DeleteAll(userID); deleteWordlistsErr != nil {
		common.Logger.Error("failed to delete user wordlists", "userId", userID, "error", deleteWordlistsErr)
	}
	err := userRepository.Delete(userID)
	return err
}

func UpdateProfile(userID int64, firstName, lastName, country, preferredLanguage, profilePictureUrl, password *string, dateOfBirth *time.Time) (*User, error) {
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
		ID:                userID,
		FirstName:         firstName,
		LastName:          lastName,
		Country:           country,
		DateOfBirth:       dateOfBirth,
		PreferredLanguage: preferredLanguage,
		ProfilePictureURL: profilePictureUrl,
		Password:          password,
	}

	user, err := userRepository.UpdateUserProfile(args)
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
