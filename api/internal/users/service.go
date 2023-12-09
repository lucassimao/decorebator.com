package users

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"decorebator.com/internal/common"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

var repository *UserRepository

// Claims struct that will be encoded to a JWT.
// jwt.StandardClaims is an embedded type to provide expiry time, issued at time, etc.
type Claims struct {
	UserID int64 `json:"userId"`
	jwt.StandardClaims
}

func generateJWT(userID int64) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour) // Token is valid for 24 hour
	claims := &jwt.StandardClaims{
		ExpiresAt: expirationTime.Unix(),
		Subject:   fmt.Sprint(userID),
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
		log.Fatal("failed to open db connection: ", err)
	}
	repository = &UserRepository{db}
}

func SaveUser(firstName, lastName, password, email string) (*User, error) {
	user, err := repository.save(firstName, lastName, password, email)
	if err != nil {
		log.Println("Failure at SaveUser:", err)
		return nil, errors.New("Could not save new user")
	}
	return user, nil
}

func LoginUser(email, password string) (string, error) {
	args := FindArgs{
		email: &email,
	}
	results, err := repository.find(args)
	if err != nil {
		log.Println("Failure at LoginUser:", err)
		return "", errors.New("Could not start process login. Try again later")
	}

	if len(results) != 1 {
		return "", errors.New("Invalid combination of email and/or password.")
	}

	user := results[0]

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err == nil {
		return generateJWT(user.ID)
	} else {
		return "", errors.New("Invalid combination of email and/or password.")
	}

}
