package service

import (
	"context"
	"errors"
	"math/rand"
	"os"
	"regexp"
	"strings"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LeitnerSystemStrategy struct{}
type Quiz = model.Quiz
type QuizType = model.QuizType

type NextDefinition struct {
	Definition       *model.Definition
	LeitnerSystemID  int64
	BoxID            int64
	WordID           int64
	ImageUrl         string
	ImageDescription string
}

func getNextDefinition(userID, wordlistID int64) (*NextDefinition, error) {

	// Grouping by box_id and getting the earliest updated_at
	query := `
		WITH words_queue AS (
			SELECT wd.word_id, max(lst.updated_at) max_lst_updated_at
				FROM leitner_system_tracking lst
				JOIN word_definitions wd ON wd.definition_id = lst.definition_id
				JOIN words w ON wd.word_id = w.id
				WHERE 
					w.wordlist_id=$2 
				GROUP BY 
					wd.word_id
				ORDER BY
					max_lst_updated_at ASC NULLS FIRST
				LIMIT 1		
		),
		earliest_per_box AS (
			SELECT def.id, lst.id AS lst_id, def.token, 
					def.part_of_speech,  def.meaning, def.examples, 
					def.inflections , lst.box_id, def.sounds, def.phonetic_notations, 
					di.url as image_url, di.description as image_description, 
					wd.word_id AS word_id
			FROM leitner_system_tracking lst 
			JOIN definitions def ON lst.definition_id = def.id
			JOIN word_definitions wd ON def.id = wd.definition_id
			LEFT JOIN definition_images di ON di.definition_id = def.id AND di.is_visible=TRUE
			WHERE 
				lst.user_id = $1
				AND wd.word_id = (select word_id FROM words_queue)
				AND def.meaning IS NOT NULL 
				AND array_length(def.examples,1) > 0
			ORDER BY
				lst.box_id ASC, lst.updated_at ASC NULLS FIRST, wd.word_id ASC
			LIMIT 1
		)

		SELECT id, token, part_of_speech, meaning, examples, inflections, lst_id, 
				box_id, sounds,phonetic_notations, COALESCE(image_url,''), word_id, COALESCE(image_description,'')
		FROM earliest_per_box;
	`

	db, err := common.GetDBConnection()
	if err != nil {
		common.Logger.Error("failed to open db connection", "error", err)
		os.Exit(1)
	}

	rows, err := db.Query(context.Background(), query, userID, wordlistID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	hasRow := rows.Next()
	if !hasRow {
		return nil, errors.New("no definitions found")
	}

	definition := model.Definition{}
	result := NextDefinition{Definition: &definition}

	err = rows.Scan(&definition.ID, &definition.Token, &definition.PartOfSpeech,
		&definition.Meaning, &definition.Examples,
		&definition.Inflections, &result.LeitnerSystemID, &result.BoxID, &definition.Sounds,
		&definition.PhoneticNotations, &result.ImageUrl, &result.WordID, &result.ImageDescription)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (LeitnerSystemStrategy) IncludeDefinitions(wordId, userId int64, definitionIds []int64, tx pgx.Tx) error {

	for _, definitionId := range definitionIds {
		query := `INSERT INTO leitner_system_tracking (user_id, definition_id, box_id, word_id)
		VALUES ($1, $2, $3, $4) RETURNING id`

		row := tx.QueryRow(context.Background(), query, userId, definitionId, 1, wordId)
		var leitnerSystemTrackingId int64
		err := row.Scan(&leitnerSystemTrackingId)

		if err != nil {
			return err
		}

		_, err = tx.Exec(context.Background(), `INSERT INTO 
		leitner_system_history (at,box_id,leitner_system_tracking_id) 
		VALUES (now(), $1, $2)`, 1, leitnerSystemTrackingId)

		if err != nil {
			return err
		}
	}

	return nil
}

func (LeitnerSystemStrategy) CreateQuiz(wordlistID, userID int64) (*Quiz, error) {
	nextDefinition, err := getNextDefinition(userID, wordlistID)
	if err != nil {
		return nil, err
	}

	var options []string
	var value string
	var quizzType QuizType
	var quizAnswer string

	var word *model.Word
	boxID := nextDefinition.BoxID
	wordID := nextDefinition.WordID
	definition := nextDefinition.Definition
	leitnerSystemID := nextDefinition.LeitnerSystemID
	imageUrl := nextDefinition.ImageUrl

	word, err = GetWordById(wordID)
	if err != nil {
		return nil, err
	}

	// [TODO] Refactor and create factory methods
	if boxID%6 == 0 || boxID%5 == 0 {
		// if no audio, jump to the GuessMeaning quiz
		if word.AudioURL == "" {
			boxID = 1
		}
	}

	common.Logger.Debug("CreateChallenge", "boxID", boxID, "leitnerSystemID", leitnerSystemID, "definition", definition.ID)

	switch {
	case boxID%6 == 0:
		quizzType = model.MeaningFromAudio
		options, err = GetRandomMeanings([]int{int(definition.ID)}, 3)
		if err != nil {
			return nil, err
		}
		quizAnswer = definition.Meaning
	case boxID%5 == 0:
		quizzType = model.WordFromAudio
		options, err = GetRandomTokens([]int{int(definition.ID)}, definition.PartOfSpeech, 3)
		if err != nil {
			return nil, err
		}

		quizAnswer = word.Name
	case boxID%4 == 0:
		quizzType = model.CompleteSentence
		options, err = GetRandomTokens([]int{int(definition.ID)}, definition.PartOfSpeech, 3)
		if err != nil {
			return nil, err
		}

		i := rand.Intn(len(definition.Examples))
		value = definition.Examples[i]

		re := regexp.MustCompile(`\[(.*?)\]`)
		matches := re.FindAllStringSubmatch(value, -1)
		var tokens []string

		// Iterate over matches and extract the content inside the brackets
		for _, match := range matches {
			if len(match) > 1 {
				tokens = append(tokens, match[1]) // Append the extracted token to the slice
			}
		}

		if len(tokens) > 0 {
			quizAnswer = strings.Join(tokens, " ")
		} else {
			quizAnswer = definition.Token
		}

	case boxID%3 == 0 && imageUrl != "":
		quizzType = model.WordFromImage
		value = imageUrl
		options, err = GetRandomTokens([]int{int(definition.ID)}, definition.PartOfSpeech, 3)

		if err != nil {
			return nil, err
		}

		re := regexp.MustCompile(`\[(.*?)\]`)
		matches := re.FindStringSubmatch(nextDefinition.ImageDescription)
		if len(matches) > 1 {
			quizAnswer = matches[1]
		} else {
			quizAnswer = definition.Token
		}
	case boxID%2 == 0:
		quizzType = model.WordFromMeaning
		value = definition.Meaning
		options, err = GetRandomTokens([]int{int(definition.ID)}, definition.PartOfSpeech, 3)

		if err != nil {
			return nil, err
		}

		quizAnswer = definition.Token
	default:
		quizzType = model.GuessMeaning
		options, err = GetRandomMeanings([]int{int(definition.ID)}, 3)
		if err != nil {
			return nil, err
		}

		value = definition.Token
		quizAnswer = definition.Meaning
	}

	// append the quiz answer to the final of the list.
	options = append(options, quizAnswer)
	// Then genrate a random position to move to it
	var answerIndex = rand.Intn(len(options))
	// shifts all elements from answerIndex to the end of the slice one position to the right, making space at answerIndex
	copy(options[answerIndex+1:], options[answerIndex:])
	options[answerIndex] = quizAnswer

	challenge := &Quiz{
		Value:       value,
		Options:     options,
		AnswerIndex: answerIndex,
		ID:          leitnerSystemID,
		Type:        quizzType,
		// Sounds:       definition.Sounds,
		// Notations:    definition.PhoneticNotations,
		PartOfSpeech:     definition.PartOfSpeech,
		ImageDescription: nextDefinition.ImageDescription,
		AudioURL:         word.AudioURL,
		WordID:           wordID,
		DefinitionID:     nextDefinition.Definition.ID,
	}

	return challenge, nil
}

func (LeitnerSystemStrategy) SaveQuizResult(leitnerSystemTrackingId int64, success bool, transactionPtr *pgx.Tx) error {

	var tx pgx.Tx
	var err error

	if transactionPtr == nil {
		var db *pgxpool.Pool
		db, err = common.GetDBConnection()
		if err != nil {
			return err
		}

		ctx := context.Background()

		tx, err = db.Begin(ctx)
		if err != nil {
			return err
		}
		// only handle transaction if created internally
		defer func() {
			if err == nil {
				tx.Commit(ctx)
			} else {
				tx.Rollback(ctx)
			}
		}()

	} else {
		tx = *transactionPtr
	}

	query := `UPDATE leitner_system_tracking 
	SET 
		updated_at = now(), 
		box_id = CASE WHEN $1 THEN box_id + 1 ELSE 1 END 
	WHERE id = $2
	RETURNING box_id`

	var boxId int64
	row := tx.QueryRow(context.Background(), query, success, leitnerSystemTrackingId)
	err = row.Scan(&boxId)

	if err != nil {
		return err
	}

	_, err = tx.Exec(context.Background(), `INSERT INTO 
		leitner_system_history (at,box_id,leitner_system_tracking_id) 
		VALUES (now(), $1, $2)`, boxId, leitnerSystemTrackingId)

	if err != nil {
		return err
	}

	return nil
}
