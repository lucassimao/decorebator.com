package common

import (
	"encoding/base64"
	"errors"
	"strings"
)

// DecodeImageBase64 takes a Base64 string (with or without a `data:` URI prefix)
// and returns the decoded bytes and, if present, the MIME content type.
//   - If the input is a data URI like "data:image/png;base64,AAA…", it extracts
//     "image/png" as contentType and decodes the payload.
//   - If there's no "data:" prefix, contentType == "" and it just decodes.
//
// Returns (data, contentType, nil) on success, or (nil, "", err) on failure.
func DecodeImageBase64(b64 string) ([]byte, string, error) {
	var contentType string

	// 1) Check for a `data:` URI prefix
	if strings.HasPrefix(b64, "data:") {
		// Split at the first comma: ["data:image/png;base64", "<base64data>"]
		parts := strings.SplitN(b64, ",", 2)
		if len(parts) != 2 {
			return nil, "", errors.New("invalid data URI format")
		}

		header := parts[0] // e.g. "data:image/png;base64"
		payload := parts[1]

		// Extract MIME type between "data:" and the first ";" (or end-of-string)
		// Example: header == "data:image/png;base64"
		//   semicolonIndex == len("data:image/png")
		if semicolonIndex := strings.Index(header, ";"); semicolonIndex != -1 {
			contentType = header[len("data:"):semicolonIndex]
		} else {
			// No semicolon – maybe someone gave "data:image/png,<base64>"
			contentType = header[len("data:"):]
		}

		// Replace b64 with just the Base64 payload (after the comma)
		b64 = payload
	}

	// 2) Attempt standard Base64 decode
	decoded, err := base64.StdEncoding.DecodeString(b64)
	if err == nil {
		return decoded, contentType, nil
	}

	// 3) If that fails, try URL‐safe Base64 (in case it's encoded with '-' and '_')
	decoded, err = base64.URLEncoding.DecodeString(b64)
	if err == nil {
		return decoded, contentType, nil
	}

	return nil, "", errors.New("invalid base64 data")
}
