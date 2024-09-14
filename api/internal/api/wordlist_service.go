package api

import (
	"errors"
	"os"

	"decorebator.com/internal/model"
	repo "decorebator.com/internal/repository"

	"decorebator.com/internal/common"
)

var wordlistRepository *repo.WordlistRepository

type Wordlist = model.Wordlist

func init() {
	db, err := common.GetDBConnection()
	if err != nil {
		common.Logger.Error("failed to open db connection", "error", err)
		os.Exit(1)
	}
	wordlistRepository = &repo.WordlistRepository{Db: db}
}

func GetUserWordlists(userId int64) ([]*Wordlist, error) {
	args := repo.FindWordlistArgs{
		OwnerId: &userId,
	}
	result, err := wordlistRepository.Find(args)
	if err != nil {
		common.Logger.Error("failed to get wordlists", "error", err, "userId", userId)
		return nil, errors.New("could not get user wordlists")
	}
	return result, nil
}

func SaveWordlist(newWordlist *Wordlist) (*Wordlist, error) {
	wordlist, err := wordlistRepository.Save(newWordlist.Name, newWordlist.Description, newWordlist.UserID)
	if err != nil {
		common.Logger.Error("failed to save wordlist", "error", err)
		return nil, errors.New("could not save your wordlist")
	}
	return wordlist, nil
}

func GetWordlistById(id, userId int64) (*Wordlist, error) {
	args := repo.FindWordlistArgs{
		Id:      &id,
		OwnerId: &userId,
	}
	result, err := wordlistRepository.Find(args)
	if err != nil {
		common.Logger.Error("failed to get wordlist", "error", err, "wordlistId", id)
		return nil, errors.New("failed to find wordlist")
	}

	if len(result) != 1 {
		return nil, common.NotFoundError{ID: id, Entity: "Wordlist"}
	}

	wordlist := result[0]
	return wordlist, nil
}

func DeleteWordlist(id, userId int64) (int64, error) {
	count, err := wordlistRepository.Delete(id, userId)
	if err != nil {
		common.Logger.Error("failed to delete wordlist", "error", err, "wordlistId", id)
		return 0, errors.New("gailed to delete wordlist")
	}

	if count == 0 {
		return 0, common.NotFoundError{ID: id, Entity: "Wordlist"}
	}

	return count, nil
}

func UpdateWordlist(wordlist *Wordlist) error {
	count, err := wordlistRepository.Update(wordlist)
	if err != nil {
		common.Logger.Error("failed to update wordlist", "error", err, "wordlistId", wordlist.ID)
		return errors.New("failed to update wordlist")
	}

	if count == 0 {
		return common.NotFoundError{ID: wordlist.ID, Entity: "Wordlist"}
	}
	return nil

}
