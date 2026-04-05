package java

import (
	"encoding/json"
	"strings"
	"github.com/cubexteam/gomc-ping/models"
)

type StatusResponse struct {
	Version struct {
		Name     string `json:"name"`
		Protocol int    `json:"protocol"`
	} `json:"version"`
	Players struct {
		Max    int `json:"max"`
		Online int `json:"online"`
		Sample []struct {
			Name string `json:"name"`
			ID   string `json:"id"`
		} `json:"sample"`
	} `json:"players"`
	Description interface{} `json:"description"`
	Favicon     string      `json:"favicon"`
}

func (s *StatusResponse) ExtractMOTD() string {
	return parseDescription(s.Description)
}

func (s *StatusResponse) GetSample() []models.Player {
	var sample []models.Player
	for _, p := range s.Players.Sample {
		sample = append(sample, models.Player{
			Name: p.Name,
			ID:   p.ID,
		})
	}
	return sample
}

func parseDescription(desc interface{}) string {
	switch v := desc.(type) {
	case string:
		return v
	case map[string]interface{}:
		var b strings.Builder
		if text, ok := v["text"].(string); ok {
			b.WriteString(text)
		}
		if extra, ok := v["extra"].([]interface{}); ok {
			for _, item := range extra {
				b.WriteString(parseDescription(item))
			}
		}
		return b.String()
	default:
		return ""
	}
}
