package service

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"

	"decorebator.com/internal/common"
	"decorebator.com/internal/openai"
	"decorebator.com/internal/repository"
	"github.com/riverqueue/river"
	"golang.org/x/image/draw"
)

type PublicQuizOgArgs struct {
	QuizID int64 `json:"quizId"`
}

func (PublicQuizOgArgs) Kind() string { return "public_quiz_og" }

type PublicQuizOgWorker struct {
	river.WorkerDefaults[PublicQuizOgArgs]
	repo *repository.PublicQuizRepository
}

func NewPublicQuizOgWorker(repo *repository.PublicQuizRepository) *PublicQuizOgWorker {
	return &PublicQuizOgWorker{repo: repo}
}

func (w *PublicQuizOgWorker) Work(ctx context.Context, job *river.Job[PublicQuizOgArgs]) error {
	// Load quiz
	onlyActive := true
	limit := 1
	list, err := w.repo.Find(ctx, repository.FindPublicQuizArgs{ID: &job.Args.QuizID, OnlyActive: &onlyActive, Limit: &limit})
	if err != nil || len(list) == 0 {
		return nil // nothing to do
	}
	pq := list[0]
	if pq.PreviewImageURL != nil && *pq.PreviewImageURL != "" {
		return nil // already has OG image
	}
	// Build prompt
	prompt := fmt.Sprintf("Create a clean, minimal, friendly cover image for a vocabulary quiz titled '%s'. Soft gradients, subtle educational vibe, no text, bright accent.", pq.Title)
	if pq.Description != nil && *pq.Description != "" {
		prompt = fmt.Sprintf("%s Description hint: %s", prompt, *pq.Description)
	}
	resp, genErr := openai.GenerateImage(prompt)
	if genErr != nil || resp == nil || resp.Error != nil || resp.Data == nil || len(*resp.Data) == 0 {
		return nil
	}
	first := (*resp.Data)[0]
	data, contentType, decErr := common.DecodeImageBase64(first.Base64Json)
	if decErr != nil {
		return nil
	}
	// Upload master
	masterObject := fmt.Sprintf("og/public-quizzes/%s-master.jpg", pq.Slug)
	ctype := contentType
	if ctype == "" {
		ctype = "image/jpeg"
	}
	url, upErr := common.Upload(ctx, data, "decorebator", masterObject, ctype)
	if upErr != nil {
		return nil
	}
	_ = w.repo.UpdatePreviewImageURL(ctx, pq.ID, url)

	// Decode master for resizing
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil
	}

	// Prepare variants for major networks
	type variant struct {
		name string
		w    int
		h    int
	}
	variants := []variant{
		{name: "twitter_x", w: 1200, h: 630},           // X/Twitter, Facebook, Telegram, WhatsApp
		{name: "facebook", w: 1200, h: 630},            // Facebook OG
		{name: "telegram", w: 1200, h: 630},            // Telegram link preview
		{name: "whatsapp", w: 1200, h: 630},            // WhatsApp preview
		{name: "square_1080", w: 1080, h: 1080},        // Instagram feed
		{name: "portrait_1080x1920", w: 1080, h: 1920}, // Stories/Reels/TikTok
	}

	for _, v := range variants {
		resized := resizeCover(img, v.w, v.h)
		var buf bytes.Buffer
		// Use high quality but reasonable size
		_ = jpeg.Encode(&buf, resized, &jpeg.Options{Quality: 85})
		object := fmt.Sprintf("og/public-quizzes/%s-%s.jpg", pq.Slug, v.name)
		_, _ = common.Upload(ctx, buf.Bytes(), "decorebator", object, "image/jpeg")
	}
	return nil
}

// resizeCover scales the source image to fully cover the target size, then center-crops
func resizeCover(src image.Image, targetW, targetH int) image.Image {
	srcW := src.Bounds().Dx()
	srcH := src.Bounds().Dy()
	if srcW == 0 || srcH == 0 || targetW == 0 || targetH == 0 {
		return src
	}
	scaleX := float64(targetW) / float64(srcW)
	scaleY := float64(targetH) / float64(srcH)
	scale := scaleX
	if scaleY > scaleX {
		scale = scaleY
	}
	// Scaled intermediate size
	interW := int(float64(srcW) * scale)
	interH := int(float64(srcH) * scale)
	// Resize to intermediate
	inter := image.NewRGBA(image.Rect(0, 0, interW, interH))
	draw.CatmullRom.Scale(inter, inter.Bounds(), src, src.Bounds(), draw.Over, nil)
	// Center crop to target
	offX := (interW - targetW) / 2
	offY := (interH - targetH) / 2
	crop := image.NewRGBA(image.Rect(0, 0, targetW, targetH))
	draw.Copy(crop, image.Pt(0, 0), inter, image.Rect(offX, offY, offX+targetW, offY+targetH), draw.Over, nil)
	return crop
}
