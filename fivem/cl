package fivem

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cubexteam/gomc-ping/models"
)

type InfoResponse struct {
	Server string `json:"server"`
	Vars   struct {
		Hostname   string `json:"sv_hostname"`
		MaxClients string `json:"sv_maxClients"`
		MapName    string `json:"mapname"`
		Gametype   string `json:"gametype"`
	} `json:"vars"`
	Version int `json:"version"`
}

func Ping(host string, port uint16, timeout time.Duration) (*models.Response, error) {
	start := time.Now()
	client := &http.Client{Timeout: timeout}

	// Fetch basic info
	infoUrl := fmt.Sprintf("http://%s:%d/info.json", host, port)
	resp, err := client.Get(infoUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var info InfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	// Fetch players count
	playersUrl := fmt.Sprintf("http://%s:%d/players.json", host, port)
	pResp, err := client.Get(playersUrl)
	var playersCount int
	if err == nil {
		defer pResp.Body.Close()
		var players []interface{}
		_ = json.NewDecoder(pResp.Body).Decode(&players)
		playersCount = len(players)
	}

	var maxPlayers int
	fmt.Sscanf(info.Vars.MaxClients, "%d", &maxPlayers)

	return &models.Response{
		Online:     true,
		Host:       host,
		Port:       port,
		MOTD:       models.CleanMOTD(info.Vars.Hostname),
		PlayersOn:  playersCount,
		PlayersMax: maxPlayers,
		Map:        info.Vars.MapName,
		Software:   info.Server,
		Version:    fmt.Sprintf("v%d", info.Version),
		Latency:    time.Since(start),
		Edition:    "FiveM",
	}, nil
}
