package samp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"github.com/cubexteam/gomc-ping/models"
)

func Ping(host string, port uint16, timeout time.Duration) (*models.Response, error) {
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := net.DialTimeout("udp", addr, timeout)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(timeout))

	remoteAddr := conn.RemoteAddr().(*net.UDPAddr)
	ip := remoteAddr.IP.To4()
	if ip == nil {
		return nil, fmt.Errorf("SA-MP requires IPv4, but got %s", remoteAddr.IP)
	}

	var buf bytes.Buffer
	buf.WriteString("SAMP")
	buf.Write(ip)
	binary.Write(&buf, binary.LittleEndian, port)
	buf.WriteByte('i')

	start := time.Now()
	if _, err := conn.Write(buf.Bytes()); err != nil {
		return nil, err
	}

	resp := make([]byte, 1024)
	n, err := conn.Read(resp)
	if err != nil {
		return nil, err
	}
	latency := time.Since(start)

	if n < 11 || !bytes.Equal(resp[:4], []byte("SAMP")) {
		return nil, fmt.Errorf("invalid samp response")
	}

	r := bytes.NewReader(resp[11:])

	password, err := r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("samp: read password flag: %w", err)
	}
	players, err := readUint16(r)
	if err != nil {
		return nil, fmt.Errorf("samp: read players: %w", err)
	}
	maxPlayers, err := readUint16(r)
	if err != nil {
		return nil, fmt.Errorf("samp: read max players: %w", err)
	}
	hostname, err := readString(r)
	if err != nil {
		return nil, fmt.Errorf("samp: read hostname: %w", err)
	}
	gamemode, err := readString(r)
	if err != nil {
		return nil, fmt.Errorf("samp: read gamemode: %w", err)
	}
	mapName, err := readString(r)
	if err != nil {
		return nil, fmt.Errorf("samp: read map: %w", err)
	}

	result := &models.Response{
		Online:     true,
		Host:       host,
		Port:       port,
		MOTD:       hostname,
		PlayersOn:  int(players),
		PlayersMax: int(maxPlayers),
		Map:        mapName,
		Software:   gamemode,
		Edition:    "SA-MP",
		Password:   password == 1,
	}
	result.SetLatency(latency)
	return result, nil
}

func readUint16(r *bytes.Reader) (uint16, error) {
	var v uint16
	err := binary.Read(r, binary.LittleEndian, &v)
	return v, err
}

func readString(r *bytes.Reader) (string, error) {
	var length uint32
	if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
		return "", err
	}
	b := make([]byte, length)
	if _, err := r.Read(b); err != nil {
		return "", err
	}
	return string(b), nil
}
