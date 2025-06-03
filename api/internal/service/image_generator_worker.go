package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
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
	var longestExample string

	if job.Args.CustomPrompt != "" {
		prompt = job.Args.CustomPrompt
		longestExample = job.Args.CustomPrompt
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

		longestExample = ""
		// using the longest example as the image description
		for _, item := range definition.Examples {
			if len(item) > len(longestExample) {
				longestExample = item
			}
		}

		// Regular expression to find text within square brackets
		re := regexp.MustCompile(`\[(.*?)\]`)
		// Replace [word] with word (without brackets)
		result := re.ReplaceAllString(longestExample, "$1")

		matches := re.FindStringSubmatch(longestExample)
		var token string
		if matches != nil {
			token = matches[1]
		} else {
			token = definition.Token
		}

		prompt = fmt.Sprintf(`
			Render an image representing the following sentence: %s
			In that sentence, the word %s must convey %s
			You MUST NOT include any references to that sentence neither to the word %s in the generated image.
			`, result, token, definition.Meaning, token)
		fmt.Println(prompt)
	}

	data, err := generateWithOpenAI(prompt)

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
		Api:          model.OPENAI,
		URL:          url,
		Description:  longestExample,
		Model:        "gpt-image-1",
		Prompt:       prompt,
		DefinitionId: definitionID,
	})

	return err
}

func generateWithOpenAI(prompt string) ([]byte, error) {
	var logger = common.Logger.With("func", "generateWithOpenAI")

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
