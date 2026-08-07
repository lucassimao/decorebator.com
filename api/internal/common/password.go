package common

import (
	"fmt"
	"os"
	"strconv"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

const (
	PasswordMinCharacters       = 8
	PasswordMaxBytes            = 72
	ProductionBcryptMinimumCost = bcrypt.DefaultCost
)

func ValidatePassword(password string) error {
	if utf8.ValidString(password) && utf8.RuneCountInString(password) < PasswordMinCharacters {
		return fmt.Errorf("password must contain at least %d characters", PasswordMinCharacters)
	}
	if !utf8.ValidString(password) || len([]byte(password)) > PasswordMaxBytes {
		return fmt.Errorf("password must contain at most %d UTF-8 bytes", PasswordMaxBytes)
	}
	return nil
}

func ValidateBcryptConfiguration(environment string) error {
	costValue := os.Getenv("BCRYPT_COST")
	if costValue == "" {
		return nil
	}
	cost, err := strconv.Atoi(costValue)
	if err != nil || cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		return fmt.Errorf("BCRYPT_COST must be an integer between %d and %d", bcrypt.MinCost, bcrypt.MaxCost)
	}
	if environment == "production" && cost < ProductionBcryptMinimumCost {
		return fmt.Errorf("production BCRYPT_COST must be at least %d", ProductionBcryptMinimumCost)
	}
	return nil
}
