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
	ciphertext, _ := hex.DecodeString(encrypted)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

type ResetPasswordPayload struct {
	UserId    int64     `json:"userId"`
	ExpiresAt time.Time `json:"expiresAt"`
}

func createResetPasswordToken(userId int64) (encryptedPayload string, err error) {

	futureTime := time.Now().Add(time.Duration(30) * time.Minute)
	payload := ResetPasswordPayload{UserId: userId, ExpiresAt: futureTime}
	encodedPayload, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	key := []byte(os.Getenv("RESET_PASSWORD_PRIVATE_KEY"))
	encryptedPayload, err = encryptAES([]byte(key), string(encodedPayload))
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
