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

// ErrorReportService encapsulates all error reporting functionality
type ErrorReportService struct {
	// Dependencies only
	db                     *pgxpool.Pool
	repo                   *repository.ErrorReportRepository
	definitionService      *DefinitionService
	wordService            *WordService
	leitnerTrackingService *LeitnerTrackingService
	jobService             JobService
	logger                 *slog.Logger
}

// NewErrorReportService creates a new error report service with dependencies
func NewErrorReportService(db *pgxpool.Pool, definitionService *DefinitionService, wordService *WordService, leitnerTrackingService *LeitnerTrackingService, jobService JobService) *ErrorReportService {
	return &ErrorReportService{
		db:                     db,
		repo:                   repository.NewErrorReportRepository(db),
		definitionService:      definitionService,
		wordService:            wordService,
		leitnerTrackingService: leitnerTrackingService,
		jobService:             jobService,
		logger:                 common.Logger,
	}
}

// ReportError is the main entry point for error reporting with request data as parameters
func (s *ErrorReportService) ReportError(ctx context.Context, errorType ErrorReportType, wordID int64, definitionID *int64, userID int64, quizDetails *QuizDetails) error {
	logger := s.logger.With("errorType", errorType, "wordID", wordID, "definitionID", definitionID, "userID", userID)

	// Log error report attempt for monitoring
	logger.Info("Error report attempt",
		"action", "error_report_attempt",
		"timestamp", time.Now().Unix(),
	)

	// Create request context
	requestCtx := &errorReportContext{
		ctx:          ctx,
		errorType:    errorType,
		wordID:       wordID,
		definitionID: definitionID,
		userID:       userID,
		quizDetails:  quizDetails,
		logger:       logger,
		service:      s,
	}

	// Validate user owns the word
	if err := requestCtx.validateUserOwnsWord(); err != nil {
		return err
	}

	// Check cooldown period
	if err := requestCtx.checkCooldown(); err != nil {
		return err
	}

	// Execute error report in transaction
	return requestCtx.executeTransaction()
}

// errorReportContext holds request-specific data for processing
type errorReportContext struct {
	ctx          context.Context
	errorType    ErrorReportType
	wordID       int64
	definitionID *int64
	userID       int64
	quizDetails  *QuizDetails
	logger       *slog.Logger
	service      *ErrorReportService
}

// validateUserOwnsWord ensures the user has permission to report errors for this word
func (ctx *errorReportContext) validateUserOwnsWord() error {
	isValid, err := ctx.service.definitionService.didUserCreateWord(ctx.wordID, ctx.userID)
	if err != nil || !isValid {
		if err != nil {
			ctx.logger.Error("validation failed", "error", err)
		}
		return errors.New("validation failed")
	}
	return nil
}

// checkCooldown verifies if the user is within the cooldown period for this error type
func (ctx *errorReportContext) checkCooldown() error {
	cooldownUntil, err := ctx.service.repo.CheckCooldown(ctx.ctx, ctx.userID, ctx.wordID, ctx.definitionID, string(ctx.errorType))
	if err != nil {
		ctx.logger.Error("failed to check cooldown", "error", err)
		return err
	}

	if cooldownUntil != nil {
		retryAfter := cooldownUntil.Sub(time.Now())
		ctx.logger.Warn("Error report blocked by cooldown",
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
func (ctx *errorReportContext) executeTransaction() error {
	tx, err := ctx.service.db.Begin(ctx.ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			if commitErr := tx.Commit(ctx.ctx); commitErr != nil {
				common.Logger.Error("failed to commit transaction in error reporting", "error", commitErr)
			}
		} else {
			if rollbackErr := tx.Rollback(ctx.ctx); rollbackErr != nil {
				common.Logger.Error("failed to rollback transaction in error reporting", "error", rollbackErr)
			}
		}
	}()

	// Process the error type and trigger appropriate workers
	report, err := ctx.processErrorType(tx)
	if err != nil {
		common.Logger.Error("failed to trigger jobs", "error", err)
		return err
	}

	// Handle post-processing steps
	return ctx.completeReport(tx, report)
}

// processErrorType handles the specific error type and triggers the appropriate workers
func (ctx *errorReportContext) processErrorType(tx pgx.Tx) (ErrorReport, error) {
	var report ErrorReport
	var err error

	switch ctx.errorType {
	case SoundNotPlaying:
		report = ErrorReport{WordId: &ctx.wordID, UserId: ctx.userID}
		_, err = ctx.service.jobService.ScheduleAudioJob(ctx.wordID, &ctx.userID, &report, &tx)

	case UnrelatedImage, MissingImage:
		if ctx.definitionID == nil {
			return report, fmt.Errorf("definition ID required for image-related errors")
		}
		report = ErrorReport{DefinitionId: ctx.definitionID, UserId: ctx.userID}
		_, err = ctx.service.jobService.ScheduleImageJob(*ctx.definitionID, &ctx.userID, &report, &tx)

	case UnrelatedExample, UnrelatedMeaning:
		err = ctx.service.definitionService.DeleteWordDefinitions(ctx.wordID, &tx)
		if err == nil {
			report = ErrorReport{WordId: &ctx.wordID, UserId: ctx.userID}
			_, err = ctx.service.jobService.ScheduleDefinitionJob(ctx.wordID, &ctx.userID, &report, &tx)
		}

	case ProcessingFailed:
		err = ctx.service.definitionService.DeleteWordDefinitions(ctx.wordID, &tx)
		if err == nil {
			report = ErrorReport{WordId: &ctx.wordID, UserId: ctx.userID}
			_, err = ctx.service.jobService.ScheduleDefinitionJob(ctx.wordID, &ctx.userID, &report, &tx)
		}

	default:
		err = fmt.Errorf("invalid error type %s", ctx.errorType)
	}

	return report, err
}

// completeReport handles all the post-processing steps after the main error processing
func (ctx *errorReportContext) completeReport(tx pgx.Tx, report ErrorReport) error {
	// Update last regenerated timestamp
	if err := ctx.updateLastRegeneratedTimestamp(tx); err != nil {
		ctx.logger.Error("failed to update last regenerated timestamp", "error", err)
		// Don't fail the whole operation for this
	}

	// Set cooldown for this error report
	if err := ctx.setCooldownPeriod(tx); err != nil {
		ctx.logger.Error("failed to set cooldown", "error", err)
		// Don't fail the whole operation for this
	}

	// Mark definition temporarily as unavailable
	if err := ctx.service.leitnerTrackingService.SetTemporarySkip(report, tx); err != nil {
		common.Logger.Error("failed to report error", "error", err)
		return err
	}

	// Build content snapshot before potential deletion
	contentSnapshot, finalDefinitionID, err := ctx.buildContentSnapshot(tx)
	if err != nil {
		ctx.logger.Error("failed to build content snapshot", "error", err)
		return err
	}

	// Upsert error report with content snapshot
	if err := ctx.service.repo.UpsertErrorReport(ctx.ctx, tx, ctx.userID, finalDefinitionID, ctx.wordID, string(ctx.errorType), contentSnapshot); err != nil {
		common.Logger.Error("failed to save error report", "error", err)
		return err
	}

	// Log successful error report for monitoring
	ctx.logger.Info("Error report processed successfully",
		"action", "error_report_success",
		"timestamp", time.Now().Unix(),
		"regeneration_type", ctx.errorType,
	)

	return nil
}

// updateLastRegeneratedTimestamp updates the timestamp for when content was last regenerated
func (ctx *errorReportContext) updateLastRegeneratedTimestamp(tx pgx.Tx) error {
	switch ctx.errorType {
	case SoundNotPlaying:
		return ctx.service.repo.UpdateWordAudioLastRegeneratedAt(ctx.ctx, tx, ctx.wordID)
	case UnrelatedImage, MissingImage, UnrelatedExample, UnrelatedMeaning:
		if ctx.definitionID != nil {
			return ctx.service.repo.UpdateDefinitionLastRegeneratedAt(ctx.ctx, tx, *ctx.definitionID)
		}
	}
	return nil
}

// setCooldownPeriod sets the cooldown period to prevent spam reporting
func (ctx *errorReportContext) setCooldownPeriod(tx pgx.Tx) error {
	cooldownDuration := time.Hour // 1 hour cooldown as agreed
	cooldownUntilTime := time.Now().Add(cooldownDuration)
	return ctx.service.repo.SetCooldown(ctx.ctx, tx, ctx.userID, ctx.wordID, ctx.definitionID, string(ctx.errorType), cooldownUntilTime)
}

// buildContentSnapshot creates a content snapshot and determines final foreign key strategy
func (ctx *errorReportContext) buildContentSnapshot(tx pgx.Tx) (map[string]interface{}, *int64, error) {
	var snapshot map[string]any
	var finalDefinitionID *int64

	switch ctx.errorType {
	case UnrelatedMeaning, UnrelatedExample:
		// These will delete definitions, so capture full snapshot and nullify foreign key
		if ctx.definitionID == nil {
			return nil, nil, fmt.Errorf("definition ID required for %s", ctx.errorType)
		}

		defSnapshot, err := ctx.fetchCompleteDefinitionSnapshot(*ctx.definitionID)
		if err != nil {
			return nil, nil, err
		}

		snapshot = defSnapshot

		// Add quiz details if available
		if ctx.quizDetails != nil {
			snapshot["quiz_context"] = map[string]interface{}{
				"quiz_type":    ctx.quizDetails.QuizType,
				"value":        ctx.quizDetails.Value,
				"options":      ctx.quizDetails.Options,
				"answer_index": ctx.quizDetails.AnswerIndex,
				"context":      ctx.quizDetails.Context,
				"captured_at":  time.Now().Format(time.RFC3339),
			}
		}

		finalDefinitionID = nil // Set to NULL since definition will be deleted

	case UnrelatedImage, MissingImage:
		// These regenerate images but keep definitions, so preserve foreign key
		if ctx.definitionID == nil {
			return nil, nil, fmt.Errorf("definition ID required for %s", ctx.errorType)
		}

		defSnapshot, err := ctx.fetchCompleteDefinitionSnapshot(*ctx.definitionID)
		if err != nil {
			return nil, nil, err
		}

		snapshot = defSnapshot

		// Add quiz details if available
		if ctx.quizDetails != nil {
			snapshot["quiz_context"] = map[string]interface{}{
				"quiz_type":    ctx.quizDetails.QuizType,
				"value":        ctx.quizDetails.Value,
				"options":      ctx.quizDetails.Options,
				"answer_index": ctx.quizDetails.AnswerIndex,
				"context":      ctx.quizDetails.Context,
				"captured_at":  time.Now().Format(time.RFC3339),
			}
		}

		finalDefinitionID = ctx.definitionID // Keep foreign key

	case SoundNotPlaying:
		// This regenerates audio but keeps word, so preserve foreign key
		wordSnapshot, err := ctx.fetchWordSnapshot(tx)
		if err != nil {
			return nil, nil, err
		}

		snapshot = wordSnapshot

		// Add quiz details if available
		if ctx.quizDetails != nil {
			snapshot["quiz_context"] = map[string]interface{}{
				"quiz_type":    ctx.quizDetails.QuizType,
				"value":        ctx.quizDetails.Value,
				"options":      ctx.quizDetails.Options,
				"answer_index": ctx.quizDetails.AnswerIndex,
				"context":      ctx.quizDetails.Context,
				"captured_at":  time.Now().Format(time.RFC3339),
			}
		}

		finalDefinitionID = ctx.definitionID // Keep foreign key (might be nil for word-level errors)

	case ProcessingFailed:
		// This is a processing retry, no content quality issue
		snapshot = make(map[string]interface{})

		// Add quiz details if available, even for processing failures
		if ctx.quizDetails != nil {
			snapshot["quiz_context"] = map[string]interface{}{
				"quiz_type":    ctx.quizDetails.QuizType,
				"value":        ctx.quizDetails.Value,
				"options":      ctx.quizDetails.Options,
				"answer_index": ctx.quizDetails.AnswerIndex,
				"context":      ctx.quizDetails.Context,
				"captured_at":  time.Now().Format(time.RFC3339),
			}
		} else {
			// If no quiz details, keep the snapshot nil as before
			snapshot = nil
		}

		finalDefinitionID = nil // Set to NULL since definition will be deleted

	default:
		return nil, nil, fmt.Errorf("unsupported error type: %s", ctx.errorType)
	}

	return snapshot, finalDefinitionID, nil
}

// fetchCompleteDefinitionSnapshot fetches complete definition data for historical preservation
func (ctx *errorReportContext) fetchCompleteDefinitionSnapshot(definitionID int64) (map[string]interface{}, error) {
	// Use definition service to fetch the definition
	definition, err := ctx.service.definitionService.GetDefinitionByID(definitionID)
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
func (ctx *errorReportContext) fetchWordSnapshot(tx pgx.Tx) (map[string]interface{}, error) {
	word, err := ctx.service.wordService.GetWordByID(ctx.wordID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch word: %w", err)
	}
	if word == nil {
		return nil, fmt.Errorf("word not found with ID: %d", ctx.wordID)
	}

	// Get language from wordlist - we still need this query since word doesn't contain language directly
	var language string
	err = tx.QueryRow(ctx.ctx, `
		SELECT wl.language_code
		FROM wordlists wl
		WHERE wl.id = $1
	`, word.WordlistID).Scan(&language)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch wordlist language: %w", err)
	}

	// Build word snapshot
	snapshot := map[string]interface{}{
		"word_id":     ctx.wordID,
		"token":       word.Name,
		"language":    language,
		"audio_url":   word.AudioURL,
		"captured_at": time.Now().Format(time.RFC3339),
	}

	return snapshot, nil
}

// DeleteUserErrorReports deletes all error reports for a specific user
func (s *ErrorReportService) DeleteUserErrorReports(userID int64) (int64, error) {
	return s.repo.DeleteUserErrorReports(context.Background(), userID)
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
