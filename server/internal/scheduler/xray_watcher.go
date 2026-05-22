package scheduler

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/routex/routex/internal/service"
	"github.com/routex/routex/internal/xray"
	"go.uber.org/zap"
)

// XrayWatcher 周期 Ping Xray gRPC，从 "宕→恢复" 跳变时触发一次 SyncAll，
// 把 DB 里所有启用用户重新 push 到 Xray，避免 xray 容器重启后丢运行态。
// 同时也让 RouterService 重新激活 proxy-out（Xray 重启后 outbound 表会清空）。
type XrayWatcher struct {
	xc       *xray.Client
	users    *service.UserService
	router   *service.RouterService
	log      *zap.Logger
	interval time.Duration
	healthy  atomic.Bool
}

func NewXrayWatcher(xc *xray.Client, users *service.UserService, router *service.RouterService, log *zap.Logger, interval time.Duration) *XrayWatcher {
	if interval <= 0 {
		interval = 15 * time.Second
	}
	w := &XrayWatcher{xc: xc, users: users, router: router, log: log, interval: interval}
	w.healthy.Store(true) // 乐观初值；首次 tick 会立刻校正
	return w
}

// SyncNow 立即跑一次同步，给启动时调用。
func (w *XrayWatcher) SyncNow(ctx context.Context) {
	if err := w.xc.Ping(ctx); err != nil {
		w.healthy.Store(false)
		w.log.Warn("xray initial ping failed; watcher will retry", zap.Error(err))
		return
	}
	added, skipped, err := w.users.SyncAll(ctx)
	if err != nil {
		w.log.Warn("xray initial SyncAll failed", zap.Error(err))
		return
	}
	w.log.Info("xray initial sync done", zap.Int("added", added), zap.Int("skipped", skipped))
	if w.router != nil {
		w.router.EnsureInitial(ctx)
	}
}

func (w *XrayWatcher) Run(ctx context.Context) {
	w.log.Info("xray watcher started", zap.Duration("interval", w.interval))
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			w.log.Info("xray watcher stopped")
			return
		case <-ticker.C:
			w.tick(ctx)
		}
	}
}

func (w *XrayWatcher) tick(ctx context.Context) {
	err := w.xc.Ping(ctx)
	wasHealthy := w.healthy.Load()
	if err != nil {
		if wasHealthy {
			w.log.Warn("xray became unhealthy", zap.Error(err))
		}
		w.healthy.Store(false)
		return
	}
	if !wasHealthy {
		w.log.Info("xray recovered, resyncing all users")
		added, skipped, sErr := w.users.SyncAll(ctx)
		if sErr != nil {
			w.log.Warn("xray resync failed", zap.Error(sErr))
		} else {
			w.log.Info("xray resync done", zap.Int("added", added), zap.Int("skipped", skipped))
		}
		// Xray 重启会清空 outbound 表，proxy-out 也得重新激活
		if w.router != nil {
			if err := w.router.SyncBest(ctx); err != nil {
				w.log.Warn("xray recovered: router resync failed", zap.Error(err))
			}
		}
	}
	w.healthy.Store(true)
}
