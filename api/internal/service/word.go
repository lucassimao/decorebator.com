package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	repo "decorebator.com/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Word = model.Word

// WordService handles word-related operations with dependency injection
type WordService struct {
	repository             *repo.WordRepository
	moderationService      ModerationService
	definitionService      *DefinitionService
	jobService             JobService
	leitnerTrackingService *LeitnerTrackingService
}

// NewWordService creates a new word service with dependencies
func NewWordService(db *pgxpool.Pool, definitionService *DefinitionService, moderationService ModerationService, jobService JobService, leitnerTrackingService *LeitnerTrackingService) *WordService {
	return &WordService{
		repository:             &repo.WordRepository{Db: db},
		moderationService:      moderationService,
		definitionService:      definitionService,
		jobService:             jobService,
		leitnerTrackingService: leitnerTrackingService,
	}
}

// GetWordByWordlist returns words from wordlist with optional filtering
func (ws *WordService) GetWordByWordlist(wordlistID, userID int64, onlyWithDefinitions bool) ([]Word, error) {
	return ws.repository.GetWordsByWordlist(wordlistID, userID, onlyWithDefinitions)
}

func (ws *WordService) GetWordByID(id int64) (*Word, error) {
	return ws.repository.GetByID(id)
}

// UpdateProcessingStatus updates the processing status and related fields for a word
func (ws *WordService) UpdateProcessingStatus(wordID int64, status string, errorMsg string, tx *pgx.Tx) error {
	updates := map[string]interface{}{
		"processing_status": status,
	}
	if status == "failed" && errorMsg != "" {
		updates["processing_error"] = errorMsg
	}
	return ws.repository.UpdateFields(wordID, updates, tx)
}

// GetWordlistLanguageAndPronunciation retrieves the language code and pronunciation system for a wordlist from a word ID
func (ws *WordService) GetWordlistLanguageAndPronunciation(wordID int64) (string, model.PronunciationSystem, error) {
	query := `
		SELECT w.language_code, w.pronunciation_system
		FROM words wd 
		JOIN wordlists w ON wd.wordlist_id = w.id 
		WHERE wd.id = $1`

	var languageCode string
	var pronunciationSystem model.PronunciationSystem
	err := ws.repository.Db.QueryRow(context.Background(), query, wordID).Scan(&languageCode, &pronunciationSystem)
	if err != nil {
		return "", "", err
	}

	return languageCode, pronunciationSystem, nil
}

// UpdatePronunciation updates the pronunciation for a word
func (ws *WordService) UpdatePronunciation(wordID int64, pronunciation string, tx *pgx.Tx) error {
	updates := map[string]interface{}{
		"pronunciation": pronunciation,
	}
	return ws.repository.UpdateFields(wordID, updates, tx)
}

// UpdateWordFields provides a flexible way to update specific word fields
func (ws *WordService) UpdateWordFields(wordID int64, updates map[string]interface{}, tx *pgx.Tx) error {
	if len(updates) == 0 {
		return fmt.Errorf("no fields specified for update")
	}
	return ws.repository.UpdateFields(wordID, updates, tx)
}

func (ws *WordService) SaveWord(ctx context.Context, dto *Word) (*Word, error) {
	var lowerCasedName = strings.ToLower(dto.Name)
	var trimmedName = strings.TrimSpace(lowerCasedName)

	// count runes (Unicode characters), not bytes
	if utf8.RuneCountInString(trimmedName) > 15 {
		return nil, common.BusinessError{Message: "words must be limited to 15 chars"}
	}

	// Validate word content using moderation service
	filterResult := ws.moderationService.Validate(trimmedName)
	if !filterResult.IsAppropriate {
		return nil, common.BusinessError{
			Message: fmt.Sprintf("Word content not appropriate: %s", filterResult.Reason),
		}
	}

	// Validate notes content if provided
	if dto.Notes != "" {
		notesResult := ws.moderationService.Validate(dto.Notes)
		if !notesResult.IsAppropriate {
			return nil, common.BusinessError{
				Message: fmt.Sprintf("Word notes not appropriate: %s", notesResult.Reason),
			}
		}
	}

	tx, err := ws.repository.Db.Begin(ctx)

	if err != nil {
		return nil, err
	}

	// db transaction mgmt
	defer func() {
		if err == nil {
			if commitErr := tx.Commit(ctx); commitErr != nil {
				common.Logger.Error("failed to commit transaction", "error", commitErr)
			}
		} else {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				common.Logger.Error("failed to rollback transaction", "error", rollbackErr)
			}
		}
	}()

	word, err := ws.repository.Save(trimmedName, dto.Notes, dto.UserID, dto.WordlistID, &tx)
	if err != nil {
		return nil, err
	}

	// check if there are definitions for this word already
	definitions, _ := ws.definitionService.findDefinitionsByName(word.Name)

	if len(definitions) > 0 {
		definitionIds := []int64{}

		for _, def := range definitions {
			definitionIds = append(definitionIds, def.ID)
		}
		if reuseErr := ws.repository.ReuseDefinitions(word.ID, definitionIds, tx); reuseErr != nil {
			common.Logger.Error("failed to reuse definitions", "wordId", word.ID, "error", reuseErr)
		}

		if includeErr := ws.leitnerTrackingService.IncludeDefinitions(word.ID, word.UserID, definitionIds, tx); includeErr != nil {
			common.Logger.Error("failed to include definitions in tracking system", "wordId", word.ID, "error", includeErr)
		}

		var latestAudioURL string
		latestAudioURL, err = ws.repository.GetLatestAudioURL(trimmedName)

		if err != nil {
			_, _ = ws.jobService.ScheduleAudioJob(word.ID, &word.UserID, nil, &tx)
			err = nil // fine if triggering the worker fails somehow
		} else {
			word.AudioURL = latestAudioURL
			err = ws.UpdateWord(word, &tx)
		}
	} else {
		_, _ = ws.jobService.ScheduleDefinitionJob(word.ID, &word.UserID, nil, &tx)
		_, _ = ws.jobService.ScheduleAudioJob(word.ID, &word.UserID, nil, &tx)
	}

	if err != nil {
		return nil, err
	}

	return word, nil
}

func (ws *WordService) DeleteWord(id, userID int64) (int64, error) {
	word, err := ws.GetWordByID(id)
	if err != nil {
		return 0, err
	}

	if word == nil {
		return 0, common.NotFoundError{ID: id, Entity: "Word"}
	}

	count, err := ws.repository.Delete(userID, id)
	if err != nil {
		common.Logger.Error("failed to delete word", "error", err)
		return 0, errors.New("failed to delete word")
	}

	return count, nil
}

func (ws *WordService) UpdateWord(word *Word, tx *pgx.Tx) error {
	count, err := ws.repository.Update(word, tx)
	if err != nil {
		return err
	}

	if count == 0 {
		return common.NotFoundError{ID: word.ID, Entity: "Word"}
	}
	return nil
}
