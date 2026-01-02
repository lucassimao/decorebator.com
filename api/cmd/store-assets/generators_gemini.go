package main

import (
	"bytes"
	"context"
	"errors"
	"image"
	"os"
	"time"

	"google.golang.org/genai"
)

func (g GeminiImageGenerator) Name() string {
	return "gemini:" + g.Model
}

func (g GeminiImageGenerator) Generate(req ImageRequest) (image.Image, error) {
	return geminiGenerateImage(g, req)
}

func geminiGenerateImage(generator GeminiImageGenerator, req ImageRequest) (image.Image, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, errors.New("GEMINI_API_KEY is not set")
	}

	var err error
	var preparedScreenshotPath string
	var cleanup func()
	if req.ScreenshotPath != "" {
		preparedScreenshotPath, cleanup, err = prepareScreenshotForEdit(req)
		if err != nil {
			return nil, err
		}
		defer cleanup()
	}

	var screenshotData []byte
	var screenshotMime string
	if preparedScreenshotPath != "" {
		screenshotData, screenshotMime, err = readFileBytes(preparedScreenshotPath)
		if err != nil {
			return nil, err
		}
	}

	var guideData []byte
	var guideMime string
	if req.GuidePath != "" {
		guideData, guideMime, err = readFileBytes(req.GuidePath)
		if err != nil {
			return nil, err
		}
	}
	var frameData []byte
	var frameMime string
	if req.UseFrame && req.FramePath != "" {
		frameData, frameMime, err = readFileBytes(req.FramePath)
		if err != nil {
			return nil, err
		}
	}

	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	parts := []*genai.Part{
		{Text: req.Prompt},
	}
	if len(screenshotData) > 0 {
		parts = append(parts, &genai.Part{InlineData: &genai.Blob{Data: screenshotData, MIMEType: screenshotMime}})
	}
	if len(guideData) > 0 {
		parts = append(parts, &genai.Part{InlineData: &genai.Blob{Data: guideData, MIMEType: guideMime}})
	}
	if req.UseFrame && len(frameData) > 0 {
		parts = append(parts, &genai.Part{InlineData: &genai.Blob{Data: frameData, MIMEType: frameMime}})
	}

	resp, err := client.Models.GenerateContent(ctx, generator.Model, []*genai.Content{
		{
			Role:  "user",
			Parts: parts,
		},
	}, &genai.GenerateContentConfig{
		ResponseModalities: []string{"TEXT", "IMAGE"},
		ImageConfig: &genai.ImageConfig{
			AspectRatio: generator.AspectRatio,
			ImageSize:   generator.ImageSize,
		},
	})
	if err != nil {
		return nil, err
	}

	for _, candidate := range resp.Candidates {
		for _, part := range candidate.Content.Parts {
			if part.InlineData != nil && len(part.InlineData.Data) > 0 {
				return decodeImage(bytes.NewReader(part.InlineData.Data))
			}
		}
	}

	return nil, errors.New("gemini image response missing inline image data")
}
