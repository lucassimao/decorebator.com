package workers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/definitions"
	"github.com/riverqueue/river"
)

var logger = common.Logger.With("module", "imagegenerator")

type ImageGeneratorArgs struct {
	DefinitionId int64 `json:"definitionId"`
}

func (ImageGeneratorArgs) Kind() string { return "ImageGenerator" }

func (ImageGeneratorArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: IMAGE_GENERATOR_QUEUE,
	}
}

type ImageGeneratorWorker struct {
	river.WorkerDefaults[ImageGeneratorArgs]
}

func (w *ImageGeneratorWorker) Work(ctx context.Context, job *river.Job[ImageGeneratorArgs]) error {

	definition, err := definitions.GetById(job.Args.DefinitionId)
	if err != nil {

		if errors.Is(err, &common.NotFoundError{}) {
			logger.Error("skipping image generation: inexisting definition", "definitionId", job.Args.DefinitionId)
			return nil
		}

		logger.Error("failed to get definition by id", "definitionId", job.Args.DefinitionId, "error", err)
		return err
	}

	var prompt = fmt.Sprintf("Illustrate the %s %s based on the following meaning: %s", definition.PartOfSpeech, definition.Token, definition.Meaning)
	response, err := callOpenAIImageGeneration(prompt)

	if err != nil {
		// track potential causes here and decide if return nil or not
		logger.Error("failed to generate image", "error", err)
		return err
	}

	// track potential causes here and decide if return nil or not
	if response.Error != nil {
		logger.Error("failed to generate image", "body", response.Error)

		if response.Error.Code == "billing_hard_limit_reached" {
			// notification here elsewhere
			logger.Warn(response.Error.Message)
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
		fmt.Sprintf("definition-%d-%d.png", definition.ID, time.Now().Unix()), "image/png")

	if err != nil {
		logger.Error("failed to upload image", "error", err)
		return err
	}

	logger.Debug("image generated", "definitionId", definition.ID, "url", url)
	return definitions.SetImage(definition.ID, url)
}

type ImageGenerationResponse struct {
	Created *int64 `json:"created"`
	Data    *[]struct {
		Base64Json string `json:"b64_json"`
	} `json:"data"`

	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Param   string `json:"param"`
		Type    string `json:"type"`
	} `json:"error"`
}

func callOpenAIImageGeneration(prompt string) (*ImageGenerationResponse, error) {
	var requestBodyStruct = map[string]any{
		"model":           "dall-e-3",
		"prompt":          prompt,
		"n":               1,
		"size":            "1024x1024",
		"style":           "natural",
		"response_format": "b64_json",
	}

	var requestBody, err = json.Marshal(requestBodyStruct)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request data: %w", err)
	}

	req, err := http.NewRequest("POST", common.Env.OpenaiImageGenerationApiEndpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", common.Env.OpenaiApiKey))

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request API endpoint: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var imageGenerationResponse ImageGenerationResponse
	err = json.Unmarshal(body, &imageGenerationResponse)

	if err != nil {
		return nil, fmt.Errorf("failed to unmarshall response body: %w", err)
	}

	return &imageGenerationResponse, nil
}
