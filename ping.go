package gomcping

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cubexteam/gomc-ping/bedrock"
	"github.com/cubexteam/gomc-ping/cache"
	"github.com/cubexteam/gomc-ping/fivem"
	"github.com/cubexteam/gomc-ping/java"
	"github.com/cubexteam/gomc-ping/models"
	"github.com/cubexteam/gomc-ping/samp"
	"github.com/cubexteam/gomc-ping/source"
	"github.com/cubexteam/gomc-ping/terraria"
	"github.com/cubexteam/gomc-ping/utils"
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
		EnableFiveM:  false,
		EnableSAMP:   false,
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

	// Minecraft Java — enrich with Query asynchronously so it doesn't delay result
	wg.Add(1)
	go func() {
		defer wg.Done()
		resp, err := java.Ping(targetHost, targetPort, host, cfg)
		if err != nil {
			return
		}
		resp.Host = host
		resp.MOTD = models.CleanMOTD(resp.MOTD)
		sendResult(ctx, resultChan, resp)
		go enrichJava(resp, targetHost, targetPort, cfg)
	}()

	// Minecraft Bedrock
	wg.Add(1)
	go func() {
		defer wg.Done()
		resp, err := bedrock.Ping(targetHost, targetPort, cfg)
		if err != nil {
			return
		}
		resp.Host = host
		resp.MOTD = models.CleanMOTD(resp.MOTD)
		sendResult(ctx, resultChan, resp)
	}()

	// Source Engine (Rust, CS2, DayZ, ARK, Valheim, Unturned)
	wg.Add(1)
	go func() {
		defer wg.Done()
		resp, err := source.Ping(targetHost, targetPort, cfg.Timeout)
		if err != nil {
			return
		}
		resp.Host = host
		sendResult(ctx, resultChan, resp)
	}()

	// Terraria
	wg.Add(1)
	go func() {
		defer wg.Done()
		resp, err := terraria.Ping(targetHost, targetPort, cfg)
		if err != nil {
			return
		}
		resp.Host = host
		sendResult(ctx, resultChan, resp)
	}()

	// FiveM (GTA V) — Only if enabled or standard port
	if cfg.EnableFiveM || targetPort == 30120 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := fivem.Ping(targetHost, targetPort, cfg.Timeout)
			if err != nil {
				return
			}
			resp.Host = host
			sendResult(ctx, resultChan, resp)
		}()
	}

	// SA-MP (GTA SA) — Only if enabled or standard port
	if cfg.EnableSAMP || targetPort == 7777 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := samp.Ping(targetHost, targetPort, cfg.Timeout)
			if err != nil {
				return
			}
			resp.Host = host
			sendResult(ctx, resultChan, resp)
		}()
	}

	// Close channel when all goroutines finish
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	select {
	case res, ok := <-resultChan:
		if ok && res != nil {
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
		return nil, fmt.Errorf("server %s:%d unreachable", host, port)
	}
	return nil, fmt.Errorf("server %s:%d unreachable", host, port)
}

func sendResult(ctx context.Context, ch chan *models.Response, res *models.Response) {
	select {
	case ch <- res:
	case <-ctx.Done():
	}
}

// enrichJava runs a GameSpy4 Query in the background and fills in extra fields.
// It mutates the already-returned Response, so callers see updated data on next access.
func enrichJava(resp *models.Response, tHost string, tPort uint16, cfg *models.Config) {
	qResp, err := java.Query(tHost, tPort, cfg.Timeout)
	if err != nil {
		return
	}
	if qResp.Software != "" {
		resp.Software = qResp.Software
	}
	if len(qResp.Plugins) > 0 {
		resp.Plugins = qResp.Plugins
	}
	if qResp.Map != "" {
		resp.Map = qResp.Map
	}
}

// pingWithCache wraps a typed ping call with GlobalCache.
// Key includes edition so PingRust / PingCS2 etc. have isolated entries.
func pingWithCache(host string, port uint16, edition string, fn func() (*models.Response, error)) (*models.Response, error) {
	cacheKey := fmt.Sprintf("%s:%d:%s", host, port, edition)
	if resp, ok := GlobalCache.Get(cacheKey); ok {
		return resp, nil
	}
	resp, err := fn()
	if err != nil {
		return nil, err
	}
	GlobalCache.Set(cacheKey, resp)
	return resp, nil
}

func PingRust(host string, port uint16) (*models.Response, error) {
	return pingWithCache(host, port, "Rust", func() (*models.Response, error) {
		resp, err := source.Ping(host, port, DefaultTimeout)
		if err == nil {
			resp.Edition = "Rust"
		}
		return resp, err
	})
}

func PingCS2(host string, port uint16) (*models.Response, error) {
	return pingWithCache(host, port, "CS2", func() (*models.Response, error) {
		resp, err := source.Ping(host, port, DefaultTimeout)
		if err == nil {
			resp.Edition = "CS2"
		}
		return resp, err
	})
}

func PingDayZ(host string, port uint16) (*models.Response, error) {
	return pingWithCache(host, port, "DayZ", func() (*models.Response, error) {
		resp, err := source.Ping(host, port, DefaultTimeout)
		if err == nil {
			resp.Edition = "DayZ"
		}
		return resp, err
	})
}

func PingARK(host string, port uint16) (*models.Response, error) {
	return pingWithCache(host, port, "ARK", func() (*models.Response, error) {
		resp, err := source.Ping(host, port, DefaultTimeout)
		if err == nil {
			resp.Edition = "ARK"
		}
		return resp, err
	})
}

func PingValheim(host string, port uint16) (*models.Response, error) {
	return pingWithCache(host, port, "Valheim", func() (*models.Response, error) {
		resp, err := source.Ping(host, port, DefaultTimeout)
		if err == nil {
			resp.Edition = "Valheim"
		}
		return resp, err
	})
}

func PingUnturned(host string, port uint16) (*models.Response, error) {
	return pingWithCache(host, port, "Unturned", func() (*models.Response, error) {
		resp, err := source.Ping(host, port, DefaultTimeout)
		if err == nil {
			resp.Edition = "Unturned"
		}
		return resp, err
	})
}

func PingTerraria(host string, port uint16) (*models.Response, error) {
	return pingWithCache(host, port, "Terraria", func() (*models.Response, error) {
		return terraria.Ping(host, port, NewConfig())
	})
}

func PingFiveM(host string, port uint16) (*models.Response, error) {
	return pingWithCache(host, port, "FiveM", func() (*models.Response, error) {
		return fivem.Ping(host, port, DefaultTimeout)
	})
}

func PingSAMP(host string, port uint16) (*models.Response, error) {
	return pingWithCache(host, port, "SAMP", func() (*models.Response, error) {
		return samp.Ping(host, port, DefaultTimeout)
	})
}

func SaveFavicon(data string, path string) error {
	return utils.SaveFavicon(data, path)
}

func Close() {
	GlobalCache.Close()
}
