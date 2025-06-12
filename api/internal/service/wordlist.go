package service

import (
	"fmt"
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

// include wordsCount and WordsLearnedCount for each scholarship
func GetUserWordlistsWithWordStats(userID int64) ([]*Wordlist, error) {
	args := repo.FindWordlistArgs{
		OwnerID:                  &userID,
		ComputeWordsCount:        true,
		ComputeWordsLearnedCount: true,
	}
	result, err := wordlistRepository.Find(args)
	if err != nil {
		wrappedErr := fmt.Errorf(
			"failed to get all wordlists: %w", err,
		)
		return nil, wrappedErr
	}
	return result, nil
}

func SaveWordlist(newWordlist *Wordlist) (*Wordlist, error) {
	wordlist, err := wordlistRepository.Save(newWordlist.Name, newWordlist.Description, newWordlist.LanguageCode, newWordlist.UserID)
	if err != nil {
		wrappedErr := fmt.Errorf(
			"failed to save wordlist: %w", err,
		)
		return nil, wrappedErr
	}

	return wordlist, nil
}

func GetWordlistByID(id, userID int64) (*Wordlist, error) {
	args := repo.FindWordlistArgs{
		ID:      &id,
		OwnerID: &userID,
	}
	result, err := wordlistRepository.Find(args)
	if err != nil {
		wrappedErr := fmt.Errorf(
			"failed to get wordlist %d by id: %w", id, err,
		)
		return nil, wrappedErr
	}

	if len(result) != 1 {
		return nil, common.NotFoundError{ID: id, Entity: "Wordlist"}
	}

	wordlist := result[0]
	return wordlist, nil
}

func DeleteWordlist(id, userID int64) (int64, error) {
	count, err := wordlistRepository.Delete(id, userID)
	if err != nil {
		wrappedErr := fmt.Errorf(
			"failed to delete wordlist %d : %w", id, err,
		)
		return 0, wrappedErr
	}

	if count == 0 {
		return 0, common.NotFoundError{ID: id, Entity: "Wordlist"}
	}

	return count, nil
}

func UpdateWordlist(wordlist *Wordlist) error {
	count, err := wordlistRepository.Update(wordlist)
	if err != nil {
		wrappedErr := fmt.Errorf(
			"failed to update wordlist %d : %w", wordlist.ID, err,
		)
		return wrappedErr
	}

	if count == 0 {
		return common.NotFoundError{ID: wordlist.ID, Entity: "Wordlist"}
	}
	return nil
}
