package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
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
	isValid, err := ctx.service.definitionService.didUserCreateWord(ctx.ctx, ctx.wordID, ctx.userID)
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
	return ctx.cooldownResult(cooldownUntil, err)
}

func (ctx *errorReportContext) checkCooldownTx(tx pgx.Tx) error {
	cooldownUntil, err := ctx.service.repo.CheckCooldownTx(ctx.ctx, tx, ctx.userID, ctx.wordID, ctx.definitionID, string(ctx.errorType))
	return ctx.cooldownResult(cooldownUntil, err)
}

func (ctx *errorReportContext) cooldownResult(cooldownUntil *time.Time, err error) error {
	if err != nil {
		ctx.logger.Error("failed to check cooldown", "error", err)
		return err
	}

	if cooldownUntil != nil {
		retryAfter := time.Until(*cooldownUntil)
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
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx.ctx)) }()

	if validationErr := ctx.lockAndValidateTarget(tx); validationErr != nil {
		return validationErr
	}
	if cooldownErr := ctx.checkCooldownTx(tx); cooldownErr != nil {
		return cooldownErr
	}

	// Process the error type and trigger appropriate workers
	report, err := ctx.processErrorType(tx)
	if err != nil {
		common.Logger.Error("failed to trigger jobs", "error", err)
		return err
	}

	// Handle post-processing steps
	if err := ctx.completeReport(tx, report); err != nil {
		return err
	}
	if err := tx.Commit(ctx.ctx); err != nil {
		return fmt.Errorf("failed to commit error report transaction: %w", err)
	}
	return nil
}

func (ctx *errorReportContext) lockAndValidateTarget(tx pgx.Tx) error {
	var lockedWordID int64
	if err := tx.QueryRow(ctx.ctx, `
		SELECT id FROM words
		WHERE id=$1 AND user_id=$2
		FOR UPDATE
	`, ctx.wordID, ctx.userID).Scan(&lockedWordID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("validation failed")
		}
		return fmt.Errorf("lock reported word: %w", err)
	}

	requiresDefinition := ctx.errorType == UnrelatedImage || ctx.errorType == MissingImage ||
		ctx.errorType == UnrelatedMeaning || ctx.errorType == UnrelatedExample
	if !requiresDefinition {
		return nil
	}
	if ctx.definitionID == nil {
		return fmt.Errorf("definition ID required for %s", ctx.errorType)
	}
	var lockedDefinitionID int64
	if err := tx.QueryRow(ctx.ctx, `
		SELECT d.id
		FROM word_definitions wd
		JOIN definitions d ON d.id=wd.definition_id
		WHERE wd.word_id=$1 AND wd.definition_id=$2
		FOR UPDATE OF wd, d
	`, ctx.wordID, *ctx.definitionID).Scan(&lockedDefinitionID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("validation failed")
		}
		return fmt.Errorf("lock reported definition relationship: %w", err)
	}
	return nil
}

// processErrorType handles the specific error type and triggers the appropriate workers
func (ctx *errorReportContext) processErrorType(tx pgx.Tx) (ErrorReport, error) {
	var report ErrorReport
	var err error

	switch ctx.errorType {
	case SoundNotPlaying:
		report = ErrorReport{WordID: &ctx.wordID, UserID: ctx.userID, ErrorType: string(ctx.errorType)}
		_, err = ctx.service.jobService.ScheduleAudioJob(ctx.ctx, ctx.wordID, &ctx.userID, &report, &tx)

	case UnrelatedImage, MissingImage:
		if ctx.definitionID == nil {
			return report, fmt.Errorf("definition ID required for image-related errors")
		}
		report = ErrorReport{DefinitionID: ctx.definitionID, WordID: &ctx.wordID, UserID: ctx.userID, ErrorType: string(ctx.errorType)}
		_, err = ctx.service.jobService.ScheduleImageJob(ctx.ctx, *ctx.definitionID, &ctx.userID, &report, &tx)

	case UnrelatedExample, UnrelatedMeaning:
		report = ErrorReport{DefinitionID: ctx.definitionID, WordID: &ctx.wordID, UserID: ctx.userID, ErrorType: string(ctx.errorType)}
		_, err = ctx.service.jobService.ScheduleDefinitionJob(ctx.ctx, ctx.wordID, &ctx.userID, &report, &tx)

	case ProcessingFailed:
		report = ErrorReport{WordID: &ctx.wordID, UserID: ctx.userID, ErrorType: string(ctx.errorType)}
		_, err = ctx.service.jobService.ScheduleDefinitionJob(ctx.ctx, ctx.wordID, &ctx.userID, &report, &tx)

	default:
		err = fmt.Errorf("invalid error type %s", ctx.errorType)
	}

	return report, err
}

// completeReport handles all the post-processing steps after the main error processing
func (ctx *errorReportContext) completeReport(tx pgx.Tx, report ErrorReport) error {
	// Update last regenerated timestamp
	if err := ctx.updateLastRegeneratedTimestamp(tx); err != nil {
		return fmt.Errorf("update last regenerated timestamp: %w", err)
	}

	// Set cooldown for this error report
	if err := ctx.setCooldownPeriod(tx); err != nil {
		return fmt.Errorf("set error report cooldown: %w", err)
	}

	// Mark definition temporarily as unavailable
	if err := ctx.service.leitnerTrackingService.SetTemporarySkip(ctx.ctx, report, tx); err != nil {
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
		// Capture the reported content while retaining the original relationship until
		// the replacement transaction commits.
		if ctx.definitionID == nil {
			return nil, nil, fmt.Errorf("definition ID required for %s", ctx.errorType)
		}

		defSnapshot, err := ctx.fetchCompleteDefinitionSnapshot(tx, *ctx.definitionID)
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

		finalDefinitionID = ctx.definitionID // Preserve the reported definition until replacement commits.

	case UnrelatedImage, MissingImage:
		// These regenerate images but keep definitions, so preserve foreign key
		if ctx.definitionID == nil {
			return nil, nil, fmt.Errorf("definition ID required for %s", ctx.errorType)
		}

		defSnapshot, err := ctx.fetchCompleteDefinitionSnapshot(tx, *ctx.definitionID)
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

		finalDefinitionID = nil // Processing failures are word-level reports.

	default:
		return nil, nil, fmt.Errorf("unsupported error type: %s", ctx.errorType)
	}

	return snapshot, finalDefinitionID, nil
}

// fetchCompleteDefinitionSnapshot fetches complete definition data for historical preservation
func (ctx *errorReportContext) fetchCompleteDefinitionSnapshot(tx pgx.Tx, definitionID int64) (map[string]interface{}, error) {
	var definition model.Definition
	if err := tx.QueryRow(ctx.ctx, `
		SELECT token, language, part_of_speech, is_verb_type, meaning,
			examples, inflections, source
		FROM definitions WHERE id=$1
	`, definitionID).Scan(
		&definition.Token, &definition.Language, &definition.PartOfSpeech,
		&definition.IsVerbType, &definition.Meaning, &definition.Examples,
		&definition.Inflections, &definition.Source,
	); err != nil {
		return nil, fmt.Errorf("failed to fetch definition snapshot: %w", err)
	}

	// Build comprehensive snapshot
	snapshot := map[string]interface{}{
		"definition_id":  definitionID,
		"meaning":        definition.Meaning,
		"part_of_speech": definition.PartOfSpeech,
		"source":         definition.Source,
		"language":       definition.Language,
		"token":          definition.Token,
		"examples":       definition.GetAllExamples(),
		"captured_at":    time.Now().Format(time.RFC3339),
	}

	return snapshot, nil
}

// fetchWordSnapshot fetches word data for audio-related errors
func (ctx *errorReportContext) fetchWordSnapshot(tx pgx.Tx) (map[string]interface{}, error) {
	var token, language, audioURL string
	if err := tx.QueryRow(ctx.ctx, `
		SELECT w.name, wl.language_code, COALESCE(w.audio_url, '')
		FROM words w
		JOIN wordlists wl ON wl.id=w.wordlist_id
		WHERE w.id=$1 AND w.user_id=$2
	`, ctx.wordID, ctx.userID).Scan(&token, &language, &audioURL); err != nil {
		return nil, fmt.Errorf("failed to fetch word snapshot: %w", err)
	}

	// Build word snapshot
	snapshot := map[string]interface{}{
		"word_id":     ctx.wordID,
		"token":       token,
		"language":    language,
		"audio_url":   audioURL,
		"captured_at": time.Now().Format(time.RFC3339),
	}

	return snapshot, nil
}

// DeleteUserErrorReports deletes all error reports for a specific user
func (s *ErrorReportService) DeleteUserErrorReports(ctx context.Context, userID int64) (int64, error) {
	return s.repo.DeleteUserErrorReports(ctx, userID)
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
