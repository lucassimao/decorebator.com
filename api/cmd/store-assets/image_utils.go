package main

import (
	"image"
	imagedraw "image/draw"

	"golang.org/x/image/draw"
)

func resizeForStore(img image.Image, mode, store string, width, height int) image.Image {
	if mode == "exact" {
		return resizeExact(img, width, height)
	}
	if mode == "openai" {
		return scaleToFill(img, width, height)
	}
	if store == "appstore" {
		return letterbox(img, width, height)
	}
	return scaleToFill(img, width, height)
}

func scaleToFill(img image.Image, width, height int) image.Image {
	srcW := img.Bounds().Dx()
	srcH := img.Bounds().Dy()
	scale := maxFloat(float64(width)/float64(srcW), float64(height)/float64(srcH))
	newW := int(float64(srcW) * scale)
	newH := int(float64(srcH) * scale)

	tmp := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.CatmullRom.Scale(tmp, tmp.Bounds(), img, img.Bounds(), draw.Over, nil)

	offsetX := (newW - width) / 2
	offsetY := (newH - height) / 2
	crop := image.NewRGBA(image.Rect(0, 0, width, height))
	imagedraw.Draw(crop, crop.Bounds(), tmp, image.Point{X: offsetX, Y: offsetY}, imagedraw.Src)
	return crop
}

func letterbox(img image.Image, width, height int) image.Image {
	srcW := img.Bounds().Dx()
	srcH := img.Bounds().Dy()
	if srcW == 0 || srcH == 0 {
		return img
	}
	scale := minFloat(float64(width)/float64(srcW), float64(height)/float64(srcH))
	newW := int(float64(srcW) * scale)
	newH := int(float64(srcH) * scale)
	if newW <= 0 || newH <= 0 {
		return img
	}

	canvas := image.NewRGBA(image.Rect(0, 0, width, height))
	scaled := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.CatmullRom.Scale(scaled, scaled.Bounds(), img, img.Bounds(), draw.Over, nil)

	offsetX := (width - newW) / 2
	offsetY := (height - newH) / 2
	imagedraw.Draw(canvas, image.Rect(offsetX, offsetY, offsetX+newW, offsetY+newH), scaled, image.Point{}, imagedraw.Over)

	// Edge-extend padding for top/bottom/left/right to blend with the image.
	extendEdges(canvas, scaled, offsetX, offsetY, width, height)
	return canvas
}

func extendEdges(canvas *image.RGBA, scaled *image.RGBA, offsetX, offsetY, width, height int) {
	// Top padding
	if offsetY > 0 {
		topStrip := scaled.SubImage(image.Rect(0, 0, scaled.Bounds().Dx(), 1))
		drawScaledStrip(canvas, topStrip, image.Rect(0, 0, width, offsetY))
	}
	// Bottom padding
	if offsetY+scaled.Bounds().Dy() < height {
		bottomStrip := scaled.SubImage(image.Rect(0, scaled.Bounds().Dy()-1, scaled.Bounds().Dx(), scaled.Bounds().Dy()))
		drawScaledStrip(canvas, bottomStrip, image.Rect(0, offsetY+scaled.Bounds().Dy(), width, height))
	}
	// Left padding (if any)
	if offsetX > 0 {
		leftStrip := scaled.SubImage(image.Rect(0, 0, 1, scaled.Bounds().Dy()))
		drawScaledStrip(canvas, leftStrip, image.Rect(0, offsetY, offsetX, offsetY+scaled.Bounds().Dy()))
	}
	// Right padding (if any)
	if offsetX+scaled.Bounds().Dx() < width {
		rightStrip := scaled.SubImage(image.Rect(scaled.Bounds().Dx()-1, 0, scaled.Bounds().Dx(), scaled.Bounds().Dy()))
		drawScaledStrip(canvas, rightStrip, image.Rect(offsetX+scaled.Bounds().Dx(), offsetY, width, offsetY+scaled.Bounds().Dy()))
	}
}

func drawScaledStrip(canvas *image.RGBA, strip image.Image, dest image.Rectangle) {
	if dest.Dx() <= 0 || dest.Dy() <= 0 {
		return
	}
	tmp := image.NewRGBA(image.Rect(0, 0, dest.Dx(), dest.Dy()))
	draw.CatmullRom.Scale(tmp, tmp.Bounds(), strip, strip.Bounds(), draw.Over, nil)
	imagedraw.Draw(canvas, dest, tmp, image.Point{}, imagedraw.Over)
}

func scaleToFit(img image.Image, maxWidth, maxHeight int) image.Image {
	srcW := img.Bounds().Dx()
	srcH := img.Bounds().Dy()
	if srcW == 0 || srcH == 0 {
		return img
	}
	scale := minFloat(float64(maxWidth)/float64(srcW), float64(maxHeight)/float64(srcH))
	newW := int(float64(srcW) * scale)
	newH := int(float64(srcH) * scale)
	if newW <= 0 || newH <= 0 {
		return img
	}

	out := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.CatmullRom.Scale(out, out.Bounds(), img, img.Bounds(), draw.Over, nil)
	return out
}

func resizeExact(img image.Image, width, height int) image.Image {
	out := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(out, out.Bounds(), img, img.Bounds(), draw.Over, nil)
	return out
}
