package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/color"
	imagedraw "image/draw"
	"os"
	"path/filepath"
	"strings"
)

func generateGuideAndMask(cfg *Config, invertMask, dryRun, overwrite bool) error {
	_, _, err := ensureGuideAndMask(cfg, invertMask, dryRun, overwrite)
	return err
}

func ensureGuideAndMask(cfg *Config, invertMask, dryRun, overwrite bool) (string, string, error) {
	guidePath := filepath.Join(cfg.Defaults.GuideDir, "layout-guide.png")
	maskPath := filepath.Join(cfg.Defaults.MaskDir, "layout-mask.png")
	hashPath := filepath.Join(cfg.Defaults.GuideDir, "layout.hash")

	configHash, err := hashFile(cfg)
	if err != nil {
		return "", "", err
	}

	if !overwrite {
		if fileExists(guidePath) && fileExists(maskPath) && hashMatches(hashPath, configHash) {
			return guidePath, maskPath, nil
		}
	}

	if dryRun {
		fmt.Printf("generate guide -> %s\n", guidePath)
		fmt.Printf("generate mask -> %s\n", maskPath)
		return guidePath, maskPath, nil
	}

	if err := os.MkdirAll(cfg.Defaults.GuideDir, 0o755); err != nil {
		return "", "", err
	}
	if err := os.MkdirAll(cfg.Defaults.MaskDir, 0o755); err != nil {
		return "", "", err
	}

	guide := image.NewRGBA(image.Rect(0, 0, cfg.Defaults.BaseWidth, cfg.Defaults.BaseHeight))
	mask := image.NewNRGBA(image.Rect(0, 0, cfg.Defaults.BaseWidth, cfg.Defaults.BaseHeight))

	topHeight := int(float64(cfg.Defaults.BaseHeight) * cfg.Layout.TextBackdropHeightPct)
	textRect := image.Rect(0, 0, cfg.Defaults.BaseWidth, topHeight)
	fillRGBA(guide, textRect, color.NRGBA{R: 255, G: 255, B: 255, A: 40})

	screenshotRect := screenshotArea(cfg.Defaults.BaseWidth, cfg.Defaults.BaseHeight, cfg.Layout)
	fillRGBA(guide, screenshotRect, color.NRGBA{R: 255, G: 153, B: 0, A: 50})

	// Mask: opaque where we want to preserve screenshot, transparent elsewhere.
	if invertMask {
		fillRGBA(mask, image.Rect(0, 0, cfg.Defaults.BaseWidth, cfg.Defaults.BaseHeight), color.NRGBA{A: 255})
		fillRGBA(mask, screenshotRect, color.NRGBA{A: 0})
	} else {
		fillRGBA(mask, image.Rect(0, 0, cfg.Defaults.BaseWidth, cfg.Defaults.BaseHeight), color.NRGBA{A: 0})
		fillRGBA(mask, screenshotRect, color.NRGBA{A: 255})
	}

	if err := writeImage(guidePath, "png", guide); err != nil {
		return "", "", err
	}
	if err := writeImage(maskPath, "png", mask); err != nil {
		return "", "", err
	}
	if err := os.WriteFile(hashPath, []byte(configHash), 0o600); err != nil {
		return "", "", err
	}

	return guidePath, maskPath, nil
}

func fillRGBA(img imagedraw.Image, rect image.Rectangle, c color.NRGBA) {
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			img.Set(x, y, c)
		}
	}
}

func screenshotArea(canvasW, canvasH int, layout LayoutConfig) image.Rectangle {
	maxWidth := int(float64(canvasW) * layout.ScreenshotWidthPct)
	top := int(float64(canvasH) * layout.ScreenshotTopPct)
	bottom := int(float64(canvasH) * layout.ScreenshotBottomPct)
	maxHeight := canvasH - top - bottom

	areaW := maxWidth
	areaH := maxHeight
	if areaW > canvasW {
		areaW = canvasW
	}
	if areaH > canvasH {
		areaH = canvasH
	}

	x := (canvasW - areaW) / 2
	y := top + (maxHeight-areaH)/2
	return image.Rect(x, y, x+areaW, y+areaH)
}

func hashFile(cfg *Config) (string, error) {
	if cfg.ConfigPath == "" {
		return "", errors.New("config path is missing")
	}
	content, err := os.ReadFile(cfg.ConfigPath)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:]), nil
}

func hashMatches(path, expected string) bool {
	content, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(content)) == expected
}
