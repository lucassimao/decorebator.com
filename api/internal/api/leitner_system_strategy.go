package api

import (
	"context"
	"errors"
	"math/rand"
	"os"
	"regexp"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
)

type LeitnerSystemStrategy struct{}
type Quiz = model.Quiz
type QuizType = model.QuizType

func getNextDefinition(userID, wordlistID int64) (*Definition, int64, int64, error) {

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
					def.inflections , lst.box_id, def.sounds, def.phonetic_notations, def.image_url
			FROM leitner_system_tracking lst 
			JOIN definitions def ON lst.definition_id = def.id
			JOIN word_definitions wd ON def.id = wd.definition_id
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
				box_id, sounds,phonetic_notations, COALESCE(image_url,'')
		FROM earliest_per_box;
	`

	db, err := common.GetDBConnection()
	if err != nil {
		common.Logger.Error("failed to open db connection", "error", err)
		os.Exit(1)
	}

	rows, err := db.Query(context.Background(), query, userID, wordlistID)
	if err != nil {
		return nil, -1, -1, err
	}

	defer rows.Close()

	hasRow := rows.Next()
	if !hasRow {
		return nil, -1, -1, errors.New("no definitions found")
	}

	definition := Definition{}
	var leitnerSystemID int64
	var boxID int64

	err = rows.Scan(&definition.ID, &definition.Token, &definition.PartOfSpeech,
		&definition.Meaning, &definition.Examples,
		&definition.Inflections, &leitnerSystemID, &boxID, &definition.Sounds,
		&definition.PhoneticNotations, &definition.ImageUrl)

	if err != nil {
		return nil, -1, -1, err
	}

	return &definition, leitnerSystemID, boxID, nil
}

func (LeitnerSystemStrategy) IncludeDefinitions(userID int64, definitions []Definition) error {
	db, err := common.GetDBConnection()
	if err != nil {
		common.Logger.Error("failed to open db connection", "error", err)
		os.Exit(1)
	}

	tx, err := db.Begin(context.Background())
	if err != nil {
		return err
	}

	for _, definition := range definitions {
		query := `INSERT INTO leitner_system_tracking (user_id, definition_id, box_id)
		VALUES ($1, $2, $3)`

		_, err := tx.Exec(context.Background(), query, userID, definition.ID, 1)
		if err != nil {
			tx.Rollback(context.Background())
			return err
		}
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func (LeitnerSystemStrategy) CreateQuiz(wordlistID, userID int64) (*Quiz, error) {
	definition, leitnerSystemID, boxID, err := getNextDefinition(userID, wordlistID)
	if err != nil {
		return nil, err
	}

	var options []string
	var value string
	var quizzType QuizType
	var quizAnswer string

	common.Logger.Debug("CreateChallenge", "boxID", boxID, "leitnerSystemID", leitnerSystemID, "definition", definition.ID)

	switch {
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
		if len(matches) > 0 {
			quizAnswer = matches[0][1]
		} else {
			quizAnswer = definition.Token
		}
	case boxID%3 == 0 && definition.ImageUrl != "":
		quizzType = model.WordFromImage
		value = definition.ImageUrl
		options, err = GetRandomTokens([]int{int(definition.ID)}, definition.PartOfSpeech, 3)

		if err != nil {
			return nil, err
		}

		quizAnswer = definition.Token
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
		Value:        value,
		Options:      options,
		AnswerIndex:  answerIndex,
		ID:           leitnerSystemID,
		Type:         quizzType,
		Sounds:       definition.Sounds,
		Notations:    definition.PhoneticNotations,
		PartOfSpeech: definition.PartOfSpeech,
	}

	return challenge, nil
}

func (LeitnerSystemStrategy) SaveQuizResult(id int64, success bool) error {

	query := `UPDATE leitner_system_tracking 
	SET 
		updated_at = now(), 
		box_id = CASE WHEN $1 THEN LEAST(box_id + 1,4) ELSE 1 END 
	WHERE id = $2`

	db, err := common.GetDBConnection()
	if err != nil {
		common.Logger.Error("failed to open db connection", "error", err)
		os.Exit(1)
	}

	_, err = db.Exec(context.Background(), query, success, id)
	if err != nil {
		return err
	}

	return nil
}
