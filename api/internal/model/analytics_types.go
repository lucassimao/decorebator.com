package model

import (
	"fmt"
	"time"
)

// ProgressSummaryResponse contains progress data for all user's wordlists
type ProgressSummaryResponse struct {
	Wordlists []WordlistProgress `json:"wordlists"`
}

// WordlistProgress contains summary progress data for a single wordlist
type WordlistProgress struct {
	WordlistID       int64      `json:"wordlistId"`
	WordlistName     string     `json:"wordlistName"`
	LanguageCode     string     `json:"languageCode"`
	TotalWords       int        `json:"totalWords"`
	WordsMastered    int        `json:"wordsMastered"`
	ProgressPercent  float64    `json:"progressPercent"`
	CurrentStreak    int        `json:"currentStreak"`
	LastActivityDate *time.Time `json:"lastActivityDate,omitempty"`
}

// CacheableAnalyticsData represents any analytics data that can be cached
type CacheableAnalyticsData interface {
	GetCacheKey(userID, wordlistID int64) string
	GetDataType() string
}

// WordlistStats represents statistics for a wordlist
type WordlistStats struct {
	TotalWords        int      `json:"totalWords"`
	WordsMastered     int      `json:"wordsMastered"`
	AverageMastery    *float64 `json:"averageMastery"`
	BestStreak        *int     `json:"bestStreak"`
	WordsStudiedToday int      `json:"wordsStudiedToday"`
	QuizzesToday      int      `json:"quizzesToday"`
	AccuracyToday     *float64 `json:"accuracyToday"`
	CurrentStreak     int      `json:"currentStreak"`
}

// GetCacheKey implements CacheableAnalyticsData for WordlistStats
func (w WordlistStats) GetCacheKey(userID, wordlistID int64) string {
	return fmt.Sprintf("analytics:stats:%d:%d", userID, wordlistID)
}

// GetDataType implements CacheableAnalyticsData for WordlistStats
func (w WordlistStats) GetDataType() string {
	return "stats"
}

// WordMasteryStats represents word mastery statistics
type WordMasteryStats struct {
	WordID       int64      `json:"wordId"`
	Word         string     `json:"word"`
	MasteryLevel float64    `json:"masteryLevel"`
	Accuracy     float64    `json:"accuracy"`
	StreakCount  int        `json:"streakCount"`
	LastSeenAt   *time.Time `json:"lastSeenAt"`
	HighestBox   int        `json:"highestBox"`
}

// QuizTypePerformance represents quiz type performance statistics
type QuizTypePerformance struct {
	QuizType      string    `json:"quizType"`
	TotalAttempts int       `json:"totalAttempts"`
	SuccessRate   float64   `json:"successRate"`
	AvgResponseMs int       `json:"avgResponseMs"`
	LastUpdated   time.Time `json:"lastUpdated"`
}

// LearningProgressStats represents daily learning progress statistics
type LearningProgressStats struct {
	Date          time.Time `json:"date"`
	WordsStudied  int       `json:"wordsStudied"`
	WordsMastered int       `json:"wordsMastered"`
	TotalAttempts int       `json:"totalAttempts"`
	AccuracyRate  float64   `json:"accuracyRate"`
	AvgResponseMs int       `json:"avgResponseMs"`
}

// BoxDistribution represents the current distribution of words across Leitner boxes
type BoxDistribution struct {
	Box1Count  int `json:"box1"`
	Box2Count  int `json:"box2"`
	Box3Count  int `json:"box3"`
	Box4Count  int `json:"box4"`
	Box5Count  int `json:"box5"`
	Box6Count  int `json:"box6"`
	Box7Count  int `json:"box7"`
	TotalWords int `json:"totalWords"`
}

// PracticeTimeStats represents daily practice time statistics
type PracticeTimeStats struct {
	Date                time.Time `json:"date"`
	PracticeTimeMs      int       `json:"practiceTimeMs"`
	PracticeTimeMinutes float64   `json:"practiceTimeMinutes"`
	QuizCount           int       `json:"quizCount"`
}
