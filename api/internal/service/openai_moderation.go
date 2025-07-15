package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"decorebator.com/internal/common"
)

// ModerationCategory represents OpenAI moderation categories
type ModerationCategory struct {
	Harassment            bool `json:"harassment"`
	HarassmentThreatening bool `json:"harassment/threatening"`
	Hate                  bool `json:"hate"`
	HateThreatening       bool `json:"hate/threatening"`
	SelfHarm              bool `json:"self-harm"`
	SelfHarmIntent        bool `json:"self-harm/intent"`
	SelfHarmInstructions  bool `json:"self-harm/instructions"`
	Sexual                bool `json:"sexual"`
	SexualMinors          bool `json:"sexual/minors"`
	Violence              bool `json:"violence"`
	ViolenceGraphic       bool `json:"violence/graphic"`
}

// ModerationCategoryScore represents confidence scores for each category
type ModerationCategoryScore struct {
	Harassment            float64 `json:"harassment"`
	HarassmentThreatening float64 `json:"harassment/threatening"`
	Hate                  float64 `json:"hate"`
	HateThreatening       float64 `json:"hate/threatening"`
	SelfHarm              float64 `json:"self-harm"`
	SelfHarmIntent        float64 `json:"self-harm/intent"`
	SelfHarmInstructions  float64 `json:"self-harm/instructions"`
	Sexual                float64 `json:"sexual"`
	SexualMinors          float64 `json:"sexual/minors"`
	Violence              float64 `json:"violence"`
	ViolenceGraphic       float64 `json:"violence/graphic"`
}

// ModerationResult represents a single moderation result
type ModerationResult struct {
	Categories     ModerationCategory      `json:"categories"`
	CategoryScores ModerationCategoryScore `json:"category_scores"`
	Flagged        bool                    `json:"flagged"`
}

// ModerationResponse represents the OpenAI moderation API response
type ModerationResponse struct {
	ID      string             `json:"id"`
	Model   string             `json:"model"`
	Results []ModerationResult `json:"results"`
}

// ModerationRequest represents the OpenAI moderation API request
type ModerationRequest struct {
	Input string `json:"input"`
	Model string `json:"model,omitempty"` // Optional, defaults to "omni-moderation-latest"
}

// ContentFilterResult maintains compatibility with existing code
type ContentFilterResult struct {
	IsAppropriate bool
	Reason        string
	FlaggedWords  []string
}

// ModerationService interface for content moderation
type ModerationService interface {
	Validate(ctx context.Context, text string) ContentFilterResult
}

// OpenAIModerationService handles content moderation using OpenAI's moderation API
type OpenAIModerationService struct {
	apiKey     string
	httpClient *http.Client
	enabled    bool
	apiURL     string
}

var _ ModerationService = (*OpenAIModerationService)(nil)

// NewOpenAIModerationService creates a new OpenAI moderation service
func NewOpenAIModerationService() *OpenAIModerationService {
	apiKey := os.Getenv("OPENAI_API_KEY")
	isProduction := os.Getenv("ENV") == "production"
	apiURL := os.Getenv("OPENAI_API_URL")
	if apiURL == "" {
		apiURL = "https://api.openai.com/v1/moderations"
	}

	return &OpenAIModerationService{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		enabled: isProduction && apiKey != "",
		apiURL:  apiURL,
	}
}

// Validate moderates content using OpenAI's moderation API
func (s *OpenAIModerationService) Validate(ctx context.Context, text string) ContentFilterResult {
	// Basic validation first (empty, too long, etc.)
	if text == "" {
		return ContentFilterResult{
			IsAppropriate: false,
			Reason:        "Content cannot be empty",
			FlaggedWords:  []string{},
		}
	}

	if len(text) > 2000 {
		return ContentFilterResult{
			IsAppropriate: false,
			Reason:        "Content is too long (maximum 2000 characters)",
			FlaggedWords:  []string{},
		}
	}

	// If not in production or API key not available, use basic validation
	if !s.enabled {
		common.Logger.Debug("OpenAI moderation disabled - using basic validation", "text", text)
		return ContentFilterResult{
			IsAppropriate: true,
			Reason:        "",
			FlaggedWords:  []string{},
		}
	}

	// Call OpenAI moderation API
	result, err := s.moderateWithOpenAI(ctx, text)
	if err != nil {
		// Enhanced logging to differentiate timeout vs other errors
		if strings.Contains(err.Error(), "timed out after 10 seconds") {
			common.Logger.Warn("OpenAI moderation API timed out, allowing content creation", "error", err, "text", text)
		} else {
			common.Logger.Error("OpenAI moderation API failed, allowing content creation", "error", err, "text", text)
		}
		return ContentFilterResult{
			IsAppropriate: true,
			Reason:        "",
			FlaggedWords:  []string{},
		}
	}

	if result.Flagged {
		// Track rejection in Redis
		s.trackRejection(ctx, result)

		// Determine the most relevant flagged category
		reason := s.buildRejectionReason(result)

		return ContentFilterResult{
			IsAppropriate: false,
			Reason:        reason,
			FlaggedWords:  []string{}, // OpenAI doesn't provide specific flagged words
		}
	}

	return ContentFilterResult{
		IsAppropriate: true,
		Reason:        "",
		FlaggedWords:  []string{},
	}
}

// moderateWithOpenAI calls the OpenAI moderation API
func (s *OpenAIModerationService) moderateWithOpenAI(ctx context.Context, text string) (*ModerationResult, error) {
	reqBody := ModerationRequest{
		Input: text,
		Model: "omni-moderation-latest",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	// Create context with 10-second timeout (or use parent context if it has shorter timeout)
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(timeoutCtx, "POST", s.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		// Check if error is due to timeout
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("OpenAI moderation API request timed out after 10 seconds: %w", err)
		}
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API returned status %d", resp.StatusCode)
	}

	var moderationResp ModerationResponse
	if err := json.NewDecoder(resp.Body).Decode(&moderationResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	if len(moderationResp.Results) == 0 {
		return nil, fmt.Errorf("no results in moderation response")
	}

	return &moderationResp.Results[0], nil
}

// buildRejectionReason creates a human-readable rejection reason
func (s *OpenAIModerationService) buildRejectionReason(result *ModerationResult) string {
	categories := result.Categories

	if categories.Sexual || categories.SexualMinors {
		return "Content contains sexual or adult material not appropriate for educational use"
	}
	if categories.Hate || categories.HateThreatening {
		return "Content contains hate speech or threatening language"
	}
	if categories.Harassment || categories.HarassmentThreatening {
		return "Content contains harassment or threatening behavior"
	}
	if categories.Violence || categories.ViolenceGraphic {
		return "Content contains violent or graphic material"
	}
	if categories.SelfHarm || categories.SelfHarmIntent || categories.SelfHarmInstructions {
		return "Content contains self-harm related material"
	}

	return "Content flagged as inappropriate for educational use"
}

// trackRejection increments the daily rejection counter in Redis
func (s *OpenAIModerationService) trackRejection(ctx context.Context, result *ModerationResult) {
	// Get today's date for the Redis key
	today := time.Now().Format("2006-01-02")
	redisKey := fmt.Sprintf("content:rejections:%s", today)

	// Try to get Redis client
	redisClient, err := common.GetRedisClient()
	if err != nil || redisClient == nil {
		// Redis not available, just log the rejection
		common.Logger.Warn("Content rejected by OpenAI moderation - Redis unavailable for tracking",
			"date", today,
			"reason", s.buildRejectionReason(result))
		return
	}

	// Increment counter with expiration (keep for 30 days)
	// Use provided context for Redis operations
	pipe := redisClient.Pipeline()
	pipe.Incr(ctx, redisKey)
	pipe.Expire(ctx, redisKey, 30*24*time.Hour) // 30 days

	_, pipeErr := pipe.Exec(ctx)
	if pipeErr != nil {
		common.Logger.Error("Failed to track content rejection in Redis",
			"error", pipeErr,
			"key", redisKey,
			"reason", s.buildRejectionReason(result))
	} else {
		common.Logger.Info("Content rejected by OpenAI moderation",
			"date", today,
			"reason", s.buildRejectionReason(result))
	}
}

// GetDailyRejectionCount returns the number of rejections for a specific date
func (s *OpenAIModerationService) GetDailyRejectionCount(ctx context.Context, date string) (int, error) {
	redisClient, err := common.GetRedisClient()
	if err != nil {
		return 0, fmt.Errorf("Redis client not available: %w", err)
	}

	redisKey := fmt.Sprintf("content:rejections:%s", date)
	val, err := redisClient.Get(ctx, redisKey).Result()
	if err != nil {
		if err.Error() == "redis: nil" {
			return 0, nil // No rejections for this date
		}
		return 0, err
	}

	count, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("invalid count value in Redis: %w", err)
	}

	return count, nil
}
