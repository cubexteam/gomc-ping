package terraria

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/cubexteam/gomc-ping/models"
)

type TShockStatus struct {
	Name       string   `json:"name"`
	Port       int      `json:"port"`
	Players    []string `json:"players"`
	MaxPlayers int      `json:"maxplayers"`
	World      string   `json:"world"`
	Version    string   `json:"version"`
}

func Ping(host string, port uint16, cfg *models.Config) (*models.Response, error) {
	start := time.Now()

	// Try TShock REST API on the provided port first
	if resp, err := tryTShock(host, int(port), cfg.Timeout); err == nil {
		resp.Latency = time.Since(start)
		return resp, nil
	}

	// Only if the provided port isn't 7878, try the default TShock port
	if port != 7878 {
		if resp, err := tryTShock(host, 7878, cfg.Timeout); err == nil {
			resp.Latency = time.Since(start)
			return resp, nil
		}
	}

	// Fallback: Only if enabled in config
	if cfg.TerrariaFallback {
		addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
		conn, err := net.DialTimeout("tcp", addr, cfg.Timeout)
		if err != nil {
			return nil, err
		}
		defer conn.Close()

		return &models.Response{
			Online:  true,
			Host:    host,
			Port:    port,
			MOTD:    "Terraria Server",
			Edition: "Terraria (Unverified)",
			Latency: time.Since(start),
		}, nil
	}

	return nil, fmt.Errorf("terraria api unreachable on %s:%d", host, port)
}

func tryTShock(host string, port int, timeout time.Duration) (*models.Response, error) {
	client := &http.Client{Timeout: timeout}
	url := fmt.Sprintf("http://%s:%d/v2/server/status", host, port)

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	var status TShockStatus
	limitReader := io.LimitReader(resp.Body, 1024*1024) // 1MB limit
	if err := json.NewDecoder(limitReader).Decode(&status); err != nil {
		return nil, err
	}

	return &models.Response{
		Online:     true,
		Host:       host,
		Port:       uint16(status.Port),
		MOTD:       status.Name,
		World:      status.World,
		PlayersOn:  len(status.Players),
		PlayersMax: status.MaxPlayers,
		Version:    status.Version,
		Edition:    "Terraria (TShock)",
	}, nil
}
