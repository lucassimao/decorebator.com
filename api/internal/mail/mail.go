package mail

import (
	"context"
	"fmt"
	"html/template"
	"os"
	"strconv"
	"strings"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/resend/resend-go/v3"

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

//go:embed reset_password.html
var resetPasswordEmailTemplate string

// SendResetPasswordEmail sends a password reset email to the specified address
func (m *MailService) SendResetPasswordEmail(ctx context.Context, email string, deliveryKeys ...string) error {
	logger := common.Logger.With("func", "SendResetPasswordEmail")

	if !m.shouldSendEmails() {
		logger.Warn("emails disabled via DISABLE_EMAILS flag. skipping reset password email")
		return nil
	}

	canonicalEmail, err := common.NormalizeEmail(email)
	if err != nil {
		logger.Debug("discarding reset request without a deliverable account")
		return nil
	}
	result, err := m.userRepo.Find(ctx, repository.FindUserArgs{Email: &canonicalEmail})

	if err != nil || len(result) != 1 {
		if err != nil {
			return fmt.Errorf("find reset recipient: %w", err)
		}
		logger.Debug("discarding reset request without a deliverable account")
		return nil
	}

	user := result[0]
	logger = logger.With("user_id", user.ID)
	encryptedPayload, err := IssueResetPasswordToken(ctx, m.db, user.ID, deliveryKeys...)
	if err != nil {
		logger.Error("failed to create reset password token", "error", err)
		return err
	}

	tmpl, err := template.New("email").Parse(resetPasswordEmailTemplate)
	if err != nil {
		logger.Error("failed to parse reset password template", "error", err)
		return err
	}

	data := map[string]string{
		"FirstName": user.FirstName,
		"ResetLink": fmt.Sprintf("https://decorebator.com/reset-password?token=%s", encryptedPayload),
	}
	var sb strings.Builder
	err = tmpl.Execute(&sb, data)
	if err != nil {
		logger.Error("failed to execute reset password template", "error", err)
		return err
	}

	subject := "Reset Your Password for Decorebator"
	fullName := fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	plainTextContent := sb.String()
	htmlContent := sb.String()

	client, err := m.newResendClient()
	if err != nil {
		logger.Error("failed to create Resend client", "error", err)
		return err
	}

	request := &resend.SendEmailRequest{
		From:    "Decorebator <support@decorebator.com>",
		To:      []string{fmt.Sprintf("%s <%s>", fullName, user.Email)},
		Subject: subject,
		Text:    plainTextContent,
		Html:    htmlContent,
	}
	if len(deliveryKeys) > 0 {
		_, err = client.Emails.SendWithOptions(ctx, request, &resend.SendEmailOptions{
			IdempotencyKey: "reset-password/" + deliveryKeys[0],
		})
	} else {
		_, err = client.Emails.SendWithContext(ctx, request)
	}
	if err != nil {
		logger.Error("failed to send reset password email", "error", err)
		return err
	}

	logger.Info("reset password email sent successfully")
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

//go:embed welcome_en.html
var welcomeEmailTemplateEN string

//go:embed welcome_es.html
var welcomeEmailTemplateES string

//go:embed welcome_fr.html
var welcomeEmailTemplateFR string

//go:embed welcome_de.html
var welcomeEmailTemplateDE string

//go:embed welcome_it.html
var welcomeEmailTemplateIT string

//go:embed welcome_pt.html
var welcomeEmailTemplatePT string

//go:embed welcome_ja.html
var welcomeEmailTemplateJA string

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
func (m *MailService) SendSubscriptionActivatedEmail(user *model.User, data SubscriptionEmailData) error { //nolint:dupl // Removed with legacy billing emails.
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

	subject := "Welcome to Decorebator Premium! 🎉"
	fullName := fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	plainTextContent := "Your Decorebator subscription is now active!"
	htmlContent := sb.String()

	client, err := m.newResendClient()
	if err != nil {
		logger.Error("failed to create Resend client", "error", err)
		return err
	}

	_, err = client.Emails.Send(&resend.SendEmailRequest{
		From:    "Decorebator <support@decorebator.com>",
		To:      []string{fmt.Sprintf("%s <%s>", fullName, user.Email)},
		Subject: subject,
		Text:    plainTextContent,
		Html:    htmlContent,
	})
	if err != nil {
		logger.Error("failed to send email", "error", err)
		return err
	}

	logger.Info("subscription activated email sent successfully")
	return nil
}

// SendSubscriptionRenewedEmail sends a confirmation email when subscription is renewed
func (m *MailService) SendSubscriptionRenewedEmail(user *model.User, data SubscriptionEmailData) error { //nolint:dupl // Removed with legacy billing emails.
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

	subject := "Subscription Renewed Successfully ✅"
	fullName := fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	plainTextContent := "Your Decorebator subscription has been renewed."
	htmlContent := sb.String()

	client, err := m.newResendClient()
	if err != nil {
		logger.Error("failed to create Resend client", "error", err)
		return err
	}

	_, err = client.Emails.Send(&resend.SendEmailRequest{
		From:    "Decorebator <support@decorebator.com>",
		To:      []string{fmt.Sprintf("%s <%s>", fullName, user.Email)},
		Subject: subject,
		Text:    plainTextContent,
		Html:    htmlContent,
	})
	if err != nil {
		logger.Error("failed to send email", "error", err)
		return err
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

	subject := "Subscription Renewal Reminder ⏰"
	fullName := fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	plainTextContent := fmt.Sprintf("Your Decorebator subscription will renew on %s", templateData["RenewalDate"])
	htmlContent := sb.String()

	client, err := m.newResendClient()
	if err != nil {
		logger.Error("failed to create Resend client", "error", err)
		return err
	}

	_, err = client.Emails.Send(&resend.SendEmailRequest{
		From:    "Decorebator <support@decorebator.com>",
		To:      []string{fmt.Sprintf("%s <%s>", fullName, user.Email)},
		Subject: subject,
		Text:    plainTextContent,
		Html:    htmlContent,
	})
	if err != nil {
		logger.Error("failed to send email", "error", err)
		return err
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

	subject := "Subscription Canceled"
	fullName := fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	plainTextContent := fmt.Sprintf("Your subscription has been cancelled. You'll have access until %s", templateData["AccessUntil"]) //nolint:misspell // Preserve legacy customer-facing copy.
	htmlContent := sb.String()

	client, err := m.newResendClient()
	if err != nil {
		logger.Error("failed to create Resend client", "error", err)
		return err
	}

	_, err = client.Emails.Send(&resend.SendEmailRequest{
		From:    "Decorebator <support@decorebator.com>",
		To:      []string{fmt.Sprintf("%s <%s>", fullName, user.Email)},
		Subject: subject,
		Text:    plainTextContent,
		Html:    htmlContent,
	})
	if err != nil {
		logger.Error("failed to send email", "error", err)
		return err
	}

	logger.Info("subscription cancelled email sent successfully") //nolint:misspell // Preserve the existing observability event text.
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

	subject := "Payment Failed - Action Required ⚠️"
	fullName := fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	plainTextContent := "Your payment failed. Please update your payment method to avoid service interruption."
	htmlContent := sb.String()

	client, err := m.newResendClient()
	if err != nil {
		logger.Error("failed to create Resend client", "error", err)
		return err
	}

	_, err = client.Emails.Send(&resend.SendEmailRequest{
		From:    "Decorebator <support@decorebator.com>",
		To:      []string{fmt.Sprintf("%s <%s>", fullName, user.Email)},
		Subject: subject,
		Text:    plainTextContent,
		Html:    htmlContent,
	})
	if err != nil {
		logger.Error("failed to send email", "error", err)
		return err
	}

	logger.Info("payment failed email sent successfully")
	return nil
}

// SendWelcomeEmail sends a welcome email to new users
func (m *MailService) SendWelcomeEmail(ctx context.Context, email string) error {
	logger := common.Logger.With("func", "SendWelcomeEmail")

	if !m.shouldSendEmails() {
		logger.Warn("emails disabled via DISABLE_EMAILS flag. skipping welcome email")
		return nil
	}

	if os.Getenv("ENV") != "production" {
		logger.Info("non-production environment. skipping", "env", os.Getenv("ENV"))
		return nil
	}

	canonicalEmail, err := common.NormalizeEmail(email)
	if err != nil {
		return common.ErrInvalidEmailAddress
	}
	result, err := m.userRepo.Find(ctx, repository.FindUserArgs{Email: &canonicalEmail})

	if err != nil || len(result) != 1 {
		logger.Warn("user not found for welcome email", "error", err, "matchCount", len(result))
		return fmt.Errorf("no user found")
	}

	user := result[0]
	logger = logger.With("user_id", user.ID)

	templateSource := resolveWelcomeTemplate(user.PreferredLanguage)
	subject := resolveWelcomeSubject(user.PreferredLanguage)

	tmpl, err := template.New("email").Parse(templateSource)
	if err != nil {
		logger.Error("failed to parse welcome email template", "error", err)
		return err
	}

	data := map[string]string{
		"FirstName": user.FirstName,
	}
	var sb strings.Builder
	err = tmpl.Execute(&sb, data)
	if err != nil {
		logger.Error("failed to execute welcome email template", "error", err)
		return err
	}

	fullName := fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	plainTextContent := sb.String()
	htmlContent := sb.String()

	client, err := m.newResendClient()
	if err != nil {
		logger.Error("failed to create Resend client", "error", err)
		return err
	}

	_, err = client.Emails.Send(&resend.SendEmailRequest{
		From:    "Decorebator <support@decorebator.com>",
		To:      []string{fmt.Sprintf("%s <%s>", fullName, user.Email)},
		Subject: subject,
		Text:    plainTextContent,
		Html:    htmlContent,
	})
	if err != nil {
		logger.Error("failed to send welcome email", "error", err)
		return err
	}

	logger.Info("welcome email sent successfully")
	return nil
}

func (m *MailService) newResendClient() (*resend.Client, error) {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("RESEND_API_KEY is required")
	}
	return resend.NewClient(apiKey), nil
}

func normalizeLanguageTag(lang *string) string {
	if lang == nil {
		return "en"
	}
	normalized := strings.ToLower(strings.TrimSpace(*lang))
	if normalized == "" {
		return "en"
	}
	normalized = strings.ReplaceAll(normalized, "_", "-")
	return normalized
}

func resolveWelcomeTemplate(lang *string) string {
	templates := map[string]string{
		"en": welcomeEmailTemplateEN,
		"es": welcomeEmailTemplateES,
		"fr": welcomeEmailTemplateFR,
		"de": welcomeEmailTemplateDE,
		"it": welcomeEmailTemplateIT,
		"pt": welcomeEmailTemplatePT,
		"ja": welcomeEmailTemplateJA,
	}

	tag := normalizeLanguageTag(lang)
	if tmpl, ok := templates[tag]; ok {
		return tmpl
	}
	if idx := strings.Index(tag, "-"); idx > 0 {
		if tmpl, ok := templates[tag[:idx]]; ok {
			return tmpl
		}
	}
	return templates["en"]
}

func resolveWelcomeSubject(lang *string) string {
	subjects := map[string]string{
		"en": "Welcome to Decorebator!",
		"es": "¡Bienvenido a Decorebator!",
		"fr": "Bienvenue sur Decorebator !",
		"de": "Willkommen bei Decorebator!",
		"it": "Benvenuto su Decorebator!",
		"pt": "Bem-vindo ao Decorebator!",
		"ja": "Decorebatorへようこそ！",
	}

	tag := normalizeLanguageTag(lang)
	if subject, ok := subjects[tag]; ok {
		return subject
	}
	if idx := strings.Index(tag, "-"); idx > 0 {
		if subject, ok := subjects[tag[:idx]]; ok {
			return subject
		}
	}
	return subjects["en"]
}
