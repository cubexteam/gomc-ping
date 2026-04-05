package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Response struct {
	Online     bool          `json:"online"`
	Host       string        `json:"host"`
	Port       uint16        `json:"port"`
	MOTD       string        `json:"motd"`
	PlayersMax int           `json:"players_max"`
	PlayersOn  int           `json:"players_on"`
	Version    string        `json:"version"`
	Protocol   int           `json:"protocol"`
	Latency    time.Duration `json:"latency"`
	Edition    string        `json:"edition"`
	Software   string        `json:"software,omitempty"`
	Favicon    string        `json:"favicon,omitempty"`
	Sample     []Player      `json:"sample,omitempty"`
	Map        string        `json:"map,omitempty"`
	World      string        `json:"world,omitempty"`
	Plugins    []string      `json:"plugins,omitempty"`
	Password   bool          `json:"password,omitempty"`
}

type Player struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type Config struct {
	Timeout          time.Duration
	SRV              bool
	JavaProtocol     int
	DisableCache     bool
	TerrariaFallback bool
	EnableFiveM      bool
	EnableSAMP       bool
}

var (
	ErrConnectionFailed = fmt.Errorf("failed to connect to server")
	ErrInvalidResponse  = fmt.Errorf("invalid response from server")
	ErrTimeout          = fmt.Errorf("connection timeout")
)

func (r *Response) String() string {
	if !r.Online {
		return fmt.Sprintf("❌ Server %s:%d is offline", r.Host, r.Port)
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("🎮 [%s] %s (%s)\n", r.Edition, r.MOTD, r.Version))
	b.WriteString(fmt.Sprintf("👥 Players: %d/%d\n", r.PlayersOn, r.PlayersMax))
	b.WriteString(fmt.Sprintf("📡 Latency: %v\n", r.Latency))
	if r.World != "" {
		b.WriteString(fmt.Sprintf("🌍 World: %s\n", r.World))
	}
	return b.String()
}

func (r *Response) JSON() string {
	b, _ := json.Marshal(r)
	return string(b)
}

func CleanMOTD(motd string) string {
	var b strings.Builder
	runes := []rune(motd)
	for i := 0; i < len(runes); i++ {
		if runes[i] == '§' {
			i++
			continue
		}
		if runes[i] == '&' && i+1 < len(runes) {
			next := runes[i+1]
			if (next >= '0' && next <= '9') || (next >= 'a' && next <= 'f') ||
			   (next >= 'k' && next <= 'o') || next == 'r' {
				i++
				continue
			}
		}
		b.WriteRune(runes[i])
	}
	return b.String()
}
