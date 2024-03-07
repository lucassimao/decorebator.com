package spacedrepetion

import (
	"context"
	"log"
	"math/rand"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/definitions"
)

type LeitnerSystemAlgorithm struct{}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func getNextDefinition(userID, wordlistID int64) (*definitions.Definition, error) {
	// Grouping by box_id and getting the earliest updated_at
	query := `
		WITH earliest_per_box AS (
			SELECT DISTINCT ON (lst.box_id) def.id, def.token, def.part_of_speech, def.meaning, def.examples, def.inflections , lst.updated_at
				FROM leitner_system_tracking lst JOIN definitions def ON lst.definition_id = def.id
			JOIN word_definitions wd ON def.id = wd.definition_id
			JOIN words w ON wd.word_id = w.id
			WHERE 
				lst.user_id =$1 AND w.wordlist_id=$2
			ORDER BY
				lst.box_id, lst.updated_at ASC NULLS FIRST
		)
		SELECT id, token, part_of_speech, meaning, examples, inflections 
		FROM earliest_per_box ORDER BY updated_at ASC NULLS FIRST
		LIMIT 1;
	`

	db, err := common.GetDBConnection()
	if err != nil {
		log.Fatal("failed to open db connection: ", err)
	}

	rows, err := db.Query(context.Background(), query, userID, wordlistID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	rows.Next()
	definition := definitions.Definition{}
	err = rows.Scan(&definition.ID, &definition.Token, &definition.PartOfSpeech,
		&definition.Meaning, &definition.Examples, &definition.Inflections)

	if err != nil {
		return nil, err
	}

	return &definition, nil
}

func (LeitnerSystemAlgorithm) IncludeDefinitions(userID int64, definitions []definitions.Definition) error {
	db, err := common.GetDBConnection()
	if err != nil {
		log.Fatal("failed to open db connection:", err)
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
			log.Printf("Failed to insert definition %v into leitner_system_tracking: %v\n", definition, err)
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
	next, err := getNextDefinition(userID, wordlistID)
	if err != nil {
		log.Println("Error in CreateChallenge:", err)
		return nil, err
	}

	randomMeanings, err := definitions.GetRandomMeanings([]int{int(next.ID)}, 3)

	if err != nil {
		log.Println("Error getting random meanings:", err)
		return nil, err
	}

	index := rand.Intn(4)

	randomMeanings = append(randomMeanings, "")
	copy(randomMeanings[index+1:], randomMeanings[index:])
	randomMeanings[index] = next.Meaning

	challenge := &Challenge{
		Token:        next.Token,
		Options:      randomMeanings,
		AnswerIndex:  index,
		DefinitionID: next.ID,
	}

	return challenge, nil
}

func (LeitnerSystemAlgorithm) SaveChallengeResult(definitionID int64, success bool) error {

	query := `UPDATE leitner_system_tracking SET updated_at=now(), box_id = CASE WHEN $1 THEN box_id + 1 ELSE 1 END WHERE definition_id = $2`

	db, err := common.GetDBConnection()
	if err != nil {
		log.Fatal("failed to open db connection:", err)
	}

	_, err = db.Exec(context.Background(), query, success, definitionID)
	if err != nil {
		return err
	}
	return nil
}
