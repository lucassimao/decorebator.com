package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/openai"
	"decorebator.com/internal/stability_ai"
	"github.com/riverqueue/river"
)

var logger = common.Logger.With("worker", "imagegenerator")

type ImageGeneratorArgs struct {
	DefinitionId int64  `json:"definitionId"`
	CustomPrompt string `json:"customPrompt"`
}

func (ImageGeneratorArgs) Kind() string { return "ImageGenerator" }

type ImageGeneratorWorker struct {
	river.WorkerDefaults[ImageGeneratorArgs]
}

func (w *ImageGeneratorWorker) Work(ctx context.Context, job *river.Job[ImageGeneratorArgs]) error {

	var prompt string
	var definitionID = job.Args.DefinitionId
	var description string

	if job.Args.CustomPrompt != "" {
		prompt = job.Args.CustomPrompt
		description = job.Args.CustomPrompt
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

		description = ""
		// using the longest example as the image description
		for _, item := range definition.Examples {
			if len(item) > len(description) {
				description = item
			}
		}

		endWrappedToken := strings.Index(description, "]")
		if endWrappedToken == -1 { // no word between square brackets founds
			prompt = fmt.Sprintf("%s. %s meaning %s.", description, definition.Token, definition.Meaning)
		} else {
			// removing the square brackets and adding the term's meaning close to the term and wrapped by parentheses"
			prompt = description
			pos := strings.Index(prompt, "]")
			prompt = fmt.Sprintf("%s (%s) %s", prompt[:pos+1], definition.Meaning, prompt[pos+1:])

			prompt = strings.Replace(prompt, "[", "", 1)
			prompt = strings.Replace(prompt, "]", "", 1)
		}
	}

	data, err := generateWithStabilityAI(prompt)

	if err != nil {
		return err
	}

	url, err := common.Upload(data, "decorebator",
		fmt.Sprintf("images/definition-%d-%d.png", definitionID, time.Now().Unix()), "image/png")

	if err != nil {
		logger.Error("failed to upload image", "error", err)
		return err
	}

	logger.Debug("image generated", "definitionId", definitionID, "url", url)

	_, err = SaveDefinitionImage(model.CreateDefinitionImageDTO{
		Api:          model.STABILITY_AI,
		URL:          url,
		Description:  description,
		Model:        "stable-image-core",
		Prompt:       prompt,
		DefinitionId: definitionID,
	})

	return err
}

func generateWithStabilityAI(prompt string) ([]byte, error) {

	response, err := stability_ai.GenerateImage(prompt)

	if err != nil {
		logger.Error("failed to generate image", "error", err, "details", response.Errors)

		switch err {
		case stability_ai.ErrInvalidParams:
			return nil, river.JobCancel(err)
		case stability_ai.ErrContentFlagged:
			return nil, river.JobCancel(err)
		case stability_ai.ErrRejected:
			return nil, river.JobCancel(err)
		case stability_ai.ErrTooManyRequets:
			// snoozing between 1 and 2min
			return nil, river.JobSnooze(time.Minute + (time.Duration(rand.Intn(60)) * time.Second))
		case stability_ai.ErrInternalError:
			// snoozing between 1 and 2min
			return nil, river.JobSnooze(time.Minute + (time.Duration(rand.Intn(60)) * time.Second))
		default:
			return nil, river.JobCancel(err)
		}
	}

	data, err := base64.StdEncoding.DecodeString(*response.Image)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 data: %v", err)
	}

	return data, nil
}

func generateWithOpenAI(prompt string) ([]byte, error) {

	response, err := openai.GenerateImage(prompt)

	if err != nil {
		// [TODO] track potential causes here and decide if return nil or not
		logger.Error("failed to generate image", "error", err)
		return nil, err
	}

	// [TODO] track potential causes here and decide if return nil or not
	if response.Error != nil {
		logger.Error("failed to generate image", "body", response.Error)

		switch response.Error.Code {
		case "rate_limit_exceeded":
			// snoozing between 1 and 2min
			return nil, river.JobSnooze(time.Minute + (time.Duration(rand.Intn(60)) * time.Second))
		case "billing_hard_limit_reached":
			// [TODO] notification here elsewhere
			logger.Warn(response.Error.Message)
		case "content_policy_violation":
			jsonData, err := json.Marshal(response.Error)

			if err != nil {
				return nil, river.JobCancel(err)
			}

			// aborting job
			return nil, river.JobCancel(
				fmt.Errorf("content_policy_violation: %q", string(jsonData)))
		}

		return nil, fmt.Errorf("OpenAI error: %s", response.Error.Message)
	}

	firstItem := *(response.Data)
	data, err := base64.StdEncoding.DecodeString(firstItem[0].Base64Json)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 data: %v", err)
	}

	return data, nil
}
