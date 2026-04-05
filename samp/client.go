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

	ip := net.ParseIP(host).To4()
	if ip == nil {
		ips, err := net.LookupIP(host)
		if err != nil || len(ips) == 0 {
			return nil, fmt.Errorf("failed to resolve host")
		}
		ip = ips[0].To4()
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

	password, _ := r.ReadByte()
	players, _ := readUint16(r)
	maxPlayers, _ := readUint16(r)
	hostname, _ := readString(r)
	gamemode, _ := readString(r)
	mapName, _ := readString(r)

	return &models.Response{
		Online:     true,
		Host:       host,
		Port:       port,
		MOTD:       hostname,
		PlayersOn:  int(players),
		PlayersMax: int(maxPlayers),
		Map:        mapName,
		Software:   gamemode,
		Latency:    latency,
		Edition:    "SA-MP",
		Password:   password == 1,
	}, nil
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
