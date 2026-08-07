package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	AuthIssuer             = "decorebator-api"
	AuthAudience           = "decorebator-clients"
	AccessTokenDuration    = 15 * time.Minute
	minimumJWTSigningBytes = 32
)

type AuthConfig struct {
	SigningKey  []byte
	Environment string
	Issuer      string
	Audience    string
	AccessTTL   time.Duration
}

func LoadAuthConfig() (AuthConfig, error) {
	return LoadAuthConfigFrom(os.LookupEnv)
}

func LoadAuthConfigFrom(lookup EnvironmentLookup) (AuthConfig, error) {
	if lookup == nil {
		return AuthConfig{}, fmt.Errorf("auth environment lookup is required")
	}
	key, ok := lookup("JWT_KEY")
	if !ok || key == "" {
		return AuthConfig{}, fmt.Errorf("JWT_KEY is required")
	}
	if strings.TrimSpace(key) != key {
		return AuthConfig{}, fmt.Errorf("JWT_KEY must not contain surrounding whitespace")
	}
	if len([]byte(key)) < minimumJWTSigningBytes {
		return AuthConfig{}, fmt.Errorf("JWT_KEY must contain at least %d bytes", minimumJWTSigningBytes)
	}
	lowerKey := strings.ToLower(key)
	for _, marker := range []string{"your_jwt", "change_me", "changeme", "secret_here", "placeholder"} {
		if strings.Contains(lowerKey, marker) {
			return AuthConfig{}, fmt.Errorf("JWT_KEY must not use placeholder signing material")
		}
	}
	environment, ok := lookup("ENV")
	environment = strings.TrimSpace(environment)
	if !ok || environment == "" {
		return AuthConfig{}, fmt.Errorf("ENV is required for auth token binding")
	}
	return AuthConfig{
		SigningKey: []byte(key), Environment: environment,
		Issuer: AuthIssuer, Audience: AuthAudience, AccessTTL: AccessTokenDuration,
	}, nil
}
