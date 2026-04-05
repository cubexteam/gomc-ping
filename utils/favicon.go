package utils

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

// SaveFavicon saves a base64 encoded Minecraft favicon to a file.
// It verifies that the data is a valid PNG image.
func SaveFavicon(data string, path string) error {
	if data == "" {
		return fmt.Errorf("favicon data is empty")
	}

	const prefix = "data:image/png;base64,"
	if strings.HasPrefix(data, prefix) {
		data = data[len(prefix):]
	}

	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return fmt.Errorf("failed to decode base64: %v", err)
	}

	// Verify PNG signature: \x89PNG\r\n\x1a\n
	pngSignature := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	if len(decoded) < 8 || !bytes.Equal(decoded[:8], pngSignature) {
		return fmt.Errorf("invalid png signature")
	}

	return os.WriteFile(path, decoded, 0644)
}
