package api

import (
	"fmt"
	"os"

	"decorebator.com/internal/common"
)

var definitionImageRepository *DefinitionImageRepository

func init() {
	db, err := common.GetDBConnection()
	if err != nil {
		common.Logger.Error("failed to open db connection: ", "error", err)
		os.Exit(1)
	}
	definitionImageRepository = &DefinitionImageRepository{db}
}

func SaveDefinitionImage(dto CreateDefinitionImageDTO) (DefinitionImage, error) {

	definitionImage, err := definitionImageRepository.save(dto)
	if err != nil {
		return DefinitionImage{}, fmt.Errorf("failed to save definitions: %w", err)
	}

	return definitionImage, nil
}
