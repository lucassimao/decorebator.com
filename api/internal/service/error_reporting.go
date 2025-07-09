package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ErrorReportType string

const (
	UnrelatedImage   ErrorReportType = "_unrelated_image"
	MissingImage     ErrorReportType = "_missing_image"
	UnrelatedMeaning ErrorReportType = "_unrelated_meaning"
	UnrelatedExample ErrorReportType = "_unrelated_example"
	SoundNotPlaying  ErrorReportType = "_sound_not_playing"
	ProcessingFailed ErrorReportType = "_processing_failed" // For retrying failed word processing
)

type QuizDetails struct {
	QuizType    string   `json:"quizType"`
	Value       string   `json:"value"`
	Options     []string `json:"options"`
	AnswerIndex int      `json:"answerIndex"`
	Context     string   `json:"context"` // "quiz" or "flashcard"
}

// ErrorReportService encapsulates all error reporting functionality with zero-parameter methods
type ErrorReportService struct {
	// Dependencies
	db     *pgxpool.Pool
	repo   *repository.ErrorReportRepository
	logger *slog.Logger

	// Request data (embedded directly)
	ctx          context.Context
	errorType    ErrorReportType
	wordID       int64
	definitionID *int64
	userID       int64
	quizDetails  *QuizDetails
}

// NewErrorReportService creates a new error report service with all dependencies and request data
func NewErrorReportService(ctx context.Context, errorType ErrorReportType, wordID int64, definitionID *int64, userID int64, quizDetails *QuizDetails) (*ErrorReportService, error) {
	db := common.GetDBConnection()

	return &ErrorReportService{
		db:           db,
		repo:         repository.NewErrorReportRepository(db),
		logger:       common.Logger.With("errorType", errorType, "wordID", wordID, "definitionID", definitionID, "userID", userID),
		ctx:          ctx,
		errorType:    errorType,
		wordID:       wordID,
		definitionID: definitionID,
		userID:       userID,
		quizDetails:  quizDetails,
	}, nil
}

// ProcessReport is the main entry point that orchestrates the entire error reporting process
func (s *ErrorReportService) ProcessReport() error {
	// Log error report attempt for monitoring
	s.logger.Info("Error report attempt",
		"action", "error_report_attempt",
		"timestamp", time.Now().Unix(),
	)

	// Validate user owns the word
	if err := s.validateUserOwnsWord(); err != nil {
		return err
	}

	// Check cooldown period
	if err := s.checkCooldown(); err != nil {
		return err
	}

	// Execute error report in transaction
	return s.executeTransaction()
}

// validateUserOwnsWord ensures the user has permission to report errors for this word
func (s *ErrorReportService) validateUserOwnsWord() error {
	isValid, err := didUserCreateWord(s.wordID, s.userID)
	if err != nil || !isValid {
		if err != nil {
			s.logger.Error("validation failed", "error", err)
		}
		return errors.New("validation failed")
	}
	return nil
}

// checkCooldown verifies if the user is within the cooldown period for this error type
func (s *ErrorReportService) checkCooldown() error {
	cooldownUntil, err := s.repo.CheckCooldown(s.ctx, s.userID, s.wordID, s.definitionID, string(s.errorType))
	if err != nil {
		s.logger.Error("failed to check cooldown", "error", err)
		return err
	}

	if cooldownUntil != nil {
		retryAfter := cooldownUntil.Sub(time.Now())
		s.logger.Warn("Error report blocked by cooldown",
			"action", "error_report_cooldown_blocked",
			"timestamp", time.Now().Unix(),
		)
		return CooldownError{
			Message:       fmt.Sprintf("Please wait before reporting this error again. You can retry in %d minutes.", int(retryAfter.Minutes())),
			CooldownUntil: *cooldownUntil,
			RetryAfter:    retryAfter,
		}
	}

	return nil
}

// executeTransaction manages the database transaction for the error report process
func (s *ErrorReportService) executeTransaction() error {
	tx, err := s.db.Begin(s.ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			if commitErr := tx.Commit(s.ctx); commitErr != nil {
				common.Logger.Error("failed to commit transaction in error reporting", "error", commitErr)
			}
		} else {
			if rollbackErr := tx.Rollback(s.ctx); rollbackErr != nil {
				common.Logger.Error("failed to rollback transaction in error reporting", "error", rollbackErr)
			}
		}
	}()

	// Process the error type and trigger appropriate workers
	report, err := s.processErrorType(tx)
	if err != nil {
		common.Logger.Error("failed to trigger jobs", "error", err)
		return err
	}

	// Handle post-processing steps
	return s.completeReport(tx, report)
}

// processErrorType handles the specific error type and triggers the appropriate workers
func (s *ErrorReportService) processErrorType(tx pgx.Tx) (ErrorReport, error) {
	var report ErrorReport
	var err error

	switch s.errorType {
	case SoundNotPlaying:
		report = ErrorReport{WordId: &s.wordID, UserId: s.userID}
		_, err = TriggerTextToSpeechWorker(s.wordID, &s.userID, &report, &tx)

	case UnrelatedImage, MissingImage:
		if s.definitionID == nil {
			return report, fmt.Errorf("definition ID required for image-related errors")
		}
		report = ErrorReport{DefinitionId: s.definitionID, UserId: s.userID}
		_, err = TriggerGenerateImageWorker(*s.definitionID, &s.userID, &report, &tx)

	case UnrelatedExample, UnrelatedMeaning:
		err = DeleteWordDefinitions(s.wordID, &tx)
		if err == nil {
			report = ErrorReport{WordId: &s.wordID, UserId: s.userID}
			_, err = TriggerFetchDefinitionWorker(s.wordID, &s.userID, &report, &tx)
		}

	case ProcessingFailed:
		err = DeleteWordDefinitions(s.wordID, &tx)
		if err == nil {
			report = ErrorReport{WordId: &s.wordID, UserId: s.userID}
			_, err = TriggerFetchDefinitionWorker(s.wordID, &s.userID, &report, &tx)
		}

	default:
		err = fmt.Errorf("invalid error type %s", s.errorType)
	}

	return report, err
}

// completeReport handles all the post-processing steps after the main error processing
func (s *ErrorReportService) completeReport(tx pgx.Tx, report ErrorReport) error {
	// Update last regenerated timestamp
	if err := s.updateLastRegeneratedTimestamp(tx); err != nil {
		s.logger.Error("failed to update last regenerated timestamp", "error", err)
		// Don't fail the whole operation for this
	}

	// Set cooldown for this error report
	if err := s.setCooldownPeriod(tx); err != nil {
		s.logger.Error("failed to set cooldown", "error", err)
		// Don't fail the whole operation for this
	}

	// Mark definition temporarily as unavailable
	strategy := DefaultLeitnerSystemStrategy()
	if err := strategy.ReportError(s.userID, report, tx, s.ctx); err != nil {
		common.Logger.Error("failed to report error", "error", err)
		return err
	}

	// Build content snapshot before potential deletion
	contentSnapshot, finalDefinitionID, err := s.buildContentSnapshot(tx)
	if err != nil {
		s.logger.Error("failed to build content snapshot", "error", err)
		return err
	}

	// Upsert error report with content snapshot
	if err := s.repo.UpsertErrorReport(s.ctx, tx, s.userID, finalDefinitionID, s.wordID, string(s.errorType), contentSnapshot); err != nil {
		common.Logger.Error("failed to save error report", "error", err)
		return err
	}

	// Log successful error report for monitoring
	s.logger.Info("Error report processed successfully",
		"action", "error_report_success",
		"timestamp", time.Now().Unix(),
		"regeneration_type", s.errorType,
	)

	return nil
}

// updateLastRegeneratedTimestamp updates the timestamp for when content was last regenerated
func (s *ErrorReportService) updateLastRegeneratedTimestamp(tx pgx.Tx) error {
	switch s.errorType {
	case SoundNotPlaying:
		return s.repo.UpdateWordAudioLastRegeneratedAt(s.ctx, tx, s.wordID)
	case UnrelatedImage, MissingImage, UnrelatedExample, UnrelatedMeaning:
		if s.definitionID != nil {
			return s.repo.UpdateDefinitionLastRegeneratedAt(s.ctx, tx, *s.definitionID)
		}
	}
	return nil
}

// setCooldownPeriod sets the cooldown period to prevent spam reporting
func (s *ErrorReportService) setCooldownPeriod(tx pgx.Tx) error {
	cooldownDuration := time.Hour // 1 hour cooldown as agreed
	cooldownUntilTime := time.Now().Add(cooldownDuration)
	return s.repo.SetCooldown(s.ctx, tx, s.userID, s.wordID, s.definitionID, string(s.errorType), cooldownUntilTime)
}

// buildContentSnapshot creates a content snapshot and determines final foreign key strategy
func (s *ErrorReportService) buildContentSnapshot(tx pgx.Tx) (map[string]interface{}, *int64, error) {
	var snapshot map[string]any
	var finalDefinitionID *int64

	switch s.errorType {
	case UnrelatedMeaning, UnrelatedExample:
		// These will delete definitions, so capture full snapshot and nullify foreign key
		if s.definitionID == nil {
			return nil, nil, fmt.Errorf("definition ID required for %s", s.errorType)
		}

		defSnapshot, err := s.fetchCompleteDefinitionSnapshot(*s.definitionID)
		if err != nil {
			return nil, nil, err
		}

		snapshot = defSnapshot

		// Add quiz details if available
		if s.quizDetails != nil {
			snapshot["quiz_context"] = map[string]interface{}{
				"quiz_type":    s.quizDetails.QuizType,
				"value":        s.quizDetails.Value,
				"options":      s.quizDetails.Options,
				"answer_index": s.quizDetails.AnswerIndex,
				"context":      s.quizDetails.Context,
				"captured_at":  time.Now().Format(time.RFC3339),
			}
		}

		finalDefinitionID = nil // Set to NULL since definition will be deleted

	case UnrelatedImage, MissingImage:
		// These regenerate images but keep definitions, so preserve foreign key
		if s.definitionID == nil {
			return nil, nil, fmt.Errorf("definition ID required for %s", s.errorType)
		}

		defSnapshot, err := s.fetchCompleteDefinitionSnapshot(*s.definitionID)
		if err != nil {
			return nil, nil, err
		}

		snapshot = defSnapshot

		// Add quiz details if available
		if s.quizDetails != nil {
			snapshot["quiz_context"] = map[string]interface{}{
				"quiz_type":    s.quizDetails.QuizType,
				"value":        s.quizDetails.Value,
				"options":      s.quizDetails.Options,
				"answer_index": s.quizDetails.AnswerIndex,
				"context":      s.quizDetails.Context,
				"captured_at":  time.Now().Format(time.RFC3339),
			}
		}

		finalDefinitionID = s.definitionID // Keep foreign key

	case SoundNotPlaying:
		// This regenerates audio but keeps word, so preserve foreign key
		wordSnapshot, err := s.fetchWordSnapshot(tx)
		if err != nil {
			return nil, nil, err
		}

		snapshot = wordSnapshot

		// Add quiz details if available
		if s.quizDetails != nil {
			snapshot["quiz_context"] = map[string]interface{}{
				"quiz_type":    s.quizDetails.QuizType,
				"value":        s.quizDetails.Value,
				"options":      s.quizDetails.Options,
				"answer_index": s.quizDetails.AnswerIndex,
				"context":      s.quizDetails.Context,
				"captured_at":  time.Now().Format(time.RFC3339),
			}
		}

		finalDefinitionID = s.definitionID // Keep foreign key (might be nil for word-level errors)

	case ProcessingFailed:
		// This is a processing retry, no content quality issue
		snapshot = make(map[string]interface{})

		// Add quiz details if available, even for processing failures
		if s.quizDetails != nil {
			snapshot["quiz_context"] = map[string]interface{}{
				"quiz_type":    s.quizDetails.QuizType,
				"value":        s.quizDetails.Value,
				"options":      s.quizDetails.Options,
				"answer_index": s.quizDetails.AnswerIndex,
				"context":      s.quizDetails.Context,
				"captured_at":  time.Now().Format(time.RFC3339),
			}
		} else {
			// If no quiz details, keep the snapshot nil as before
			snapshot = nil
		}

		finalDefinitionID = nil // Set to NULL since definition will be deleted

	default:
		return nil, nil, fmt.Errorf("unsupported error type: %s", s.errorType)
	}

	return snapshot, finalDefinitionID, nil
}

// fetchCompleteDefinitionSnapshot fetches complete definition data for historical preservation
func (s *ErrorReportService) fetchCompleteDefinitionSnapshot(definitionID int64) (map[string]interface{}, error) {
	// Use definition service to fetch the definition
	definition, err := GetDefinitionById(definitionID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch definition: %w", err)
	}
	if definition == nil {
		return nil, fmt.Errorf("definition not found with ID: %d", definitionID)
	}

	// Build comprehensive snapshot
	snapshot := map[string]interface{}{
		"definition_id":  definitionID,
		"meaning":        definition.Meaning,
		"part_of_speech": definition.PartOfSpeech,
		"source":         definition.Source,
		"language":       definition.Language,
		"token":          definition.Token,
		"examples":       definition.Examples,
		"captured_at":    time.Now().Format(time.RFC3339),
	}

	return snapshot, nil
}

// fetchWordSnapshot fetches word data for audio-related errors
func (s *ErrorReportService) fetchWordSnapshot(tx pgx.Tx) (map[string]interface{}, error) {
	wordService := NewWordService(s.db, nil) // Pass db connection, nil for moderation service since we're just reading
	word, err := wordService.GetWordByID(s.wordID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch word: %w", err)
	}
	if word == nil {
		return nil, fmt.Errorf("word not found with ID: %d", s.wordID)
	}

	// Get language from wordlist - we still need this query since word doesn't contain language directly
	var language string
	err = tx.QueryRow(s.ctx, `
		SELECT wl.language_code
		FROM wordlists wl
		WHERE wl.id = $1
	`, word.WordlistID).Scan(&language)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch wordlist language: %w", err)
	}

	// Build word snapshot
	snapshot := map[string]interface{}{
		"word_id":     s.wordID,
		"token":       word.Name,
		"language":    language,
		"audio_url":   word.AudioURL,
		"captured_at": time.Now().Format(time.RFC3339),
	}

	return snapshot, nil
}

// ReportError is the main public entry point for error reporting
func ReportError(ctx context.Context, errorType ErrorReportType, wordID int64, definitionID *int64, userID int64, quizDetails *QuizDetails) error {
	service, err := NewErrorReportService(ctx, errorType, wordID, definitionID, userID, quizDetails)
	if err != nil {
		return err
	}

	return service.ProcessReport()
}

// DeleteUserErrorReports deletes all error reports for a specific user
func DeleteUserErrorReports(userID int64) (int64, error) {
	db := common.GetDBConnection()

	repo := repository.NewErrorReportRepository(db)
	return repo.DeleteUserErrorReports(context.Background(), userID)
}

// CooldownError represents an error when a cooldown is active
type CooldownError struct {
	Message       string
	CooldownUntil time.Time
	RetryAfter    time.Duration
}

func (e CooldownError) Error() string {
	return e.Message
}
