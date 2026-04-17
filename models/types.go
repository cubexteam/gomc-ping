package models

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// Response holds the result of a server ping.
// All fields are safe to read after the Ping call returns.
// Fields enriched asynchronously (Software, Plugins, Map for Java) are
// protected by mu — use the accessor methods if you read them concurrently.
type Response struct {
	mu sync.RWMutex

	Online     bool     `json:"online"`
	Host       string   `json:"host"`
	Port       uint16   `json:"port"`
	MOTD       string   `json:"motd"`
	PlayersMax int      `json:"players_max"`
	PlayersOn  int      `json:"players_on"`
	Version    string   `json:"version"`
	Protocol   int      `json:"protocol"`
	LatencyMs  int64    `json:"latency_ms"`
	Latency    time.Duration `json:"-"`
	Edition    string   `json:"edition"`
	Software   string   `json:"software,omitempty"`
	Favicon    string   `json:"favicon,omitempty"`
	Sample     []Player `json:"sample,omitempty"`
	Map        string   `json:"map,omitempty"`
	World      string   `json:"world,omitempty"`
	Plugins    []string `json:"plugins,omitempty"`
	Password   bool     `json:"password,omitempty"`
}

// SetLatency sets both Latency and the derived LatencyMs field.
func (r *Response) SetLatency(d time.Duration) {
	r.Latency = d
	r.LatencyMs = d.Milliseconds()
}

// Enrich safely overwrites enrichment fields that arrive asynchronously
// (Java Query: Software, Plugins, Map).
func (r *Response) Enrich(software string, plugins []string, mapName string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if software != "" {
		r.Software = software
	}
	if len(plugins) > 0 {
		r.Plugins = plugins
	}
	if mapName != "" {
		r.Map = mapName
	}
}

// GetSoftware returns Software safely.
func (r *Response) GetSoftware() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Software
}

// GetPlugins returns a copy of Plugins safely.
func (r *Response) GetPlugins() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, len(r.Plugins))
	copy(out, r.Plugins)
	return out
}

// GetMap returns Map safely.
func (r *Response) GetMap() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Map
}

type Player struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

// Config controls the behaviour of a ping call.
type Config struct {
	Timeout          time.Duration
	SRV              bool
	JavaProtocol     int
	DisableCache     bool
	TerrariaFallback bool
	EnableFiveM      bool
	EnableSAMP       bool
	Retries          int
	RetryDelay       time.Duration
}

// WithTimeout returns a copy of cfg with Timeout set.
func (c *Config) WithTimeout(d time.Duration) *Config {
	cp := *c
	cp.Timeout = d
	return &cp
}

// WithoutCache returns a copy of cfg with caching disabled.
func (c *Config) WithoutCache() *Config {
	cp := *c
	cp.DisableCache = true
	return &cp
}

// WithRetries returns a copy of cfg with retry settings.
func (c *Config) WithRetries(n int, delay time.Duration) *Config {
	cp := *c
	cp.Retries = n
	cp.RetryDelay = delay
	return &cp
}

// WithSRV returns a copy of cfg with SRV resolution toggled.
func (c *Config) WithSRV(enabled bool) *Config {
	cp := *c
	cp.SRV = enabled
	return &cp
}

var (
	ErrConnectionFailed = fmt.Errorf("failed to connect to server")
	ErrInvalidResponse  = fmt.Errorf("invalid response from server")
	ErrTimeout          = fmt.Errorf("connection timeout")
)

func (r *Response) String() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.Online {
		return fmt.Sprintf("❌ Server %s:%d is offline", r.Host, r.Port)
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("🎮 [%s] %s (%s)\n", r.Edition, r.MOTD, r.Version))
	b.WriteString(fmt.Sprintf("👥 Players: %d/%d\n", r.PlayersOn, r.PlayersMax))
	b.WriteString(fmt.Sprintf("📡 Latency: %dms\n", r.LatencyMs))
	if r.World != "" {
		b.WriteString(fmt.Sprintf("🌍 World: %s\n", r.World))
	}
	if r.Map != "" {
		b.WriteString(fmt.Sprintf("🗺  Map: %s\n", r.Map))
	}
	if r.Software != "" {
		b.WriteString(fmt.Sprintf("🖥  Software: %s\n", r.Software))
	}
	if len(r.Plugins) > 0 {
		b.WriteString(fmt.Sprintf("🔌 Plugins (%d): %s\n", len(r.Plugins), strings.Join(r.Plugins, ", ")))
	}
	return b.String()
}

func (r *Response) JSON() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	b, _ := json.Marshal(r)
	return string(b)
}

// CleanMOTD strips ANSI escape codes, Minecraft § and & color codes,
// and trims the resulting string.
func CleanMOTD(motd string) string {
	motd = ansiRegex.ReplaceAllString(motd, "")

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
	return strings.TrimSpace(b.String())
}
