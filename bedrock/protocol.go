package bedrock

import (
	"encoding/binary"
	"fmt"
	"strings"
)

func ParseResponse(data []byte) (motd string, playersOn, playersMax int, version, edition string, protocol int, err error) {
	if len(data) < 35 {
		return "", 0, 0, "", "", 0, fmt.Errorf("packet too short")
	}

	if data[0] != 0x1c {
		return "", 0, 0, "", "", 0, fmt.Errorf("invalid magic byte")
	}

	// Read string length at offset 33
	strLen := int(binary.BigEndian.Uint16(data[33:35]))
	if len(data) < 35+strLen {
		return "", 0, 0, "", "", 0, fmt.Errorf("string length mismatch")
	}

	s := string(data[35 : 35+strLen])
	parts := strings.Split(s, ";")

	if len(parts) < 6 {
		return "", 0, 0, "", "", 0, fmt.Errorf("invalid response format")
	}

	edition = parts[0]
	motd = parts[1]
	fmt.Sscanf(parts[2], "%d", &protocol)
	version = parts[3]
	fmt.Sscanf(parts[4], "%d", &playersOn)
	fmt.Sscanf(parts[5], "%d", &playersMax)

	return motd, playersOn, playersMax, version, edition, protocol, nil
}
