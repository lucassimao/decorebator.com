package mail

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"math/rand"
	"strings"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"

	_ "embed"
)

// https://www.twilio.com/docs/sendgrid/api-reference/contacts/add-or-update-a-contact
func AddContactToList(user *model.User) {
	logger := common.Logger.With("func", "AddContactToList", "user", user.ID)

	if common.Env.Env != common.Production {
		logger.Debug("non-production environment. skipping")
		return
	}

	request := sendgrid.GetRequest(common.Env.SendGridApiKey, "/v3/marketing/contacts", "")
	request.Method = "PUT"
	request.Body = []byte(fmt.Sprintf(`
	{
		"contacts":[
			{
				"email": "%s",
				"external_id": "%v",
				"first_name": "%s",
				"last_name": "%s"
			}
		]
	}
	`, user.Email, user.ID, user.FirstName, user.LastName))

	response, err := sendgrid.API(request)
	if err != nil || response.StatusCode != 202 {
		attrs := []any{}
		if err != nil {
			attrs = append(attrs, slog.String("error", err.Error()))
		} else {
			attrs = append(attrs, slog.Int("StatusCode", response.StatusCode))
			attrs = append(attrs, slog.String("Body", response.Body))
		}
		logger.Error("failed to add user to sendgrid's contact list", attrs...)
	}
}

type ResetPasswordPayload struct {
	UserId    int64     `json:"userId"`
	ExpiresAt time.Time `json:"expiresAt"`
}

//go:embed reset_password.html
var resetPasswordEmailTemplate string

func ResetPassword(user *model.User) error {
	logger := common.Logger.With("func", "ResetPassword", "user", user.ID)
	key := []byte(common.Env.ResetPasswordPrivateKey)

	rand.Seed(time.Now().UnixNano())
	// Intn(10) gives a number between 0 and 9, so +1 makes it between 1 and 10
	randomNum := rand.Intn(10) + 1

	now := time.Now()
	// between 31 and 40 min from now
	futureTime := now.Add(time.Duration(30+randomNum) * time.Minute)
	payload := ResetPasswordPayload{UserId: user.ID, ExpiresAt: futureTime}
	encodedPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	encryptedPayload, err := encryptAES([]byte(key), string(encodedPayload))
	if err != nil {
		return err
	}

	tmpl, err := template.New("email").Parse(resetPasswordEmailTemplate)
	if err != nil {
		return err
	}

	data := map[string]string{
		"FirstName": user.FirstName,
		"ResetLink": fmt.Sprintf("https://decorebator.com/reset?token=%s", encryptedPayload),
	}
	var sb strings.Builder
	err = tmpl.Execute(&sb, data)
	if err != nil {
		return err
	}

	from := mail.NewEmail("Decorebator", "support@decorebator.com")
	subject := "Reset Your Password for Decorebator"
	fullName := fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	to := mail.NewEmail(fullName, user.Email)

	plainTextContent := sb.String()
	htmlContent := sb.String()
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(common.Env.SendGridApiKey)
	response, err := client.Send(message)

	if err != nil || response.StatusCode != 202 {
		attrs := []any{}
		if err != nil {
			attrs = append(attrs, slog.String("error", err.Error()))
		} else {
			attrs = append(attrs, slog.Int("StatusCode", response.StatusCode))
			attrs = append(attrs, slog.String("Body", response.Body))
		}
		logger.Error("failed to sent reset password email", attrs...)
	}

	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	encrypted := ""
	fmt.Println("Encrypted:", encrypted)

	// Usage
	decrypted, err := decryptAES(key, encrypted)
	if err != nil {
		panic(err)
	}
	fmt.Println("Decrypted:", decrypted)

	// Usage
	isValid, err := validateToken(decrypted)
	if !isValid {
		fmt.Println("Invalid or expired token")
	} else {
		fmt.Println("Token is valid")
	}

	return nil
}
