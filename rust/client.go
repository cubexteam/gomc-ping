package rust

import (
	"time"

	"github.com/cubexteam/gomc-ping/models"
	"github.com/cubexteam/gomc-ping/source"
)

func Ping(host string, port uint16, timeout time.Duration) (*models.Response, error) {
	resp, err := source.Ping(host, port, timeout)
	if err != nil {
		return nil, err
	}
	resp.Edition = "Rust"
	return resp, nil
}
