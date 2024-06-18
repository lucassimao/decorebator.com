package spacedrepetion

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/definitions"
)

type LeitnerSystemAlgorithm struct{}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func getNextDefinition(userID, wordlistID int64) (*definitions.Definition, int64, int64, error) {
	// Grouping by box_id and getting the earliest updated_at
	query := `
		-- rouding robin by updated_at
		WITH words_queue AS (
			SELECT wd.word_id, max(lst.updated_at) max_lst_updated_at
			FROM leitner_system_tracking lst
			JOIN word_definitions wd ON wd.definition_id = lst.definition_id
			JOIN words w ON wd.word_id = w.id
			WHERE 
				w.user_id =$1 AND w.wordlist_id=$2 
			GROUP BY 
				wd.word_id
			ORDER BY
				max_lst_updated_at ASC NULLS FIRST
			LIMIT 1
		),
		earliest_per_box AS (
			SELECT def.id, lst.id AS lst_id, def.token, 
					def.part_of_speech,  def.meaning, def.examples, 
					def.inflections , lst.box_id, def.sounds, def.phonetic_notations
			FROM leitner_system_tracking lst 
			JOIN definitions def ON lst.definition_id = def.id
			JOIN word_definitions wd ON def.id = wd.definition_id
			WHERE 
				lst.user_id = $1
				AND wd.word_id = (select word_id FROM words_queue)
				AND def.meaning IS NOT NULL AND array_length(def.examples,1) > 0
			ORDER BY
				lst.box_id ASC, lst.updated_at ASC NULLS FIRST, wd.word_id ASC
		)
		SELECT id, token, part_of_speech, meaning, examples, inflections, lst_id, box_id, sounds,phonetic_notations
		FROM earliest_per_box 
		LIMIT 1;
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

	rows.Next()
	definition := definitions.Definition{}
	var leitnerSystemID int64
	var boxID int64

	err = rows.Scan(&definition.ID, &definition.Token, &definition.PartOfSpeech,
		&definition.Meaning, &definition.Examples,
		&definition.Inflections, &leitnerSystemID, &boxID, &definition.Sounds,
		&definition.PhoneticNotations)

	if err != nil {
		return nil, -1, -1, err
	}

	return &definition, leitnerSystemID, boxID, nil
}

func (LeitnerSystemAlgorithm) IncludeDefinitions(userID int64, definitions []definitions.Definition) error {
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
func (LeitnerSystemAlgorithm) CreateChallenge(wordlistID, userID int64) (*Challenge, error) {
	definition, leitnerSystemID, boxID, err := getNextDefinition(userID, wordlistID)
	if err != nil {
		return nil, err
	}

	var options []string
	var value string
	var index int
	var quizzType ChallengeType

	// boxID been odd, will make each first quiz of a definition of type GUESS_MEANING, otherwise a COMPLETE_SENTENCE
	fmt.Printf("box id %v leitnerSystemID %v", boxID, leitnerSystemID)
	if boxID%2 == 1 {
		randomMeanings, err := definitions.GetRandomMeanings([]int{int(definition.ID)}, 3)
		if err != nil {
			return nil, err
		}

		quizzType = GUESS_MEANING
		value = definition.Token
		options = append(randomMeanings, definition.Meaning)
	} else {
		fmt.Println("def %v", definition)
		randomTokens, err := definitions.GetRandomTokens([]int{int(definition.ID)}, definition.PartOfSpeech, 3)
		if err != nil {
			return nil, err
		}

		quizzType = COMPLETE_SENTENCE
		i := rand.Intn(len(definition.Examples))
		value = definition.Examples[i]

		re := regexp.MustCompile(`\[(.*?)\]`)
		matches := re.FindAllStringSubmatch(value, -1)
		if len(matches) > 0 {
			options = append(randomTokens, matches[0][1])
		} else {
			options = append(randomTokens, definition.Token)
		}

	}

	lastItem := options[len(options)-1]
	index = rand.Intn(len(options))
	copy(options[index+1:], options[index:])
	options[index] = lastItem

	challenge := &Challenge{
		Value:        value,
		Options:      options,
		AnswerIndex:  index,
		ID:           leitnerSystemID,
		Type:         quizzType,
		Sounds:       definition.Sounds,
		Notations:    definition.PhoneticNotations,
		PartOfSpeech: definition.PartOfSpeech,
	}

	return challenge, nil
}

func (LeitnerSystemAlgorithm) SaveChallengeResult(id int64, success bool) error {

	query := `UPDATE leitner_system_tracking 
	SET updated_at=now(), box_id = CASE WHEN $1 THEN box_id + 1 ELSE box_id END 
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
