package model

// ContentStatus represents the content moderation status
type ContentStatus string

const (
	ContentStatusApproved ContentStatus = "approved"
	ContentStatusPending  ContentStatus = "pending"
	ContentStatusRejected ContentStatus = "rejected"
)

// ContentFilterResult represents the result of content filtering
type ContentFilterResult struct {
	IsAppropriate bool     `json:"isAppropriate"`
	Severity      string   `json:"severity,omitempty"`
	Reason        string   `json:"reason,omitempty"`
	FlaggedWords  []string `json:"flaggedWords,omitempty"`
}

// ContentModerationLog represents a content moderation event
type ContentModerationLog struct {
	ID           int64          `json:"id"`
	EntityType   string         `json:"entityType"`
	EntityID     int64          `json:"entityId"`
	UserID       int64          `json:"userId"`
	OldStatus    *ContentStatus `json:"oldStatus,omitempty"`
	NewStatus    ContentStatus  `json:"newStatus"`
	Reason       *string        `json:"reason,omitempty"`
	FlaggedWords []string       `json:"flaggedWords,omitempty"`
	Severity     *string        `json:"severity,omitempty"`
	Automated    bool           `json:"automated"`
	ReviewedBy   *int64         `json:"reviewedBy,omitempty"`
	CreatedAt    string         `json:"createdAt"`
	UpdatedAt    string         `json:"updatedAt"`
}
