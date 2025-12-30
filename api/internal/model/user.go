package model

import (
	"encoding/json"
	"time"

	"github.com/jackc/pgx/pgtype"
)

type User struct {
	ID                   int64               `json:"id"`
	FirstName            string              `json:"firstName"`
	LastName             string              `json:"lastName"`
	PasswordHash         string              `json:"passwordHash"`
	Email                string              `json:"email"`
	ProfilePictureURL    *string             `json:"profilePictureUrl,omitempty"`
	Country              *string             `json:"country,omitempty"`
	DateOfBirth          *time.Time          `json:"dateOfBirth,omitempty"`
	PreferredLanguage    *string             `json:"preferredLanguage,omitempty"`
	SubscriptionPlan     SubscriptionPlan    `json:"subscriptionPlan"`
	SubscriptionStatus   *SubscriptionStatus `json:"subscriptionStatus,omitempty"`
	StripeCustomerID     *string             `json:"stripeCustomerId,omitempty"`
	Platform             *PlatformType       `json:"platform,omitempty"`
	SubscriptionEndsAt   *time.Time          `json:"subscriptionEndsAt,omitempty"`
	NotificationsEnabled bool                `json:"notificationsEnabled"`
	CreatedAt            pgtype.Timestamp    `json:"createdAt"`
	UpdatedAt            pgtype.Timestamp    `json:"updatedAt"`
}

func (u User) MarshalJSON() ([]byte, error) {
	userMap := map[string]interface{}{
		"id":               u.ID,
		"firstName":        u.FirstName,
		"lastName":         u.LastName,
		"passwordHash":     u.PasswordHash,
		"email":            u.Email,
		"subscriptionPlan": u.SubscriptionPlan,
	}

	if u.ProfilePictureURL != nil {
		userMap["profilePictureUrl"] = *u.ProfilePictureURL
	}

	if u.Country != nil {
		userMap["country"] = *u.Country
	}

	if u.DateOfBirth != nil {
		userMap["dateOfBirth"] = u.DateOfBirth.UTC().Format("2006-01-02")
	}

	if u.PreferredLanguage != nil {
		userMap["preferredLanguage"] = *u.PreferredLanguage
	}

	if u.SubscriptionStatus != nil {
		userMap["subscriptionStatus"] = *u.SubscriptionStatus
	}

	if u.StripeCustomerID != nil {
		userMap["stripeCustomerId"] = *u.StripeCustomerID
	}

	if u.Platform != nil {
		userMap["platform"] = *u.Platform
	}

	if u.SubscriptionEndsAt != nil {
		userMap["subscriptionEndsAt"] = u.SubscriptionEndsAt.UTC().Format(time.RFC3339)
	}

	userMap["notificationsEnabled"] = u.NotificationsEnabled

	if u.CreatedAt.Status == pgtype.Present {
		userMap["createdAt"] = u.CreatedAt.Time.UTC().Format(time.RFC3339)
	}

	if u.UpdatedAt.Status == pgtype.Present {
		userMap["updatedAt"] = u.UpdatedAt.Time.UTC().Format(time.RFC3339)
	}

	return json.Marshal(userMap)
}
