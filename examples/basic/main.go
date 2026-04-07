package main

import (
	"fmt"
	"github.com/cubexteam/gomc-ping"
	"github.com/cubexteam/gomc-ping/models"
)

func main() {
	host := "mc.hypixel.net"
	port := uint16(25565)

	cfg := &models.Config{
		Timeout: 5 * 1000 * 1000 * 1000,
		SRV:     true,
	}

	fmt.Printf("Pinging %s...\n", host)

	resp, err := gomcping.PingWithConfig(host, port, cfg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println(resp.String())

	if resp.Favicon != "" {
		_ = gomcping.SaveFavicon(resp.Favicon, "hypixel_favicon.png")
		fmt.Println("Favicon saved to hypixel_favicon.png")
	}
}
