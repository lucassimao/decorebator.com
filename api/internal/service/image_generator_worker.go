package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
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
	DefinitionId int64        `json:"definitionId"`
	UserID       int64        `json:"userId"`
	ErrorReport  *ErrorReport `json:"errorReport"`
}

func (ImageGeneratorArgs) Kind() string { return "ImageGenerator" }

type ImageGeneratorWorker struct {
	river.WorkerDefaults[ImageGeneratorArgs]
}

func (w *ImageGeneratorWorker) Work(ctx context.Context, job *river.Job[ImageGeneratorArgs]) error {
	var logger = common.Logger.With("worker", "imagegenerator")

	var (
		definitionID = job.Args.DefinitionId
		userID       = job.Args.UserID
	)

	// Validate user eligibility before processing
	if err := ValidateUserEligibilityForWorkers(userID); err != nil {
		logger.Warn("User not eligible for image generation",
			"userId", userID, "definitionId", definitionID, "error", err)
		// Cancel job permanently - user needs to upgrade
		return river.JobCancel(err)
	}

	definition, err := GetDefinitionById(job.Args.DefinitionId)
	if err != nil {
		logger.Error("failed to get definition by id", "definitionId", job.Args.DefinitionId, "error", err)
		return river.JobCancel(err)
	}

	// Get language for language-aware image generation
	languageCode := definition.Language
	if languageCode == "" {
		return river.JobCancel(fmt.Errorf("no definition language for definition %d", definitionID))
	}

	var longestExample string
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

	prompt, err := buildImagePrompt(result, token, definition.Meaning, languageCode)
	if err != nil {
		return river.JobCancel(err)
	}
	logger.Debug(prompt)

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

	if err != nil {
		logger.Error("failed to save definition image", "err", err)
		return err
	}

	// if this job was triggered by an error report, then mark the issue as solved
	if job.Args.ErrorReport != nil {
		var strategy LeitnerSystemStrategy
		return strategy.MarkErrorResolved(*job.Args.ErrorReport)
	}
	return nil
}

// buildImagePrompt creates a language-specific prompt for image generation
func buildImagePrompt(sentence, token, meaning, languageCode string) (string, error) {
	// Language-specific prompt templates
	templates := map[string]string{
		"en": `Render an image representing the following sentence: %s
			In that sentence, the word %s must convey %s
			You MUST NOT include any references to that sentence neither to the word %s in the generated image.`,
		"es": `Renderiza una imagen que represente la siguiente oración: %s
			En esa oración, la palabra %s debe transmitir %s
			NO DEBES incluir ninguna referencia a esa oración ni a la palabra %s en la imagen generada.`,
		"fr": `Créez une image représentant la phrase suivante : %s
			Dans cette phrase, le mot %s doit véhiculer %s
			Vous NE DEVEZ PAS inclure de références à cette phrase ni au mot %s dans l'image générée.`,
		"de": `Erstellen Sie ein Bild, das den folgenden Satz darstellt: %s
			In diesem Satz muss das Wort %s %s vermitteln
			Sie DÜRFEN KEINE Verweise auf diesen Satz oder das Wort %s in das generierte Bild einbeziehen.`,
		"it": `Crea un'immagine che rappresenti la seguente frase: %s
			In quella frase, la parola %s deve trasmettere %s
			NON DEVI includere riferimenti a quella frase né alla parola %s nell'immagine generata.`,
		"pt": `Crie uma imagem representando a seguinte frase: %s
			Nessa frase, a palavra %s deve transmitir %s
			Você NÃO DEVE incluir referências a essa frase nem à palavra %s na imagem gerada.`,
		"ja": `次の文を表す画像を作成してください：%s
			その文で、単語 %s は %s を伝える必要があります
			生成された画像にその文や単語 %s への言及を含めてはいけません。`,
	}

	template, exists := templates[languageCode]
	if !exists {
		return "", fmt.Errorf("unsupported language: %s", languageCode)
	}

	return fmt.Sprintf(template, sentence, token, meaning, token), nil
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
