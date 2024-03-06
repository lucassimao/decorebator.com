package definitions

import (
	"log"

	"decorebator.com/internal/common"
)

var repository *DefinitionRepository

func init() {
	db, err := common.GetDBConnection()
	if err != nil {
		log.Fatal("failed to open db connection: ", err)
	}
	repository = &DefinitionRepository{db}
}

type TokenDefiner func(string) ([]Definition, error)

func FetchAndSave(token string, tokenId int64, definerFunc TokenDefiner) ([]Definition, error) {
	var definitions, err = definerFunc(token)

	if err != nil {
		log.Printf("Could not define token %s(%d): %v\n", token, tokenId, err)
		return nil, err
	}

	definitions, err = repository.save(tokenId, definitions)
	if err != nil {
		log.Println("Could not save definitions:", err)
		return nil, err
	}

	return definitions, nil
}
