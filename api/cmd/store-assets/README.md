# Store Assets Generator

Local-only CLI for generating App Store and Google Play screenshots using GPT Image edits. The model composes the full image (background + typography) while preserving your screenshot via a mask.

## Setup

- Set `OPENAI_API_KEY` in your environment (or in `api/.env`).
- Add your iPhone 11 screenshots to:
  - `api/cmd/store-assets/screenshots/{locale}/iphone11/{slot}.png|.jpg|.jpeg`
- Add device frame images (optional):
  - App Store (iPhone 11): `api/cmd/store-assets/frames/iphone11.png`
  - Play Store (Pixel 8 Pro): `api/cmd/store-assets/frames/pixel8pro.png`
- Edit copy in `api/cmd/store-assets/store-copy/en.json` (English). Other locales are auto-generated.
- Run `go mod tidy` in `api/` after dependency changes.

## Commands

Translate copy:

```
cd api
go run ./cmd/store-assets -action translate -locale all
```

Generate layout guide + mask (auto-created on render if missing):

```
cd api
go run ./cmd/store-assets -action guide
```

Generate iPhone frame (auto for App Store use):

```
cd api
go run ./cmd/store-assets -action frame -provider openai
```

Render assets (full GPT assembly, max 3 attempts per slot):

```
cd api
go run ./cmd/store-assets -action render -store all -locale all -slot all -attempts 3 -creative 2
```

Use Gemini instead of OpenAI:

```
cd api
go run ./cmd/store-assets -action render -provider gemini -gemini-model gemini-3-pro-image-preview -gemini-size 2K -gemini-aspect 9:16 -store all -locale all -slot all -attempts 3 -creative 3
```

## Outputs

- Final images: `api/.local/store-assets/{store}/{locale}/`
- If `-attempts` > 1, outputs are saved as `slot_try1.png`, `slot_try2.png`, etc.

## How it works

- The tool generates (or reuses) a 1024x1536 layout guide and mask.
- The guide/mask are regenerated automatically if `config.json` changes (hash check).
- Your screenshot is pre-composited onto a 1024x1536 transparent canvas in the expected area so the mask size matches the first image.
- The Images edits API is called with:
  - `image[0]`: prepared screenshot canvas
  - `image[1]`: guide image (layout reference only)
  - `mask`: preserves the screenshot area
- The output is resized to App Store (1242x2688) with letterboxing and edge-extended padding (no crop), and to Play Store (1080x1920) with scale-to-fill.

## Configuration

- `api/cmd/store-assets/config.json` controls sizes, layout zones, and paths.
- `models.image` is set to `gpt-image-1.5` by default. Use `-openai-model` to override.
- Gemini uses `GEMINI_API_KEY` and supports 1K/2K/4K with aspect ratios like 9:16.

## Notes

- The model is asked to render text directly; occasional errors are expected.
- Use `-invert-mask` if the API treats the mask as the inverse of what you expect.
- Provider flags:
  - `-provider openai|gemini`
  - `-openai-model <model>`
  - `-gemini-model <model>`
  - `-gemini-size 1K|2K|4K`
  - `-gemini-aspect 9:16`
  - `-creative 0-5` (higher = more expressive visuals)
