package words

import (
	"errors"
	"log"
	"strings"

	"decorebator.com/internal/common"
	"decorebator.com/internal/definitions"
	"decorebator.com/internal/definitions/openai"
	"decorebator.com/internal/definitions/wiktionary"
	spacedrepetion "decorebator.com/internal/quizzes/spacedrepetition"
)

var repository *WordRepository

func init() {
	db, err := common.GetDBConnection()
	if err != nil {
		log.Fatal("failed to open db connection: ", err)
	}
	repository = &WordRepository{db}
}

func GetWordsFromWordlist(wordlistId, userId int64) ([]Word, error) {
	return repository.getAllFromWordlist(wordlistId, userId)
}

func SaveWord(dto *Word) (*Word, error) {
	var lowerCasedName = strings.ToLower(dto.Name)
	var trimmedName = strings.TrimSpace(lowerCasedName)
	word, err := repository.save(trimmedName, dto.UserID, dto.WordlistID)

	if err != nil {
		log.Println("Failure at SaveWord:", err)
		return nil, errors.New("could not save word")
	}

	go func() {
		defs, err := definitions.FetchAndSave(word.Name, word.ID, wiktionary.GetDefinition)
		if err != nil || len(defs) == 0 {
			log.Println("Falling back to ChatGPT")

			defs, err = definitions.FetchAndSave(word.Name, word.ID, openai.GetDefinition)
			if err != nil {
				log.Printf("Failed to fetch from openai: %v\n", err)
			}
		}

		if len(defs) > 0 {
			algorithm := spacedrepetion.LeitnerSystemAlgorithm{}
			algorithm.IncludeDefinitions(dto.UserID, defs)
			log.Printf("Fetched and saved %v definitions for word %s\n", len(defs), dto.Name)
		}
	}()

	return word, nil
}

func DeleteWord(id, userId int64) (int64, error) {
	count, err := repository.delete(userId, id)
	if err != nil {
		log.Println("Failure in DeleteWord:", err)
		return 0, errors.New("failed to delete word")
	}

	if count == 0 {
		return 0, common.NotFoundError{ID: id, Entity: "Word"}
	}

	return count, nil
}

func UpdateWord(word *Word) error {
	count, err := repository.update(word)
	if err != nil {
		log.Println("Failure in UpdateWord:", err)
		return errors.New("failed to update word")
	}

	if count == 0 {
		return common.NotFoundError{ID: word.ID, Entity: "Word"}
	}
	return nil

}
