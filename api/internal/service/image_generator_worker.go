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
	"github.com/getsentry/sentry-go"
	"github.com/riverqueue/river"
)

type ImageGeneratorArgs struct {
	DefinitionID int64        `json:"definitionId"`
	UserID       *int64       `json:"userId"`
	ErrorReport  *ErrorReport `json:"errorReport"`
}

func (ImageGeneratorArgs) Kind() string { return "ImageGenerator" }

func (w *ImageGeneratorWorker) Timeout(*river.Job[ImageGeneratorArgs]) time.Duration {
	return 4 * time.Minute
}

type ImageGeneratorWorker struct {
	river.WorkerDefaults[ImageGeneratorArgs]
	definitionService      *DefinitionService
	definitionImageService *DefinitionImageService
	userService            *UserService
	leitnerSystemStrategy  *LeitnerSystemStrategy
}

func NewImageGeneratorWorker(definitionService *DefinitionService, definitionImageService *DefinitionImageService, userService *UserService, leitnerSystemStrategy *LeitnerSystemStrategy) *ImageGeneratorWorker {
	return &ImageGeneratorWorker{
		definitionService:      definitionService,
		definitionImageService: definitionImageService,
		userService:            userService,
		leitnerSystemStrategy:  leitnerSystemStrategy,
	}
}

func (w *ImageGeneratorWorker) Work(ctx context.Context, job *river.Job[ImageGeneratorArgs]) error {
	var logger = common.Logger.With("worker", "imagegenerator")

	txn := sentry.StartTransaction(ctx, "worker.ImageGenerator", sentry.WithOpName("queue.process"), sentry.WithTransactionSource(sentry.SourceTask))
	defer txn.Finish()
	ctx = txn.Context()

	var (
		definitionID = job.Args.DefinitionID
		userID       = job.Args.UserID
	)

	// Validate user eligibility before processing (skip for admin/system jobs)
	if userID != nil {
		if err := w.userService.ValidateUserEligibilityForWorkers(ctx, *userID); err != nil {
			logger.Warn("User not eligible for image generation",
				"userId", *userID, "definitionId", definitionID, "error", err)
			// Cancel job permanently - user needs to upgrade
			return river.JobCancel(err)
		}
	}

	definition, err := w.definitionService.GetDefinitionByID(ctx, job.Args.DefinitionID)
	if err != nil {
		logger.Error("failed to get definition by id", "definitionId", job.Args.DefinitionID, "error", err)
		return imageDefinitionLookupError(err)
	}

	// Get language for language-aware image generation
	languageCode := definition.Language
	if languageCode == "" {
		return river.JobCancel(fmt.Errorf("no definition language for definition %d", definitionID))
	}

	// Using the longest example as the image description
	longestExample := definition.GetLongestExample()
	if longestExample == "" {
		return river.JobCancel(fmt.Errorf("no examples available for image generation, definition ID: %d", definitionID))
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

	span := sentry.StartSpan(ctx, "openai.image.generate", sentry.WithDescription("openai.GenerateImage"))
	data, err := generateWithOpenAI(span.Context(), prompt)
	span.Finish()

	if err != nil {
		return err
	}

	span = sentry.StartSpan(ctx, "storage.upload", sentry.WithDescription("minio.Upload"))
	url, err := common.Upload(span.Context(), data, "decorebator",
		fmt.Sprintf("images/definition-%d-%d.png", definitionID, time.Now().Unix()), "image/png")
	span.Finish()

	if err != nil {
		logger.Error("failed to upload image", "error", err)
		return err
	}

	logger.Debug("image generated", "definitionId", definitionID, "url", url)

	span = sentry.StartSpan(ctx, "db.definition_image.save", sentry.WithDescription("definitionImageService.SaveDefinitionImage"))
	_, err = w.definitionImageService.SaveDefinitionImage(span.Context(), model.CreateDefinitionImageDTO{
		API:          model.OPENAI,
		URL:          url,
		Description:  longestExample,
		Model:        "gpt-image-1.5",
		Prompt:       prompt,
		DefinitionID: definitionID,
	})
	span.Finish()

	if err != nil {
		logger.Error("failed to save definition image", "err", err)
		return err
	}

	// if this job was triggered by an error report, then mark the issue as solved
	if job.Args.ErrorReport != nil {
		return w.resolveErrorReport(ctx, definitionID, *job.Args.ErrorReport)
	}
	return nil
}

func imageDefinitionLookupError(err error) error {
	var notFoundErr common.NotFoundError
	if errors.As(err, &notFoundErr) {
		return river.JobCancel(err)
	}
	return err
}

func (w *ImageGeneratorWorker) resolveErrorReport(ctx context.Context, definitionID int64, report ErrorReport) error {
	if w.leitnerSystemStrategy == nil {
		return river.JobCancel(fmt.Errorf("no leitner strategy configured for definition %d", definitionID))
	}
	return w.leitnerSystemStrategy.MarkErrorResolved(ctx, report)
}

// buildImagePrompt creates a language-specific prompt for image generation
func buildImagePrompt(sentence, token, meaning, languageCode string) (string, error) {
	// Language-specific prompt templates
	templates := map[string]string{
		"en": `Create a single, clear, realistic scene that visually conveys the meaning of the word as used in the sentence below.
			The scene should be easy to understand at a glance and focused on one main action or situation.
			Do NOT depict the target word itself in the image or include the sentence text.
			Avoid violence and avoid logos or brand names.

			Sentence context: %s
			Target word: %s
			Intended meaning: %s`,
		"es": `Crea una escena realista, única y clara que transmita visualmente el significado de la palabra tal como se usa en la oración de abajo.
			La escena debe ser fácil de entender de un vistazo y centrarse en una sola acción o situación principal.
			NO muestres la palabra objetivo en la imagen ni incluyas el texto de la oración.
			Evita la violencia y evita logotipos o nombres de marcas.

			Contexto de la oración: %s
			Palabra objetivo: %s
			Significado previsto: %s`,
		"fr": `Créez une scène réaliste, unique et claire qui transmet visuellement le sens du mot tel qu’il est utilisé dans la phrase ci‑dessous.
			La scène doit être facile à comprendre d’un coup d’œil et se concentrer sur une seule action ou situation principale.
			Ne représentez PAS le mot cible dans l’image et n’incluez pas le texte de la phrase.
			Évitez la violence et évitez les logos ou noms de marque.

			Contexte de la phrase : %s
			Mot cible : %s
			Sens visé : %s`,
		"de": `Erstellen Sie eine einzelne, klare und realistische Szene, die die Bedeutung des Wortes im untenstehenden Satz visuell vermittelt.
			Die Szene soll auf einen Blick verständlich sein und sich auf eine Hauptaktion oder -situation konzentrieren.
			Das Zielwort nicht im Bild darstellen und den Satztext nicht einblenden.
			Vermeiden Sie Gewalt sowie Logos oder Markennamen.

			Satzkontext: %s
			Zielwort: %s
			Gemeinte Bedeutung: %s`,
		//nolint:misspell // "significato" is correct Italian prompt text.
		"it": `Crea una scena realistica, unica e chiara che trasmetta visivamente il significato della parola così come usata nella frase qui sotto.
			La scena deve essere facile da capire a colpo d’occhio e concentrarsi su un’unica azione o situazione principale.
			NON raffigurare la parola obiettivo nell’immagine e NON includere il testo della frase.
			Evita la violenza ed evita loghi o nomi di marca.

			Contesto della frase: %s
			Parola obiettivo: %s
			Significato previsto: %s`,
		"pt": `Crie uma cena realista, única e clara que transmita visualmente o significado da palavra conforme ela é usada na frase abaixo.
			A cena deve ser fácil de entender de relance e focar em uma única ação ou situação principal.
			NÃO represente a palavra‑alvo na imagem e NÃO inclua o texto da frase.
			Evite violência e evite logotipos ou nomes de marcas.

			Contexto da frase: %s
			Palavra‑alvo: %s
			Significado pretendido: %s`,
		"ja": `以下の文で使われている語の意味を、視覚的に伝える「単一で明確な現実的シーン」を作成してください。
			一目で理解できるように、主要な行動または状況を1つに絞ってください。
			対象語を画像内に描写せず、文のテキストも入れないでください。
			暴力表現やロゴ・ブランド名は避けてください。

			文の文脈: %s
			対象語: %s
			意図する意味: %s`,
	}

	template, exists := templates[languageCode]
	if !exists {
		return "", fmt.Errorf("unsupported language: %s", languageCode)
	}

	return fmt.Sprintf(template, sentence, token, meaning), nil
}

func generateWithOpenAI(ctx context.Context, prompt string) ([]byte, error) {
	var logger = common.Logger.With("func", "generateWithOpenAI")

	response, err := openai.GenerateImage(ctx, prompt)

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
			//nolint:gosec // G404 - using weak random for rate limiting jitter, not cryptographic security
			return nil, river.JobSnooze(time.Minute + (time.Duration(rand.Intn(60)) * time.Second))
		case "billing_hard_limit_reached":
			// [TODO] notification here elsewhere
			logger.Warn(response.Error.Message)
		case "content_policy_violation":
			jsonData, marshalErr := json.Marshal(response.Error)

			if marshalErr != nil {
				return nil, river.JobCancel(marshalErr)
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
