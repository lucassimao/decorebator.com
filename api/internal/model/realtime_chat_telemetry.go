package model

import "time"

type RealtimeChatTurn struct {
	UserSpeechMs       *int            `json:"userSpeechMs,omitempty"`
	AssistantLatencyMs *int            `json:"assistantLatencyMs,omitempty"`
	Usage              *TelemetryUsage `json:"usage,omitempty"`
}

type TelemetryUsage struct {
	InputTokens       int `json:"inputTokens"`
	OutputTokens      int `json:"outputTokens"`
	AudioInputTokens  int `json:"audioInputTokens"`
	AudioOutputTokens int `json:"audioOutputTokens"`
	TextInputTokens   int `json:"textInputTokens"`
	TextOutputTokens  int `json:"textOutputTokens"`
}

type RateLimitSnapshot struct {
	Name         string  `json:"name"`
	Limit        int     `json:"limit"`
	Remaining    int     `json:"remaining"`
	ResetSeconds float64 `json:"resetSeconds"`
	RecordedAt   string  `json:"recordedAt"`
}

type RealtimeChatTelemetry struct {
	UserID              int64
	WordlistID          *int64
	SessionUUID         string
	ConversationID      *string
	LanguageCode        *string
	Model               *string
	StartedAt           time.Time
	EndedAt             time.Time
	DurationMs          int
	TotalAudioTokensIn  int
	TotalAudioTokensOut int
	TotalTextTokensIn   int
	TotalTextTokensOut  int
	TotalTurns          int
	Turns               []RealtimeChatTurn
	RateLimits          []RateLimitSnapshot
	ErrorCount          int
	ClientVersion       *string
	DeviceInfo          map[string]any
}
