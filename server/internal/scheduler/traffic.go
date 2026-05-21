package scheduler

import (
	"context"
	"time"

	"github.com/routex/routex/internal/model"
	"github.com/routex/routex/internal/service"
	"github.com/routex/routex/internal/xray"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// TrafficCollector 周期性地从 Xray 拉取每个启用用户的流量计数（reset=true 拉完即清零），
// 累加到 users.used_traffic，并写一条 user_traffic 明细。
// 同时检查超额和到期，触发 Disable。
type TrafficCollector struct {
	db       *gorm.DB
	xc       *xray.Client
	users    *service.UserService
	log      *zap.Logger
	interval time.Duration
}

func NewTrafficCollector(db *gorm.DB, xc *xray.Client, users *service.UserService, log *zap.Logger, interval time.Duration) *TrafficCollector {
	if interval <= 0 {
		interval = 60 * time.Second
	}
	return &TrafficCollector{db: db, xc: xc, users: users, log: log, interval: interval}
}

func (t *TrafficCollector) Run(ctx context.Context) {
	t.log.Info("traffic collector started", zap.Duration("interval", t.interval))
	ticker := time.NewTicker(t.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			t.log.Info("traffic collector stopped")
			return
		case <-ticker.C:
			t.tick(ctx)
		}
	}
}

func (t *TrafficCollector) tick(ctx context.Context) {
	var users []model.User
	if err := t.db.WithContext(ctx).Where("status = ?", 1).Find(&users).Error; err != nil {
		t.log.Warn("collector: list users", zap.Error(err))
		return
	}
	now := time.Now()
	for i := range users {
		u := &users[i]
		t.collectOne(ctx, u, now)
	}
}

func (t *TrafficCollector) collectOne(ctx context.Context, u *model.User, now time.Time) {
	stat, err := t.xc.QueryUserTraffic(ctx, xray.EmailOf(u.ID), true)
	if err != nil {
		t.log.Warn("collector: query user traffic", zap.Int64("user_id", u.ID), zap.Error(err))
		return
	}
	delta := stat.Uplink + stat.Downlink

	if delta > 0 {
		_ = t.db.WithContext(ctx).Create(&model.UserTraffic{
			UserID:   u.ID,
			Upload:   stat.Uplink,
			Download: stat.Downlink,
		}).Error
		// 原子累加，避免 read-modify-write 竞态
		_ = t.db.WithContext(ctx).Model(u).
			UpdateColumn("used_traffic", gorm.Expr("used_traffic + ?", delta)).Error
		u.UsedTraffic += delta
	}

	// 超额 / 到期 → 禁用
	if u.TrafficLimit > 0 && u.UsedTraffic >= u.TrafficLimit {
		if err := t.users.Disable(ctx, u.ID); err != nil {
			t.log.Warn("collector: disable over-quota", zap.Int64("user_id", u.ID), zap.Error(err))
		} else {
			t.log.Info("user disabled by traffic quota",
				zap.Int64("user_id", u.ID), zap.Int64("used", u.UsedTraffic), zap.Int64("limit", u.TrafficLimit))
		}
		return
	}
	if !u.ExpireTime.IsZero() && now.After(u.ExpireTime) {
		if err := t.users.Disable(ctx, u.ID); err != nil {
			t.log.Warn("collector: disable expired", zap.Int64("user_id", u.ID), zap.Error(err))
		} else {
			t.log.Info("user disabled by expiration", zap.Int64("user_id", u.ID), zap.Time("expired_at", u.ExpireTime))
		}
	}
}
