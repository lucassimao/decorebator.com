package mail

import (
	"fmt"
	"log/slog"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"github.com/sendgrid/sendgrid-go"
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

func ResetPassword(user *model.User) error {
	logger := common.Logger.With("func", "ResetPassword", "user", user.ID)

	// Usage
	key := []byte("your-secret-key-must-be-32-bytes!")
	encrypted, err := encryptAES(key, `{userId:, valid: }`)
	if err != nil {
		panic(err)
	}
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
