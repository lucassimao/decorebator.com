# Store Assets CLI

## Scope
Local-only tool for generating App Store and Google Play cover images from real app screenshots + localized copy using AI image models.

## Core Features
- **Providers**: OpenAI (`gpt-image-1.5`) and Gemini (`gemini-3-pro-image-preview`).
- **Slots**: dashboard, quiz, ai-content, flashcards, analytics, premium, voice-coach.
- **Multi-language**: source English copy with auto-translation via OpenAI.
- **Guided layouts**: guide + mask, auto-regenerated when config changes.
- **Frames**: optional iPhone (App Store) + Android (Play Store) device frames.
- **Creativity knob**: `-creative 0-5` controls prompt style intensity.
- **Parallel renders**: default concurrency matches attempts.

## Inputs
- Screenshots: `api/cmd/store-assets/screenshots/{locale}/iphone11/{slot}.png|.jpg|.jpeg`
- Copy: `api/cmd/store-assets/store-copy/{locale}.json`
- Frames (optional):
  - `api/cmd/store-assets/frames/iphone11.png`
  - `api/cmd/store-assets/frames/pixel8pro.png`

## Outputs
- Rendered images: `api/.local/store-assets/{store}/{locale}/`
- Guides/masks: `api/cmd/store-assets/guides/`

## Key Commands
- Render:
  - `go run ./cmd/store-assets -action render -provider openai -store all -locale en -slot all -attempts 3 -creative 2`
  - `go run ./cmd/store-assets -action render -provider gemini -gemini-size 2K -gemini-aspect 9:16 -store all -locale en -slot all -attempts 3 -creative 3`
- Translate:
  - `go run ./cmd/store-assets -action translate -locale all`
- Guides:
  - `go run ./cmd/store-assets -action guide`
- Frames:
  - `go run ./cmd/store-assets -action frame -provider openai`

## Notes
- OpenAI uses masked edits; Gemini uses reference images + layout guide.
- App Store sizes are configured in `config.json` and must match Apple’s accepted dimensions.
- App Store can render with or without frames (frame optional). Play Store uses Android frame if present.
