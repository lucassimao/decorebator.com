package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func main() {
	var (
		configPath   = flag.String("config", "cmd/store-assets/config.json", "Path to config JSON")
		action       = flag.String("action", "render", "Action: render, translate, guide, frame")
		store        = flag.String("store", "all", "Store: appstore, play, all")
		locale       = flag.String("locale", "en", "Locale code or all")
		slot         = flag.String("slot", "all", "Slot key or all")
		provider     = flag.String("provider", "openai", "Image provider: openai or gemini")
		openaiModel  = flag.String("openai-model", "", "Override OpenAI image model")
		geminiModel  = flag.String("gemini-model", "gemini-3-pro-image-preview", "Gemini image model")
		geminiSize   = flag.String("gemini-size", "2K", "Gemini image size: 1K, 2K, 4K")
		geminiAspect = flag.String("gemini-aspect", "9:16", "Gemini aspect ratio")
		creative     = flag.Int("creative", 2, "Creativity level (0-5)")
		attempts     = flag.Int("attempts", 3, "Max attempts per slot")
		invertMask   = flag.Bool("invert-mask", false, "Invert the generated mask")
		dryRun       = flag.Bool("dry-run", false, "Plan without writing output files")
		overwrite    = flag.Bool("overwrite", false, "Overwrite existing outputs")
	)
	flag.Parse()

	loadEnvFile()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		fatal(err)
	}

	switch *action {
	case "render":
		generator, genErr := buildGenerator(cfg, *provider, *openaiModel, *geminiModel, *geminiSize, *geminiAspect)
		if genErr != nil {
			err = genErr
			break
		}
		resizeMode := resizeModeForProvider(*provider)
		err = renderAssets(cfg, generator, resizeMode, *provider, *creative, *store, *locale, *slot, *attempts, *invertMask, *dryRun, *overwrite)
	case "translate":
		err = translateCopies(cfg, *locale, *dryRun)
	case "guide":
		err = generateGuideAndMask(cfg, *invertMask, *dryRun, *overwrite)
	case "frame":
		generator, genErr := buildGenerator(cfg, *provider, *openaiModel, *geminiModel, *geminiSize, *geminiAspect)
		if genErr != nil {
			err = genErr
			break
		}
		err = generateFrame(cfg, generator, *dryRun, *overwrite)
	default:
		err = fmt.Errorf("unknown action: %s", *action)
	}
	if err != nil {
		fatal(err)
	}
}

func loadEnvFile() {
	_ = godotenv.Load(".env")
}

func loadConfig(path string) (*Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(content, &cfg); err != nil {
		return nil, err
	}
	cfg.ConfigPath = path
	return &cfg, nil
}

func buildGenerator(cfg *Config, provider, openaiModel, geminiModel, geminiSize, geminiAspect string) (ImageGenerator, error) {
	switch strings.ToLower(provider) {
	case "openai":
		model := cfg.Models.Image
		if openaiModel != "" {
			model = openaiModel
		}
		return OpenAIImageGenerator{
			Model:      model,
			Quality:    "high",
			Background: "opaque",
		}, nil
	case "gemini":
		return GeminiImageGenerator{
			Model:       geminiModel,
			AspectRatio: geminiAspect,
			ImageSize:   geminiSize,
		}, nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}
}

func resizeModeForProvider(provider string) string {
	switch strings.ToLower(provider) {
	case "gemini":
		return "exact"
	default:
		return "openai"
	}
}
