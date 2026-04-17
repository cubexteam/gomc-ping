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

	if resp, err := tryTShock(host, int(port), port, cfg.Timeout); err == nil {
		resp.SetLatency(time.Since(start))
		return resp, nil
	}

	// Only if the provided port isn't 7878, try the default TShock port
	if port != 7878 {
		if resp, err := tryTShock(host, 7878, port, cfg.Timeout); err == nil {
			resp.SetLatency(time.Since(start))
			return resp, nil
		}
	}

	// Fallback: raw TCP — server is up but no TShock REST API
	if cfg.TerrariaFallback {
		addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
		conn, err := net.DialTimeout("tcp", addr, cfg.Timeout)
		if err != nil {
			return nil, err
		}
		defer conn.Close()

		resp := &models.Response{
			Online:  true,
			Host:    host,
			Port:    port,
			MOTD:    "Terraria Server",
			Edition: "Terraria (Unverified)",
		}
		resp.SetLatency(time.Since(start))
		return resp, nil
	}

	return nil, fmt.Errorf("terraria api unreachable on %s:%d", host, port)
}

// tryTShock attempts a TShock REST API call on apiPort, but always records
// the user-supplied port in the returned Response.
func tryTShock(host string, apiPort int, originalPort uint16, timeout time.Duration) (*models.Response, error) {
	client := &http.Client{Timeout: timeout}
	url := fmt.Sprintf("http://%s:%d/v2/server/status", host, apiPort)

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	var status TShockStatus
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1024*1024)).Decode(&status); err != nil {
		return nil, err
	}

	return &models.Response{
		Online:     true,
		Host:       host,
		Port:       originalPort, // always use the port the caller requested
		MOTD:       status.Name,
		World:      status.World,
		PlayersOn:  len(status.Players),
		PlayersMax: status.MaxPlayers,
		Version:    status.Version,
		Edition:    "Terraria (TShock)",
	}, nil
}
