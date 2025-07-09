package service

import (
	"fmt"
	"runtime/debug"

	"decorebator.com/internal/common"
	rep "decorebator.com/internal/repository"
)

var definitionImageRepository *rep.DefinitionImageRepository

func init() {
	db, err := common.GetDBConnection()
	if err != nil {
		common.Logger.Error("failed to open db connection: ", "error", err)
		// Don't crash the application - let services handle the error gracefully
		return
	}
	definitionImageRepository = rep.NewDefinitionImageRepository(db)
}

func SaveDefinitionImage(dto rep.CreateDefinitionImageDTO) (*rep.DefinitionImage, error) {

	definitionImage, err := definitionImageRepository.Save(dto)
	if err != nil {
		msg := "failed to save definition image"
		common.Logger.Error(msg, "error", err, "stacktrace", string(debug.Stack()))
		return nil, fmt.Errorf(msg+": %w", err)
	}

	return definitionImage, nil
}
