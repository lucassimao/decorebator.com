package common

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var ErrDecodedDataTooLarge = errors.New("decoded data is too large")

// GenerateRandomString returns a URL-safe random string of n bytes (hex-encoded length 2n)
func GenerateRandomString(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	const hexdigits = "0123456789abcdef"
	out := make([]byte, n*2)
	for i, by := range b {
		out[i*2] = hexdigits[by>>4]
		out[i*2+1] = hexdigits[by&0x0f]
	}
	return string(out)
}

// DecodeImageBase64 takes a Base64 string (with or without a `data:` URI prefix)
// and returns the decoded bytes and, if present, the MIME content type.
//   - If the input is a data URI like "data:image/png;base64,AAA…", it extracts
//     "image/png" as contentType and decodes the payload.
//   - If there's no "data:" prefix, contentType == "" and it just decodes.
//
// Returns (data, contentType, nil) on success, or (nil, "", err) on failure.
func DecodeImageBase64(b64 string) ([]byte, string, error) {
	return DecodeImageBase64Bounded(b64, 0)
}

// DecodeImageBase64Bounded rejects encoded input before allocation when its
// maximum decoded size can exceed maxDecodedBytes. A nonpositive limit keeps
// the legacy unbounded behavior for non-user-controlled callers.
func DecodeImageBase64Bounded(b64 string, maxDecodedBytes int) ([]byte, string, error) {
	var contentType string

	// 1) Check for a `data:` URI prefix
	if strings.HasPrefix(b64, "data:") {
		// Split at the first comma: ["data:image/png;base64", "<base64data>"]
		parts := strings.SplitN(b64, ",", 2)
		if len(parts) != 2 {
			return nil, "", errors.New("invalid data URI format")
		}

		header := parts[0]
		payload := parts[1]
		switch header {
		case "data:image/jpeg;base64":
			contentType = "image/jpeg"
		case "data:image/jpg;base64":
			contentType = "image/jpg"
		case "data:image/png;base64":
			contentType = "image/png"
		default:
			return nil, "", errors.New("unsupported data URI header")
		}
		b64 = payload
	}
	// Select the alphabet before allocating so URL-safe payloads do not first
	// allocate a full failed standard-decoding buffer.
	encoding, ok := strictPaddedBase64Encoding(b64)
	if !ok {
		return nil, "", errors.New("invalid base64 data")
	}
	if maxDecodedBytes > 0 && exactBase64DecodedLen(b64) > maxDecodedBytes {
		return nil, "", fmt.Errorf("%w: limit is %d bytes", ErrDecodedDataTooLarge, maxDecodedBytes)
	}
	decoded, err := encoding.Strict().DecodeString(b64)
	if err == nil && (maxDecodedBytes <= 0 || len(decoded) <= maxDecodedBytes) {
		return decoded, contentType, nil
	}

	return nil, "", errors.New("invalid base64 data")
}

func strictPaddedBase64Encoding(encoded string) (*base64.Encoding, bool) { //nolint:gocyclo // canonical alphabet and padding grammar is explicit
	if encoded == "" || len(encoded)%4 != 0 || strings.ContainsAny(encoded, "\r\n\t ") {
		return nil, false
	}
	paddingStart := strings.IndexByte(encoded, '=')
	if paddingStart < 0 {
		paddingStart = len(encoded)
	}
	padding := len(encoded) - paddingStart
	if padding > 2 || strings.Contains(encoded[:paddingStart], "=") {
		return nil, false
	}
	for _, character := range encoded[paddingStart:] {
		if character != '=' {
			return nil, false
		}
	}
	hasStandard, hasURL := false, false
	for _, character := range encoded[:paddingStart] {
		switch {
		case character >= 'A' && character <= 'Z',
			character >= 'a' && character <= 'z',
			character >= '0' && character <= '9':
		case character == '+' || character == '/':
			hasStandard = true
		case character == '-' || character == '_':
			hasURL = true
		default:
			return nil, false
		}
	}
	if hasStandard && hasURL {
		return nil, false
	}
	if hasURL {
		return base64.URLEncoding, true
	}
	return base64.StdEncoding, true
}

func exactBase64DecodedLen(encoded string) int {
	decodedLen := base64.StdEncoding.DecodedLen(len(encoded))
	if strings.HasSuffix(encoded, "==") {
		return decodedLen - 2
	}
	if strings.HasSuffix(encoded, "=") {
		return decodedLen - 1
	}
	return decodedLen
}

// GetBcryptCost returns the appropriate bcrypt cost factor based on the environment
// and any environment variable override. This allows for fast tests (cost 4) while
// maintaining production security (default cost 10).
func GetBcryptCost() int {
	// Check for explicit environment variable override first
	if costStr := os.Getenv("BCRYPT_COST"); costStr != "" {
		if cost, err := strconv.Atoi(costStr); err == nil {
			// Validate cost is within bcrypt limits
			if cost >= bcrypt.MinCost && cost <= bcrypt.MaxCost {
				if os.Getenv("ENV") == "production" && cost < ProductionBcryptMinimumCost {
					return ProductionBcryptMinimumCost
				}
				return cost
			}
		}
	}

	// Use MinCost for non-production environments (test, development)
	env := os.Getenv("ENV")
	if env == "test" || env == "development" {
		return bcrypt.MinCost // Cost 4 - fastest for non-production
	}

	// Production default
	return bcrypt.DefaultCost // Cost 10 - production security
}
