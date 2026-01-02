package main

import "image"

type Config struct {
	Locales    []string               `json:"locales"`
	Devices    []string               `json:"devices"`
	Slots      []SlotConfig           `json:"slots"`
	Stores     map[string]StoreConfig `json:"stores"`
	Layout     LayoutConfig           `json:"layout"`
	Defaults   DefaultsConfig         `json:"defaults"`
	Models     ModelsConfig           `json:"models"`
	ConfigPath string                 `json:"-"`
}

type SlotConfig struct {
	ID       int      `json:"id"`
	Key      string   `json:"key"`
	Keywords []string `json:"keywords"`
}

type StoreConfig struct {
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Format string `json:"format"`
}

type LayoutConfig struct {
	TextTopPct            float64 `json:"textTopPct"`
	TextSidePct           float64 `json:"textSidePct"`
	TextMaxWidthPct       float64 `json:"textMaxWidthPct"`
	TextBackdropHeightPct float64 `json:"textBackdropHeightPct"`
	ScreenshotWidthPct    float64 `json:"screenshotWidthPct"`
	ScreenshotTopPct      float64 `json:"screenshotTopPct"`
	ScreenshotBottomPct   float64 `json:"screenshotBottomPct"`
}

type DefaultsConfig struct {
	CopyDir       string `json:"copyDir"`
	ScreenshotDir string `json:"screenshotDir"`
	GuideDir      string `json:"guideDir"`
	MaskDir       string `json:"maskDir"`
	FramePath     string `json:"framePath"`
	FramePathApp  string `json:"framePathApp"`
	FramePathPlay string `json:"framePathPlay"`
	OutputDir     string `json:"outputDir"`
	BaseWidth     int    `json:"baseWidth"`
	BaseHeight    int    `json:"baseHeight"`
}

type ModelsConfig struct {
	Image       string `json:"image"`
	Translation string `json:"translation"`
}

type CopyEntry struct {
	Headline string `json:"headline"`
	Subhead  string `json:"subhead"`
}

type CopyFile map[string]CopyEntry

type OpenAIImageResponse struct {
	Data []struct {
		Base64Json string `json:"b64_json"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

type OpenAIChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text       string `json:"text,omitempty"`
				InlineData *struct {
					MimeType string `json:"mime_type,omitempty"`
					Data     string `json:"data,omitempty"`
				} `json:"inline_data,omitempty"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

type ImageRequest struct {
	Prompt                string
	BaseWidth             int
	BaseHeight            int
	Layout                LayoutConfig
	ScreenshotPath        string
	GuidePath             string
	MaskPath              string
	FramePath             string
	UseFrame              bool
	TransparentBackground bool
}

type ImageGenerator interface {
	Generate(req ImageRequest) (image.Image, error)
	Name() string
}

type OpenAIImageGenerator struct {
	Model      string
	Quality    string
	Background string
}

type GeminiImageGenerator struct {
	Model       string
	AspectRatio string
	ImageSize   string
}
