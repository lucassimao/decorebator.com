package mail

import (
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"

	_ "embed"
)

// MailService handles email operations with dependency injection
//
//revive:disable:exported
type MailService struct {
	db       *pgxpool.Pool
	userRepo *repository.UserRepository
}

// NewMailService creates a new mail service with injected dependencies
func NewMailService(db *pgxpool.Pool) *MailService {
	return &MailService{
		db:       db,
		userRepo: &repository.UserRepository{Db: db},
	}
}

// shouldSendEmails checks if emails should be sent based on environment variables
// Returns false if DISABLE_EMAILS=true or ENV is not production (for some email types)
func (m *MailService) shouldSendEmails() bool {
	// Check DISABLE_EMAILS flag (used for load testing and development)
	if disableEmails, err := strconv.ParseBool(os.Getenv("DISABLE_EMAILS")); err == nil && disableEmails {
		return false
	}
	return true
}

// AddContactToList adds a user to the SendGrid contact list
// https://www.twilio.com/docs/sendgrid/api-reference/contacts/add-or-update-a-contact
func (m *MailService) AddContactToList(user *model.User) {
	logger := common.Logger.With("func", "AddContactToList", "user", user.ID)

	if !m.shouldSendEmails() {
		logger.Debug("emails disabled via DISABLE_EMAILS flag. skipping")
		return
	}

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

// SendResetPasswordEmail sends a password reset email to the specified address
func (m *MailService) SendResetPasswordEmail(email string) error {
	if !m.shouldSendEmails() {
		common.Logger.Debug("emails disabled via DISABLE_EMAILS flag. skipping reset password email", "email", email)
		return nil
	}

	result, err := m.userRepo.Find(context.Background(), repository.FindUserArgs{Email: &email})

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

// SendSubscriptionActivatedEmail sends a welcome email when subscription is activated
func (m *MailService) SendSubscriptionActivatedEmail(user *model.User, data SubscriptionEmailData) error {
	logger := common.Logger.With("func", "SendSubscriptionActivatedEmail", "user", user.ID)

	if !m.shouldSendEmails() {
		logger.Debug("emails disabled via DISABLE_EMAILS flag. skipping subscription activated email")
		return nil
	}

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

// SendSubscriptionRenewedEmail sends a confirmation email when subscription is renewed
func (m *MailService) SendSubscriptionRenewedEmail(user *model.User, data SubscriptionEmailData) error {
	logger := common.Logger.With("func", "SendSubscriptionRenewedEmail", "user", user.ID)

	if !m.shouldSendEmails() {
		logger.Debug("emails disabled via DISABLE_EMAILS flag. skipping subscription renewed email")
		return nil
	}

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

// SendRenewalReminderEmail sends a reminder before subscription renewal
func (m *MailService) SendRenewalReminderEmail(user *model.User, data SubscriptionEmailData) error {
	logger := common.Logger.With("func", "SendRenewalReminderEmail", "user", user.ID)

	if !m.shouldSendEmails() {
		logger.Debug("emails disabled via DISABLE_EMAILS flag. skipping renewal reminder email")
		return nil
	}

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

// SendSubscriptionCancelledEmail sends a confirmation email when subscription is canceled
func (m *MailService) SendSubscriptionCancelledEmail(user *model.User, data SubscriptionEmailData) error {
	logger := common.Logger.With("func", "SendSubscriptionCancelledEmail", "user", user.ID)

	if !m.shouldSendEmails() {
		logger.Debug("emails disabled via DISABLE_EMAILS flag. skipping subscription canceled email")
		return nil
	}

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

// SendPaymentFailedEmail sends a notification when payment fails
func (m *MailService) SendPaymentFailedEmail(user *model.User, data SubscriptionEmailData) error {
	logger := common.Logger.With("func", "SendPaymentFailedEmail", "user", user.ID)

	if !m.shouldSendEmails() {
		logger.Debug("emails disabled via DISABLE_EMAILS flag. skipping payment failed email")
		return nil
	}

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

//go:embed welcome.html
var welcomeEmailTemplate string

// SendWelcomeEmail sends a welcome email to new users
func (m *MailService) SendWelcomeEmail(email string) error {
	logger := common.Logger.With("func", "SendWelcomeEmail", "email", email)

	if !m.shouldSendEmails() {
		logger.Debug("emails disabled via DISABLE_EMAILS flag. skipping welcome email")
		return nil
	}

	if os.Getenv("ENV") != "production" {
		logger.Debug("non-production environment. skipping")
		return nil
	}

	result, err := m.userRepo.Find(context.Background(), repository.FindUserArgs{Email: &email})

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
