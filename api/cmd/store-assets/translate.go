package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func translateCopies(cfg *Config, localeFilter string, dryRun bool) error {
	locales := selectLocales(cfg, localeFilter)
	if len(locales) == 0 {
		return fmt.Errorf("no locales matched: %s", localeFilter)
	}

	enCopy, err := loadCopyFile(cfg, "en")
	if err != nil {
		return err
	}

	for _, loc := range locales {
		if loc == "en" {
			continue
		}
		translated, err := translateCopy(cfg, loc, enCopy)
		if err != nil {
			return err
		}
		path := filepath.Join(cfg.Defaults.CopyDir, loc+".json")
		if dryRun {
			fmt.Printf("translate %s -> %s\n", loc, path)
			continue
		}
		if err := os.MkdirAll(cfg.Defaults.CopyDir, 0o755); err != nil {
			return err
		}
		content, err := json.MarshalIndent(translated, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(path, content, 0o644); err != nil {
			return err
		}
		fmt.Printf("wrote %s\n", path)
	}

	return nil
}

func translateCopy(cfg *Config, locale string, copyFile CopyFile) (CopyFile, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, errors.New("OPENAI_API_KEY is not set")
	}

	payload := map[string]any{
		"model": cfg.Models.Translation,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are a professional translator for app store screenshot text. Translate carefully, keep tone concise, and return only valid JSON.",
			},
			{
				"role":    "user",
				"content": buildTranslationPrompt(locale, copyFile),
			},
		},
		"temperature": 0.2,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var parsed OpenAIChatResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, err
	}
	if parsed.Error != nil {
		return nil, fmt.Errorf("openai translation error: %s", parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 {
		return nil, errors.New("openai translation response missing choices")
	}

	content := strings.TrimSpace(parsed.Choices[0].Message.Content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var translated CopyFile
	if err := json.Unmarshal([]byte(content), &translated); err != nil {
		return nil, fmt.Errorf("failed to parse translation JSON: %w", err)
	}
	return translated, nil
}

func buildTranslationPrompt(locale string, copyFile CopyFile) string {
	payload, _ := json.MarshalIndent(copyFile, "", "  ")
	return fmt.Sprintf("Translate this JSON to locale %s. Keep keys unchanged. Preserve concise marketing tone. Output JSON only.\n\n%s", locale, string(payload))
}
