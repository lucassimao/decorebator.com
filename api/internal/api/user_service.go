package api

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	repo "decorebator.com/internal/repository"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

type User = model.User

var userRepository *repo.UserRepository

const AUTH_TOKEN_DURATION = 24 * time.Hour

// jwt.StandardClaims is an embedded type to provide expiry time, issued at time, etc.
type Claims struct {
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Environment string `json:"environment"`
	jwt.StandardClaims
}

func generateJWT(user User) (string, error) {
	ginMode := os.Getenv("GIN_MODE")

	claims := &Claims{
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Environment: ginMode,
		StandardClaims: jwt.StandardClaims{
			Issuer:    "Decorebator",
			ExpiresAt: time.Now().Add(AUTH_TOKEN_DURATION).Unix(), // Token is valid for 24 hour
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

func LoginUser(email, password string) (string, error) {
	lowerCaseEmail := strings.ToLower(email)

	args := repo.FindUserArgs{
		Email: lowerCaseEmail,
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
		return generateJWT(user)
	} else {
		return "", errors.New("invalid combination of email and/or password")
	}

}
