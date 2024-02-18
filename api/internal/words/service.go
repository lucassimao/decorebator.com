package words

import (
	"errors"
	"log"

	"decorebator.com/internal/common"
	"decorebator.com/internal/openai"
)

var repository *WordRepository

func init() {
	db, err := common.GetDBConnection()
	if err != nil {
		log.Fatal("failed to open db connection: ", err)
	}
	repository = &WordRepository{db}
}

func GetWordsFromWordlist(wordlistId, userId int64) ([]*Word, error) {
	return repository.getAllFromWordlist(wordlistId, userId)
}

func SaveWord(newWord *Word) (*Word, error) {
	word, err := repository.save(newWord.Name, newWord.UserID, newWord.WordlistID)
	go openai.GetDefinition(newWord.Name, word.ID)

	if err != nil {
		log.Println("Failure at SaveWord:", err)
		return nil, errors.New("Could not save your word")
	}
	return word, nil
}

func DeleteWord(id, userId int64) (int64, error) {
	count, err := repository.delete(userId, id)
	if err != nil {
		log.Println("Failure in DeleteWord:", err)
		return 0, errors.New("Failed to delete word")
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
		return errors.New("Failed to update word")
	}

	if count == 0 {
		return common.NotFoundError{ID: word.ID, Entity: "Word"}
	}
	return nil

}
