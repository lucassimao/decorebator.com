package common

import (
	"errors"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var ErrInvalidEmailAddress = errors.New("invalid email address")

const asciiEmailWhitespace = " \t\n\r\f\v"

var emailIdentityValidator = validator.New()

var asciiEmailIdentityPattern = regexp.MustCompile("^[a-z0-9!#$%&'*+/=?^_`{|}~-]+(\\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*@([a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?\\.)+[a-z]{2,63}$")

// NormalizeEmail defines the application's identity contract. Email identity
// is ASCII-only, trims surrounding ASCII whitespace, preserves plus aliases,
// and folds all letters to lowercase. Restricting identity to ASCII keeps Go
// and PostgreSQL case folding byte-for-byte equivalent.
func NormalizeEmail(raw string) (string, error) {
	trimmed := strings.Trim(raw, asciiEmailWhitespace)
	if trimmed == "" || len(trimmed) > 254 {
		return "", ErrInvalidEmailAddress
	}
	for _, character := range trimmed {
		if character > 127 {
			return "", ErrInvalidEmailAddress
		}
	}
	canonical := strings.ToLower(trimmed)
	localPart, _, found := strings.Cut(canonical, "@")
	if !found || len(localPart) > 64 || !asciiEmailIdentityPattern.MatchString(canonical) {
		return "", ErrInvalidEmailAddress
	}
	// Use the same validator that guarded the historical HTTP signup/login
	// boundary so existing accepted ASCII syntax is not silently narrowed.
	if err := emailIdentityValidator.Var(canonical, "required,email"); err != nil {
		return "", ErrInvalidEmailAddress
	}
	return canonical, nil
}
