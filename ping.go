package gomcping

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cubexteam/gomc-ping/java"
	"github.com/cubexteam/gomc-ping/bedrock"
	"github.com/cubexteam/gomc-ping/source"
	"github.com/cubexteam/gomc-ping/terraria"
	"github.com/cubexteam/gomc-ping/fivem"
	"github.com/cubexteam/gomc-ping/samp"
	"github.com/cubexteam/gomc-ping/utils"
	"github.com/cubexteam/gomc-ping/models"
	"github.com/cubexteam/gomc-ping/cache"
)

var (
	DefaultTimeout = 5 * time.Second
	GlobalCache    = cache.New(1*time.Minute, 5*time.Minute)
)

func NewConfig() *models.Config {
	return &models.Config{
		Timeout:      DefaultTimeout,
		SRV:          true,
		JavaProtocol: 47,
		EnableFiveM:  false, // Opt-in for parallel ping
		EnableSAMP:   false, // Opt-in for parallel ping
	}
}

func Ping(host string, port uint16) (*models.Response, error) {
	return PingWithConfig(host, port, NewConfig())
}

func PingWithConfig(host string, port uint16, cfg *models.Config) (*models.Response, error) {
	cacheKey := fmt.Sprintf("%s:%d", host, port)
	if !cfg.DisableCache {
		if resp, ok := GlobalCache.Get(cacheKey); ok {
			return resp, nil
		}
	}

	targetHost, targetPort := host, port
	if cfg.SRV {
		if srvHost, srvPort, err := utils.ResolveSRV(host); err == nil {
			targetHost, targetPort = srvHost, srvPort
		}
	}

	resultChan := make(chan *models.Response, 6)
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(6)

	// Minecraft Java
	go func() {
		defer wg.Done()
		if resp, err := java.Ping(targetHost, targetPort, targetHost, cfg); err == nil {
			resp.Host = host
			processResponse(resp, host, targetHost, targetPort, cfg)
			sendResult(ctx, resultChan, resp)
		}
	}()

	// Minecraft Bedrock
	go func() {
		defer wg.Done()
		if resp, err := bedrock.Ping(targetHost, targetPort, cfg); err == nil {
			resp.Host = host
			resp.MOTD = models.CleanMOTD(resp.MOTD)
			sendResult(ctx, resultChan, resp)
		}
	}()

	// Source Engine (Rust, CS2)
	go func() {
		defer wg.Done()
		if resp, err := source.Ping(targetHost, targetPort, cfg.Timeout); err == nil {
			resp.Host = host
			sendResult(ctx, resultChan, resp)
		}
	}()

	// Terraria
	go func() {
		defer wg.Done()
		if resp, err := terraria.Ping(targetHost, targetPort, cfg); err == nil {
			resp.Host = host
			sendResult(ctx, resultChan, resp)
		}
	}()

	// FiveM (GTA V) - Only if enabled or standard port
	go func() {
		defer wg.Done()
		if cfg.EnableFiveM || targetPort == 30120 {
			if resp, err := fivem.Ping(targetHost, targetPort, cfg.Timeout); err == nil {
				resp.Host = host
				sendResult(ctx, resultChan, resp)
			}
		}
	}()

	// SA-MP (GTA SA) - Only if enabled or standard port
	go func() {
		defer wg.Done()
		if cfg.EnableSAMP || targetPort == 7777 {
			if resp, err := samp.Ping(targetHost, targetPort, cfg.Timeout); err == nil {
				resp.Host = host
				sendResult(ctx, resultChan, resp)
			}
		}
	}()

	// Cleanup goroutine
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	select {
	case res := <-resultChan:
		if res != nil {
			cancel()
			if !cfg.DisableCache {
				GlobalCache.Set(cacheKey, res)
			}
			return res, nil
		}
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("server %s:%d unreachable (timeout)", host, port)
		}
	}

	return nil, fmt.Errorf("server %s:%d unreachable", host, port)
}

func sendResult(ctx context.Context, ch chan *models.Response, res *models.Response) {
	select {
	case ch <- res:
	case <-ctx.Done():
	}
}

func PingRust(host string, port uint16) (*models.Response, error) {
	resp, err := source.Ping(host, port, DefaultTimeout)
	if err == nil { resp.Edition = "Rust" }
	return resp, err
}

func PingCS2(host string, port uint16) (*models.Response, error) {
	resp, err := source.Ping(host, port, DefaultTimeout)
	if err == nil { resp.Edition = "CS2" }
	return resp, err
}

func PingTerraria(host string, port uint16) (*models.Response, error) {
	return terraria.Ping(host, port, NewConfig())
}

func PingFiveM(host string, port uint16) (*models.Response, error) {
	return fivem.Ping(host, port, DefaultTimeout)
}

func PingSAMP(host string, port uint16) (*models.Response, error) {
	return samp.Ping(host, port, DefaultTimeout)
}

func processResponse(resp *models.Response, host, tHost string, tPort uint16, cfg *models.Config) {
	resp.MOTD = models.CleanMOTD(resp.MOTD)
	if qResp, qErr := java.Query(tHost, tPort, cfg.Timeout); qErr == nil {
		if qResp.Software != "" { resp.Software = qResp.Software }
		if len(qResp.Plugins) > 0 { resp.Plugins = qResp.Plugins }
		if qResp.Map != "" { resp.Map = qResp.Map }
	}
}

func SaveFavicon(data string, path string) error {
	return utils.SaveFavicon(data, path)
}

func Close() {
	GlobalCache.Close()
}
