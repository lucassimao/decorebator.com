package service

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/oauth2"
)

type GoogleOAuthTokenSource struct {
	source oauth2.TokenSource
}

func NewGoogleOAuthTokenSource(source oauth2.TokenSource) (*GoogleOAuthTokenSource, error) {
	if source == nil {
		return nil, fmt.Errorf("Google OAuth token source is required")
	}
	return &GoogleOAuthTokenSource{source: oauth2.ReuseTokenSource(nil, source)}, nil
}

func (s *GoogleOAuthTokenSource) AccessToken(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", &GoogleAccessTokenError{Retryable: true, Cause: err}
	}
	token, err := s.source.Token()
	if err != nil {
		return "", &GoogleAccessTokenError{Retryable: true, Cause: err}
	}
	if token == nil || strings.TrimSpace(token.AccessToken) == "" {
		return "", &GoogleAccessTokenError{Retryable: true, Cause: fmt.Errorf("Google OAuth token is empty")}
	}
	return token.AccessToken, nil
}
