package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"decorebator.com/internal/common"
)

// ModerationCategory represents OpenAI moderation categories
type ModerationCategory struct {
	Sexual         bool `json:"sexual"`
	Hate           bool `json:"hate"`
	Harassment     bool `json:"harassment"`
	SelfHarm       bool `json:"self-harm"`
	SexualMinors   bool `json:"sexual/minors"`
	HateThreat     bool `json:"hate/threatening"`
	ViolenceGore   bool `json:"violence/graphic"`
	SelfHarmIntent bool `json:"self-harm/intent"`
	SelfHarmInstr  bool `json:"self-harm/instructions"`
	HarassmentThr  bool `json:"harassment/threatening"`
	Violence       bool `json:"violence"`
}

// ModerationCategoryScore represents confidence scores for each category
type ModerationCategoryScore struct {
	Sexual         float64 `json:"sexual"`
	Hate           float64 `json:"hate"`
	Harassment     float64 `json:"harassment"`
	SelfHarm       float64 `json:"self-harm"`
	SexualMinors   float64 `json:"sexual/minors"`
	HateThreat     float64 `json:"hate/threatening"`
	ViolenceGore   float64 `json:"violence/graphic"`
	SelfHarmIntent float64 `json:"self-harm/intent"`
	SelfHarmInstr  float64 `json:"self-harm/instructions"`
	HarassmentThr  float64 `json:"harassment/threatening"`
	Violence       float64 `json:"violence"`
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
	Model string `json:"model,omitempty"` // Optional, defaults to "text-moderation-latest"
}

// ContentFilterResult maintains compatibility with existing code
type ContentFilterResult struct {
	IsAppropriate bool
	Reason        string
	FlaggedWords  []string
}

// ModerationService interface for content moderation
type ModerationService interface {
	ValidateContent(text string) ContentFilterResult
	ValidateWord(word string) ContentFilterResult
	ValidateWordlistName(name string) ContentFilterResult
	ValidateDescription(description string) ContentFilterResult
	GetContentGuidelines() []string
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

// ValidateContent moderates content using OpenAI's moderation API
func (s *OpenAIModerationService) ValidateContent(text string) ContentFilterResult {
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
	result, err := s.moderateWithOpenAI(text)
	if err != nil {
		// Log error but don't block content creation
		common.Logger.Error("OpenAI moderation API failed, allowing content", "error", err, "text", text)
		return ContentFilterResult{
			IsAppropriate: true,
			Reason:        "",
			FlaggedWords:  []string{},
		}
	}

	if result.Flagged {
		// Track rejection in Redis
		s.trackRejection(result)

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

// ValidateWord validates a single word
func (s *OpenAIModerationService) ValidateWord(word string) ContentFilterResult {
	return s.ValidateContent(word)
}

// ValidateWordlistName validates a wordlist name
func (s *OpenAIModerationService) ValidateWordlistName(name string) ContentFilterResult {
	return s.ValidateContent(name)
}

// ValidateDescription validates a description
func (s *OpenAIModerationService) ValidateDescription(description string) ContentFilterResult {
	return s.ValidateContent(description)
}

// moderateWithOpenAI calls the OpenAI moderation API
func (s *OpenAIModerationService) moderateWithOpenAI(text string) (*ModerationResult, error) {
	reqBody := ModerationRequest{
		Input: text,
		Model: "text-moderation-latest",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", s.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
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
	if categories.Hate || categories.HateThreat {
		return "Content contains hate speech or threatening language"
	}
	if categories.Harassment || categories.HarassmentThr {
		return "Content contains harassment or threatening behavior"
	}
	if categories.Violence || categories.ViolenceGore {
		return "Content contains violent or graphic material"
	}
	if categories.SelfHarm || categories.SelfHarmIntent || categories.SelfHarmInstr {
		return "Content contains self-harm related material"
	}

	return "Content flagged as inappropriate for educational use"
}

// trackRejection increments the daily rejection counter in Redis
func (s *OpenAIModerationService) trackRejection(result *ModerationResult) {
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
	ctx := context.Background()
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
func (s *OpenAIModerationService) GetDailyRejectionCount(date string) (int, error) {
	redisClient, err := common.GetRedisClient()
	if err != nil {
		return 0, fmt.Errorf("Redis client not available: %w", err)
	}

	redisKey := fmt.Sprintf("content:rejections:%s", date)
	val, err := redisClient.Get(context.Background(), redisKey).Result()
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

// GetContentGuidelines returns user-friendly content guidelines
func (s *OpenAIModerationService) GetContentGuidelines() []string {
	return []string{
		"Keep content educational and appropriate for all ages",
		"Use vocabulary words that help language learning",
		"Avoid profanity, adult content, or offensive language",
		"Don't include personal information (emails, phone numbers, URLs)",
		"Keep wordlist names and descriptions family-friendly",
		"Focus on legitimate vocabulary and educational content",
		"Respect cultural sensitivities across different languages",
		"Content is automatically reviewed to ensure it meets community standards",
	}
}
