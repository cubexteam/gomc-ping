package java

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/cubexteam/gomc-ping/models"
)

// Query retrieves server status using the GameSpy4 Query protocol
func Query(host string, port uint16, timeout time.Duration) (*models.Response, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("udp", addr, timeout)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(timeout))

	// Get Challenge Token
	sessionID := int32(0x01010101)
	buf := new(bytes.Buffer)
	buf.Write([]byte{0xFE, 0xFD, 0x09})
	binary.Write(buf, binary.BigEndian, sessionID)

	if _, err := conn.Write(buf.Bytes()); err != nil {
		return nil, err
	}

	resp := make([]byte, 2048)
	n, err := conn.Read(resp)
	if err != nil {
		return nil, err
	}
	if n < 5 {
		return nil, fmt.Errorf("response too short")
	}

	challengeToken, _ := strconv.Atoi(string(resp[5 : n-1]))

	// Request Full Stat
	buf.Reset()
	buf.Write([]byte{0xFE, 0xFD, 0x00})
	binary.Write(buf, binary.BigEndian, sessionID)
	binary.Write(buf, binary.BigEndian, int32(challengeToken))
	buf.Write([]byte{0xFF, 0xFF, 0xFF, 0x01})

	if _, err := conn.Write(buf.Bytes()); err != nil {
		return nil, err
	}

	n, err = conn.Read(resp)
	if err != nil {
		return nil, err
	}

	// Extract data from response
	data := resp[11:n]
	parts := bytes.Split(data, []byte{0x00})

	kv := make(map[string]string)
	var players []string

	isKV := true
	for i := 0; i < len(parts); i++ {
		if len(parts[i]) == 0 {
			if i+1 < len(parts) && len(parts[i+1]) == 0 {
				isKV = false
				i += 2
				continue
			}
		}
		if isKV {
			key := string(parts[i])
			if i+1 < len(parts) {
				kv[key] = string(parts[i+1])
				i++
			}
		} else {
			if name := string(parts[i]); name != "" {
				players = append(players, name)
			}
		}
	}

	resPlayers := make([]models.Player, len(players))
	for i, p := range players {
		resPlayers[i] = models.Player{Name: p}
	}

	return &models.Response{
		Online:     true,
		MOTD:       kv["hostname"],
		Version:    kv["version"],
		Software:   kv["server_mod"], // GameSpy4 standard field for server core
		PlayersMax: queryToInt(kv["maxplayers"]),
		PlayersOn:  queryToInt(kv["numplayers"]),
		Map:        kv["map"],
		Plugins:    strings.Split(kv["plugins"], ";"),
		Sample:     resPlayers,
		Edition:    "Query",
	}, nil
}

func queryToInt(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}
