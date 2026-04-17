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
		Retries:      0,
		RetryDelay:   500 * time.Millisecond,
	}
}

// Target is a host/port pair used in PingAll.
type Target struct {
	Host string
	Port uint16
}

// PingResult pairs a Target with its result or error.
type PingResult struct {
	Target Target
	Resp   *models.Response
	Err    error
}

// Ping pings a server using auto-detection with default config.
func Ping(host string, port uint16) (*models.Response, error) {
	return PingWithConfig(host, port, NewConfig())
}

// PingWithContext is like Ping but honours an external context.
func PingWithContext(ctx context.Context, host string, port uint16) (*models.Response, error) {
	return pingCtx(ctx, host, port, NewConfig())
}

// PingWithConfig pings a server with a custom Config.
func PingWithConfig(host string, port uint16, cfg *models.Config) (*models.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()
	return pingCtx(ctx, host, port, cfg)
}

// PingWithConfigContext is like PingWithConfig but honours an external context.
func PingWithConfigContext(ctx context.Context, host string, port uint16, cfg *models.Config) (*models.Response, error) {
	// Wrap with timeout only if the config has one and ctx has no deadline yet.
	if _, ok := ctx.Deadline(); !ok && cfg.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.Timeout)
		defer cancel()
	}
	return pingCtx(ctx, host, port, cfg)
}

// PingAll pings multiple targets concurrently, with at most concurrency
// goroutines running at the same time. Pass concurrency=0 to run all at once.
func PingAll(targets []Target, concurrency int) []PingResult {
	results := make([]PingResult, len(targets))

	if concurrency <= 0 {
		concurrency = len(targets)
	}
	sem := make(chan struct{}, concurrency)

	var wg sync.WaitGroup
	for i, t := range targets {
		wg.Add(1)
		go func(idx int, target Target) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			resp, err := Ping(target.Host, target.Port)
			results[idx] = PingResult{Target: target, Resp: resp, Err: err}
		}(i, t)
	}
	wg.Wait()
	return results
}

// pingCtx is the core implementation shared by all Ping variants.
func pingCtx(ctx context.Context, host string, port uint16, cfg *models.Config) (*models.Response, error) {
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

	var (
		resp *models.Response
		err  error
	)

	attempts := 1 + cfg.Retries
	for i := 0; i < attempts; i++ {
		if i > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(cfg.RetryDelay):
			}
		}
		resp, err = runProbes(ctx, host, targetHost, targetPort, cfg)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, err
	}

	if !cfg.DisableCache {
		GlobalCache.Set(cacheKey, resp)
	}
	return resp, nil
}

// runProbes fires all protocol probes concurrently and returns the first success.
func runProbes(ctx context.Context, host, targetHost string, targetPort uint16, cfg *models.Config) (*models.Response, error) {
	resultChan := make(chan *models.Response, 6)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup

	launch := func(fn func()) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fn()
		}()
	}

	// Minecraft Java
	launch(func() {
		resp, e := java.Ping(targetHost, targetPort, host, cfg)
		if e != nil {
			return
		}
		resp.Host = host
		resp.MOTD = models.CleanMOTD(resp.MOTD)
		sendResult(ctx, resultChan, resp)
		// Enrich with Query data asynchronously — uses Enrich() which is mutex-safe.
		go enrichJava(resp, targetHost, targetPort, cfg)
	})

	// Minecraft Bedrock
	launch(func() {
		resp, e := bedrock.Ping(targetHost, targetPort, cfg)
		if e != nil {
			return
		}
		resp.Host = host
		resp.MOTD = models.CleanMOTD(resp.MOTD)
		sendResult(ctx, resultChan, resp)
	})

	// Source Engine
	launch(func() {
		resp, e := source.Ping(targetHost, targetPort, cfg.Timeout)
		if e != nil {
			return
		}
		resp.Host = host
		sendResult(ctx, resultChan, resp)
	})

	// Terraria
	launch(func() {
		resp, e := terraria.Ping(targetHost, targetPort, cfg)
		if e != nil {
			return
		}
		resp.Host = host
		sendResult(ctx, resultChan, resp)
	})

	// FiveM
	if cfg.EnableFiveM || targetPort == 30120 {
		launch(func() {
			resp, e := fivem.Ping(ctx, targetHost, targetPort, cfg.Timeout)
			if e != nil {
				return
			}
			resp.Host = host
			sendResult(ctx, resultChan, resp)
		})
	}

	// SA-MP
	if cfg.EnableSAMP || targetPort == 7777 {
		launch(func() {
			resp, e := samp.Ping(targetHost, targetPort, cfg.Timeout)
			if e != nil {
				return
			}
			resp.Host = host
			sendResult(ctx, resultChan, resp)
		})
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	select {
	case res, ok := <-resultChan:
		if ok && res != nil {
			cancel()
			return res, nil
		}
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("server %s:%d unreachable (timeout)", host, targetPort)
		}
		return nil, fmt.Errorf("server %s:%d unreachable", host, targetPort)
	}
	return nil, fmt.Errorf("server %s:%d unreachable", host, targetPort)
}

func sendResult(ctx context.Context, ch chan *models.Response, res *models.Response) {
	select {
	case ch <- res:
	case <-ctx.Done():
	}
}

// enrichJava fetches Query data and writes it via the mutex-safe Enrich method.
func enrichJava(resp *models.Response, tHost string, tPort uint16, cfg *models.Config) {
	qResp, err := java.Query(tHost, tPort, cfg.Timeout)
	if err != nil {
		return
	}
	resp.Enrich(qResp.Software, qResp.Plugins, qResp.Map)
}

// pingWithCache wraps a typed ping with GlobalCache.
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

// PingJava pings a Minecraft Java Edition server explicitly.
func PingJava(host string, port uint16) (*models.Response, error) {
	return pingWithCache(host, port, "Java", func() (*models.Response, error) {
		cfg := NewConfig()
		resp, err := java.Ping(host, port, host, cfg)
		if err != nil {
			return nil, err
		}
		resp.Host = host
		resp.MOTD = models.CleanMOTD(resp.MOTD)
		return resp, nil
	})
}

// PingBedrock pings a Minecraft Bedrock Edition server explicitly.
func PingBedrock(host string, port uint16) (*models.Response, error) {
	return pingWithCache(host, port, "Bedrock", func() (*models.Response, error) {
		cfg := NewConfig()
		resp, err := bedrock.Ping(host, port, cfg)
		if err != nil {
			return nil, err
		}
		resp.Host = host
		resp.MOTD = models.CleanMOTD(resp.MOTD)
		return resp, nil
	})
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
		return fivem.Ping(context.Background(), host, port, DefaultTimeout)
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
