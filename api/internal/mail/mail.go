package mail

import (
	"fmt"
	"html/template"
	"log/slog"
	"os"
	"strings"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"

	_ "embed"
)

var userRepository *repository.UserRepository

func init() {
	db, err := common.GetDBConnection()
	if err != nil {
		common.Logger.Error("failed to open db connection", "error", err)
		os.Exit(1)
	}
	userRepository = &repository.UserRepository{Db: db}
}

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
	// Status code 202 indicates that the contacts are queued for processing
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

//go:embed reset_password.html
var resetPasswordEmailTemplate string

func SendResetPasswordEmail(email string) error {

	result, err := userRepository.Find(repository.FindUserArgs{Email: &email})

	if err != nil || len(result) != 1 {
		return fmt.Errorf("no user found")
	}

	user := result[0]
	encryptedPayload, err := createResetPasswordToken(user.ID)
	if err != nil {
		return err
	}

	tmpl, err := template.New("email").Parse(resetPasswordEmailTemplate)
	if err != nil {
		return err
	}

	data := map[string]string{
		"FirstName": user.FirstName,
		"ResetLink": fmt.Sprintf("https://decorebator.com/reset-password?token=%s", encryptedPayload),
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

	if err != nil {
		return err
	}

	// 202 status meaning Accepted
	if response.StatusCode != 202 {
		return fmt.Errorf("failed to send email: %v", response.Body)
	}

	return nil
}
