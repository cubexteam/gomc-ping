package main

import (
	"context"
	"fmt"
	"time"

	gomcping "github.com/cubexteam/gomc-ping"
)

func main() {
	// Single server with builder config
	cfg := gomcping.NewConfig().
		WithTimeout(3 * time.Second).
		WithRetries(1, 300*time.Millisecond)

	fmt.Println("Pinging mc.hypixel.net...")
	resp, err := gomcping.PingWithConfig("mc.hypixel.net", 25565, cfg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println(resp.String())
		if resp.Favicon != "" {
			_ = gomcping.SaveFavicon(resp.Favicon, "hypixel_favicon.png")
			fmt.Println("Favicon saved to hypixel_favicon.png")
		}
	}

	// Context-aware ping (e.g. HTTP handler with deadline)
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	resp2, err := gomcping.PingWithContext(ctx, "play.mineplex.com", 25565)
	if err != nil {
		fmt.Printf("Mineplex error: %v\n", err)
	} else {
		fmt.Println(resp2.String())
	}

	// Batch ping
	targets := []gomcping.Target{
		{Host: "mc.hypixel.net", Port: 25565},
		{Host: "play.mineplex.com", Port: 25565},
		{Host: "mc.cubecraft.net", Port: 25565},
	}

	fmt.Println("\nBatch ping (concurrency=2):")
	results := gomcping.PingAll(targets, 2)
	for _, r := range results {
		if r.Err != nil {
			fmt.Printf("  %s:%d — error: %v\n", r.Target.Host, r.Target.Port, r.Err)
		} else {
			fmt.Printf("  %s:%d — %s %d/%d players\n",
				r.Target.Host, r.Target.Port,
				r.Resp.Edition, r.Resp.PlayersOn, r.Resp.PlayersMax)
		}
	}
}
