package mail

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

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
	ExpiresAt time.Time `json:"expiresAt"`
}

func createResetPasswordToken(userID int64) (encryptedPayload string, err error) {
	futureTime := time.Now().Add(time.Duration(30) * time.Minute)
	payload := ResetPasswordPayload{UserID: userID, ExpiresAt: futureTime}
	encodedPayload, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	key := []byte(os.Getenv("RESET_PASSWORD_PRIVATE_KEY"))
	encryptedPayload, err = encryptAES(key, string(encodedPayload))
	if err != nil {
		return "", err
	}

	return encryptedPayload, nil
}

func ValidateResetPasswordPayload(encrypted string) (*ResetPasswordPayload, error) {
	key := []byte(os.Getenv("RESET_PASSWORD_PRIVATE_KEY"))

	decrypted, err := decryptAES(key, encrypted)
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

	return &payload, nil
}
