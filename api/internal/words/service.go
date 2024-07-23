package words

import (
	"errors"
	"strings"

	"decorebator.com/internal/common"
	"decorebator.com/internal/definitions"
	"decorebator.com/internal/definitions/openai"
	spacedrepetion "decorebator.com/internal/quizzes/spacedrepetition"
	"decorebator.com/internal/workers"
)

var repository *WordRepository

func init() {
	db, err := common.GetDBConnection()
	if err != nil {
		common.Logger.Error("failed to open db connection", "error", err)
	}
	repository = &WordRepository{db}
}

func FindByWordlist(wordlistId, userId int64) ([]Word, error) {
	return repository.getAllFromWordlist(wordlistId, userId)
}

func Save(dto *Word) (*Word, error) {
	var lowerCasedName = strings.ToLower(dto.Name)
	var trimmedName = strings.TrimSpace(lowerCasedName)
	word, err := repository.save(trimmedName, dto.UserID, dto.WordlistID)
	logger := common.Logger.With("token", dto.Name, "userId", dto.UserID, "wordId", word.ID, "token", dto.Name, "func", "SaveWord")

	if err != nil {
		logger.Error("failed to save word", "error", err)
		return nil, errors.New("could not save word")
	}

	go func() {
		// defs, err := definitions.FetchAndSave(word.Name, word.ID, wiktionary.GetDefinition)

		// if err != nil || len(defs) == 0 {
		// 	logger.Error("failed to fetch definitions using wiktionary. Falling back to chatgpt", "error", err)

		// 	defs, err = definitions.FetchAndSave(word.Name, word.ID, openai.GetDefinition)
		// 	if err != nil {
		// 		logger.Error("failed to fetch definitions using chatgpt", "error", err)
		// 	}
		// }

		defs, err := definitions.FetchAndSave(word.Name, word.ID, openai.GetDefinition)

		if err != nil || len(defs) == 0 {
			logger.Error("failed to fetch definitions using chatgpt", "error", err)
			return
		}

		if len(defs) > 0 {
			algorithm := spacedrepetion.LeitnerSystemAlgorithm{}
			algorithm.IncludeDefinitions(dto.UserID, defs)

			for _, definition := range defs {
				_, err = workers.TriggerImageGenerator(definition.ID, "")

				if err != nil {
					logger.Error("failed to trigger image generator definition", "definitionId", definition.ID, "error", err)
				}
			}

		}
		logger.Info("definitions fetched", "count", len(defs))

	}()

	return word, nil
}

func Delete(id, userId int64) (int64, error) {
	count, err := repository.delete(userId, id)
	if err != nil {
		common.Logger.Error("failed to delete word", "error", err)
		return 0, errors.New("failed to delete word")
	}

	if count == 0 {
		return 0, common.NotFoundError{ID: id, Entity: "Word"}
	}

	return count, nil
}

func Update(word *Word) error {
	count, err := repository.update(word)
	if err != nil {
		common.Logger.Error("failed to update word", "error", err)
		return errors.New("failed to update word")
	}

	if count == 0 {
		return common.NotFoundError{ID: word.ID, Entity: "Word"}
	}
	return nil
}
