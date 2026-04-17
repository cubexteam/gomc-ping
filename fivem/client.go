package fivem

import (
	"encoding/json"
	"fmt"
	"io"
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

	infoURL := fmt.Sprintf("http://%s:%d/info.json", host, port)
	infoResp, err := client.Get(infoURL)
	if err != nil {
		return nil, err
	}
	defer infoResp.Body.Close()

	if infoResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("info.json returned status %d", infoResp.StatusCode)
	}

	var info InfoResponse
	if err := json.NewDecoder(io.LimitReader(infoResp.Body, 1024*1024)).Decode(&info); err != nil {
		return nil, err
	}

	var playersCount int
	playersURL := fmt.Sprintf("http://%s:%d/players.json", host, port)
	pResp, err := client.Get(playersURL)
	if err == nil {
		defer pResp.Body.Close()
		if pResp.StatusCode == http.StatusOK {
			var players []interface{}
			if jsonErr := json.NewDecoder(io.LimitReader(pResp.Body, 4*1024*1024)).Decode(&players); jsonErr == nil {
				playersCount = len(players)
			}
		}
	}

	var maxPlayers int
	fmt.Sscanf(info.Vars.MaxClients, "%d", &maxPlayers)

	resp := &models.Response{
		Online:     true,
		Host:       host,
		Port:       port,
		MOTD:       models.CleanMOTD(info.Vars.Hostname),
		PlayersOn:  playersCount,
		PlayersMax: maxPlayers,
		Map:        info.Vars.MapName,
		Software:   info.Server,
		Version:    fmt.Sprintf("v%d", info.Version),
		Edition:    "FiveM",
	}
	resp.SetLatency(time.Since(start))
	return resp, nil
}
