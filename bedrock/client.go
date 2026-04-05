package bedrock

import (
	"fmt"
	"net"
	"time"

	"github.com/cubexteam/gomc-ping/models"
)

func Ping(host string, port uint16, config *models.Config) (*models.Response, error) {
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := net.DialTimeout("udp", addr, config.Timeout)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(config.Timeout))

	packet := []byte{
		0x01,                   // Unconnected Ping
		0x00, 0x00, 0x00, 0x00, // Time
		0x00, 0x00, 0x00, 0x00,
		0x00, 0xFF, 0xFF, 0x00, // Magic
		0xFE, 0xFE, 0xFE, 0xFE,
		0xFD, 0xFD, 0xFD, 0xFD,
		0x12, 0x34, 0x56, 0x78,
		0x00, 0x00, 0x00, 0x00, // Client ID
		0x00, 0x00, 0x00, 0x00,
	}

	start := time.Now()
	if _, err := conn.Write(packet); err != nil {
		return nil, err
	}

	resp := make([]byte, 2048)
	n, err := conn.Read(resp)
	if err != nil {
		return nil, err
	}
	latency := time.Since(start)

	motd, online, max, version, edition, proto, err := ParseResponse(resp[:n])
	if err != nil {
		return nil, err
	}

	return &models.Response{
		Online:     true,
		Host:       host,
		Port:       port,
		MOTD:       motd,
		PlayersOn:  online,
		PlayersMax: max,
		Version:    version,
		Edition:    edition,
		Protocol:   proto,
		Latency:    latency,
	}, nil
}
