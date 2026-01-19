package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"decorebator.com/internal/common"
	i18nutil "decorebator.com/internal/i18n"
	"decorebator.com/internal/repository"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

const expoPushEndpoint = "https://exp.host/--/api/v2/push/send"
const expoReceiptEndpoint = "https://exp.host/--/api/v2/push/getReceipts"

type PushNotificationService struct {
	tokenRepo    *repository.PushTokenRepository
	reminderRepo *repository.PushNotificationRepository
	eventRepo    *repository.PushNotificationEventRepository
	receiptRepo  *repository.PushReceiptRepository
	httpClient   *http.Client
	accessToken  string
}

type expoPushMessage struct {
	To             any                    `json:"to"`
	Title          string                 `json:"title,omitempty"`
	Body           string                 `json:"body,omitempty"`
	Data           map[string]interface{} `json:"data,omitempty"`
	Sound          any                    `json:"sound,omitempty"`
	TTL            *int                   `json:"ttl,omitempty"`
	Expiration     *int64                 `json:"expiration,omitempty"`
	Priority       string                 `json:"priority,omitempty"`
	Subtitle       string                 `json:"subtitle,omitempty"`
	Badge          *int                   `json:"badge,omitempty"`
	ChannelID      string                 `json:"channelId,omitempty"`
	CategoryID     string                 `json:"categoryId,omitempty"`
	MutableContent *bool                  `json:"mutableContent,omitempty"`
}

type expoPushResponse struct {
	Data   []expoPushTicket `json:"data,omitempty"`
	Errors []expoPushError  `json:"errors,omitempty"`
}

type expoReceiptResponse struct {
	Data   map[string]expoPushTicket `json:"data"`
	Errors []expoPushError           `json:"errors,omitempty"`
}

type expoPushTicket struct {
	Status  string                 `json:"status"`
	ID      string                 `json:"id,omitempty"`
	Message string                 `json:"message,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

type expoPushError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewPushNotificationService(tokenRepo *repository.PushTokenRepository, reminderRepo *repository.PushNotificationRepository, eventRepo *repository.PushNotificationEventRepository, receiptRepo *repository.PushReceiptRepository) *PushNotificationService {
	return &PushNotificationService{
		tokenRepo:    tokenRepo,
		reminderRepo: reminderRepo,
		eventRepo:    eventRepo,
		receiptRepo:  receiptRepo,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		accessToken: os.Getenv("EXPO_PUSH_ACCESS_TOKEN"),
	}
}

// SetHTTPClient overrides the HTTP client (used in tests).
func (s *PushNotificationService) SetHTTPClient(client *http.Client) {
	s.httpClient = client
}

func (s *PushNotificationService) SendDailyPracticeReminders(ctx context.Context, now time.Time) error {
	if s.tokenRepo == nil {
		return errors.New("push token repository is nil")
	}
	if s.reminderRepo == nil {
		return errors.New("push notification repository is nil")
	}
	if s.eventRepo == nil {
		return errors.New("push notification event repository is nil")
	}

	candidates, err := s.reminderRepo.FindDailyReminderCandidates(ctx, now)
	if err != nil {
		return err
	}

	if len(candidates) == 0 {
		return nil
	}

	const batchSize = 100
	for start := 0; start < len(candidates); start += batchSize {
		end := start + batchSize
		if end > len(candidates) {
			end = len(candidates)
		}
		batch := candidates[start:end]

		messages := make([]expoPushMessage, 0, len(batch))
		tokenUserIDs := make(map[string]int64, len(batch))
		for _, candidate := range batch {
			if candidate.ExpoToken == "" {
				continue
			}
			title, body := dailyReminderCopy(selectLocaleForDaily(candidate))
			messages = append(messages, expoPushMessage{
				To:       candidate.ExpoToken,
				Title:    title,
				Body:     body,
				Sound:    "default",
				Priority: "default",
				Data: map[string]interface{}{
					"type": "daily_practice_reminder",
				},
			})
			tokenUserIDs[candidate.ExpoToken] = candidate.UserID
		}

		if len(messages) == 0 {
			continue
		}

		tickets, err := s.sendBatch(ctx, messages)
		if err != nil {
			return err
		}

		s.processTickets(ctx, now, messages, tickets, tokenUserIDs, "daily_practice_reminder")
	}

	return nil
}

func (s *PushNotificationService) SendDueItemsReminders(ctx context.Context, now time.Time) error {
	if s.tokenRepo == nil {
		return errors.New("push token repository is nil")
	}
	if s.reminderRepo == nil {
		return errors.New("push notification repository is nil")
	}
	if s.eventRepo == nil {
		return errors.New("push notification event repository is nil")
	}

	candidates, err := s.reminderRepo.FindDueItemsReminderCandidates(ctx, now)
	if err != nil {
		return err
	}

	if len(candidates) == 0 {
		return nil
	}

	const batchSize = 100
	for start := 0; start < len(candidates); start += batchSize {
		end := start + batchSize
		if end > len(candidates) {
			end = len(candidates)
		}
		batch := candidates[start:end]

		messages := make([]expoPushMessage, 0, len(batch))
		tokenUserIDs := make(map[string]int64, len(batch))
		for _, candidate := range batch {
			if candidate.ExpoToken == "" {
				continue
			}
			title, body := dueItemsReminderCopy(selectLocaleForDueItems(candidate), candidate.WordlistName, candidate.DueCount)
			messages = append(messages, expoPushMessage{
				To:       candidate.ExpoToken,
				Title:    title,
				Body:     body,
				Sound:    "default",
				Priority: "default",
				Data: map[string]interface{}{
					"type":       "due_items_reminder",
					"dueCount":   candidate.DueCount,
					"wordlistId": candidate.WordlistID,
					"reminder":   "due_items",
					"timestamp":  now.Unix(),
				},
			})
			tokenUserIDs[candidate.ExpoToken] = candidate.UserID
		}

		if len(messages) == 0 {
			continue
		}

		tickets, err := s.sendBatch(ctx, messages)
		if err != nil {
			return err
		}

		s.processTickets(ctx, now, messages, tickets, tokenUserIDs, "due_items_reminder")
	}

	return nil
}

func (s *PushNotificationService) CheckReceipts(ctx context.Context, now time.Time) error {
	if s.receiptRepo == nil {
		return errors.New("push receipt repository is nil")
	}

	const batchSize = 100
	for {
		pending, err := s.receiptRepo.FetchPending(ctx, batchSize)
		if err != nil {
			return err
		}
		if len(pending) == 0 {
			return nil
		}

		ids := make([]string, 0, len(pending))
		tokenByID := make(map[string]string, len(pending))
		for _, row := range pending {
			if row.TicketID == "" {
				continue
			}
			ids = append(ids, row.TicketID)
			tokenByID[row.TicketID] = row.ExpoToken
		}

		if len(ids) == 0 {
			return nil
		}

		receipts, err := s.fetchReceipts(ctx, ids)
		if err != nil {
			return err
		}
		if len(receipts) == 0 {
			common.Logger.Warn("empty expo receipt response", "pending", len(ids))
			return nil
		}

		var deactivateTokens []string
		for id, ticket := range receipts {
			if err := s.receiptRepo.UpdateReceipt(ctx, id, ticket.Status, ticket.Message, ticket.Details, now); err != nil {
				common.Logger.Error("failed to update push receipt", "error", err, "ticketId", id)
			}
			if ticket.Status == "error" && isDeviceNotRegistered(ticket.Details) {
				if token, ok := tokenByID[id]; ok {
					deactivateTokens = append(deactivateTokens, token)
				}
			}
		}

		if err := s.tokenRepo.DeactivateByTokens(ctx, deactivateTokens); err != nil {
			common.Logger.Error("failed to deactivate push tokens from receipts", "error", err)
		}

		if len(receipts) < len(ids) {
			common.Logger.Warn("partial expo receipt response", "expected", len(ids), "received", len(receipts))
			return nil
		}
	}
}

func selectLocaleForDaily(candidate repository.DailyReminderCandidate) string {
	if candidate.Language != nil && *candidate.Language != "" {
		return normalizeLocale(*candidate.Language)
	}
	if candidate.Locale != nil && *candidate.Locale != "" {
		return normalizeLocale(*candidate.Locale)
	}
	return "en"
}

func selectLocaleForDueItems(candidate repository.DueItemsReminderCandidate) string {
	if candidate.Language != nil && *candidate.Language != "" {
		return normalizeLocale(*candidate.Language)
	}
	if candidate.Locale != nil && *candidate.Locale != "" {
		return normalizeLocale(*candidate.Locale)
	}
	return "en"
}

func normalizeLocale(locale string) string {
	locale = strings.ToLower(strings.TrimSpace(locale))
	if locale == "" {
		return "en"
	}
	if idx := strings.Index(locale, "-"); idx > 0 {
		locale = locale[:idx]
	}
	if idx := strings.Index(locale, "_"); idx > 0 {
		locale = locale[:idx]
	}
	return locale
}

func dailyReminderCopy(locale string) (string, string) {
	localizer := i18nutil.LocalizerForLocale(locale)

	title, titleErr := localizer.Localize(&i18n.LocalizeConfig{MessageID: "daily.title"})
	body, bodyErr := localizer.Localize(&i18n.LocalizeConfig{MessageID: "daily.body"})

	if titleErr != nil {
		title = "Time to practice"
	}
	if bodyErr != nil {
		body = "Your last quiz was a while ago. Come back for a quick practice."
	}

	return title, body
}

func dueItemsReminderCopy(locale string, wordlistName string, dueCount int) (string, string) {
	if wordlistName == "" {
		wordlistName = "your wordlist"
	}
	localizer := i18nutil.LocalizerForLocale(locale)
	templateData := map[string]interface{}{
		"WordlistName": wordlistName,
		"DueCount":     dueCount,
	}

	title, titleErr := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    "due_items.title",
		TemplateData: templateData,
		PluralCount:  dueCount,
	})
	body, bodyErr := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    "due_items.body",
		TemplateData: templateData,
		PluralCount:  dueCount,
	})

	if titleErr != nil {
		if dueCount == 1 {
			title = "You have 1 review due"
		} else {
			title = "You have reviews due"
		}
	}
	if bodyErr != nil {
		if dueCount == 1 {
			body = "You have 1 item ready in " + wordlistName + ". Come back for a quick session."
		} else {
			body = "You have items ready in " + wordlistName + ". Come back for a quick session."
		}
	}

	return title, body
}

func (s *PushNotificationService) sendBatch(ctx context.Context, messages []expoPushMessage) ([]expoPushTicket, error) {
	payload, err := json.Marshal(messages)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, expoPushEndpoint, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if s.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+s.accessToken)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, errors.New("expo push request failed with status " + resp.Status)
	}

	var result expoPushResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Errors) > 0 {
		common.Logger.Warn("expo push returned errors", "errors", result.Errors)
	}

	return result.Data, nil
}

func (s *PushNotificationService) fetchReceipts(ctx context.Context, ids []string) (map[string]expoPushTicket, error) {
	payload, err := json.Marshal(map[string][]string{"ids": ids})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, expoReceiptEndpoint, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if s.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+s.accessToken)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, errors.New("expo receipt request failed with status " + resp.Status)
	}

	var result expoReceiptResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Errors) > 0 {
		common.Logger.Warn("expo receipt response errors", "errors", result.Errors)
	}
	if result.Data == nil {
		return map[string]expoPushTicket{}, nil
	}
	return result.Data, nil
}

func (s *PushNotificationService) processTickets(ctx context.Context, now time.Time, messages []expoPushMessage, tickets []expoPushTicket, tokenUserIDs map[string]int64, notificationType string) {
	if len(tickets) == 0 {
		return
	}

	limit := len(tickets)
	if len(messages) < limit {
		limit = len(messages)
	}

	var notifiedTokens []string
	var deactivateTokens []string
	var receiptInserts []repository.PushReceiptInsert
	uniqueUserIDs := make(map[int64]struct{})

	for i := 0; i < limit; i++ {
		message := messages[i]
		ticket := tickets[i]
		token := pushTokenFromTo(message.To)
		if token == "" {
			continue
		}

		switch ticket.Status {
		case "ok":
			notifiedTokens = append(notifiedTokens, token)
			if userID, ok := tokenUserIDs[token]; ok {
				uniqueUserIDs[userID] = struct{}{}
			}
			if ticket.ID != "" && s.receiptRepo != nil {
				receiptInserts = append(receiptInserts, repository.PushReceiptInsert{
					TicketID:  ticket.ID,
					ExpoToken: token,
				})
			}
		case "error":
			if isDeviceNotRegistered(ticket.Details) {
				deactivateTokens = append(deactivateTokens, token)
			}
			common.Logger.Warn("expo push ticket error",
				"message", ticket.Message,
				"details", ticket.Details,
				"token", token)
		default:
			common.Logger.Warn("unexpected expo push ticket status",
				"status", ticket.Status,
				"token", token)
		}
	}

	if err := s.tokenRepo.MarkNotified(ctx, notifiedTokens, now); err != nil {
		common.Logger.Error("failed to mark push notifications as sent", "error", err)
	}
	if err := s.tokenRepo.DeactivateByTokens(ctx, deactivateTokens); err != nil {
		common.Logger.Error("failed to deactivate push tokens", "error", err)
	}
	if s.eventRepo != nil && len(uniqueUserIDs) > 0 {
		userIDs := make([]int64, 0, len(uniqueUserIDs))
		for userID := range uniqueUserIDs {
			userIDs = append(userIDs, userID)
		}
		if err := s.eventRepo.InsertEvents(ctx, userIDs, notificationType, now); err != nil {
			common.Logger.Error("failed to insert push notification events", "error", err)
		}
	}
	if len(receiptInserts) > 0 && s.receiptRepo != nil {
		if err := s.receiptRepo.InsertTickets(ctx, receiptInserts); err != nil {
			common.Logger.Error("failed to insert push receipts", "error", err)
		}
	}
}

func isDeviceNotRegistered(details map[string]interface{}) bool {
	if details == nil {
		return false
	}
	if value, ok := details["error"].(string); ok {
		return value == "DeviceNotRegistered"
	}
	return false
}

func pushTokenFromTo(to any) string {
	switch value := to.(type) {
	case string:
		return value
	case []string:
		if len(value) == 0 {
			return ""
		}
		return value[0]
	default:
		return ""
	}
}
