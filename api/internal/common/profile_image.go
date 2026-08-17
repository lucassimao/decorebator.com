package common

import (
	"bytes"
	"compress/zlib"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"image"
	standarddraw "image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"strings"

	xdraw "golang.org/x/image/draw"
)

const (
	MaxProfileImageBytes           = 5 << 20
	MaxProfileImageWidth           = 5000
	MaxProfileImageHeight          = 5000
	MaxProfileImageInputPixels     = 13_000_000
	MaxProfileImagePixels          = 4_194_304
	MaxProfileImageOutputDimension = 2048
	profileImageWorkers            = 1
	maxProfilePNGChunks            = 512
	maxProfilePNGIDATChunks        = 128
	maxProfileJPEGSegments         = 512
	maxProfileJPEGHuffmanTables    = 8
	maxProfileJPEGHuffmanSymbols   = 256
	profileFormatJPEG              = "jpeg"
	profileFormatPNG               = "png"
	profileContentTypeJPEG         = "image/jpeg"
	profileContentTypePNG          = "image/png"
)

var ErrInvalidProfileImage = errors.New("invalid profile image")
var ErrProfileImageTooLarge = errors.New("profile image is too large")
var ErrProfileImageBusy = errors.New("profile image processing is busy")
var profileImageSlots = make(chan struct{}, profileImageWorkers)

type ProfileImage struct {
	Data        []byte
	ContentType string
	Extension   string
}

// NormalizeProfileImage accepts only static JPEG and PNG images. It validates
// the decoded dimensions before full decoding, downsizes legacy camera input
// to the canonical output budget, and re-encodes pixels without metadata.
func NormalizeProfileImage(ctx context.Context, encoded string) (ProfileImage, error) { //nolint:gocyclo // explicit fail-closed media pipeline
	select {
	case profileImageSlots <- struct{}{}:
		defer func() { <-profileImageSlots }()
	default:
		if err := ctx.Err(); err != nil {
			return ProfileImage{}, err
		}
		return ProfileImage{}, ErrProfileImageBusy
	}

	data, declaredContentType, err := DecodeImageBase64Bounded(encoded, MaxProfileImageBytes)
	if err != nil {
		if errors.Is(err, ErrDecodedDataTooLarge) {
			return ProfileImage{}, ErrProfileImageTooLarge
		}
		return ProfileImage{}, fmt.Errorf("%w: %v", ErrInvalidProfileImage, err)
	}

	config, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || !supportedProfileImageFormat(format) {
		return ProfileImage{}, ErrInvalidProfileImage
	}
	if config.Width <= 0 || config.Height <= 0 ||
		config.Width > MaxProfileImageWidth || config.Height > MaxProfileImageHeight ||
		int64(config.Width)*int64(config.Height) > MaxProfileImageInputPixels {
		return ProfileImage{}, ErrInvalidProfileImage
	}
	if format == profileFormatPNG {
		if validationErr := validateStrictPNGContainer(ctx, data); validationErr != nil {
			if errors.Is(validationErr, context.Canceled) || errors.Is(validationErr, context.DeadlineExceeded) {
				return ProfileImage{}, validationErr
			}
			return ProfileImage{}, ErrInvalidProfileImage
		}
	}
	if format == profileFormatJPEG {
		if validationErr := validateStrictJPEGContainer(ctx, data); validationErr != nil {
			if errors.Is(validationErr, context.Canceled) || errors.Is(validationErr, context.DeadlineExceeded) {
				return ProfileImage{}, validationErr
			}
			return ProfileImage{}, ErrInvalidProfileImage
		}
	}
	if !profileContentTypeMatches(format, declaredContentType) {
		return ProfileImage{}, ErrInvalidProfileImage
	}
	if contextErr := ctx.Err(); contextErr != nil {
		return ProfileImage{}, contextErr
	}
	decoded, err := decodeProfileImage(ctx, data, format)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return ProfileImage{}, err
		}
		return ProfileImage{}, ErrInvalidProfileImage
	}
	if contextErr := ctx.Err(); contextErr != nil {
		return ProfileImage{}, contextErr
	}
	decoded, err = downscaleProfileImage(ctx, decoded)
	if err != nil {
		return ProfileImage{}, err
	}
	if format == profileFormatJPEG {
		decoded, err = applyEXIFOrientation(ctx, decoded, jpegEXIFOrientation(data))
		if err != nil {
			return ProfileImage{}, err
		}
	}
	if format == profileFormatPNG {
		decoded, err = profileImageNRGBA8(ctx, decoded)
		if err != nil {
			return ProfileImage{}, err
		}
	}
	if contextErr := ctx.Err(); contextErr != nil {
		return ProfileImage{}, contextErr
	}

	normalized := &boundedImageBuffer{ctx: ctx, remaining: MaxProfileImageBytes}
	result := ProfileImage{}
	switch format {
	case profileFormatJPEG:
		err = encodeWithCancellation(ctx, func() error {
			return jpeg.Encode(normalized, decoded, &jpeg.Options{Quality: 90})
		})
		result.ContentType = profileContentTypeJPEG
		result.Extension = "jpg"
	case profileFormatPNG:
		err = encodeWithCancellation(ctx, func() error {
			return png.Encode(normalized, decoded)
		})
		result.ContentType = profileContentTypePNG
		result.Extension = "png"
	default:
		return ProfileImage{}, ErrInvalidProfileImage
	}
	if errors.Is(err, io.ErrShortBuffer) {
		return ProfileImage{}, ErrProfileImageTooLarge
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return ProfileImage{}, err
	}
	if err != nil {
		return ProfileImage{}, ErrInvalidProfileImage
	}
	if err := ctx.Err(); err != nil {
		return ProfileImage{}, err
	}
	result.Data = normalized.Bytes()
	return result, nil
}

func profileContentTypeMatches(format, declared string) bool {
	if declared == "" {
		return true
	}
	declared = strings.ToLower(strings.TrimSpace(declared))
	switch format {
	case profileFormatJPEG:
		return declared == profileContentTypeJPEG || declared == "image/jpg"
	case profileFormatPNG:
		return declared == profileContentTypePNG
	default:
		return false
	}
}

type boundedImageBuffer struct {
	ctx       context.Context
	buffer    bytes.Buffer
	remaining int
}

func (b *boundedImageBuffer) Write(data []byte) (int, error) {
	if err := b.ctx.Err(); err != nil {
		return 0, err
	}
	if len(data) > b.remaining {
		return 0, io.ErrShortBuffer
	}
	written, err := b.buffer.Write(data)
	b.remaining -= written
	return written, err
}

func (b *boundedImageBuffer) Bytes() []byte { return b.buffer.Bytes() }

func supportedProfileImageFormat(format string) bool {
	return format == profileFormatJPEG || format == profileFormatPNG
}

func decodeProfileImage(ctx context.Context, data []byte, format string) (image.Image, error) {
	reader := bytes.NewReader(data)
	checkedReader := &contextReader{ctx: ctx, reader: reader}
	var decoded image.Image
	var err error
	switch format {
	case profileFormatJPEG:
		decoded, err = jpeg.Decode(checkedReader)
	case profileFormatPNG:
		decoded, err = png.Decode(checkedReader)
	default:
		return nil, ErrInvalidProfileImage
	}
	if err != nil {
		return nil, err
	}
	if reader.Len() != 0 {
		return nil, ErrInvalidProfileImage
	}
	return decoded, nil
}

func downscaleProfileImage(ctx context.Context, source image.Image) (image.Image, error) {
	bounds := source.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width <= MaxProfileImageOutputDimension && height <= MaxProfileImageOutputDimension && int64(width)*int64(height) <= MaxProfileImagePixels {
		return source, ctx.Err()
	}
	scale := math.Min(
		float64(MaxProfileImageOutputDimension)/float64(width),
		float64(MaxProfileImageOutputDimension)/float64(height),
	)
	targetWidth := max(1, int(math.Round(float64(width)*scale)))
	targetHeight := max(1, int(math.Round(float64(height)*scale)))
	target := image.NewNRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	// ApproxBiLinear keeps profile-photo quality while avoiding CatmullRom's
	// deadline-breaking cost on legitimate legacy camera images.
	// Scaling is bounded by the validated 13 MP source limit. Avoid wrapping
	// every pixel read with a context check: that made legitimate legacy-camera
	// uploads exceed the request deadline under race/coverage instrumentation.
	xdraw.ApproxBiLinear.Scale(target, target.Bounds(), source, bounds, xdraw.Over, nil)
	return target, ctx.Err()
}

func profileImageNRGBA8(ctx context.Context, source image.Image) (image.Image, error) {
	switch source.(type) {
	case *image.NRGBA, *image.RGBA, *image.Gray:
		return source, ctx.Err()
	}
	bounds := source.Bounds()
	target := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	for y := 0; y < bounds.Dy(); y++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		row := image.Rect(0, y, bounds.Dx(), y+1)
		standarddraw.Draw(target, row, source, image.Pt(bounds.Min.X, bounds.Min.Y+y), standarddraw.Src)
	}
	return target, nil
}

func applyEXIFOrientation(ctx context.Context, source image.Image, orientation uint16) (image.Image, error) {
	if orientation < 2 || orientation > 8 {
		return source, ctx.Err()
	}
	bounds := source.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	targetWidth, targetHeight := width, height
	if orientation >= 5 {
		targetWidth, targetHeight = height, width
	}
	target := image.NewNRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	for y := 0; y < height; y++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		for x := 0; x < width; x++ {
			var targetX, targetY int
			switch orientation {
			case 2:
				targetX, targetY = width-1-x, y
			case 3:
				targetX, targetY = width-1-x, height-1-y
			case 4:
				targetX, targetY = x, height-1-y
			case 5:
				targetX, targetY = y, x
			case 6:
				targetX, targetY = height-1-y, x
			case 7:
				targetX, targetY = height-1-y, width-1-x
			case 8:
				targetX, targetY = y, width-1-x
			}
			target.Set(targetX, targetY, source.At(bounds.Min.X+x, bounds.Min.Y+y))
		}
	}
	return target, nil
}

func encodeWithCancellation(ctx context.Context, encode func() error) (err error) {
	if contextErr := ctx.Err(); contextErr != nil {
		return contextErr
	}
	err = encode()
	if err != nil {
		return err
	}
	return ctx.Err()
}

func jpegEXIFOrientation(data []byte) uint16 {
	if len(data) < 4 || data[0] != 0xff || data[1] != 0xd8 {
		return 1
	}
	for offset := 2; offset < len(data); {
		marker, next, ok := nextJPEGMarker(data, offset)
		if !ok || marker == 0xd9 || marker == 0xda {
			return 1
		}
		offset = next
		if marker == 0x01 || marker >= 0xd0 && marker <= 0xd7 {
			continue
		}
		if offset+2 > len(data) {
			return 1
		}
		segmentLength := int(binary.BigEndian.Uint16(data[offset : offset+2]))
		if segmentLength < 2 || offset+segmentLength > len(data) {
			return 1
		}
		if marker == 0xe1 {
			if orientation, ok := parseEXIFOrientation(data[offset+2 : offset+segmentLength]); ok {
				return orientation
			}
		}
		offset += segmentLength
	}
	return 1
}

func parseEXIFOrientation(payload []byte) (uint16, bool) {
	if len(payload) < 14 || !bytes.Equal(payload[:6], []byte("Exif\x00\x00")) {
		return 0, false
	}
	tiff := payload[6:]
	var order binary.ByteOrder
	switch string(tiff[:2]) {
	case "II":
		order = binary.LittleEndian
	case "MM":
		order = binary.BigEndian
	default:
		return 0, false
	}
	if order.Uint16(tiff[2:4]) != 42 {
		return 0, false
	}
	ifdOffset := uint64(order.Uint32(tiff[4:8]))
	if ifdOffset > uint64(len(tiff))-2 {
		return 0, false
	}
	entryCount := uint64(order.Uint16(tiff[ifdOffset : ifdOffset+2]))
	entriesStart := ifdOffset + 2
	if entryCount > (uint64(len(tiff))-entriesStart)/12 {
		return 0, false
	}
	for index := uint64(0); index < entryCount; index++ {
		entry := tiff[entriesStart+index*12 : entriesStart+(index+1)*12]
		if order.Uint16(entry[:2]) != 0x0112 {
			continue
		}
		if order.Uint16(entry[2:4]) != 3 || order.Uint32(entry[4:8]) != 1 {
			return 0, false
		}
		orientation := order.Uint16(entry[8:10])
		return orientation, orientation >= 1 && orientation <= 8
	}
	return 0, false
}

func validateStrictPNGContainer(ctx context.Context, data []byte) error { //nolint:gocyclo // PNG chunk grammar is intentionally explicit
	const signatureLength = 8
	if len(data) < signatureLength || !bytes.Equal(data[:signatureLength], []byte("\x89PNG\r\n\x1a\n")) {
		return ErrInvalidProfileImage
	}
	offset := signatureLength
	seenIHDR, seenPLTE, seenIDAT, endedIDAT := false, false, false, false
	pngColorType, paletteEntries := -1, 0
	seenAncillary := make(map[string]bool)
	var idatSegments [][]byte
	var expectedInflated int64
	chunkCount, idatCount := 0, 0
	for offset < len(data) {
		chunkCount++
		if chunkCount > maxProfilePNGChunks {
			return ErrInvalidProfileImage
		}
		if offset+12 > len(data) {
			return ErrInvalidProfileImage
		}
		length := int64(binary.BigEndian.Uint32(data[offset : offset+4]))
		chunkEnd := int64(offset) + 12 + length
		if chunkEnd > int64(len(data)) {
			return ErrInvalidProfileImage
		}
		typeStart := offset + 4
		dataStart := offset + 8
		dataEnd := dataStart + int(length)
		if !validPNGChunkType(data[typeStart:dataStart]) {
			return ErrInvalidProfileImage
		}
		chunkType := string(data[typeStart:dataStart])
		wantCRC := binary.BigEndian.Uint32(data[dataEnd : dataEnd+4])
		if crc32.ChecksumIEEE(data[typeStart:dataEnd]) != wantCRC {
			return ErrInvalidProfileImage
		}
		if !seenIHDR && chunkType != "IHDR" {
			return ErrInvalidProfileImage
		}
		switch chunkType {
		case "IHDR":
			if seenIHDR || length != 13 || data[dataStart+8] != 8 {
				return ErrInvalidProfileImage
			}
			var ok bool
			expectedInflated, ok = pngExpectedInflatedBytes(data[dataStart:dataEnd])
			if !ok {
				return ErrInvalidProfileImage
			}
			seenIHDR = true
			pngColorType = int(data[dataStart+9])
		case "PLTE":
			if !seenIHDR || seenPLTE || seenIDAT || length == 0 || length > 256*3 || length%3 != 0 ||
				pngColorType == 0 || pngColorType == 4 {
				return ErrInvalidProfileImage
			}
			seenPLTE = true
			paletteEntries = int(length / 3)
		case "IDAT":
			if !seenIHDR || endedIDAT || pngColorType == 3 && !seenPLTE {
				return ErrInvalidProfileImage
			}
			seenIDAT = true
			idatCount++
			if idatCount > maxProfilePNGIDATChunks {
				return ErrInvalidProfileImage
			}
			idatSegments = append(idatSegments, data[dataStart:dataEnd])
		case "IEND":
			if !seenIDAT || length != 0 || chunkEnd != int64(len(data)) {
				return ErrInvalidProfileImage
			}
			return validateExactZlibStream(ctx, idatSegments, expectedInflated)
		case "acTL", "fcTL", "fdAT":
			return ErrInvalidProfileImage
		default:
			// Unknown critical chunks are not valid profile-image metadata.
			if chunkType[0]&0x20 == 0 {
				return ErrInvalidProfileImage
			}
			if !validPNGAncillaryChunk(chunkType, data[dataStart:dataEnd], seenAncillary, seenPLTE, seenIDAT, pngColorType, paletteEntries) {
				return ErrInvalidProfileImage
			}
			if seenIDAT {
				endedIDAT = true
			}
		}
		offset = int(chunkEnd)
	}
	return ErrInvalidProfileImage
}

//nolint:gocyclo // fixed-shape ancillary policy is clearer as one exhaustive switch
func validPNGAncillaryChunk(
	chunkType string,
	payload []byte,
	seen map[string]bool,
	seenPLTE, seenIDAT bool,
	colorType, paletteEntries int,
) bool {
	singleton := false
	switch chunkType {
	case "cHRM":
		singleton = true
		if seenPLTE || seenIDAT || len(payload) != 32 {
			return false
		}
	case "gAMA":
		singleton = true
		if seenPLTE || seenIDAT || len(payload) != 4 || binary.BigEndian.Uint32(payload) == 0 {
			return false
		}
	case "sRGB":
		singleton = true
		if seenPLTE || seenIDAT || len(payload) != 1 || payload[0] > 3 {
			return false
		}
	case "pHYs":
		singleton = true
		if seenIDAT || len(payload) != 9 || payload[8] > 1 {
			return false
		}
	case "eXIf":
		singleton = true
		if seenIDAT || len(payload) > 65527 || !validBoundedTIFF(payload) {
			return false
		}
	case "tRNS":
		singleton = true
		if seenIDAT {
			return false
		}
		switch colorType {
		case 0:
			if len(payload) != 2 || binary.BigEndian.Uint16(payload) > 255 {
				return false
			}
		case 2:
			if len(payload) != 6 || binary.BigEndian.Uint16(payload[0:2]) > 255 ||
				binary.BigEndian.Uint16(payload[2:4]) > 255 || binary.BigEndian.Uint16(payload[4:6]) > 255 {
				return false
			}
		case 3:
			if !seenPLTE || len(payload) == 0 || len(payload) > paletteEntries {
				return false
			}
		default:
			return false
		}
	default:
		// Reject text, comments, compressed profiles, EXIF, and unknown private
		// chunks. Accepted color/physical metadata above has fixed binary shape
		// and is stripped during canonical re-encoding.
		return false
	}
	if singleton && seen[chunkType] {
		return false
	}
	if singleton {
		seen[chunkType] = true
	}
	return true
}

func validPNGChunkType(chunkType []byte) bool {
	if len(chunkType) != 4 || chunkType[2]&0x20 != 0 {
		return false
	}
	for _, character := range chunkType {
		if (character < 'A' || character > 'Z') && (character < 'a' || character > 'z') {
			return false
		}
	}
	return true
}

func validateExactZlibStream(ctx context.Context, segments [][]byte, expected int64) error {
	reader := &segmentedByteReader{segments: segments}
	zlibReader, err := zlib.NewReader(reader)
	if err != nil {
		return ErrInvalidProfileImage
	}
	bounded := &contextReader{ctx: ctx, reader: zlibReader}
	written, copyErr := io.CopyN(io.Discard, bounded, expected+1)
	closeErr := zlibReader.Close()
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if written != expected || !errors.Is(copyErr, io.EOF) || closeErr != nil || reader.remaining() != 0 {
		return ErrInvalidProfileImage
	}
	return nil
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r *contextReader) Read(data []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(data)
}

type segmentedByteReader struct {
	segments [][]byte
	segment  int
	offset   int
}

func (r *segmentedByteReader) Read(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}
	for r.segment < len(r.segments) && r.offset == len(r.segments[r.segment]) {
		r.segment++
		r.offset = 0
	}
	if r.segment == len(r.segments) {
		return 0, io.EOF
	}
	written := copy(data, r.segments[r.segment][r.offset:])
	r.offset += written
	return written, nil
}

func (r *segmentedByteReader) ReadByte() (byte, error) {
	var data [1]byte
	_, err := r.Read(data[:])
	return data[0], err
}

func (r *segmentedByteReader) remaining() int {
	remaining := 0
	for index := r.segment; index < len(r.segments); index++ {
		start := 0
		if index == r.segment {
			start = r.offset
		}
		remaining += len(r.segments[index]) - start
	}
	return remaining
}

func pngExpectedInflatedBytes(ihdr []byte) (int64, bool) {
	width := int64(binary.BigEndian.Uint32(ihdr[0:4]))
	height := int64(binary.BigEndian.Uint32(ihdr[4:8]))
	bitDepth, colorType, interlace := int64(ihdr[8]), ihdr[9], ihdr[12]
	channels := int64(0)
	switch colorType {
	case 0, 3:
		channels = 1
	case 2:
		channels = 3
	case 4:
		channels = 2
	case 6:
		channels = 4
	default:
		return 0, false
	}
	if width <= 0 || height <= 0 || (interlace != 0 && interlace != 1) {
		return 0, false
	}
	rowBytes := func(passWidth int64) int64 { return 1 + (passWidth*channels*bitDepth+7)/8 }
	if interlace == 0 {
		return height * rowBytes(width), true
	}
	startsX := [...]int64{0, 4, 0, 2, 0, 1, 0}
	stepsX := [...]int64{8, 8, 4, 4, 2, 2, 1}
	startsY := [...]int64{0, 0, 4, 0, 2, 0, 1}
	stepsY := [...]int64{8, 8, 8, 4, 4, 2, 2}
	var total int64
	for index := range startsX {
		passWidth := max(int64(0), (width-startsX[index]+stepsX[index]-1)/stepsX[index])
		passHeight := max(int64(0), (height-startsY[index]+stepsY[index]-1)/stepsY[index])
		if passWidth > 0 && passHeight > 0 {
			total += passHeight * rowBytes(passWidth)
		}
	}
	return total, true
}

type baselineJPEGComponent struct {
	id byte
	h  int
	v  int
}

type baselineJPEGHuffmanTable struct {
	valid       bool
	counts      [17]int
	firstCode   [17]int
	firstSymbol [17]int
	symbols     [maxProfileJPEGHuffmanSymbols]byte
}

type baselineJPEGScanComponent struct {
	component baselineJPEGComponent
	dcTable   int
	acTable   int
}

// validateStrictJPEGContainer validates the complete, single-scan baseline JPEG entropy
// stream rather than relying on image/jpeg's intentionally tolerant decoder.
// That lets us reject payload bytes hidden between the final MCU and EOI.
func validateStrictJPEGContainer(ctx context.Context, data []byte) error { //nolint:gocyclo // JPEG marker grammar is intentionally explicit
	if len(data) < 4 || data[0] != 0xff || data[1] != 0xd8 {
		return ErrInvalidProfileImage
	}
	offset := 2
	var width, height, maxH, maxV int
	components := make(map[byte]baselineJPEGComponent)
	var dcTables, acTables [4]baselineJPEGHuffmanTable
	restartInterval := 0
	segmentCount, huffmanTableCount := 0, 0
	seenJFIF, seenEXIF := false, false
	for offset < len(data) {
		if err := ctx.Err(); err != nil {
			return err
		}
		segmentCount++
		if segmentCount > maxProfileJPEGSegments {
			return ErrInvalidProfileImage
		}
		marker, next, ok := nextJPEGMarker(data, offset)
		if !ok {
			return ErrInvalidProfileImage
		}
		offset = next
		if marker == 0xd8 || marker == 0xd9 || marker == 0x01 || marker >= 0xd0 && marker <= 0xd7 {
			return ErrInvalidProfileImage
		}
		if marker != 0xc0 && marker != 0xc4 && marker != 0xda && marker != 0xdb && marker != 0xdd &&
			marker != 0xe0 && marker != 0xe1 {
			// Reject reserved, arithmetic, hierarchical, DNL, and non-baseline markers.
			return ErrInvalidProfileImage
		}
		if offset+2 > len(data) {
			return ErrInvalidProfileImage
		}
		segmentLength := int(binary.BigEndian.Uint16(data[offset : offset+2]))
		if segmentLength < 2 || offset+segmentLength > len(data) {
			return ErrInvalidProfileImage
		}
		segment := data[offset+2 : offset+segmentLength]
		offset += segmentLength
		switch marker {
		case 0xe0:
			if seenJFIF || len(components) != 0 || !validJPEGJFIFMetadata(segment) {
				return ErrInvalidProfileImage
			}
			seenJFIF = true
		case 0xe1:
			if seenEXIF || len(components) != 0 || !validBoundedJPEGEXIFMetadata(segment) {
				return ErrInvalidProfileImage
			}
			seenEXIF = true
		case 0xc0:
			if len(components) != 0 || len(segment) < 9 || segment[0] != 8 {
				return ErrInvalidProfileImage
			}
			height = int(binary.BigEndian.Uint16(segment[1:3]))
			width = int(binary.BigEndian.Uint16(segment[3:5]))
			componentCount := int(segment[5])
			if width <= 0 || height <= 0 || componentCount < 1 || componentCount > 4 || len(segment) != 6+3*componentCount {
				return ErrInvalidProfileImage
			}
			for index := 0; index < componentCount; index++ {
				base := 6 + 3*index
				id, sampling, quantizationTable := segment[base], segment[base+1], segment[base+2]
				h, v := int(sampling>>4), int(sampling&0x0f)
				if id == 0 || h < 1 || h > 4 || v < 1 || v > 4 || quantizationTable > 3 {
					return ErrInvalidProfileImage
				}
				if _, duplicate := components[id]; duplicate {
					return ErrInvalidProfileImage
				}
				components[id] = baselineJPEGComponent{id: id, h: h, v: v}
				maxH, maxV = max(maxH, h), max(maxV, v)
			}
		case 0xc4:
			if !parseBaselineJPEGHuffmanTables(ctx, segment, &dcTables, &acTables, &huffmanTableCount) {
				return jpegContainerValidationError(ctx)
			}
		case 0xdb:
			if !validBaselineJPEGQuantizationTables(segment) {
				return ErrInvalidProfileImage
			}
		case 0xdd:
			if len(segment) != 2 {
				return ErrInvalidProfileImage
			}
			restartInterval = int(binary.BigEndian.Uint16(segment))
		case 0xda:
			scan, ok := parseBaselineJPEGScan(segment, components, dcTables, acTables)
			if !ok || len(scan) != len(components) {
				return ErrInvalidProfileImage
			}
			reader := baselineJPEGBitReader{ctx: ctx, data: data, offset: offset}
			mcuColumns := (width + 8*maxH - 1) / (8 * maxH)
			mcuRows := (height + 8*maxV - 1) / (8 * maxV)
			if !decodeBaselineJPEGScan(&reader, scan, dcTables, acTables, mcuColumns*mcuRows, restartInterval) {
				return jpegContainerValidationError(ctx)
			}
			if !reader.consumeTerminalEOI() || reader.offset != len(data) {
				return jpegContainerValidationError(ctx)
			}
			return nil
		}
	}
	return ErrInvalidProfileImage
}

func validJPEGJFIFMetadata(segment []byte) bool {
	return len(segment) == 14 && bytes.Equal(segment[:5], []byte("JFIF\x00")) &&
		segment[5] == 1 && segment[7] <= 2 && segment[12] == 0 && segment[13] == 0
}

func validBoundedJPEGEXIFMetadata(segment []byte) bool {
	if len(segment) < 20 || !bytes.Equal(segment[:6], []byte("Exif\x00\x00")) {
		return false
	}
	return validBoundedTIFF(segment[6:])
}

func validBoundedTIFF(tiff []byte) bool {
	if len(tiff) < 14 {
		return false
	}
	var order binary.ByteOrder
	if bytes.Equal(tiff[:2], []byte("II")) {
		order = binary.LittleEndian
	} else if bytes.Equal(tiff[:2], []byte("MM")) {
		order = binary.BigEndian
	} else {
		return false
	}
	if order.Uint16(tiff[2:4]) != 42 {
		return false
	}
	state := exifValidationState{seenIFDs: make(map[uint32]bool), covered: make([]bool, len(tiff))}
	markEXIFRange(state.covered, 0, 8)
	if !validateEXIFDirectory(tiff, order, order.Uint32(tiff[4:8]), &state, 0) {
		return false
	}
	for index, covered := range state.covered {
		if !covered && tiff[index] != 0 {
			return false
		}
	}
	return true
}

type exifValidationState struct {
	seenIFDs     map[uint32]bool
	totalEntries int
	covered      []bool
}

func validateEXIFDirectory(tiff []byte, order binary.ByteOrder, offset uint32, state *exifValidationState, depth int) bool { //nolint:gocyclo // bounded TIFF grammar is explicit
	if depth > 3 || state.seenIFDs[offset] || uint64(offset)+2 > uint64(len(tiff)) {
		return false
	}
	state.seenIFDs[offset] = true
	rawEntryCount := order.Uint16(tiff[offset : offset+2])
	entryCount := uint64(rawEntryCount)
	state.totalEntries += int(rawEntryCount)
	if entryCount > 128 || state.totalEntries > 384 {
		return false
	}
	entriesStart := uint64(offset) + 2
	entriesEnd := entriesStart + entryCount*12
	if entriesEnd+4 > uint64(len(tiff)) {
		return false
	}
	markEXIFRange(state.covered, uint64(offset), entriesEnd+4)
	seenTags := make(map[uint16]bool, entryCount)
	for index := uint64(0); index < entryCount; index++ {
		entryOffset := entriesStart + index*12
		entry := tiff[entryOffset : entryOffset+12]
		tag := order.Uint16(entry[0:2])
		if seenTags[tag] || tag == 0x0201 || tag == 0x0202 {
			return false
		}
		seenTags[tag] = true
		typeSize := exifTypeSize(order.Uint16(entry[2:4]))
		count := uint64(order.Uint32(entry[4:8]))
		if typeSize == 0 || count == 0 || count > 1<<20 || count > ^uint64(0)/typeSize {
			return false
		}
		valueSize := count * typeSize
		if valueSize > 4 {
			valueOffset := uint64(order.Uint32(entry[8:12]))
			if valueOffset > uint64(len(tiff)) || valueSize > uint64(len(tiff))-valueOffset {
				return false
			}
			markEXIFRange(state.covered, valueOffset, valueOffset+valueSize)
		}
		if tag == 0x8769 || tag == 0x8825 || tag == 0xa005 {
			if order.Uint16(entry[2:4]) != 4 || count != 1 ||
				!validateEXIFDirectory(tiff, order, order.Uint32(entry[8:12]), state, depth+1) {
				return false
			}
		}
	}
	return order.Uint32(tiff[entriesEnd:entriesEnd+4]) == 0
}

func markEXIFRange(covered []bool, start, end uint64) {
	for index := start; index < end; index++ {
		covered[index] = true
	}
}

func exifTypeSize(valueType uint16) uint64 {
	switch valueType {
	case 1, 2, 7:
		return 1
	case 3:
		return 2
	case 4, 9, 11:
		return 4
	case 5, 10, 12:
		return 8
	default:
		return 0
	}
}

func jpegContainerValidationError(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return ErrInvalidProfileImage
}

func validBaselineJPEGQuantizationTables(segment []byte) bool {
	seen := false
	for len(segment) > 0 {
		if len(segment) < 65 || segment[0]>>4 != 0 || segment[0]&0x0f > 3 {
			return false
		}
		seen = true
		segment = segment[65:]
	}
	return seen
}

func parseBaselineJPEGHuffmanTables(
	ctx context.Context,
	segment []byte,
	dcTables, acTables *[4]baselineJPEGHuffmanTable,
	tableCount *int,
) bool {
	seen := false
	for len(segment) > 0 {
		if ctx.Err() != nil {
			return false
		}
		if len(segment) < 17 {
			return false
		}
		class, id := segment[0]>>4, int(segment[0]&0x0f)
		if class > 1 || id > 3 {
			return false
		}
		count := 0
		for _, value := range segment[1:17] {
			count += int(value)
		}
		if count == 0 || count > maxProfileJPEGHuffmanSymbols || len(segment) < 17+count {
			return false
		}
		*tableCount++
		if *tableCount > maxProfileJPEGHuffmanTables {
			return false
		}
		target := &dcTables[id]
		if class == 1 {
			target = &acTables[id]
		}
		if target.valid {
			return false
		}
		table := baselineJPEGHuffmanTable{valid: true}
		code, symbolOffset := 0, 0
		for length := 1; length <= 16; length++ {
			entries := int(segment[length])
			if code+entries > 1<<length {
				return false
			}
			table.counts[length] = entries
			table.firstCode[length] = code
			table.firstSymbol[length] = symbolOffset
			for index := 0; index < entries; index++ {
				if symbolOffset >= len(table.symbols) {
					return false
				}
				table.symbols[symbolOffset] = segment[17+symbolOffset]
				symbolOffset++
			}
			code = (code + entries) << 1
		}
		*target = table
		seen = true
		segment = segment[17+count:]
	}
	return seen
}

func parseBaselineJPEGScan(
	segment []byte,
	components map[byte]baselineJPEGComponent,
	dcTables, acTables [4]baselineJPEGHuffmanTable,
) ([]baselineJPEGScanComponent, bool) {
	if len(segment) < 6 {
		return nil, false
	}
	count := int(segment[0])
	if count < 1 || len(segment) != 1+2*count+3 || segment[len(segment)-3] != 0 ||
		segment[len(segment)-2] != 63 || segment[len(segment)-1] != 0 {
		return nil, false
	}
	result := make([]baselineJPEGScanComponent, 0, count)
	seen := make(map[byte]bool, count)
	for index := 0; index < count; index++ {
		id, selectors := segment[1+2*index], segment[2+2*index]
		component, exists := components[id]
		dc, ac := int(selectors>>4), int(selectors&0x0f)
		if !exists || seen[id] || dc > 3 || ac > 3 || !dcTables[dc].valid || !acTables[ac].valid {
			return nil, false
		}
		seen[id] = true
		result = append(result, baselineJPEGScanComponent{component: component, dcTable: dc, acTable: ac})
	}
	return result, true
}

type baselineJPEGBitReader struct {
	ctx       context.Context
	data      []byte
	offset    int
	value     byte
	remaining uint8
}

func (r *baselineJPEGBitReader) readBit() (int, bool) {
	if r.remaining == 0 {
		if r.ctx.Err() != nil {
			return 0, false
		}
		if r.offset >= len(r.data) {
			return 0, false
		}
		value := r.data[r.offset]
		r.offset++
		if value == 0xff {
			if r.offset >= len(r.data) || r.data[r.offset] != 0x00 {
				return 0, false
			}
			r.offset++
		}
		r.value = value
		r.remaining = 8
	}
	r.remaining--
	return int(r.value>>r.remaining) & 1, true
}

func (r *baselineJPEGBitReader) skipBits(count int) bool {
	for range count {
		if _, ok := r.readBit(); !ok {
			return false
		}
	}
	return true
}

func (r *baselineJPEGBitReader) alignWithOnePadding() bool {
	if r.remaining > 0 && r.value&byte((1<<r.remaining)-1) != byte((1<<r.remaining)-1) {
		return false
	}
	r.remaining = 0
	return true
}

func (r *baselineJPEGBitReader) consumeMarker(expected byte) bool {
	if !r.alignWithOnePadding() || r.offset >= len(r.data) || r.data[r.offset] != 0xff {
		return false
	}
	for r.offset < len(r.data) && r.data[r.offset] == 0xff {
		r.offset++
	}
	if r.offset >= len(r.data) || r.data[r.offset] != expected {
		return false
	}
	r.offset++
	return true
}

func (r *baselineJPEGBitReader) consumeTerminalEOI() bool { return r.consumeMarker(0xd9) }

func decodeBaselineJPEGScan(
	reader *baselineJPEGBitReader,
	scan []baselineJPEGScanComponent,
	dcTables, acTables [4]baselineJPEGHuffmanTable,
	mcuCount, restartInterval int,
) bool {
	restart := byte(0xd0)
	for mcu := 0; mcu < mcuCount; mcu++ {
		for _, scanComponent := range scan {
			for block := 0; block < scanComponent.component.h*scanComponent.component.v; block++ {
				if !decodeBaselineJPEGBlock(reader, &dcTables[scanComponent.dcTable], &acTables[scanComponent.acTable]) {
					return false
				}
			}
		}
		if restartInterval > 0 && (mcu+1)%restartInterval == 0 && mcu+1 < mcuCount {
			if !reader.consumeMarker(restart) {
				return false
			}
			restart = 0xd0 + (restart-0xd0+1)%8
		}
	}
	return reader.alignWithOnePadding()
}

func decodeBaselineJPEGBlock(reader *baselineJPEGBitReader, dcTable, acTable *baselineJPEGHuffmanTable) bool {
	dcSize, ok := decodeBaselineJPEGHuffmanSymbol(reader, dcTable)
	if !ok || dcSize > 11 || !reader.skipBits(int(dcSize)) {
		return false
	}
	coefficient := 1
	for coefficient < 64 {
		symbol, ok := decodeBaselineJPEGHuffmanSymbol(reader, acTable)
		if !ok {
			return false
		}
		run, size := int(symbol>>4), int(symbol&0x0f)
		if size == 0 {
			if run == 0 {
				return true
			}
			if run != 15 {
				return false
			}
			coefficient += 16
			if coefficient > 64 {
				return false
			}
			continue
		}
		if size > 10 {
			return false
		}
		coefficient += run
		if coefficient >= 64 || !reader.skipBits(size) {
			return false
		}
		coefficient++
	}
	return true
}

func decodeBaselineJPEGHuffmanSymbol(reader *baselineJPEGBitReader, table *baselineJPEGHuffmanTable) (byte, bool) {
	code := 0
	for length := 1; length <= 16; length++ {
		bit, ok := reader.readBit()
		if !ok {
			return 0, false
		}
		code = code<<1 | bit
		first, count := table.firstCode[length], table.counts[length]
		if code >= first && code < first+count {
			return table.symbols[table.firstSymbol[length]+code-first], true
		}
	}
	return 0, false
}

func nextJPEGMarker(data []byte, offset int) (byte, int, bool) {
	if offset >= len(data) || data[offset] != 0xff {
		return 0, offset, false
	}
	for offset < len(data) && data[offset] == 0xff {
		offset++
	}
	if offset >= len(data) || data[offset] == 0x00 {
		return 0, offset, false
	}
	return data[offset], offset + 1, true
}
