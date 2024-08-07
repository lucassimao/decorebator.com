package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/openai"
	"github.com/riverqueue/river"
)

type ImageGeneratorArgs struct {
	DefinitionId int64  `json:"definitionId"`
	CustomPrompt string `json:"customPrompt"`
}

func (ImageGeneratorArgs) Kind() string { return "ImageGenerator" }

type ImageGeneratorWorker struct {
	river.WorkerDefaults[ImageGeneratorArgs]
}

func (w *ImageGeneratorWorker) Work(ctx context.Context, job *river.Job[ImageGeneratorArgs]) error {
	var logger = common.Logger.With("worker", "imagegenerator")

	var prompt string
	var definitionID = job.Args.DefinitionId

	if job.Args.CustomPrompt != "" {
		prompt = job.Args.CustomPrompt
	} else {
		definition, err := GetDefinitionById(job.Args.DefinitionId)
		if err != nil {

			if errors.Is(err, &common.NotFoundError{}) {
				logger.Error("skipping image generation: inexisting definition", "definitionId", job.Args.DefinitionId)
				return nil
			}

			logger.Error("failed to get definition by id", "definitionId", job.Args.DefinitionId, "error", err)
			return err
		}
		var longestExample string
		for _, item := range definition.Examples {
			if len(item) > len(longestExample) {
				longestExample = item
			}
		}
		prompt = fmt.Sprintf("Illustrate: %s", longestExample)
	}

	response, err := openai.GenerateImage(prompt)

	if err != nil {
		// [TODO] track potential causes here and decide if return nil or not
		logger.Error("failed to generate image", "error", err)
		return err
	}

	// [TODO] track potential causes here and decide if return nil or not
	if response.Error != nil {
		logger.Error("failed to generate image", "body", response.Error)

		switch response.Error.Code {
		case "rate_limit_exceeded":
			// snoozing between 1 and 2min
			return river.JobSnooze(time.Minute + (time.Duration(rand.Intn(60)) * time.Second))
		case "billing_hard_limit_reached":
			// [TODO] notification here elsewhere
			logger.Warn(response.Error.Message)
		case "content_policy_violation":
			jsonData, err := json.Marshal(response.Error)

			if err != nil {
				return river.JobCancel(err)
			}

			// aborting job
			return river.JobCancel(
				fmt.Errorf("content_policy_violation: %q", string(jsonData)))
		}

		return fmt.Errorf("OpenAI error: %s", response.Error.Message)
	}

	firstItem := *(response.Data)
	// Decode base64 data
	data, err := base64.StdEncoding.DecodeString(firstItem[0].Base64Json)
	if err != nil {
		return fmt.Errorf("failed to decode base64 data: %v", err)
	}

	url, err := common.Upload(data, "images",
		fmt.Sprintf("definition-%d-%d.png", definitionID, time.Now().Unix()), "image/png")

	if err != nil {
		logger.Error("failed to upload image", "error", err)
		return err
	}

	logger.Debug("image generated", "definitionId", definitionID, "url", url)
	return SetDefinitionImage(definitionID, url)
}
