package common

import (
	"bytes"
	"compress/zlib"
	"context"
	"encoding/base64"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/color"
	standarddraw "image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type colorChannel string

const (
	redChannel  colorChannel = "red"
	blueChannel colorChannel = "blue"
)

func dominantChannel(value color.Color) colorChannel {
	red, _, blue, _ := value.RGBA()
	if red > blue {
		return redChannel
	}
	return blueChannel
}

func jpegWithEXIFOrientation(t *testing.T, source []byte, orientation uint16) []byte {
	t.Helper()
	require.GreaterOrEqual(t, len(source), 2)
	require.Equal(t, []byte{0xff, 0xd8}, source[:2])
	payload := []byte{
		'E', 'x', 'i', 'f', 0, 0,
		'I', 'I', 42, 0, 8, 0, 0, 0,
		1, 0,
		0x12, 0x01, 3, 0, 1, 0, 0, 0, byte(orientation), byte(orientation >> 8), 0, 0,
		0, 0, 0, 0,
	}
	segment := []byte{0xff, 0xe1, 0, 0}
	binary.BigEndian.PutUint16(segment[2:4], uint16(len(payload)+2))
	result := append([]byte(nil), source[:2]...)
	result = append(result, segment...)
	result = append(result, payload...)
	return append(result, source[2:]...)
}

func TestNormalizeProfileImageCanonicalizesSupportedFormats(t *testing.T) {
	tests := []struct {
		name        string
		encode      func(*bytes.Buffer) error
		contentType string
		extension   string
	}{
		{name: "jpeg", encode: func(buffer *bytes.Buffer) error {
			return jpeg.Encode(buffer, testProfileImage(), &jpeg.Options{Quality: 70})
		}, contentType: "image/jpeg", extension: "jpg"},
		{name: "png", encode: func(buffer *bytes.Buffer) error {
			return png.Encode(buffer, testProfileImage())
		}, contentType: "image/png", extension: "png"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var source bytes.Buffer
			require.NoError(t, test.encode(&source))
			result, err := NormalizeProfileImage(context.Background(), base64.StdEncoding.EncodeToString(source.Bytes()))
			require.NoError(t, err)
			assert.Equal(t, test.contentType, result.ContentType)
			assert.Equal(t, test.extension, result.Extension)
			assert.NotContains(t, string(result.Data), "polyglot")
			_, decodedFormat, decodeErr := image.Decode(bytes.NewReader(result.Data))
			require.NoError(t, decodeErr)
			assert.Equal(t, test.name, decodedFormat)
		})
	}
}

func TestNativeProfileImageFixturesSatisfyServerContract(t *testing.T) {
	found := false
	for _, environmentVariable := range []string{"PROFILE_IMAGE_ANDROID_FIXTURE", "PROFILE_IMAGE_IOS_FIXTURE"} {
		fixturePath := os.Getenv(environmentVariable)
		if fixturePath == "" {
			continue
		}
		found = true
		fixture, err := os.ReadFile(fixturePath)
		require.NoError(t, err)
		result, err := NormalizeProfileImage(context.Background(), base64.StdEncoding.EncodeToString(fixture))
		require.NoError(t, err, environmentVariable)
		assert.Equal(t, "image/jpeg", result.ContentType)
		assert.NotEmpty(t, result.Data)
	}
	if !found {
		t.Skip("native compatibility fixtures were not supplied")
	}
}

func TestNormalizeProfileImageAppliesJPEGEXIFOrientation(t *testing.T) {
	sourceImage := image.NewNRGBA(image.Rect(0, 0, 80, 40))
	standarddraw.Draw(sourceImage, sourceImage.Bounds(), &image.Uniform{C: color.NRGBA{R: 255, A: 255}}, image.Point{}, standarddraw.Src)
	standarddraw.Draw(sourceImage, image.Rect(40, 0, 80, 40), &image.Uniform{C: color.NRGBA{B: 255, A: 255}}, image.Point{}, standarddraw.Src)

	for _, test := range []struct {
		name        string
		orientation uint16
		width       int
		height      int
		first       colorChannel
		last        colorChannel
	}{
		{name: "mirrored", orientation: 2, width: 80, height: 40, first: blueChannel, last: redChannel},
		{name: "rotated", orientation: 6, width: 40, height: 80, first: redChannel, last: blueChannel},
	} {
		t.Run(test.name, func(t *testing.T) {
			var encoded bytes.Buffer
			require.NoError(t, jpeg.Encode(&encoded, sourceImage, &jpeg.Options{Quality: 100}))
			withOrientation := jpegWithEXIFOrientation(t, encoded.Bytes(), test.orientation)
			result, err := NormalizeProfileImage(context.Background(), base64.StdEncoding.EncodeToString(withOrientation))
			require.NoError(t, err)
			decoded, err := jpeg.Decode(bytes.NewReader(result.Data))
			require.NoError(t, err)
			assert.Equal(t, image.Rect(0, 0, test.width, test.height), decoded.Bounds())
			assert.Equal(t, test.first, dominantChannel(decoded.At(5, 5)))
			assert.Equal(t, test.last, dominantChannel(decoded.At(test.width-6, test.height-6)))
		})
	}
}

func TestNormalizeProfileImageAcceptsBoundedLegacyAndroidEXIF(t *testing.T) {
	var source bytes.Buffer
	require.NoError(t, jpeg.Encode(&source, testProfileImage(), &jpeg.Options{Quality: 80}))
	legacy := jpegWithLegacyAndroidEXIF(t, source.Bytes())

	result, err := NormalizeProfileImage(context.Background(), base64.StdEncoding.EncodeToString(legacy))
	require.NoError(t, err)
	assert.Equal(t, profileContentTypeJPEG, result.ContentType)
	assert.NotContains(t, string(result.Data), "Canon")
	assert.NotContains(t, string(result.Data), "MakerNote")
}

func TestNormalizeProfileImageAlwaysEmitsEightBitPNG(t *testing.T) {
	palette := make(color.Palette, 256)
	for index := range palette {
		palette[index] = color.NRGBA{R: uint8(index), A: 255}
	}
	sourceImage := image.NewPaletted(image.Rect(0, 0, 2, 2), palette)
	var encoded bytes.Buffer
	require.NoError(t, png.Encode(&encoded, sourceImage))
	require.Equal(t, byte(8), encoded.Bytes()[24])

	result, err := NormalizeProfileImage(context.Background(), base64.StdEncoding.EncodeToString(encoded.Bytes()))
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(result.Data), 26)
	assert.Equal(t, byte(8), result.Data[24])
	assert.NotEqual(t, byte(3), result.Data[25])
}

func TestNormalizeProfileImageAcceptsStandardIndexedTransparency(t *testing.T) {
	palette := make(color.Palette, 256)
	for index := range palette {
		palette[index] = color.NRGBA{A: 255}
	}
	palette[0] = color.NRGBA{R: 255, A: 255}
	palette[1] = color.NRGBA{B: 255, A: 0}
	indexed := image.NewPaletted(image.Rect(0, 0, 2, 1), palette)
	indexed.SetColorIndex(0, 0, 0)
	indexed.SetColorIndex(1, 0, 1)
	var encoded bytes.Buffer
	require.NoError(t, png.Encode(&encoded, indexed))
	require.True(t, bytes.Contains(encoded.Bytes(), []byte("tRNS")))

	result, err := NormalizeProfileImage(context.Background(), base64.StdEncoding.EncodeToString(encoded.Bytes()))
	require.NoError(t, err)
	decoded, err := png.Decode(bytes.NewReader(result.Data))
	require.NoError(t, err)
	transparent := color.NRGBAModel.Convert(decoded.At(1, 0)).(color.NRGBA)
	assert.Zero(t, transparent.A)
}

func TestNormalizeProfileImageAcceptsBoundedLegacyAndroidPNGEXIF(t *testing.T) {
	var pngSource bytes.Buffer
	require.NoError(t, png.Encode(&pngSource, testProfileImage()))
	var jpegSource bytes.Buffer
	require.NoError(t, jpeg.Encode(&jpegSource, testProfileImage(), &jpeg.Options{Quality: 80}))
	tiff := legacyAndroidTIFF(t, jpegSource.Bytes())
	withEXIF := insertPNGChunkBefore(t, pngSource.Bytes(), "IDAT", "eXIf", tiff)

	result, err := NormalizeProfileImage(context.Background(), base64.StdEncoding.EncodeToString(withEXIF))
	require.NoError(t, err)
	assert.Equal(t, profileContentTypePNG, result.ContentType)
	assert.NotContains(t, string(result.Data), "Canon")
}

func TestProfileImageEightBitConversionReusesCanonicalRaster(t *testing.T) {
	source := image.NewNRGBA(image.Rect(0, 0, 2048, 2048))
	converted, err := profileImageNRGBA8(context.Background(), source)
	require.NoError(t, err)
	assert.Same(t, source, converted)
}

func TestProfileImageRasterStagesHonorCancellation(t *testing.T) {
	large := image.NewNRGBA(image.Rect(0, 0, 3000, 3000))

	_, err := applyEXIFOrientation(&cancelAfterChecksContext{cancelAt: 2}, large, 6)
	assert.ErrorIs(t, err, context.Canceled)

	_, err = downscaleProfileImage(&cancelAfterChecksContext{cancelAt: 1}, large)
	assert.ErrorIs(t, err, context.Canceled)

	_, err = profileImageNRGBA8(&cancelAfterChecksContext{cancelAt: 2}, image.NewPaletted(
		image.Rect(0, 0, 2, 2),
		color.Palette{color.Black, color.White},
	))
	assert.ErrorIs(t, err, context.Canceled)

	encodeContext := &cancelAfterChecksContext{cancelAt: 2}
	buffer := &boundedImageBuffer{ctx: encodeContext, remaining: MaxProfileImageBytes}
	err = encodeWithCancellation(encodeContext, func() error {
		return jpeg.Encode(buffer, large, &jpeg.Options{Quality: 90})
	})
	assert.ErrorIs(t, err, context.Canceled)
}

func TestProfileImageDecodeCancellationPromptlyReleasesRasterSlot(t *testing.T) {
	var source bytes.Buffer
	require.NoError(t, jpeg.Encode(
		&source,
		image.NewNRGBA(image.Rect(0, 0, 1000, 1000)),
		&jpeg.Options{Quality: 90},
	))
	ctx := &cancelAfterChecksContext{cancelAt: 2}
	_, err := NormalizeProfileImage(ctx, base64.StdEncoding.EncodeToString(source.Bytes()))
	assert.ErrorIs(t, err, context.Canceled)

	var small bytes.Buffer
	require.NoError(t, jpeg.Encode(&small, testProfileImage(), &jpeg.Options{Quality: 90}))
	_, err = NormalizeProfileImage(context.Background(), base64.StdEncoding.EncodeToString(small.Bytes()))
	require.NoError(t, err)
}

func TestStrictJPEGParserBoundsMarkerFanoutAndHonorsCancellation(t *testing.T) {
	var encoded bytes.Buffer
	require.NoError(t, jpeg.Encode(&encoded, testProfileImage(), &jpeg.Options{Quality: 90}))
	valid := encoded.Bytes()
	dht := firstJPEGSegment(t, valid, 0xc4)
	redefinedDHT := append(append(append([]byte(nil), valid[:2]...), dht...), valid[2:]...)
	tooManySymbols := append(append(append([]byte(nil), valid[:2]...), oversizedJPEGDHT()...), valid[2:]...)

	repeatedDQT := firstJPEGSegment(t, valid, 0xdb)
	segmentFanout := append([]byte(nil), valid[:2]...)
	for range maxProfileJPEGSegments + 1 {
		segmentFanout = append(segmentFanout, repeatedDQT...)
	}
	segmentFanout = append(segmentFanout, valid[2:]...)

	for _, hostile := range [][]byte{redefinedDHT, tooManySymbols, segmentFanout} {
		assert.NotPanics(t, func() {
			_, err := NormalizeProfileImage(context.Background(), base64.StdEncoding.EncodeToString(hostile))
			assert.ErrorIs(t, err, ErrInvalidProfileImage)
		})
	}

	ctx := &cancelAfterChecksContext{cancelAt: 3}
	_, err := NormalizeProfileImage(ctx, base64.StdEncoding.EncodeToString(segmentFanout))
	assert.ErrorIs(t, err, context.Canceled)

	// Cancellation or policy rejection must not retain the process-wide slot.
	_, err = NormalizeProfileImage(context.Background(), base64.StdEncoding.EncodeToString(valid))
	require.NoError(t, err)
}

func oversizedJPEGDHT() []byte {
	counts := make([]byte, 16)
	counts[8] = 255
	counts[9] = 2
	payload := append([]byte{0}, counts...)
	payload = append(payload, make([]byte, 257)...)
	result := []byte{0xff, 0xc4, 0, 0}
	binary.BigEndian.PutUint16(result[2:4], uint16(len(payload)+2))
	return append(result, payload...)
}

type cancelAfterChecksContext struct {
	checks   int
	cancelAt int
}

func (c *cancelAfterChecksContext) Deadline() (time.Time, bool) { return time.Time{}, false }
func (c *cancelAfterChecksContext) Done() <-chan struct{}       { return nil }
func (c *cancelAfterChecksContext) Err() error {
	c.checks++
	if c.checks >= c.cancelAt {
		return context.Canceled
	}
	return nil
}
func (c *cancelAfterChecksContext) Value(any) any { return nil }

func TestNormalizeProfileImageRejectsUnsupportedAnimatedOversizedAndInvalidImages(t *testing.T) {
	var animated bytes.Buffer
	require.NoError(t, gif.EncodeAll(&animated, &gif.GIF{
		Image: []*image.Paletted{
			image.NewPaletted(image.Rect(0, 0, 1, 1), color.Palette{color.Black}),
			image.NewPaletted(image.Rect(0, 0, 1, 1), color.Palette{color.White}),
		},
		Delay: []int{1, 1},
	}))

	var validPNG bytes.Buffer
	require.NoError(t, png.Encode(&validPNG, testProfileImage()))
	tooWide := pngWithDimensions(t, validPNG.Bytes(), MaxProfileImageWidth+1, 1)
	tooManyPixels := pngWithDimensions(t, validPNG.Bytes(), 4000, 3300)
	lowBitDepth := pngWithIHDRFields(t, validPNG.Bytes(), 4, 0)
	var highBitDepth bytes.Buffer
	require.NoError(t, png.Encode(&highBitDepth, image.NewNRGBA64(image.Rect(0, 0, 2, 2))))
	var validJPEG bytes.Buffer
	require.NoError(t, jpeg.Encode(&validJPEG, testProfileImage(), &jpeg.Options{Quality: 80}))
	polyglot := append(append([]byte(nil), validPNG.Bytes()...), []byte("<script>trailing</script>")...)
	jpegShortSuffix := append(append([]byte(nil), validJPEG.Bytes()...), byte('x'))
	jpegLargeSuffix := append(append([]byte(nil), validJPEG.Bytes()...), bytes.Repeat([]byte("payload"), 1024)...)
	concatenatedJPEG := append(append([]byte(nil), validJPEG.Bytes()...), validJPEG.Bytes()...)
	jpegPreEOIPayload := append(append(append([]byte(nil), validJPEG.Bytes()[:validJPEG.Len()-2]...), []byte("hidden-payload")...), 0xff, 0xd9)
	jpegReservedMarker := append([]byte{0xff, 0xd8, 0xff, 0x02, 0x00, 0x02}, validJPEG.Bytes()[2:]...)
	jpegAPPPolyglot := insertJPEGSegment(t, validJPEG.Bytes(), 0xe2, []byte("PK\x03\x04embedded.zip"))
	jpegCommentPolyglot := insertJPEGSegment(t, validJPEG.Bytes(), 0xfe, []byte("<script>embedded</script>"))
	jpegMalformedEXIF := insertJPEGSegment(t, validJPEG.Bytes(), 0xe1, []byte("Exif\x00\x00PK\x03\x04"))
	legacyEXIF := jpegWithLegacyAndroidEXIF(t, validJPEG.Bytes())
	legacyEXIFBadOffset := append([]byte(nil), legacyEXIF...)
	binary.LittleEndian.PutUint32(legacyEXIFBadOffset[42:46], uint32(len(legacyEXIF)+1))
	legacyEXIFThumbnail := append([]byte(nil), legacyEXIF...)
	binary.LittleEndian.PutUint16(legacyEXIFThumbnail[22:24], 0x0201)
	legacyEXIFDuplicateTag := append([]byte(nil), legacyEXIF...)
	binary.LittleEndian.PutUint16(legacyEXIFDuplicateTag[34:36], 0x0112)
	legacyTIFF := legacyAndroidTIFF(t, validJPEG.Bytes())
	pngEXIFBadOffsetPayload := append([]byte(nil), legacyTIFF...)
	binary.LittleEndian.PutUint32(pngEXIFBadOffsetPayload[30:34], uint32(len(legacyTIFF)+1))
	pngEXIFThumbnailPayload := append([]byte(nil), legacyTIFF...)
	binary.LittleEndian.PutUint16(pngEXIFThumbnailPayload[10:12], 0x0201)
	pngEXIFPolyglotPayload := append(append([]byte(nil), legacyTIFF...), []byte("PK\x03\x04embedded.zip")...)
	pngEXIFValid := insertPNGChunkBefore(t, validPNG.Bytes(), "IDAT", "eXIf", legacyTIFF)
	pngEXIFDuplicate := insertPNGChunkBefore(t, pngEXIFValid, "IDAT", "eXIf", legacyTIFF)
	pngEXIFMisplaced := insertPNGChunkBefore(t, validPNG.Bytes(), "IEND", "eXIf", legacyTIFF)
	pngEXIFBadOffset := insertPNGChunkBefore(t, validPNG.Bytes(), "IDAT", "eXIf", pngEXIFBadOffsetPayload)
	pngEXIFThumbnail := insertPNGChunkBefore(t, validPNG.Bytes(), "IDAT", "eXIf", pngEXIFThumbnailPayload)
	pngEXIFPolyglot := insertPNGChunkBefore(t, validPNG.Bytes(), "IDAT", "eXIf", pngEXIFPolyglotPayload)
	pngEXIFOversized := insertPNGChunkBefore(t, validPNG.Bytes(), "IDAT", "eXIf", make([]byte, 65528))
	progressiveJPEG := replaceJPEGMarker(t, validJPEG.Bytes(), 0xc0, 0xc2)
	apng := withAPNGAnimationChunk(t, validPNG.Bytes())
	frameControl := insertPNGChunkBefore(t, validPNG.Bytes(), "IDAT", "fcTL", make([]byte, 26))
	frameData := insertPNGChunkBefore(t, validPNG.Bytes(), "IEND", "fdAT", []byte{0, 0, 0, 1, 1})
	misplacedAncillary := make([][]byte, 0)
	duplicatedAncillary := make([][]byte, 0)
	for chunkType, payload := range map[string][]byte{
		"gAMA": make([]byte, 4),
		"cHRM": make([]byte, 32),
		"iCCP": {0},
		"sRGB": {0},
		"pHYs": make([]byte, 9),
		"eXIf": make([]byte, 4),
	} {
		misplacedAncillary = append(misplacedAncillary,
			insertPNGChunkBefore(t, validPNG.Bytes(), "IEND", chunkType, payload))
		once := insertPNGChunkBefore(t, validPNG.Bytes(), "IDAT", chunkType, payload)
		duplicatedAncillary = append(duplicatedAncillary,
			insertPNGChunkBefore(t, once, "IDAT", chunkType, payload))
	}
	trailingIDAT := insertPNGChunkBefore(t, validPNG.Bytes(), "IEND", "IDAT", []byte("ignored-garbage"))
	invalidChunkCharacter := insertPNGChunkBefore(t, validPNG.Bytes(), "IDAT", "a\x01CD", nil)
	lowercaseReservedChunk := insertPNGChunkBefore(t, validPNG.Bytes(), "IDAT", "abcd", nil)
	pngTextPolyglot := insertPNGChunkBefore(t, validPNG.Bytes(), "IDAT", "tEXt", []byte("Comment\x00PK\x03\x04embedded.zip"))
	pngUnknownPolyglot := insertPNGChunkBefore(t, validPNG.Bytes(), "IDAT", "raNd", []byte("<script>embedded</script>"))
	pngMalformedICCP := insertPNGChunkBefore(t, validPNG.Bytes(), "IDAT", "iCCP", []byte("profile\x00\x00not-zlib"))
	pngMalformedZTXT := insertPNGChunkBefore(t, validPNG.Bytes(), "IDAT", "zTXt", []byte("Comment\x00\x00not-zlib"))
	opaqueColors := make(color.Palette, 256)
	for index := range opaqueColors {
		opaqueColors[index] = color.Black
	}
	opaquePalette := image.NewPaletted(image.Rect(0, 0, 2, 1), opaqueColors)
	var opaqueIndexedPNG bytes.Buffer
	require.NoError(t, png.Encode(&opaqueIndexedPNG, opaquePalette))
	oversizedTRNS := insertPNGChunkBefore(t, opaqueIndexedPNG.Bytes(), "IDAT", "tRNS", make([]byte, 257))
	var excessiveCompressed bytes.Buffer
	excessiveWriter := zlib.NewWriter(&excessiveCompressed)
	_, err := excessiveWriter.Write(bytes.Repeat([]byte{0}, 1<<20))
	require.NoError(t, err)
	require.NoError(t, excessiveWriter.Close())
	excessiveInflation := replaceSinglePNGIDAT(t, validPNG.Bytes(), excessiveCompressed.Bytes())
	idatFanout := validPNG.Bytes()
	for range maxProfilePNGIDATChunks {
		idatFanout = insertPNGChunkBefore(t, idatFanout, "IEND", "IDAT", nil)
	}

	tests := []string{
		base64.StdEncoding.EncodeToString(animated.Bytes()),
		base64.StdEncoding.EncodeToString(tooWide),
		base64.StdEncoding.EncodeToString(tooManyPixels),
		base64.StdEncoding.EncodeToString(lowBitDepth),
		base64.StdEncoding.EncodeToString(highBitDepth.Bytes()),
		base64.StdEncoding.EncodeToString(polyglot),
		base64.StdEncoding.EncodeToString(jpegShortSuffix),
		base64.StdEncoding.EncodeToString(jpegLargeSuffix),
		base64.StdEncoding.EncodeToString(concatenatedJPEG),
		base64.StdEncoding.EncodeToString(jpegPreEOIPayload),
		base64.StdEncoding.EncodeToString(jpegReservedMarker),
		base64.StdEncoding.EncodeToString(jpegAPPPolyglot),
		base64.StdEncoding.EncodeToString(jpegCommentPolyglot),
		base64.StdEncoding.EncodeToString(jpegMalformedEXIF),
		base64.StdEncoding.EncodeToString(legacyEXIFBadOffset),
		base64.StdEncoding.EncodeToString(legacyEXIFThumbnail),
		base64.StdEncoding.EncodeToString(legacyEXIFDuplicateTag),
		base64.StdEncoding.EncodeToString(pngEXIFDuplicate),
		base64.StdEncoding.EncodeToString(pngEXIFMisplaced),
		base64.StdEncoding.EncodeToString(pngEXIFBadOffset),
		base64.StdEncoding.EncodeToString(pngEXIFThumbnail),
		base64.StdEncoding.EncodeToString(pngEXIFPolyglot),
		base64.StdEncoding.EncodeToString(pngEXIFOversized),
		base64.StdEncoding.EncodeToString(progressiveJPEG),
		base64.StdEncoding.EncodeToString(apng),
		base64.StdEncoding.EncodeToString(frameControl),
		base64.StdEncoding.EncodeToString(frameData),
		base64.StdEncoding.EncodeToString(trailingIDAT),
		base64.StdEncoding.EncodeToString(invalidChunkCharacter),
		base64.StdEncoding.EncodeToString(lowercaseReservedChunk),
		base64.StdEncoding.EncodeToString(pngTextPolyglot),
		base64.StdEncoding.EncodeToString(pngUnknownPolyglot),
		base64.StdEncoding.EncodeToString(pngMalformedICCP),
		base64.StdEncoding.EncodeToString(pngMalformedZTXT),
		base64.StdEncoding.EncodeToString(oversizedTRNS),
		base64.StdEncoding.EncodeToString(excessiveInflation),
		base64.StdEncoding.EncodeToString(idatFanout),
		"data:text/html;base64," + base64.StdEncoding.EncodeToString(validPNG.Bytes()),
		"data:image/png," + base64.StdEncoding.EncodeToString(validPNG.Bytes()),
		"data:image/png;charset=utf-8;base64," + base64.StdEncoding.EncodeToString(validPNG.Bytes()),
		"data:image/png;base64;base64," + base64.StdEncoding.EncodeToString(validPNG.Bytes()),
		base64.StdEncoding.EncodeToString([]byte("not an image")),
	}
	for _, invalidAncillary := range append(misplacedAncillary, duplicatedAncillary...) {
		tests = append(tests, base64.StdEncoding.EncodeToString(invalidAncillary))
	}
	for _, encoded := range tests {
		_, err := NormalizeProfileImage(context.Background(), encoded)
		assert.ErrorIs(t, err, ErrInvalidProfileImage)
	}
}

func TestOversizedPNGDimensionsRejectBeforeContainerInflation(t *testing.T) {
	var valid bytes.Buffer
	require.NoError(t, png.Encode(&valid, testProfileImage()))
	oversized := pngWithDimensions(t, valid.Bytes(), 5000, 5000)
	oversized = replaceSinglePNGIDAT(t, oversized, []byte("not-a-zlib-stream"))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := NormalizeProfileImage(ctx, base64.StdEncoding.EncodeToString(oversized))
	assert.ErrorIs(t, err, ErrInvalidProfileImage)
	assert.NotErrorIs(t, err, context.Canceled)
}

func replaceJPEGMarker(t *testing.T, source []byte, from, to byte) []byte {
	t.Helper()
	result := append([]byte(nil), source...)
	for offset := 0; offset+1 < len(result); offset++ {
		if result[offset] == 0xff && result[offset+1] == from {
			result[offset+1] = to
			return result
		}
	}
	t.Fatalf("JPEG marker ff%02x not found", from)
	return nil
}

func insertJPEGSegment(t *testing.T, source []byte, marker byte, payload []byte) []byte {
	t.Helper()
	require.GreaterOrEqual(t, len(source), 2)
	require.Equal(t, []byte{0xff, 0xd8}, source[:2])
	require.LessOrEqual(t, len(payload)+2, int(^uint16(0)))
	segment := []byte{0xff, marker, 0, 0}
	binary.BigEndian.PutUint16(segment[2:4], uint16(len(payload)+2))
	segment = append(segment, payload...)
	result := append([]byte(nil), source[:2]...)
	result = append(result, segment...)
	return append(result, source[2:]...)
}

func jpegWithLegacyAndroidEXIF(t *testing.T, source []byte) []byte {
	t.Helper()
	const (
		rootOffset    = 8
		makeOffset    = 62
		modelOffset   = 68
		exifOffset    = 74
		dateOffset    = 116
		makerOffset   = 136
		commentOffset = 144
	)
	tiff := make([]byte, 160)
	copy(tiff[:2], "II")
	binary.LittleEndian.PutUint16(tiff[2:4], 42)
	binary.LittleEndian.PutUint32(tiff[4:8], rootOffset)
	binary.LittleEndian.PutUint16(tiff[rootOffset:rootOffset+2], 4)
	writeEXIFEntry(tiff[10:22], 0x0112, 3, 1, 6)
	writeEXIFEntry(tiff[22:34], 0x010f, 2, 6, makeOffset)
	writeEXIFEntry(tiff[34:46], 0x0110, 2, 6, modelOffset)
	writeEXIFEntry(tiff[46:58], 0x8769, 4, 1, exifOffset)
	copy(tiff[makeOffset:modelOffset], "Canon\x00")
	copy(tiff[modelOffset:exifOffset], "Pixel\x00")
	binary.LittleEndian.PutUint16(tiff[exifOffset:exifOffset+2], 3)
	writeEXIFEntry(tiff[76:88], 0x9003, 2, 20, dateOffset)
	writeEXIFEntry(tiff[88:100], 0x927c, 7, 8, makerOffset)
	writeEXIFEntry(tiff[100:112], 0x9286, 7, 16, commentOffset)
	copy(tiff[dateOffset:makerOffset], "2026:08:08 12:00:00\x00")
	copy(tiff[makerOffset:commentOffset], "MakerNot")
	copy(tiff[commentOffset:], "ASCII\x00\x00\x00ordinary")
	payload := append([]byte("Exif\x00\x00"), tiff...)
	return insertJPEGSegment(t, source, 0xe1, payload)
}

func legacyAndroidTIFF(t *testing.T, source []byte) []byte {
	t.Helper()
	withEXIF := jpegWithLegacyAndroidEXIF(t, source)
	require.GreaterOrEqual(t, len(withEXIF), 172)
	return append([]byte(nil), withEXIF[12:172]...)
}

func writeEXIFEntry(target []byte, tag, valueType uint16, count, value uint32) {
	binary.LittleEndian.PutUint16(target[0:2], tag)
	binary.LittleEndian.PutUint16(target[2:4], valueType)
	binary.LittleEndian.PutUint32(target[4:8], count)
	if valueType == 3 && count == 1 {
		binary.LittleEndian.PutUint16(target[8:10], uint16(value))
		return
	}
	binary.LittleEndian.PutUint32(target[8:12], value)
}

func replaceSinglePNGIDAT(t *testing.T, source, payload []byte) []byte {
	t.Helper()
	for offset := 8; offset+12 <= len(source); {
		length := int(binary.BigEndian.Uint32(source[offset : offset+4]))
		chunkEnd := offset + 12 + length
		require.LessOrEqual(t, chunkEnd, len(source))
		if string(source[offset+4:offset+8]) == "IDAT" {
			chunk := make([]byte, 12+len(payload))
			binary.BigEndian.PutUint32(chunk[:4], uint32(len(payload)))
			copy(chunk[4:8], "IDAT")
			copy(chunk[8:8+len(payload)], payload)
			binary.BigEndian.PutUint32(chunk[8+len(payload):], crc32.ChecksumIEEE(chunk[4:8+len(payload)]))
			result := append([]byte(nil), source[:offset]...)
			result = append(result, chunk...)
			return append(result, source[chunkEnd:]...)
		}
		offset = chunkEnd
	}
	t.Fatal("PNG IDAT chunk not found")
	return nil
}

func insertPNGChunkBefore(t *testing.T, source []byte, beforeType, chunkType string, payload []byte) []byte {
	t.Helper()
	require.Len(t, chunkType, 4)
	for offset := 8; offset+12 <= len(source); {
		length := int(binary.BigEndian.Uint32(source[offset : offset+4]))
		chunkEnd := offset + 12 + length
		require.LessOrEqual(t, chunkEnd, len(source))
		if string(source[offset+4:offset+8]) == beforeType {
			chunk := make([]byte, 12+len(payload))
			binary.BigEndian.PutUint32(chunk[:4], uint32(len(payload)))
			copy(chunk[4:8], chunkType)
			copy(chunk[8:8+len(payload)], payload)
			binary.BigEndian.PutUint32(chunk[8+len(payload):], crc32.ChecksumIEEE(chunk[4:8+len(payload)]))
			result := append([]byte(nil), source[:offset]...)
			result = append(result, chunk...)
			return append(result, source[offset:]...)
		}
		offset = chunkEnd
	}
	t.Fatalf("PNG chunk %q not found", beforeType)
	return nil
}

func TestNormalizeProfileImageDownscalesLegacyCameraDimensions(t *testing.T) {
	var source bytes.Buffer
	require.NoError(t, jpeg.Encode(
		&source,
		image.NewNRGBA(image.Rect(0, 0, 3000, 3000)),
		&jpeg.Options{Quality: 80},
	))
	result, err := NormalizeProfileImage(context.Background(), base64.StdEncoding.EncodeToString(source.Bytes()))
	require.NoError(t, err)
	config, err := jpeg.DecodeConfig(bytes.NewReader(result.Data))
	require.NoError(t, err)
	assert.Equal(t, MaxProfileImageOutputDimension, config.Width)
	assert.Equal(t, MaxProfileImageOutputDimension, config.Height)
}

func TestNormalizeProfileImageReturnsTypedDecodedSizeError(t *testing.T) {
	_, err := NormalizeProfileImage(
		context.Background(),
		strings.Repeat("A", base64.StdEncoding.EncodedLen(MaxProfileImageBytes+1)),
	)
	assert.ErrorIs(t, err, ErrProfileImageTooLarge)
}

func TestNormalizeProfileImageReturnsBusyWithoutWaitingForRasterSlot(t *testing.T) {
	profileImageSlots <- struct{}{}
	defer func() { <-profileImageSlots }()
	started := time.Now()
	_, err := NormalizeProfileImage(context.Background(), "aW1hZ2U=")
	assert.ErrorIs(t, err, ErrProfileImageBusy)
	assert.Less(t, time.Since(started), 50*time.Millisecond)
}

func TestNormalizeProfileImagePrioritizesCancellationOverBusyRasterSlot(t *testing.T) {
	profileImageSlots <- struct{}{}
	defer func() { <-profileImageSlots }()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := NormalizeProfileImage(ctx, "aW1hZ2U=")
	assert.ErrorIs(t, err, context.Canceled)
	assert.NotErrorIs(t, err, ErrProfileImageBusy)
}

func TestDecodeImageBase64BoundedAcceptsExactDecodedLimit(t *testing.T) {
	source := bytes.Repeat([]byte{0x7f}, MaxProfileImageBytes)
	decoded, _, err := DecodeImageBase64Bounded(base64.StdEncoding.EncodeToString(source), MaxProfileImageBytes)
	require.NoError(t, err)
	assert.Len(t, decoded, MaxProfileImageBytes)
}

func TestDecodeImageBase64RequiresCanonicalPaddedGrammar(t *testing.T) {
	for _, encoded := range []string{
		"YQ==",
		"data:image/jpeg;base64,YQ==",
		"YWJj",
		"-_8=",
		"data:image/png;base64,-_8=",
	} {
		decoded, _, err := DecodeImageBase64Bounded(encoded, 3)
		require.NoError(t, err, encoded)
		assert.NotEmpty(t, decoded)
	}

	for _, encoded := range []string{
		"",
		"YQ",
		"YQ=",
		"YQ===",
		"Y=Q=",
		"YQ==\n",
		"YQ\r\n==",
		"YQ ==",
		"+/8_",
		"YR==", // Non-zero discarded bits: canonical encoding is YQ==.
		"data:image/png;base64,YQ",
	} {
		_, _, err := DecodeImageBase64Bounded(encoded, 3)
		assert.Error(t, err, encoded)
		assert.NotErrorIs(t, err, ErrDecodedDataTooLarge, encoded)
	}
}

func withAPNGAnimationChunk(t *testing.T, source []byte) []byte {
	t.Helper()
	require.GreaterOrEqual(t, len(source), 33)
	chunk := make([]byte, 20)
	binary.BigEndian.PutUint32(chunk[:4], 8)
	copy(chunk[4:8], "acTL")
	// The CRC is intentionally irrelevant: policy detection rejects acTL before decode.
	return append(append(append([]byte(nil), source[:33]...), chunk...), source[33:]...)
}

func firstJPEGSegment(t *testing.T, source []byte, wanted byte) []byte {
	t.Helper()
	for offset := 2; offset+4 <= len(source); {
		require.Equal(t, byte(0xff), source[offset])
		marker := source[offset+1]
		length := int(binary.BigEndian.Uint16(source[offset+2 : offset+4]))
		require.GreaterOrEqual(t, length, 2)
		end := offset + 2 + length
		require.LessOrEqual(t, end, len(source))
		if marker == wanted {
			return append([]byte(nil), source[offset:end]...)
		}
		offset = end
	}
	t.Fatalf("JPEG fixture has no marker ff%02x", wanted)
	return nil
}

func pngWithDimensions(t *testing.T, source []byte, width, height int) []byte {
	t.Helper()
	require.GreaterOrEqual(t, len(source), 33)
	result := append([]byte(nil), source...)
	binary.BigEndian.PutUint32(result[16:20], uint32(width))
	binary.BigEndian.PutUint32(result[20:24], uint32(height))
	binary.BigEndian.PutUint32(result[29:33], crc32.ChecksumIEEE(result[12:29]))
	return result
}

func pngWithIHDRFields(t *testing.T, source []byte, bitDepth, colorType byte) []byte {
	t.Helper()
	require.GreaterOrEqual(t, len(source), 33)
	result := append([]byte(nil), source...)
	result[24] = bitDepth
	result[25] = colorType
	binary.BigEndian.PutUint32(result[29:33], crc32.ChecksumIEEE(result[12:29]))
	return result
}

func testProfileImage() image.Image {
	result := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	result.Set(0, 0, color.NRGBA{R: 255, A: 255})
	result.Set(1, 0, color.NRGBA{G: 255, A: 255})
	result.Set(0, 1, color.NRGBA{B: 255, A: 255})
	result.Set(1, 1, color.NRGBA{R: 255, G: 255, A: 255})
	return result
}
