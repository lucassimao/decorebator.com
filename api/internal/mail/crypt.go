package mail

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	resetPasswordKeyBytes = 32
	resetPasswordV2Prefix = "v2."
)

var ErrAuthHardeningWritesDisabled = fmt.Errorf("AUTH-3 writes are disabled for rollback")

var ErrResetDeliverySuperseded = fmt.Errorf("reset delivery was superseded")

// ValidateResetPasswordConfiguration rejects missing, malformed, and obvious
// placeholder encryption material before reset links can be requested.
func ValidateResetPasswordConfiguration() error {
	key, ok := os.LookupEnv("RESET_PASSWORD_PRIVATE_KEY")
	if !ok || key == "" {
		return fmt.Errorf("RESET_PASSWORD_PRIVATE_KEY is required")
	}
	if strings.TrimSpace(key) != key {
		return fmt.Errorf("RESET_PASSWORD_PRIVATE_KEY must not contain surrounding whitespace")
	}
	if len([]byte(key)) != resetPasswordKeyBytes {
		return fmt.Errorf("RESET_PASSWORD_PRIVATE_KEY must contain exactly %d bytes", resetPasswordKeyBytes)
	}
	lowerKey := strings.ToLower(key)
	for _, marker := range []string{"your_", "change_me", "changeme", "secret_here", "placeholder"} {
		if strings.Contains(lowerKey, marker) {
			return fmt.Errorf("RESET_PASSWORD_PRIVATE_KEY must not use placeholder encryption material")
		}
	}
	return nil
}

func ValidateActivationDeliveryConfiguration(environment string) error {
	if environment != "production" {
		return nil
	}
	disabledRaw := strings.TrimSpace(os.Getenv("DISABLE_EMAILS"))
	disabled, err := strconv.ParseBool(disabledRaw)
	if err != nil {
		return fmt.Errorf("DISABLE_EMAILS must be a valid boolean in production")
	}
	if disabled {
		return fmt.Errorf("DISABLE_EMAILS must be false in production")
	}
	apiKey := strings.TrimSpace(os.Getenv("RESEND_API_KEY"))
	if apiKey == "" {
		return fmt.Errorf("RESEND_API_KEY is required in production")
	}
	if !hasResendAPIKeyShape(apiKey) {
		return fmt.Errorf("RESEND_API_KEY must use the documented re_ token format")
	}
	lowerKey := strings.ToLower(apiKey)
	for _, marker := range []string{"your_", "change_me", "changeme", "placeholder", "dummy", "example", "test", "production"} {
		if strings.Contains(lowerKey, marker) {
			return fmt.Errorf("RESEND_API_KEY must not use placeholder delivery credentials")
		}
	}
	if strings.Trim(strings.TrimPrefix(lowerKey, "re_"), "x") == "" {
		return fmt.Errorf("RESEND_API_KEY must not use placeholder delivery credentials")
	}
	return nil
}

func hasResendAPIKeyShape(apiKey string) bool {
	if !strings.HasPrefix(apiKey, "re_") || len(apiKey) <= len("re_") {
		return false
	}
	for _, char := range strings.TrimPrefix(apiKey, "re_") {
		if (char < 'a' || char > 'z') && (char < 'A' || char > 'Z') &&
			(char < '0' || char > '9') && char != '_' && char != '-' {
			return false
		}
	}
	return true
}

func encryptAES(key []byte, plaintext string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(ciphertext), nil
}

func decryptAES(key []byte, encrypted string) (string, error) {
	ciphertext, err := hex.DecodeString(encrypted)
	if err != nil {
		return "", fmt.Errorf("invalid hex encoding: %v", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short: expected at least %d bytes, got %d", nonceSize, len(ciphertext))
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

type ResetPasswordPayload struct {
	UserID    int64     `json:"userId"`
	TokenID   string    `json:"tokenId"`
	ExpiresAt time.Time `json:"expiresAt"`
	Legacy    bool      `json:"-"`
}

type IssuedResetPasswordToken struct {
	Encrypted string
	TokenID   string
	ExpiresAt time.Time
}

func CreateResetPasswordToken(userID int64) (IssuedResetPasswordToken, error) {
	var random [32]byte
	if _, err := rand.Read(random[:]); err != nil {
		return IssuedResetPasswordToken{}, fmt.Errorf("generate reset token ID: %w", err)
	}
	expiresAt := time.Now().UTC().Add(30 * time.Minute)
	payload := ResetPasswordPayload{
		UserID: userID, TokenID: base64.RawURLEncoding.EncodeToString(random[:]), ExpiresAt: expiresAt,
	}
	encodedPayload, err := json.Marshal(payload)
	if err != nil {
		return IssuedResetPasswordToken{}, err
	}

	key := []byte(os.Getenv("RESET_PASSWORD_PRIVATE_KEY"))
	encryptedPayload, err := encryptAES(key, string(encodedPayload))
	if err != nil {
		return IssuedResetPasswordToken{}, err
	}

	return IssuedResetPasswordToken{
		Encrypted: resetPasswordV2Prefix + encryptedPayload, TokenID: payload.TokenID, ExpiresAt: expiresAt,
	}, nil
}

func IssueResetPasswordToken(
	ctx context.Context,
	db *pgxpool.Pool,
	userID int64,
	deliveryKeys ...string,
) (string, error) {
	if db == nil || userID <= 0 {
		return "", fmt.Errorf("reset-token database and positive user ID are required")
	}
	issued, err := CreateResetPasswordToken(userID)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256([]byte(issued.TokenID))
	deliveryKey := issued.TokenID
	if len(deliveryKeys) > 0 {
		deliveryKey = strings.TrimSpace(deliveryKeys[0])
		if deliveryKey == "" {
			return "", fmt.Errorf("reset-token delivery key must not be empty")
		}
	}
	encryptedToken := issued.Encrypted
	storedDeliveryToken, err := encryptAES([]byte(os.Getenv("RESET_PASSWORD_PRIVATE_KEY")), encryptedToken)
	if err != nil {
		return "", fmt.Errorf("protect reset delivery token: %w", err)
	}
	err = pgx.BeginFunc(ctx, db, func(tx pgx.Tx) error {
		var writesEnabled bool
		if stateErr := tx.QueryRow(ctx, `
			SELECT writes_enabled
			FROM auth_hardening_rollout_state
			WHERE singleton=TRUE
			FOR SHARE
		`).Scan(&writesEnabled); stateErr != nil {
			return fmt.Errorf("read auth-hardening rollout state: %w", stateErr)
		}
		if !writesEnabled {
			return ErrAuthHardeningWritesDisabled
		}
		var lockedUserID int64
		if lockErr := tx.QueryRow(ctx, `SELECT id FROM users WHERE id=$1 FOR UPDATE`, userID).Scan(&lockedUserID); lockErr != nil {
			return lockErr
		}
		if len(deliveryKeys) > 0 {
			var consumedAt *time.Time
			var existingExpiresAt time.Time
			existingErr := tx.QueryRow(ctx, `
				SELECT delivery_token_ciphertext,consumed_at,expires_at
				FROM password_reset_tokens
				WHERE delivery_key=$1 AND user_id=$2
			`, deliveryKey, userID).Scan(&storedDeliveryToken, &consumedAt, &existingExpiresAt)
			if existingErr == nil {
				if consumedAt != nil || !existingExpiresAt.After(time.Now().UTC()) {
					return ErrResetDeliverySuperseded
				}
				var decryptErr error
				encryptedToken, decryptErr = decryptAES(
					[]byte(os.Getenv("RESET_PASSWORD_PRIVATE_KEY")), storedDeliveryToken,
				)
				if decryptErr != nil {
					return fmt.Errorf("recover reset delivery token: %w", decryptErr)
				}
				return nil
			}
			if !errors.Is(existingErr, pgx.ErrNoRows) {
				return existingErr
			}
		}
		if _, invalidateErr := tx.Exec(ctx, `
			UPDATE password_reset_tokens
			SET consumed_at=COALESCE(consumed_at,NOW())
			WHERE user_id=$1 AND consumed_at IS NULL
		`, userID); invalidateErr != nil {
			return invalidateErr
		}
		_, insertErr := tx.Exec(ctx, `
			INSERT INTO password_reset_tokens (
				token_hash,user_id,delivery_key,delivery_token_ciphertext,expires_at
			)
			VALUES ($1,$2,$3,$4,$5)
		`, hash[:], userID, deliveryKey, storedDeliveryToken, issued.ExpiresAt)
		return insertErr
	})
	if err != nil {
		return "", err
	}
	return encryptedToken, nil
}

func ValidateResetPasswordPayload(encrypted string) (*ResetPasswordPayload, error) {
	key := []byte(os.Getenv("RESET_PASSWORD_PRIVATE_KEY"))
	ciphertext, isV2 := strings.CutPrefix(encrypted, resetPasswordV2Prefix)
	if !isV2 {
		// Historical tokens were emitted by hex.EncodeToString, so lowercase is
		// their only canonical representation. Accepting case variants would let
		// one ciphertext produce multiple hashes in the legacy-consumption table.
		if encrypted != strings.ToLower(encrypted) {
			return nil, fmt.Errorf("legacy reset token is not canonical lowercase hex")
		}
		ciphertext = encrypted
	}

	decrypted, err := decryptAES(key, ciphertext)
	if err != nil {
		return nil, err
	}

	var payload ResetPasswordPayload
	err = json.Unmarshal([]byte(decrypted), &payload)
	if err != nil {
		return nil, err
	}

	if time.Now().After(payload.ExpiresAt) {
		return nil, fmt.Errorf("token expired")
	}
	if payload.UserID <= 0 || (isV2 && payload.TokenID == "") {
		return nil, fmt.Errorf("invalid reset token payload")
	}
	payload.Legacy = !isV2

	return &payload, nil
}
