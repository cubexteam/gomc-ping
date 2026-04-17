package utils

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

// SaveFavicon saves a base64-encoded Minecraft favicon to a file.
// Accepts both standard and URL-safe base64 (with or without padding).
func SaveFavicon(data string, path string) error {
	if data == "" {
		return fmt.Errorf("favicon data is empty")
	}

	const prefix = "data:image/png;base64,"
	if strings.HasPrefix(data, prefix) {
		data = data[len(prefix):]
	}

	// Try standard encoding first, then URL-safe, then padded variants.
	decoded, err := tryDecode(data)
	if err != nil {
		return fmt.Errorf("failed to decode base64: %v", err)
	}

	pngSignature := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	if len(decoded) < 8 || !bytes.Equal(decoded[:8], pngSignature) {
		return fmt.Errorf("invalid png signature")
	}

	return os.WriteFile(path, decoded, 0644)
}

func tryDecode(s string) ([]byte, error) {
	encodings := []*base64.Encoding{
		base64.StdEncoding,
		base64.URLEncoding,
		base64.RawStdEncoding,
		base64.RawURLEncoding,
	}
	var lastErr error
	for _, enc := range encodings {
		if b, err := enc.DecodeString(s); err == nil {
			return b, nil
		} else {
			lastErr = err
		}
	}
	return nil, lastErr
}
