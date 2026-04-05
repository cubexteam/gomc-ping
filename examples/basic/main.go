package main

import (
	"fmt"
	"github.com/cubexteam/gomc-ping"
)

func main() {
	fmt.Println("--- gomc-ping Multi-Game Example ---")

	// Universal Auto-Detection (works for any supported game)
	mcHost := "play.hypixel.net"
	resp, err := gomcping.Ping(mcHost, 25565)
	if err == nil {
		fmt.Printf("[Auto] Found %s: %s (Players: %d/%d)\n", resp.Edition, mcHost, resp.PlayersOn, resp.PlayersMax)
	}

	// Explicit Game Ping
	// Example for Rust (change IP to real one)
	rustResp, err := gomcping.PingRust("1.2.3.4", 28015)
	if err != nil {
		fmt.Println("[Rust] Server offline or IP invalid")
	} else {
		fmt.Printf("[Rust] %s: %d players\n", rustResp.MOTD, rustResp.PlayersOn)
	}

	// Saving Minecraft Favicon
	if resp != nil && resp.Favicon != "" {
		_ = gomcping.SaveFavicon(resp.Favicon, "favicon.png")
		fmt.Println("[System] Minecraft favicon saved to favicon.png")
	}
}
