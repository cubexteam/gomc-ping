package source

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"github.com/cubexteam/gomc-ping/models"
)

var a2sQueryPacket = []byte{
	0xFF, 0xFF, 0xFF, 0xFF,
	0x54,
	0x53, 0x6F, 0x75, 0x72, 0x63, 0x65, 0x20, 0x45, 0x6E, 0x67, 0x69, 0x6E, 0x65, 0x20, 0x51, 0x75, 0x65, 0x72, 0x79,
	0x00,
}

func Ping(host string, port uint16, timeout time.Duration) (*models.Response, error) {
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := net.DialTimeout("udp", addr, timeout)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(timeout))

	start := time.Now()
	if _, err := conn.Write(a2sQueryPacket); err != nil {
		return nil, err
	}

	resp := make([]byte, 2048)
	n, err := conn.Read(resp)
	if err != nil {
		return nil, err
	}

	// Handle Challenge (0x41) — build a new slice to avoid aliasing a2sQueryPacket
	if n >= 9 && resp[4] == 0x41 {
		challenge := make([]byte, len(a2sQueryPacket)+4)
		copy(challenge, a2sQueryPacket)
		copy(challenge[len(a2sQueryPacket):], resp[5:9])

		if _, err := conn.Write(challenge); err != nil {
			return nil, err
		}
		n, err = conn.Read(resp)
		if err != nil {
			return nil, err
		}
	}

	latency := time.Since(start)
	return parseA2S(resp[:n], host, port, latency)
}

func parseA2S(data []byte, host string, port uint16, latency time.Duration) (*models.Response, error) {
	if len(data) < 5 || !bytes.Equal(data[:4], []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
		return nil, fmt.Errorf("invalid a2s header")
	}
	if data[4] != 0x49 {
		return nil, fmt.Errorf("unexpected a2s response: 0x%02x", data[4])
	}

	r := bytes.NewReader(data[5:])

	if _, err := readByte(r); err != nil {
		return nil, err
	}
	name, err := readString(r)
	if err != nil {
		return nil, err
	}
	mapName, err := readString(r)
	if err != nil {
		return nil, err
	}
	folder, err := readString(r)
	if err != nil {
		return nil, err
	}
	game, err := readString(r)
	if err != nil {
		return nil, err
	}
	if _, err := readUint16(r); err != nil {
		return nil, err
	}
	players, err := readByte(r)
	if err != nil {
		return nil, err
	}
	maxPlayers, err := readByte(r)
	if err != nil {
		return nil, err
	}
	if _, err := readByte(r); err != nil {
		return nil, err
	} // Bots
	if _, err := readByte(r); err != nil {
		return nil, err
	} // Type
	if _, err := readByte(r); err != nil {
		return nil, err
	} // OS
	if _, err := readByte(r); err != nil {
		return nil, err
	} // Visibility
	if _, err := readByte(r); err != nil {
		return nil, err
	} // VAC
	version, err := readString(r)
	if err != nil {
		return nil, err
	}

	edition := "Source"
	switch {
	case folder == "dayz" || game == "DayZ":
		edition = "DayZ"
	case folder == "unturned":
		edition = "Unturned"
	case folder == "valheim":
		edition = "Valheim"
	case folder == "ark_survival_evolved" || game == "ARK: Survival Evolved":
		edition = "ARK"
	}

	resp := &models.Response{
		Online:     true,
		Host:       host,
		Port:       port,
		MOTD:       name,
		Map:        mapName,
		PlayersOn:  int(players),
		PlayersMax: int(maxPlayers),
		Software:   fmt.Sprintf("%s (%s)", game, folder),
		Version:    version,
		Edition:    edition,
	}
	resp.SetLatency(latency)
	return resp, nil
}

func readByte(r *bytes.Reader) (byte, error) {
	return r.ReadByte()
}

func readUint16(r *bytes.Reader) (uint16, error) {
	var v uint16
	err := binary.Read(r, binary.LittleEndian, &v)
	return v, err
}

func readString(r *bytes.Reader) (string, error) {
	var b []byte
	for {
		c, err := r.ReadByte()
		if err != nil {
			return "", err
		}
		if c == 0x00 {
			break
		}
		b = append(b, c)
	}
	return string(b), nil
}
