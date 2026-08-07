package mail

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestValidateResetPasswordConfiguration(t *testing.T) {
	tests := []struct {
		name  string
		value *string
	}{
		{name: "missing"},
		{name: "short", value: stringPointer("too-short")},
		{name: "long", value: stringPointer(strings.Repeat("x", 33))},
		{name: "whitespace", value: stringPointer(" " + strings.Repeat("x", 31))},
		{name: "placeholder", value: stringPointer("change_me_reset_key_material_1234")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.value == nil {
				t.Setenv("RESET_PASSWORD_PRIVATE_KEY", "")
			} else {
				t.Setenv("RESET_PASSWORD_PRIVATE_KEY", *test.value)
			}
			if err := ValidateResetPasswordConfiguration(); err == nil {
				t.Fatal("expected invalid reset-password configuration")
			}
		})
	}

	t.Setenv("RESET_PASSWORD_PRIVATE_KEY", strings.Repeat("k", resetPasswordKeyBytes))
	if err := ValidateResetPasswordConfiguration(); err != nil {
		t.Fatalf("expected valid reset-password configuration: %v", err)
	}
}

func TestResetTokenEnvelopeSeparatesNewAndLegacyBinaries(t *testing.T) {
	key := strings.Repeat("k", resetPasswordKeyBytes)
	t.Setenv("RESET_PASSWORD_PRIVATE_KEY", key)
	issued, err := CreateResetPasswordToken(42)
	if err != nil {
		t.Fatalf("create v2 token: %v", err)
	}
	if !strings.HasPrefix(issued.Encrypted, resetPasswordV2Prefix) {
		t.Fatalf("new reset token lacks version prefix: %q", issued.Encrypted)
	}
	if _, decryptErr := decryptAES([]byte(key), issued.Encrypted); decryptErr == nil {
		t.Fatal("pre-AUTH-3 decryptor must reject the non-hex v2 envelope")
	}
	v2Payload, err := ValidateResetPasswordPayload(issued.Encrypted)
	if err != nil || v2Payload.Legacy || v2Payload.TokenID == "" {
		t.Fatalf("unexpected v2 payload: %#v err=%v", v2Payload, err)
	}

	legacyJSON, err := json.Marshal(ResetPasswordPayload{
		UserID: 42, ExpiresAt: time.Now().UTC().Add(time.Minute),
	})
	if err != nil {
		t.Fatalf("marshal legacy payload: %v", err)
	}
	legacyToken, err := encryptAES([]byte(key), string(legacyJSON))
	if err != nil {
		t.Fatalf("create legacy token: %v", err)
	}
	legacyPayload, err := ValidateResetPasswordPayload(legacyToken)
	if err != nil || !legacyPayload.Legacy || legacyPayload.TokenID != "" {
		t.Fatalf("unexpected legacy payload: %#v err=%v", legacyPayload, err)
	}
}

func stringPointer(value string) *string { return &value }

func TestValidateActivationDeliveryConfiguration(t *testing.T) {
	t.Run("nonproduction may disable delivery", func(t *testing.T) {
		t.Setenv("DISABLE_EMAILS", "true")
		t.Setenv("RESEND_API_KEY", "")
		if err := ValidateActivationDeliveryConfiguration("development"); err != nil {
			t.Fatalf("unexpected nonproduction error: %v", err)
		}
	})

	for _, test := range []struct {
		name     string
		disabled string
		apiKey   string
	}{
		{name: "disabled", disabled: "true", apiKey: "production-key"},
		{name: "invalid flag", disabled: "sometimes", apiKey: "production-key"},
		{name: "missing key", disabled: "false"},
		{name: "placeholder key", disabled: "false", apiKey: "your_resend_api_key"},
		{name: "old accepted placeholder", disabled: "false", apiKey: "production-delivery-key"},
		{name: "documented placeholder", disabled: "false", apiKey: "re_xxxxxxxxx"},
		{name: "test key", disabled: "false", apiKey: "re_test_3fK8nP2vQ7mZ1sL9"},
		{name: "invalid character", disabled: "false", apiKey: "re_live secret"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("DISABLE_EMAILS", test.disabled)
			t.Setenv("RESEND_API_KEY", test.apiKey)
			if err := ValidateActivationDeliveryConfiguration("production"); err == nil {
				t.Fatal("expected production delivery configuration error")
			}
		})
	}

	t.Setenv("DISABLE_EMAILS", "false")
	t.Setenv("RESEND_API_KEY", "re_3fK8nP2vQ7mZ1sL9aC4dR6tY")
	if err := ValidateActivationDeliveryConfiguration("production"); err != nil {
		t.Fatalf("expected valid production delivery configuration: %v", err)
	}
}
