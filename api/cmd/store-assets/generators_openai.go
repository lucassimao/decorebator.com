package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	imagedraw "image/draw"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

func (g OpenAIImageGenerator) Name() string {
	return "openai:" + g.Model
}

func (g OpenAIImageGenerator) Generate(req ImageRequest) (image.Image, error) {
	if req.ScreenshotPath == "" && req.GuidePath == "" && req.MaskPath == "" {
		return openAIGenerateImage(g, req)
	}
	return openAIEditImage(g, req)
}

func openAIEditImage(generator OpenAIImageGenerator, req ImageRequest) (image.Image, error) {
	if generator.Model == "" {
		return nil, errors.New("OPENAI image model missing")
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, errors.New("OPENAI_API_KEY is not set")
	}

	preparedScreenshotPath, cleanup, err := prepareScreenshotForEdit(req)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("model", generator.Model)
	_ = writer.WriteField("prompt", req.Prompt)
	_ = writer.WriteField("size", fmt.Sprintf("%dx%d", req.BaseWidth, req.BaseHeight))
	if generator.Quality != "" {
		_ = writer.WriteField("quality", generator.Quality)
	}
	_ = writer.WriteField("output_format", "png")
	if generator.Background != "" {
		_ = writer.WriteField("background", generator.Background)
	}

	if addErr := addFormFileIndexed(writer, "image", 0, preparedScreenshotPath); addErr != nil {
		return nil, addErr
	}
	if addErr := addFormFileIndexed(writer, "image", 1, req.GuidePath); addErr != nil {
		return nil, addErr
	}
	if req.UseFrame && req.FramePath != "" {
		if addErr := addFormFileIndexed(writer, "image", 2, req.FramePath); addErr != nil {
			return nil, addErr
		}
	}
	if addErr := addFormFile(writer, "mask", req.MaskPath); addErr != nil {
		return nil, addErr
	}

	if closeErr := writer.Close(); closeErr != nil {
		return nil, closeErr
	}

	reqHTTP, err := http.NewRequest("POST", "https://api.openai.com/v1/images/edits", body)
	if err != nil {
		return nil, err
	}
	reqHTTP.Header.Set("Authorization", "Bearer "+apiKey)
	reqHTTP.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(reqHTTP)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var parsed OpenAIImageResponse
	if decodeErr := json.Unmarshal(respBody, &parsed); decodeErr != nil {
		return nil, decodeErr
	}
	if parsed.Error != nil {
		return nil, fmt.Errorf("openai image error: %s", parsed.Error.Message)
	}
	if len(parsed.Data) == 0 {
		return nil, errors.New("openai image response missing data")
	}

	decoded, err := base64.StdEncoding.DecodeString(parsed.Data[0].Base64Json)
	if err != nil {
		return nil, err
	}
	return decodeImage(bytes.NewReader(decoded))
}

func openAIGenerateImage(generator OpenAIImageGenerator, req ImageRequest) (image.Image, error) {
	if generator.Model == "" {
		return nil, errors.New("OPENAI image model missing")
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, errors.New("OPENAI_API_KEY is not set")
	}

	payload := map[string]any{
		"model":         generator.Model,
		"prompt":        req.Prompt,
		"n":             1,
		"size":          fmt.Sprintf("%dx%d", req.BaseWidth, req.BaseHeight),
		"output_format": "png",
	}
	if generator.Quality != "" {
		payload["quality"] = generator.Quality
	}
	if req.TransparentBackground {
		payload["background"] = "transparent"
	} else if generator.Background != "" {
		payload["background"] = generator.Background
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	reqHTTP, err := http.NewRequest("POST", "https://api.openai.com/v1/images/generations", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	reqHTTP.Header.Set("Authorization", "Bearer "+apiKey)
	reqHTTP.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(reqHTTP)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var parsed OpenAIImageResponse
	if decodeErr := json.Unmarshal(respBody, &parsed); decodeErr != nil {
		return nil, decodeErr
	}
	if parsed.Error != nil {
		return nil, fmt.Errorf("openai image error: %s", parsed.Error.Message)
	}
	if len(parsed.Data) == 0 {
		return nil, errors.New("openai image response missing data")
	}

	decoded, err := base64.StdEncoding.DecodeString(parsed.Data[0].Base64Json)
	if err != nil {
		return nil, err
	}
	return decodeImage(bytes.NewReader(decoded))
}

func prepareScreenshotForEdit(req ImageRequest) (string, func(), error) {
	img, err := loadImage(req.ScreenshotPath)
	if err != nil {
		return "", func() {}, err
	}

	canvas := image.NewNRGBA(image.Rect(0, 0, req.BaseWidth, req.BaseHeight))
	// Transparent background; place screenshot into the expected area.
	rect := screenshotArea(req.BaseWidth, req.BaseHeight, req.Layout)
	scaled := scaleToFit(img, rect.Dx(), rect.Dy())
	offsetX := rect.Min.X + (rect.Dx()-scaled.Bounds().Dx())/2
	offsetY := rect.Min.Y + (rect.Dy()-scaled.Bounds().Dy())/2
	imagedraw.Draw(canvas, image.Rect(offsetX, offsetY, offsetX+scaled.Bounds().Dx(), offsetY+scaled.Bounds().Dy()), scaled, image.Point{}, imagedraw.Over)

	file, err := os.CreateTemp("", "store-assets-screenshot-*.png")
	if err != nil {
		return "", func() {}, err
	}
	path := file.Name()
	file.Close()

	if err := writeImage(path, "png", canvas); err != nil {
		_ = os.Remove(path)
		return "", func() {}, err
	}

	cleanup := func() {
		_ = os.Remove(path)
	}

	return path, cleanup, nil
}
