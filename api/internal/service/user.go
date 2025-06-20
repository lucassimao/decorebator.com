package service

import (
	"errors"
	"fmt"
	"os"
	"strings"
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

const AUTH_TOKEN_DURATION = (24 * time.Hour) * 365 // 1 year

// jwt.StandardClaims is an embedded type to provide expiry time, issued at time, etc.
type Claims struct {
	Email            string                 `json:"email"`
	Environment      string                 `json:"environment"`
	SubscriptionPlan model.SubscriptionPlan `json:"subscriptionPlan"`
	jwt.StandardClaims
}

func GenerateJWT(user User) (string, error) {

	claims := &Claims{
		Email:            user.Email,
		Environment:      os.Getenv("ENV"),
		SubscriptionPlan: user.SubscriptionPlan,
		StandardClaims: jwt.StandardClaims{
			Issuer:    "Decorebator",
			ExpiresAt: time.Now().Add(AUTH_TOKEN_DURATION).Unix(), // Token is valid for 1 year
			Subject:   fmt.Sprint(user.ID),
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	var jwtKey = []byte(os.Getenv("JWT_KEY"))
	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func init() {
	db, err := common.GetDBConnection()
	if err != nil {
		common.Logger.Error("failed to open db connection", "error", err)
		os.Exit(1)
	}
	userRepository = &repo.UserRepository{Db: db}
}

func SaveUser(firstName, lastName, password, email string) (*User, error) {
	user, err := userRepository.Save(firstName, lastName, password, email)
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
	results, err := userRepository.Find(args)
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

func GetProfile(userID int64) (*User, error) {
	users, err := userRepository.Find(repository.FindUserArgs{
		ID: &userID,
	})
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, errors.New("user not found")
	}

	return &users[0], nil
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
