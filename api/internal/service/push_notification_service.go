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
	"decorebator.com/internal/repository"
)

const expoPushEndpoint = "https://exp.host/--/api/v2/push/send"
const expoReceiptEndpoint = "https://exp.host/--/api/v2/push/getReceipts"

type PushNotificationService struct {
	repo        *repository.PushTokenRepository
	receiptRepo *repository.PushReceiptRepository
	httpClient  *http.Client
	accessToken string
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

func NewPushNotificationService(repo *repository.PushTokenRepository, receiptRepo *repository.PushReceiptRepository) *PushNotificationService {
	return &PushNotificationService{
		repo:        repo,
		receiptRepo: receiptRepo,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		accessToken: os.Getenv("EXPO_PUSH_ACCESS_TOKEN"),
	}
}

func (s *PushNotificationService) SendDailyPracticeReminders(ctx context.Context, now time.Time) error {
	if s.repo == nil {
		return errors.New("push token repository is nil")
	}

	candidates, err := s.repo.FindDailyReminderCandidates(ctx, now)
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
		for _, candidate := range batch {
			if candidate.ExpoToken == "" {
				continue
			}
			title, body := dailyReminderCopy(selectLocale(candidate))
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
		}

		if len(messages) == 0 {
			continue
		}

		tickets, err := s.sendBatch(ctx, messages)
		if err != nil {
			return err
		}

		s.processTickets(ctx, now, messages, tickets)
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

		if err := s.repo.DeactivateByTokens(ctx, deactivateTokens); err != nil {
			common.Logger.Error("failed to deactivate push tokens from receipts", "error", err)
		}
	}
}

func selectLocale(candidate repository.DailyReminderCandidate) string {
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
	switch locale {
	case "it":
		return "È ora di fare pratica", "Il tuo ultimo quiz è passato da un po'. Torna per un rapido ripasso."
	case "es":
		return "Hora de practicar", "Tu último quiz fue hace un tiempo. Vuelve para una práctica rápida."
	case "fr":
		return "Il est temps de pratiquer", "Votre dernier quiz remonte à un moment. Revenez pour une courte pratique."
	case "pt":
		return "Hora de praticar", "Seu último quiz foi há algum tempo. Volte para uma prática rápida."
	case "de":
		return "Zeit zum Üben", "Dein letztes Quiz ist schon eine Weile her. Komm zurück für eine kurze Übung."
	default:
		return "Time to practice", "Your last quiz was a while ago. Come back for a quick practice."
	}
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

func (s *PushNotificationService) processTickets(ctx context.Context, now time.Time, messages []expoPushMessage, tickets []expoPushTicket) {
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

	if err := s.repo.MarkNotified(ctx, notifiedTokens, now); err != nil {
		common.Logger.Error("failed to mark push notifications as sent", "error", err)
	}
	if err := s.repo.DeactivateByTokens(ctx, deactivateTokens); err != nil {
		common.Logger.Error("failed to deactivate push tokens", "error", err)
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
