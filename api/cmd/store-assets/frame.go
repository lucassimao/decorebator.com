package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func generateFrame(cfg *Config, generator ImageGenerator, dryRun, overwrite bool) error {
	targets := []struct {
		label  string
		path   string
		prompt string
	}{
		{
			label:  "iphone11",
			path:   pickFramePath(cfg.Defaults.FramePathApp, cfg.Defaults.FramePath),
			prompt: "Front-facing iPhone 11 device frame with thin bezels and rounded corners. Transparent center area, no screen content, no text, no logos, no background. The frame should be clean, realistic, and centered.",
		},
		{
			label:  "pixel8pro",
			path:   cfg.Defaults.FramePathPlay,
			prompt: "Front-facing Google Pixel 8 Pro device frame with rounded corners, thin bezels, and a centered camera hole. Transparent center area, no screen content, no text, no logos, no background. The frame should be clean, realistic, and centered.",
		},
	}

	for _, target := range targets {
		if target.path == "" {
			continue
		}
		if !overwrite {
			if _, err := os.Stat(target.path); err == nil {
				fmt.Printf("skip existing frame: %s\n", target.path)
				continue
			}
		}

		if dryRun {
			fmt.Printf("generate frame (%s) -> %s\n", target.label, target.path)
			continue
		}

		img, err := generator.Generate(ImageRequest{
			Prompt:                target.prompt,
			BaseWidth:             cfg.Defaults.BaseWidth,
			BaseHeight:            cfg.Defaults.BaseHeight,
			Layout:                cfg.Layout,
			TransparentBackground: true,
		})
		if err != nil {
			return err
		}

		if err := os.MkdirAll(filepath.Dir(target.path), 0o755); err != nil {
			return err
		}
		if err := writeImage(target.path, "png", img); err != nil {
			return err
		}

		fmt.Printf("wrote %s\n", target.path)
	}

	return nil
}

func pickFramePath(preferred, fallback string) string {
	if preferred != "" {
		return preferred
	}
	return fallback
}
