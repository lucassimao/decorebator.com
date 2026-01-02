package main

import (
	"fmt"
	"image"
	"strings"
)

func buildFullPrompt(cfg *Config, provider string, creative int, store string, storeCfg StoreConfig, slot SlotConfig, copyEntry CopyEntry) string {
	keywords := ""
	if len(slot.Keywords) > 0 {
		keywords = strings.Join(slot.Keywords, ", ")
	}

	textSide := int(cfg.Layout.TextSidePct * 100)
	textHeight := int(cfg.Layout.TextBackdropHeightPct * 100)
	textBox := textBounds(cfg)
	storeSafe := storeSafeLine(store)
	creativeBlock := creativeGuidance(provider, creative)
	providerBlock := providerGuidance(provider, store)
	frameBlock := frameLine(store)

	return strings.TrimSpace(fmt.Sprintf(`
Create a polished %s store screenshot cover (portrait). Output is 1024x1536.
- Use the provided screenshot image as the app UI and keep it unchanged unless absolutely necessary for blending edges.
- Use the second input image only as a layout guide; do not reproduce its shapes or colors.
- The mask indicates the area to preserve.

Text to render exactly:
Headline: "%s"
Subhead: "%s"

Typography requirements:
- Large, bold headline; smaller subhead; high contrast; no curved or distorted text.
- Keep all text fully inside the top %d%% of the canvas, with at least %d%% side margins.
- Text must fit entirely inside this box (pixels on 1024x1536): left=%d, right=%d, top=%d, bottom=%d.
- Add generous top padding and do not place text near the top edge.
- If the text does not fit, reduce font size so all words are visible.
- No extra text, no logos, no watermarks.
%s

Layout guidance:
- Keep the screenshot centered within the lower area, leaving the top text zone clear.

Visual style:
- Be creative with color and composition, but keep it clean and legible.
- Avoid busy patterns behind the text; ensure readability.
- Keep background elements away from the device/screenshot area and avoid colors/patterns that visually merge with the UI.
 %s
%s
%s
%s

Store output target:
- Final will be resized to %dx%d.
`, store, copyEntry.Headline, copyEntry.Subhead, textHeight, textSide, textBox.Min.X, textBox.Max.X, textBox.Min.Y, textBox.Max.Y, storeSafe, keywordsLine(keywords), providerBlock, creativeBlock, frameBlock, storeCfg.Width, storeCfg.Height))
}

func keywordsLine(keywords string) string {
	if keywords == "" {
		return ""
	}
	return fmt.Sprintf("- Convey themes: %s.", keywords)
}

func storeSafeLine(store string) string {
	if store != "appstore" {
		return ""
	}
	return "- App Store safe area: leave at least 140px empty space above the headline; top of the headline must be below y=200px."
}

func providerGuidance(provider, store string) string {
	switch strings.ToLower(provider) {
	case "openai":
		return "- Prioritize precise typography and clean edges; avoid stylized or distorted letterforms. Ensure the background colors contrast with the screenshot so UI elements remain distinct."
	case "gemini":
		return "- Respect reference images strongly; keep the screenshot faithful and avoid altering UI elements. Ensure the background colors contrast with the screenshot so UI elements remain distinct."
	default:
		return ""
	}
}

func frameLine(store string) string {
	switch store {
	case "appstore":
		return "- Use the provided iPhone device frame image and wrap the screenshot inside it. The frame must be fully visible."
	case "play":
		return "- Use the provided Android device frame image and wrap the screenshot inside it. The frame must be fully visible."
	default:
		return ""
	}
}

func creativeGuidance(provider string, creative int) string {
	if creative < 0 {
		creative = 0
	}
	if creative > 5 {
		creative = 5
	}
	switch creative {
	case 0:
		return "- Creativity: keep composition conservative, minimal, and straightforward."
	case 1:
		return "- Creativity: subtle variation in gradients and lighting; restrained shapes."
	case 2:
		return "- Creativity: moderate; introduce soft abstract forms and layered depth."
	case 3:
		return "- Creativity: bold; use distinct color palettes, light flares, and geometric accents."
	case 4:
		return "- Creativity: high; add dynamic composition, surprising motifs, and expressive lighting."
	case 5:
		return "- Creativity: very high; cinematic contrast, rich textures, and imaginative abstract elements while preserving text clarity."
	default:
		return ""
	}
}

func textBounds(cfg *Config) image.Rectangle {
	width := cfg.Defaults.BaseWidth
	height := cfg.Defaults.BaseHeight
	left := int(float64(width) * cfg.Layout.TextSidePct)
	right := int(float64(width) * (1 - cfg.Layout.TextSidePct))
	top := int(float64(height) * cfg.Layout.TextTopPct)
	bottom := int(float64(height) * (cfg.Layout.TextTopPct + cfg.Layout.TextBackdropHeightPct))
	if bottom > height {
		bottom = height
	}
	if right > width {
		right = width
	}
	return image.Rect(left, top, right, bottom)
}
