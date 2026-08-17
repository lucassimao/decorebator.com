package http

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	analyticsrepo "decorebator.com/internal/repository/analytics"
	"github.com/gin-gonic/gin"
)

const (
	defaultPageSize  = 50
	maxPageSize      = 100
	nextCursorHeader = "X-Next-Cursor"
)

// pageRequest is intentionally keyset-only. Offset pages duplicate or skip
// items when a user adds or removes data between requests.
type pageRequest struct {
	Limit  int
	Cursor *int64
}

func parsePageRequest(c *gin.Context) (pageRequest, error) {
	limit, err := parsePageLimit(c)
	if err != nil {
		return pageRequest{}, err
	}
	page := pageRequest{Limit: limit}

	if rawCursor := c.Query("cursor"); rawCursor != "" {
		cursor, err := strconv.ParseInt(rawCursor, 10, 64)
		if err != nil || cursor <= 0 {
			return pageRequest{}, fmt.Errorf("cursor must be a positive integer")
		}
		page.Cursor = &cursor
	}
	return page, nil
}

func parsePageLimit(c *gin.Context) (int, error) {
	if rawLimit := c.Query("limit"); rawLimit != "" {
		limit, err := strconv.Atoi(rawLimit)
		if err != nil || limit < 1 || limit > maxPageSize {
			return 0, fmt.Errorf("limit must be between 1 and %d", maxPageSize)
		}
		return limit, nil
	}
	return defaultPageSize, nil
}

// pageItems removes the single look-ahead row used by repositories and exposes
// the stable keyset cursor without changing existing JSON array response shapes.
func pageItems[T any](c *gin.Context, page pageRequest, items []T, id func(T) int64) []T {
	return pageItemsWithCursor(c, page.Limit, items, func(item T) string {
		return strconv.FormatInt(id(item), 10)
	})
}

func pageItemsWithCursor[T any](c *gin.Context, limit int, items []T, cursor func(T) string) []T {
	c.Header(nextCursorHeader, "")
	if len(items) <= limit {
		return items
	}
	visible := items[:limit]
	c.Header(nextCursorHeader, cursor(visible[len(visible)-1]))
	return visible
}

func writeInvalidPage(c *gin.Context, err error) {
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}

const maxOpaqueCursorBytes = 512

type masteryCursorPayload struct {
	MasteryLevel string     `json:"masteryLevel"`
	LastSeenAt   *time.Time `json:"lastSeenAt"`
	WordID       int64      `json:"wordId"`
}

func parseMasteryCursor(raw string) (*analyticsrepo.WordMasteryCursor, error) {
	if raw == "" {
		return nil, nil
	}
	var payload masteryCursorPayload
	if err := decodeOpaqueCursor(raw, &payload); err != nil {
		return nil, err
	}
	mastery, err := strconv.ParseFloat(payload.MasteryLevel, 64)
	if err != nil || mastery < 0 || mastery > 1 ||
		strconv.FormatFloat(mastery, 'f', 2, 64) != payload.MasteryLevel ||
		(payload.LastSeenAt != nil && payload.LastSeenAt.IsZero()) || payload.WordID <= 0 {
		return nil, fmt.Errorf("cursor is invalid")
	}
	return &analyticsrepo.WordMasteryCursor{
		MasteryLevel: strconv.FormatFloat(mastery, 'f', 2, 64),
		LastSeenAt:   payload.LastSeenAt,
		WordID:       payload.WordID,
	}, nil
}

func encodeMasteryPageCursor(masteryLevel float64, lastSeenAt *time.Time, wordID int64) string {
	return encodeOpaqueCursor(masteryCursorPayload{
		MasteryLevel: strconv.FormatFloat(masteryLevel, 'f', 2, 64),
		LastSeenAt:   lastSeenAt,
		WordID:       wordID,
	})
}

type subscriptionCursorPayload struct {
	CreatedAt time.Time `json:"createdAt"`
	ID        int64     `json:"id"`
}

func parseSubscriptionCursor(raw string) (*repository.SubscriptionHistoryCursor, error) {
	if raw == "" {
		return nil, nil
	}
	var payload subscriptionCursorPayload
	if err := decodeOpaqueCursor(raw, &payload); err != nil {
		return nil, err
	}
	if payload.CreatedAt.IsZero() || payload.ID <= 0 {
		return nil, fmt.Errorf("cursor is invalid")
	}
	return &repository.SubscriptionHistoryCursor{CreatedAt: payload.CreatedAt, ID: payload.ID}, nil
}

func encodeSubscriptionCursor(subscription *model.Subscription) string {
	return encodeOpaqueCursor(subscriptionCursorPayload{CreatedAt: subscription.CreatedAt, ID: subscription.ID})
}

func decodeOpaqueCursor(raw string, target any) error {
	if len(raw) > maxOpaqueCursorBytes {
		return fmt.Errorf("cursor is too long")
	}
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil || len(decoded) == 0 || len(decoded) > maxOpaqueCursorBytes {
		return fmt.Errorf("cursor is invalid")
	}
	decoder := json.NewDecoder(bytes.NewReader(decoded))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
		return fmt.Errorf("cursor is invalid")
	}
	return nil
}

func encodeOpaqueCursor(payload any) string {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(encoded)
}
