package java

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/cubexteam/gomc-ping/models"
	"github.com/cubexteam/gomc-ping/protocol"
)

func Ping(host string, port uint16, handshakeHost string, config *models.Config) (*models.Response, error) {
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	dialer := net.Dialer{Timeout: config.Timeout}
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(config.Timeout))

	protocolVer := config.JavaProtocol
	if protocolVer == 0 {
		protocolVer = 47
	}

	// Build both packets into a single buffer to avoid append aliasing
	// and reduce the number of syscalls.
	hpb := protocol.NewPacketBuffer()
	hpb.WriteVarInt(0x00)
	hpb.WriteVarInt(protocolVer)
	hpb.WriteString(handshakeHost)
	hpb.WriteUint16(port)
	hpb.WriteVarInt(1)

	spb := protocol.NewPacketBuffer()
	spb.WriteVarInt(0x00)

	var burst bytes.Buffer
	burst.Write(hpb.Build())
	burst.Write(spb.Build())

	start := time.Now()
	if _, err := conn.Write(burst.Bytes()); err != nil {
		return nil, err
	}

	length, err := protocol.ReadVarIntFromIO(conn)
	if err != nil {
		return nil, fmt.Errorf("read length: %v", err)
	}

	body := make([]byte, length)
	if _, err := io.ReadFull(conn, body); err != nil {
		return nil, fmt.Errorf("read body: %v", err)
	}
	latency := time.Since(start)

	pr := protocol.NewPacketReader(body)
	pid, _ := pr.ReadVarInt()
	if pid != 0x00 {
		return nil, fmt.Errorf("invalid pid: %d", pid)
	}

	jsonStr, err := pr.ReadString()
	if err != nil {
		return nil, fmt.Errorf("read json: %v", err)
	}

	var status StatusResponse
	if err := json.Unmarshal([]byte(jsonStr), &status); err != nil {
		return nil, fmt.Errorf("unmarshal: %v", err)
	}

	resp := &models.Response{
		Online:     true,
		Host:       handshakeHost,
		Port:       port,
		MOTD:       status.ExtractMOTD(),
		PlayersMax: status.Players.Max,
		PlayersOn:  status.Players.Online,
		Sample:     status.GetSample(),
		Favicon:    status.Favicon,
		Version:    status.Version.Name,
		Protocol:   status.Version.Protocol,
		Edition:    "Java",
	}
	resp.SetLatency(latency)
	return resp, nil
}
