package mail

import (
	"fmt"
	"html/template"
	"log/slog"
	"os"
	"strings"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"

	_ "embed"
)

// Package-level functions for backward compatibility
// These are used by services that don't have access to MailService

// GetUserRepositoryForMail returns a user repository for mail operations
// This is a temporary solution to maintain backward compatibility
func GetUserRepositoryForMail() (*repository.UserRepository, error) {
	db, err := common.GetDBConnection()
	if err != nil {
		return nil, err
	}
	return &repository.UserRepository{Db: db}, nil
}

// https://www.twilio.com/docs/sendgrid/api-reference/contacts/add-or-update-a-contact
func AddContactToList(user *model.User) {
	logger := common.Logger.With("func", "AddContactToList", "user", user.ID)

	if os.Getenv("ENV") != "production" {
		logger.Debug("non-production environment. skipping")
		return
	}

	request := sendgrid.GetRequest(os.Getenv("SENDGRID_API_KEY"), "/v3/marketing/contacts", "")
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
	userRepo, err := GetUserRepositoryForMail()
	if err != nil {
		return fmt.Errorf("failed to get user repository: %w", err)
	}

	result, err := userRepo.Find(repository.FindUserArgs{Email: &email})

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
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
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

//go:embed subscription_activated.html
var subscriptionActivatedTemplate string

//go:embed subscription_renewed.html
var subscriptionRenewedTemplate string

//go:embed subscription_renewal_reminder.html
var subscriptionRenewalReminderTemplate string

//go:embed subscription_cancelled.html
var subscriptionCancelledTemplate string

//go:embed payment_failed.html
var paymentFailedTemplate string

type SubscriptionEmailData struct {
	FirstName        string
	PlanName         string
	AmountCents      int
	Currency         string
	NextBillingDate  time.Time
	RenewalDate      time.Time
	CancellationDate time.Time
	NextRetryDate    time.Time
	SubscriptionID   string
	PaymentError     string
}

func SendSubscriptionActivatedEmail(user *model.User, data SubscriptionEmailData) error {
	logger := common.Logger.With("func", "SendSubscriptionActivatedEmail", "user", user.ID)

	tmpl, err := template.New("email").Parse(subscriptionActivatedTemplate)
	if err != nil {
		logger.Error("failed to parse template", "error", err)
		return err
	}

	fmt.Println(data)
	// Format template data
	templateData := map[string]string{
		"FirstName":       user.FirstName,
		"PlanName":        data.PlanName,
		"Amount":          fmt.Sprintf("$%.2f", float64(data.AmountCents)/100),
		"NextBillingDate": data.NextBillingDate.Format("January 2, 2006"),
	}

	var sb strings.Builder
	err = tmpl.Execute(&sb, templateData)
	if err != nil {
		logger.Error("failed to execute template", "error", err)
		return err
	}

	from := mail.NewEmail("Decorebator", "support@decorebator.com")
	subject := "Welcome to Decorebator Premium! 🎉"
	fullName := fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	to := mail.NewEmail(fullName, user.Email)

	plainTextContent := "Your Decorebator subscription is now active!"
	htmlContent := sb.String()
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	response, err := client.Send(message)

	if err != nil {
		logger.Error("failed to send email", "error", err)
		return err
	}

	if response.StatusCode != 202 {
		logger.Error("failed to send email", "statusCode", response.StatusCode, "body", response.Body)
		return fmt.Errorf("failed to send email: %v", response.Body)
	}

	logger.Info("subscription activated email sent successfully")
	return nil
}

func SendSubscriptionRenewedEmail(user *model.User, data SubscriptionEmailData) error {
	logger := common.Logger.With("func", "SendSubscriptionRenewedEmail", "user", user.ID)

	tmpl, err := template.New("email").Parse(subscriptionRenewedTemplate)
	if err != nil {
		logger.Error("failed to parse template", "error", err)
		return err
	}

	// Format template data
	templateData := map[string]string{
		"FirstName":       user.FirstName,
		"PlanName":        data.PlanName,
		"Amount":          fmt.Sprintf("$%.2f", float64(data.AmountCents)/100),
		"NextBillingDate": data.NextBillingDate.Format("January 2, 2006"),
	}

	var sb strings.Builder
	err = tmpl.Execute(&sb, templateData)
	if err != nil {
		logger.Error("failed to execute template", "error", err)
		return err
	}

	from := mail.NewEmail("Decorebator", "support@decorebator.com")
	subject := "Subscription Renewed Successfully ✅"
	fullName := fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	to := mail.NewEmail(fullName, user.Email)

	plainTextContent := "Your Decorebator subscription has been renewed."
	htmlContent := sb.String()
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	response, err := client.Send(message)

	if err != nil {
		logger.Error("failed to send email", "error", err)
		return err
	}

	if response.StatusCode != 202 {
		logger.Error("failed to send email", "statusCode", response.StatusCode, "body", response.Body)
		return fmt.Errorf("failed to send email: %v", response.Body)
	}

	logger.Info("subscription renewed email sent successfully")
	return nil
}

func SendRenewalReminderEmail(user *model.User, data SubscriptionEmailData) error {
	logger := common.Logger.With("func", "SendRenewalReminderEmail", "user", user.ID)

	tmpl, err := template.New("email").Parse(subscriptionRenewalReminderTemplate)
	if err != nil {
		logger.Error("failed to parse template", "error", err)
		return err
	}

	// Format template data
	templateData := map[string]string{
		"FirstName":   user.FirstName,
		"PlanName":    data.PlanName,
		"Amount":      fmt.Sprintf("$%.2f", float64(data.AmountCents)/100),
		"RenewalDate": data.NextBillingDate.Format("January 2, 2006"),
	}

	var sb strings.Builder
	err = tmpl.Execute(&sb, templateData)
	if err != nil {
		logger.Error("failed to execute template", "error", err)
		return err
	}

	from := mail.NewEmail("Decorebator", "support@decorebator.com")
	subject := "Subscription Renewal Reminder ⏰"
	fullName := fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	to := mail.NewEmail(fullName, user.Email)

	plainTextContent := fmt.Sprintf("Your Decorebator subscription will renew on %s", templateData["RenewalDate"])
	htmlContent := sb.String()
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	response, err := client.Send(message)

	if err != nil {
		logger.Error("failed to send email", "error", err)
		return err
	}

	if response.StatusCode != 202 {
		logger.Error("failed to send email", "statusCode", response.StatusCode, "body", response.Body)
		return fmt.Errorf("failed to send email: %v", response.Body)
	}

	logger.Info("subscription renewal reminder email sent successfully")
	return nil
}

func SendSubscriptionCancelledEmail(user *model.User, data SubscriptionEmailData) error {
	logger := common.Logger.With("func", "SendSubscriptionCancelledEmail", "user", user.ID)

	tmpl, err := template.New("email").Parse(subscriptionCancelledTemplate)
	if err != nil {
		logger.Error("failed to parse template", "error", err)
		return err
	}

	// Format template data
	templateData := map[string]string{
		"FirstName":   user.FirstName,
		"PlanName":    data.PlanName,
		"AccessUntil": data.CancellationDate.Format("January 2, 2006"),
	}

	var sb strings.Builder
	err = tmpl.Execute(&sb, templateData)
	if err != nil {
		logger.Error("failed to execute template", "error", err)
		return err
	}

	from := mail.NewEmail("Decorebator", "support@decorebator.com")
	subject := "Subscription Canceled"
	fullName := fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	to := mail.NewEmail(fullName, user.Email)

	plainTextContent := fmt.Sprintf("Your subscription has been cancelled. You'll have access until %s", templateData["AccessUntil"])
	htmlContent := sb.String()
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	response, err := client.Send(message)

	if err != nil {
		logger.Error("failed to send email", "error", err)
		return err
	}

	if response.StatusCode != 202 {
		logger.Error("failed to send email", "statusCode", response.StatusCode, "body", response.Body)
		return fmt.Errorf("failed to send email: %v", response.Body)
	}

	logger.Info("subscription cancelled email sent successfully")
	return nil
}

func SendPaymentFailedEmail(user *model.User, data SubscriptionEmailData) error {
	logger := common.Logger.With("func", "SendPaymentFailedEmail", "user", user.ID)

	tmpl, err := template.New("email").Parse(paymentFailedTemplate)
	if err != nil {
		logger.Error("failed to parse template", "error", err)
		return err
	}

	// Format template data
	templateData := map[string]string{
		"FirstName":      user.FirstName,
		"PlanName":       data.PlanName,
		"Amount":         fmt.Sprintf("$%.2f", float64(data.AmountCents)/100),
		"AttemptDate":    data.NextRetryDate.Format("January 2, 2006"),
		"GracePeriodEnd": data.NextRetryDate.AddDate(0, 0, 7).Format("January 2, 2006"), // 7 days grace period
	}

	var sb strings.Builder
	err = tmpl.Execute(&sb, templateData)
	if err != nil {
		logger.Error("failed to execute template", "error", err)
		return err
	}

	from := mail.NewEmail("Decorebator", "support@decorebator.com")
	subject := "Payment Failed - Action Required ⚠️"
	fullName := fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	to := mail.NewEmail(fullName, user.Email)

	plainTextContent := "Your payment failed. Please update your payment method to avoid service interruption."
	htmlContent := sb.String()
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	response, err := client.Send(message)

	if err != nil {
		logger.Error("failed to send email", "error", err)
		return err
	}

	if response.StatusCode != 202 {
		logger.Error("failed to send email", "statusCode", response.StatusCode, "body", response.Body)
		return fmt.Errorf("failed to send email: %v", response.Body)
	}

	logger.Info("payment failed email sent successfully")
	return nil
}

//go:embed burst_blocked.html
var burstBlockedEmailTemplate string

//go:embed welcome.html
var welcomeEmailTemplate string

func SendBurstBlockedEmail(userID int64, email, firstName, activityType string, violations int, blockedUntil time.Time) error {
	logger := slog.With("func", "SendBurstBlockedEmail", "userId", userID, "email", email)

	if os.Getenv("ENV") != "production" {
		logger.Debug("non-production environment. skipping burst blocked email")
		return nil
	}

	tmpl, err := template.New("email").Parse(burstBlockedEmailTemplate)
	if err != nil {
		logger.Error("failed to parse burst blocked email template", "error", err)
		return err
	}

	data := map[string]interface{}{
		"FirstName":    firstName,
		"ActivityType": activityType,
		"Violations":   violations,
		"BlockedUntil": blockedUntil.Format("January 2, 2006 at 3:04 PM MST"),
	}

	var htmlBuilder strings.Builder
	err = tmpl.Execute(&htmlBuilder, data)
	if err != nil {
		logger.Error("failed to execute burst blocked email template", "error", err)
		return err
	}

	// Create plain text version
	plainTextContent := fmt.Sprintf(`Dear %s,

We've detected unusual activity on your Decorebator account that appears to be automated or scripted behavior.

Your account has been temporarily suspended due to unusual activity.

Suspension Details:
- Activity Detected: Rapid %s
- Violations Today: %d
- Account Will Be Restored: %s

This suspension is temporary and is in place to protect our service and maintain fair usage for all users. Your account will be automatically restored after 24 hours, and you'll have full access to all features again.

If you believe this is a mistake or if you were performing legitimate actions, please reply to this email with details about your usage, and we'll review your case promptly.

Best regards,
The Decorebator Team

This is an automated message sent due to unusual account activity. Please do not reply unless you need to report an issue.`,
		firstName,
		activityType,
		violations,
		blockedUntil.Format("January 2, 2006 at 3:04 PM MST"),
	)

	from := mail.NewEmail("Decorebator", "support@decorebator.com")
	subject := "Account Temporarily Suspended - Unusual Activity Detected"
	to := mail.NewEmail(firstName, email)

	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlBuilder.String())
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	response, err := client.Send(message)

	if err != nil {
		logger.Error("failed to send burst blocked email", "error", err)
		return err
	}

	if response.StatusCode > 299 {
		logger.Error("sendgrid returned error status", "statusCode", response.StatusCode, "body", response.Body)
		return fmt.Errorf("failed to send burst blocked email: %v", response.Body)
	}

	logger.Info("burst blocked email sent successfully")
	return nil
}

func SendWelcomeEmail(email string) error {
	userRepo, err := GetUserRepositoryForMail()
	if err != nil {
		return fmt.Errorf("failed to get user repository: %w", err)
	}

	result, err := userRepo.Find(repository.FindUserArgs{Email: &email})

	if err != nil || len(result) != 1 {
		return fmt.Errorf("no user found")
	}

	user := result[0]

	tmpl, err := template.New("email").Parse(welcomeEmailTemplate)
	if err != nil {
		return err
	}

	data := map[string]string{
		"FirstName": user.FirstName,
	}
	var sb strings.Builder
	err = tmpl.Execute(&sb, data)
	if err != nil {
		return err
	}

	from := mail.NewEmail("Decorebator", "support@decorebator.com")
	subject := "Welcome to Decorebator!"
	fullName := fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	to := mail.NewEmail(fullName, user.Email)

	plainTextContent := sb.String()
	htmlContent := sb.String()
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
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
