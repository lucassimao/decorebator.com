package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
)

// PublicQuizService encapsulates business logic for building public quiz questions.
type PublicQuizService struct {
	repo          *repository.PublicQuizRepository
	definitionSvc *DefinitionService
	wordlistSvc   *WordlistService
	jobs          JobService
}

func NewPublicQuizService(repo *repository.PublicQuizRepository, definitionSvc *DefinitionService, wordlistSvc *WordlistService, jobs JobService) *PublicQuizService {
	return &PublicQuizService{repo: repo, definitionSvc: definitionSvc, wordlistSvc: wordlistSvc, jobs: jobs}
}

// maskBracketedContent replaces any content wrapped in square brackets with underscores.
// Example: "They asked [wherefore] she had chosen to leave." -> "They asked _____ she had chosen to leave."
var bracketRegex = regexp.MustCompile(`\[[^\]]+\]`)

func maskBracketedContent(s string) string {
	if s == "" {
		return s
	}
	return bracketRegex.ReplaceAllString(s, "_____")
}

// BuildQuestionsForSlug generates a stable list of quiz questions for a given slug.
//
//nolint:gocyclo // Intentional complexity to combine selection and type rotation in one pass
func (s *PublicQuizService) BuildQuestionsForSlug(ctx context.Context, slug string, maxQuestions int) ([]model.Quiz, error) {
	if slug == "" {
		return nil, common.BusinessError{Message: "missing slug"}
	}

	onlyActive := true
	limit := 1
	quizzes, err := s.repo.Find(ctx, repository.FindPublicQuizArgs{Slug: &slug, OnlyActive: &onlyActive, Limit: &limit})
	if err != nil {
		common.Logger.Error("failed to find public quiz", "error", err, "slug", slug)
		return nil, err
	}
	if len(quizzes) == 0 {
		return nil, common.NotFoundError{Entity: "public quiz"}
	}
	pq := quizzes[0]

	// Fetch definitions from repository via service
	defRows, qerr := s.definitionSvc.GetDefinitionsForPublicQuiz(ctx, pq.WordlistID)
	if qerr != nil {
		common.Logger.Error("failed to load definitions for public quiz", "error", qerr)
		return nil, qerr
	}
	if len(defRows) == 0 {
		return []model.Quiz{}, nil
	}

	defs := defRows

	var questions []model.Quiz
	usedWordIDs := map[int64]struct{}{}
	usedByTypeAndDef := map[string]struct{}{}

	type builderFn func(i int, d repository.DefinitionForPublicQuiz) (*model.Quiz, bool)

	gmBuilder := func(i int, d repository.DefinitionForPublicQuiz) (*model.Quiz, bool) {
		distractors, derr := s.definitionSvc.GetRandomMeanings(ctx, []int{int(d.ID)}, 3)
		if derr != nil || len(distractors) < 3 || d.Meaning == "" {
			return nil, false
		}
		opts := append([]string{d.Meaning}, distractors...)
		rotateBy := i % len(opts)
		if rotateBy > 0 {
			opts = append(opts[rotateBy:], opts[:rotateBy]...)
		}
		answerIndex := 0
		for idx, v := range opts {
			if v == d.Meaning {
				answerIndex = idx
				break
			}
		}
		q := model.Quiz{
			Value:        d.Token,
			Options:      opts,
			AnswerIndex:  answerIndex,
			ID:           d.ID,
			Type:         model.GuessMeaning,
			PartOfSpeech: d.PartOfSpeech,
			IsVerbType:   d.IsVerbType,
			DefinitionID: d.ID,
			WordID:       d.WordID,
		}
		return &q, true
	}

	wfmBuilder := func(i int, d repository.DefinitionForPublicQuiz) (*model.Quiz, bool) {
		distractors, derr := s.definitionSvc.GetRandomTokens(ctx, []int{int(d.ID)}, d.PartOfSpeech, 3)
		if derr != nil || len(distractors) < 3 || d.Token == "" {
			return nil, false
		}
		opts := append([]string{d.Token}, distractors...)
		rotateBy := (i + 1) % len(opts)
		if rotateBy > 0 {
			opts = append(opts[rotateBy:], opts[:rotateBy]...)
		}
		answerIndex := 0
		for idx, v := range opts {
			if v == d.Token {
				answerIndex = idx
				break
			}
		}
		q := model.Quiz{
			Value:        d.Meaning,
			Options:      opts,
			AnswerIndex:  answerIndex,
			ID:           d.ID,
			Type:         model.WordFromMeaning,
			PartOfSpeech: d.PartOfSpeech,
			IsVerbType:   d.IsVerbType,
			DefinitionID: d.ID,
			WordID:       d.WordID,
		}
		return &q, true
	}

	wfiBuilder := func(i int, d repository.DefinitionForPublicQuiz) (*model.Quiz, bool) {
		if d.ImageDescription == "" || d.ImageURL == "" {
			return nil, false
		}
		distractors, derr := s.definitionSvc.GetRandomTokens(ctx, []int{int(d.ID)}, d.PartOfSpeech, 3)
		if derr != nil || len(distractors) < 3 {
			return nil, false
		}
		opts := append([]string{d.Token}, distractors...)
		rotateBy := (i + 2) % len(opts)
		if rotateBy > 0 {
			opts = append(opts[rotateBy:], opts[:rotateBy]...)
		}
		answerIndex := 0
		for idx, v := range opts {
			if v == d.Token {
				answerIndex = idx
				break
			}
		}
		// Mask any bracketed content in description to avoid revealing the answer
		desc := maskBracketedContent(d.ImageDescription)
		q := model.Quiz{
			Value:            d.ImageURL,
			Options:          opts,
			AnswerIndex:      answerIndex,
			ID:               d.ID,
			Type:             model.WordFromImage,
			PartOfSpeech:     d.PartOfSpeech,
			IsVerbType:       d.IsVerbType,
			DefinitionID:     d.ID,
			WordID:           d.WordID,
			ImageDescription: desc,
		}
		return &q, true
	}

	wfaBuilder := func(i int, d repository.DefinitionForPublicQuiz) (*model.Quiz, bool) {
		if d.WordAudio == "" {
			return nil, false
		}
		distractors, derr := s.definitionSvc.GetRandomTokens(ctx, []int{int(d.ID)}, d.PartOfSpeech, 3)
		if derr != nil || len(distractors) < 3 {
			return nil, false
		}
		opts := append([]string{d.Token}, distractors...)
		rotateBy := (i + 3) % len(opts)
		if rotateBy > 0 {
			opts = append(opts[rotateBy:], opts[:rotateBy]...)
		}
		answerIndex := 0
		for idx, v := range opts {
			if v == d.Token {
				answerIndex = idx
				break
			}
		}
		q := model.Quiz{
			Value:        d.Token,
			Options:      opts,
			AnswerIndex:  answerIndex,
			ID:           d.ID,
			Type:         model.WordFromAudio,
			PartOfSpeech: d.PartOfSpeech,
			IsVerbType:   d.IsVerbType,
			DefinitionID: d.ID,
			WordID:       d.WordID,
			AudioURL:     d.WordAudio,
		}
		return &q, true
	}

	wfeaBuilder := func(i int, d repository.DefinitionForPublicQuiz) (*model.Quiz, bool) {
		if len(d.ExampleAudioJSON) == 0 {
			return nil, false
		}
		var exArr []struct {
			ExampleText string `json:"exampleText"`
			AudioURL    string `json:"audioUrl"`
		}
		_ = json.Unmarshal(d.ExampleAudioJSON, &exArr)
		if len(exArr) == 0 || exArr[0].AudioURL == "" {
			return nil, false
		}
		distractors, derr := s.definitionSvc.GetRandomTokens(ctx, []int{int(d.ID)}, d.PartOfSpeech, 3)
		if derr != nil || len(distractors) < 3 {
			return nil, false
		}
		opts := append([]string{d.Token}, distractors...)
		rotateBy := (i + 4) % len(opts)
		if rotateBy > 0 {
			opts = append(opts[rotateBy:], opts[:rotateBy]...)
		}
		answerIndex := 0
		for idx, v := range opts {
			if v == d.Token {
				answerIndex = idx
				break
			}
		}
		q := model.Quiz{
			Value:        maskBracketedContent(exArr[0].ExampleText),
			Options:      opts,
			AnswerIndex:  answerIndex,
			ID:           d.ID,
			Type:         model.WordFromExampleAudio,
			PartOfSpeech: d.PartOfSpeech,
			IsVerbType:   d.IsVerbType,
			DefinitionID: d.ID,
			WordID:       d.WordID,
			AudioURL:     exArr[0].AudioURL,
		}
		return &q, true
	}

	wwdBuilder := func(_ int, d repository.DefinitionForPublicQuiz) (*model.Quiz, bool) {
		if d.Meaning == "" || d.Token == "" {
			return nil, false
		}
		q := model.Quiz{
			Value:        d.Meaning,
			Options:      []string{d.Token},
			AnswerIndex:  0,
			ID:           d.ID,
			Type:         model.WriteWordFromDefinition,
			PartOfSpeech: d.PartOfSpeech,
			IsVerbType:   d.IsVerbType,
			DefinitionID: d.ID,
			WordID:       d.WordID,
		}
		return &q, true
	}

	csBuilder := func(i int, d repository.DefinitionForPublicQuiz) (*model.Quiz, bool) {
		if d.Meaning == "" || d.Token == "" {
			return nil, false
		}
		distractors, derr := s.definitionSvc.GetRandomTokens(ctx, []int{int(d.ID)}, d.PartOfSpeech, 3)
		if derr != nil || len(distractors) < 3 {
			return nil, false
		}
		opts := append([]string{d.Token}, distractors...)
		rotateBy := (i + 5) % len(opts)
		if rotateBy > 0 {
			opts = append(opts[rotateBy:], opts[:rotateBy]...)
		}
		answerIndex := 0
		for idx, v := range opts {
			if v == d.Token {
				answerIndex = idx
				break
			}
		}
		q := model.Quiz{
			Value:        "_____ — " + d.Meaning,
			Options:      opts,
			AnswerIndex:  answerIndex,
			ID:           d.ID,
			Type:         model.CompleteSentence,
			PartOfSpeech: d.PartOfSpeech,
			IsVerbType:   d.IsVerbType,
			DefinitionID: d.ID,
			WordID:       d.WordID,
		}
		return &q, true
	}

	typeBuilders := []builderFn{gmBuilder, wfmBuilder, wfiBuilder, wfaBuilder, wfeaBuilder, wwdBuilder, csBuilder}

	// Phase 1: Ensure coverage — at least one question per type if possible,
	// even if this means reusing the same word across different types.
	typesCovered := map[model.QuizType]bool{}
	for tbIdx, builder := range typeBuilders {
		if len(questions) >= maxQuestions {
			break
		}
		// Scan all definitions to find the first that can build this type
		// Use a synthetic index to maintain option rotation determinism
		built := false
		for di, d := range defs {
			q, ok := builder(tbIdx+di, d)
			if !ok || q == nil {
				continue
			}
			// Avoid duplicate same (definition, type) pairs
			key := fmt.Sprintf("%d:%s", q.DefinitionID, string(q.Type))
			if _, seen := usedByTypeAndDef[key]; seen {
				continue
			}
			questions = append(questions, *q)
			usedByTypeAndDef[key] = struct{}{}
			typesCovered[q.Type] = true
			built = true
			break
		}
		// If not buildable for this type, continue — data may be missing (e.g., no image/audio)
		_ = built
	}

	if maxQuestions <= 0 || maxQuestions > 100 {
		maxQuestions = 100
	}

	for i, d := range defs {
		if len(questions) >= maxQuestions {
			break
		}
		// Prefer not to repeat words, but allow repeats if needed to reach max or increase variety
		_, wordUsed := usedWordIDs[d.WordID]

		// Rotate starting point to diversify types
		start := i % len(typeBuilders)
		var picked *model.Quiz
		for k := 0; k < len(typeBuilders); k++ {
			idx := (start + k) % len(typeBuilders)
			if q, ok := typeBuilders[idx](i, d); ok && q != nil {
				// Skip if we already created the same (definition, type) question
				key := fmt.Sprintf("%d:%s", q.DefinitionID, string(q.Type))
				if _, seen := usedByTypeAndDef[key]; seen {
					continue
				}
				picked = q
				break
			}
		}
		if picked != nil {
			questions = append(questions, *picked)
			usedByTypeAndDef[fmt.Sprintf("%d:%s", picked.DefinitionID, string(picked.Type))] = struct{}{}
			if !wordUsed {
				usedWordIDs[d.WordID] = struct{}{}
			}
		}
	}

	return questions, nil
}

// GetBySlug returns active public quiz metadata
func (s *PublicQuizService) GetBySlug(ctx context.Context, slug string) (*model.PublicQuiz, error) {
	if slug == "" {
		return nil, common.BusinessError{Message: "missing slug"}
	}
	onlyActive := true
	limit := 1
	res, err := s.repo.Find(ctx, repository.FindPublicQuizArgs{Slug: &slug, OnlyActive: &onlyActive, Limit: &limit})
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, common.NotFoundError{Entity: "public quiz"}
	}
	return res[0], nil
}

// GetLeaderboardTop returns top entries for a quiz by slug
// The returned entries use camelCase JSON via model.LeaderboardEntry
func (s *PublicQuizService) GetLeaderboardTop(ctx context.Context, slug string, top int) ([]model.LeaderboardEntry, error) {
	onlyActive := true
	limit := 1
	res, err := s.repo.Find(ctx, repository.FindPublicQuizArgs{Slug: &slug, OnlyActive: &onlyActive, Limit: &limit})
	if err != nil || len(res) == 0 {
		if err != nil {
			return nil, err
		}
		return nil, common.NotFoundError{Entity: "public quiz"}
	}
	pq := res[0]
	rows, rerr := s.repo.GetLeaderboardTopByQuizID(ctx, pq.ID, top)
	if rerr != nil {
		return nil, rerr
	}
	out := make([]model.LeaderboardEntry, 0, len(rows))
	for idx, r := range rows {
		var guestNamePtr *string
		if r.GuestName != "" {
			// capture range variable to new var to take address safely
			n := r.GuestName
			guestNamePtr = &n
		}
		out = append(out, model.LeaderboardEntry{
			QuizSlug:              pq.Slug,
			QuizTitle:             pq.Title,
			PlayerID:              nil,
			GuestName:             guestNamePtr,
			ScorePercentage:       r.ScorePercentage,
			CompletionTimeSeconds: r.CompletionTimeSeconds,
			Rank:                  idx + 1,
		})
	}
	return out, nil
}

// RecordAttempt stores an attempt and updates aggregates
func (s *PublicQuizService) RecordAttempt(ctx context.Context, slug string, tried, correct int, name *string, email *string, durationSeconds *int) error {
	onlyActive := true
	limit := 1
	res, err := s.repo.Find(ctx, repository.FindPublicQuizArgs{Slug: &slug, OnlyActive: &onlyActive, Limit: &limit})
	if err != nil || len(res) == 0 {
		if err != nil {
			return err
		}
		return common.NotFoundError{Entity: "public quiz"}
	}
	pq := res[0]
	pct := 0.0
	if tried > 0 {
		pct = float64(correct) / float64(tried) * 100.0
	}
	dur := 0
	if durationSeconds != nil && *durationSeconds > 0 {
		dur = *durationSeconds
	}
	if err := s.repo.UpdateAggregatesForAttempt(ctx, pq.ID, pct); err != nil {
		return err
	}
	return s.repo.InsertQuizAttempt(ctx, pq.ID, name, email, pct, pct, dur)
}

// PublishPublicQuiz ensures ownership, validates settings, and either reactivates/updates
// an existing quiz or creates a new one. Returns the immutable slug.
func (s *PublicQuizService) PublishPublicQuiz(ctx context.Context, wordlistID int64, userID int64, input model.PublishQuizDTO) (string, error) {
	if err := input.Validate(); err != nil {
		return "", common.BusinessError{Message: err.Error()}
	}
	// Verify ownership
	wl, err := s.wordlistSvc.GetWordlistByID(ctx, wordlistID, userID)
	if err != nil {
		var notFound common.NotFoundError
		if errors.As(err, &notFound) {
			return "", notFound
		}
		return "", err
	}
	if wl == nil {
		return "", common.NotFoundError{Entity: "wordlist"}
	}

	limit := 1
	existingList, findErr := s.repo.Find(ctx, repository.FindPublicQuizArgs{
		WordlistID: &wordlistID,
		CreatorID:  &userID,
		Limit:      &limit,
	})
	if findErr != nil {
		return "", findErr
	}
	if len(existingList) > 0 {
		existing := existingList[0]
		var desc string
		if input.Description != nil {
			desc = *input.Description
		}
		if updateErr := s.repo.ReactivateAndUpdatePublicQuiz(ctx, existing.ID, input.Title, desc, input.Difficulty, input.TimeLimitMinutes); updateErr != nil {
			return "", updateErr
		}
		// Schedule OG generation in background (worker will check/cancel if already present)
		if s.jobs != nil {
			_, _ = s.jobs.SchedulePublicQuizOgJob(ctx, existing.ID)
		}
		return existing.Slug, nil
	}

	// Create brand new with generated slug
	slug, err := s.repo.GenerateUniqueSlug(ctx, input.Title)
	if err != nil {
		return "", err
	}
	pq := &model.PublicQuiz{
		Slug:             slug,
		WordlistID:       wordlistID,
		CreatorID:        userID,
		Title:            input.Title,
		Description:      input.Description,
		Difficulty:       input.Difficulty,
		TimeLimitMinutes: input.TimeLimitMinutes,
		IsActive:         true,
	}
	if err := s.repo.CreatePublicQuiz(ctx, pq); err != nil {
		return "", err
	}
	if s.jobs != nil {
		_, _ = s.jobs.SchedulePublicQuizOgJob(ctx, pq.ID)
	}
	return pq.Slug, nil
}

// UnpublishPublicQuiz deactivates the active public quiz for an owner's wordlist.
func (s *PublicQuizService) UnpublishPublicQuiz(ctx context.Context, wordlistID int64, userID int64) error {
	// Verify ownership
	wl, err := s.wordlistSvc.GetWordlistByID(ctx, wordlistID, userID)
	if err != nil {
		return err
	}
	if wl == nil {
		return common.NotFoundError{Entity: "wordlist"}
	}
	onlyActive := true
	limit := 1
	list, err := s.repo.Find(ctx, repository.FindPublicQuizArgs{WordlistID: &wordlistID, CreatorID: &userID, OnlyActive: &onlyActive, Limit: &limit})
	if err != nil {
		return err
	}
	if len(list) == 0 {
		return nil
	}
	pq := list[0]
	return s.repo.DeactivatePublicQuiz(ctx, pq.ID)
}
