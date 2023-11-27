package wordlists

import (
	"errors"
	"log"

	"decorebator.com/internal/common"
)

var repository *WordlistRepository

func init() {
	db, err := common.GetDBConnection()
	if err != nil {
		log.Fatal("failed to open db connection: ", err)
	}
	repository = &WordlistRepository{db}
}

func GetUserWordlists(userId int64) ([]*Wordlist, error) {
	args := FindArgs{
		ownerId: &userId,
	}
	result, err := repository.find(args)
	if err != nil {
		log.Println("Failure at GetUserWordlists:", err)
		return nil, errors.New("Could not get user wordlists")
	}
	return result, nil
}

func SaveWordlist(newWordlist *Wordlist) (*Wordlist, error) {
	wordlist, err := repository.save(newWordlist.Name, newWordlist.Description, newWordlist.UserID)
	if err != nil {
		log.Println("Failure at SaveWordlist:", err)
		return nil, errors.New("Could not save your wordlist")
	}
	return wordlist, nil
}

func GetWordlistById(id, userId int64) (*Wordlist, error) {
	args := FindArgs{
		id:      &id,
		ownerId: &userId,
	}
	result, err := repository.find(args)
	if err != nil {
		log.Println("Failure in GetWordlistById:", err)
		return nil, errors.New("Failed to find wordlist")
	}

	if len(result) != 1 {
		return nil, common.NotFoundError{ID: id, Entity: "Wordlist"}
	}

	wordlist := result[0]
	return wordlist, nil
}

func DeleteWordlist(id, userId int64) (int64, error) {
	count, err := repository.delete(id, userId)
	if err != nil {
		log.Println("Failure in DeleteWordlist:", err)
		return 0, errors.New("Failed to delete wordlist")
	}

	if count == 0 {
		return 0, common.NotFoundError{ID: id, Entity: "Wordlist"}
	}

	return count, nil
}

func UpdateWordlist(wordlist *Wordlist) error {
	count, err := repository.update(wordlist)
	if err != nil {
		log.Println("Failure in UpdateWordlist:", err)
		return errors.New("Failed to update wordlist")
	}

	if count == 0 {
		return common.NotFoundError{ID: wordlist.ID, Entity: "Wordlist"}
	}
	return nil

}
