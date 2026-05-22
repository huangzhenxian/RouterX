package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/routex/routex/internal/service"
	"go.uber.org/zap"
)

// ProviderHealthChecker 周期检查所有启用的出口代理是否可达，写回延迟和失败计数。
// 连续失败超过阈值自动禁用（status=0），避免坏代理一直被路由。
// 每轮结束后通知 RouterService 重新评估最优 provider，做无感故障切换。
type ProviderHealthChecker struct {
	svc                  *service.ProviderService
	router               *service.RouterService
	log                  *zap.Logger
	interval             time.Duration
	autoDisableThreshold int
	parallelism          int
}

func NewProviderHealthChecker(svc *service.ProviderService, router *service.RouterService, log *zap.Logger, interval time.Duration) *ProviderHealthChecker {
	if interval <= 0 {
		interval = 60 * time.Second
	}
	return &ProviderHealthChecker{
		svc:                  svc,
		router:               router,
		log:                  log,
		interval:             interval,
		autoDisableThreshold: 5,
		parallelism:          5,
	}
}

func (c *ProviderHealthChecker) Run(ctx context.Context) {
	c.log.Info("provider health checker started", zap.Duration("interval", c.interval))
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			c.log.Info("provider health checker stopped")
			return
		case <-ticker.C:
			c.tick(ctx)
		}
	}
}

func (c *ProviderHealthChecker) tick(ctx context.Context) {
	providers, err := c.svc.EnabledForCheck(ctx)
	if err != nil {
		c.log.Warn("provider check: list", zap.Error(err))
		return
	}
	if len(providers) == 0 {
		return
	}

	// 限流并发，避免一次几十条 provider 瞬间冲出去
	sem := make(chan struct{}, c.parallelism)
	var wg sync.WaitGroup
	for i := range providers {
		p := &providers[i]
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			res := c.svc.Check(ctx, p)
			if err := c.svc.PersistCheck(ctx, p, res, c.autoDisableThreshold); err != nil {
				c.log.Warn("provider check: persist", zap.Int64("provider_id", p.ID), zap.Error(err))
				return
			}
			if !res.Healthy {
				c.log.Info("provider unhealthy",
					zap.Int64("provider_id", p.ID), zap.String("name", p.Name),
					zap.Int("fail_count", p.FailCount+1), zap.String("err", res.Err))
			}
		}()
	}
	wg.Wait()

	// 一轮检查写回数据后，让 router 重新挑最优 provider。
	// 不阻塞下一轮 tick：失败就记录日志，下次再来。
	if c.router != nil {
		if err := c.router.SyncBest(ctx); err != nil {
			c.log.Warn("provider check: router sync", zap.Error(err))
		}
	}
}
