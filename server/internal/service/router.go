package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/routex/routex/internal/model"
	"github.com/routex/routex/internal/xray"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// OutboundTag 是 Xray 配置里 vless-in 指向的固定出站 tag。
// RouterService 通过 RemoveOutbound + AddOutbound 持续替换它的实际实现：
//   - 有健康 provider：替换成对应 SOCKS5/HTTP outbound
//   - 没有：替换成 freedom（直连）兜底
const OutboundTag = "proxy-out"

// RouterService 把 ProviderService 选出的"最优 provider" 实际生效到 Xray 路由。
// 单实例、并发安全；当前活跃 provider 通过 currentProviderID 跟踪。
type RouterService struct {
	db  *gorm.DB
	xc  *xray.Client
	log *zap.Logger

	mu                sync.Mutex
	currentProviderID int64 // 0 表示当前是 freedom 兜底
}

func NewRouterService(db *gorm.DB, xc *xray.Client, log *zap.Logger) *RouterService {
	return &RouterService{db: db, xc: xc, log: log}
}

// CurrentProviderID 返回当前 proxy-out 实际指向的 provider ID（0 = freedom 直连）。
func (s *RouterService) CurrentProviderID() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.currentProviderID
}

// SyncBest 评估健康池，必要时切换。幂等安全：相同最优时不动 Xray。
//
// 选择策略：status=1 AND healthy=true，按 priority asc, latency_millis asc 排序，取第一条。
func (s *RouterService) SyncBest(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	best, err := s.pickBest(ctx)
	if err != nil {
		return err
	}

	var targetID int64
	if best != nil {
		targetID = best.ID
	}
	if targetID == s.currentProviderID {
		return nil
	}

	// 不论目标是 provider 还是 freedom，先尝试删除旧的 proxy-out。
	// not-found 忽略：第一次启动时 Xray 里可能就没有这个 tag。
	if err := s.xc.RemoveOutbound(ctx, OutboundTag); err != nil && !isOutboundNotFound(err) {
		s.log.Warn("router: remove old outbound failed", zap.Error(err))
	}

	if best == nil {
		if err := s.xc.AddFreedomOutbound(ctx, OutboundTag); err != nil {
			return fmt.Errorf("activate freedom fallback: %w", err)
		}
		s.currentProviderID = 0
		s.log.Info("router: no healthy provider, fell back to direct")
		return nil
	}

	if err := s.activateProvider(ctx, best); err != nil {
		// 失败时尝试回到 freedom，避免 proxy-out 缺失导致 vless 流量被丢
		s.log.Error("router: activate provider failed, falling back to freedom",
			zap.Int64("provider_id", best.ID), zap.Error(err))
		_ = s.xc.AddFreedomOutbound(ctx, OutboundTag)
		s.currentProviderID = 0
		return err
	}

	s.currentProviderID = best.ID
	s.log.Info("router: activated provider",
		zap.Int64("provider_id", best.ID), zap.String("name", best.Name),
		zap.String("type", best.Type), zap.Int("latency_ms", best.LatencyMillis))
	return nil
}

func (s *RouterService) pickBest(ctx context.Context) (*model.ProxyProvider, error) {
	var p model.ProxyProvider
	err := s.db.WithContext(ctx).
		Where("status = ? AND healthy = ?", 1, true).
		Order("priority asc, latency_millis asc").
		First(&p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (s *RouterService) activateProvider(ctx context.Context, p *model.ProxyProvider) error {
	port := uint32(p.Port)
	switch strings.ToLower(p.Type) {
	case "socks5":
		return s.xc.AddSocksOutbound(ctx, OutboundTag, p.Host, port, p.Username, p.Password)
	case "http", "https":
		// HTTPS 代理同样走 CONNECT 模型，xhttp client 不区分 TLS 与否（TLS 需要 stream settings，本 v1 不支持）
		return s.xc.AddHTTPOutbound(ctx, OutboundTag, p.Host, port, p.Username, p.Password)
	default:
		return fmt.Errorf("router: unsupported provider type %q", p.Type)
	}
}

// EnsureInitial 启动时调一次：保证 proxy-out 一定存在（即便没 provider 也是 freedom）。
// 设计目的：deploy/xray/config.json 把 proxy-out 写成 freedom 占位，但万一被改掉或我们
// 想完全脱离静态配置，这一步也兜底。
func (s *RouterService) EnsureInitial(ctx context.Context) {
	if err := s.SyncBest(ctx); err != nil {
		s.log.Warn("router: initial sync failed", zap.Error(err))
	}
}

// isOutboundNotFound 识别 "not found / not exist" 一类的错误。
// Xray 没用 gRPC status code，只能字符串匹配；版本变动时此处可能需要适配。
func isOutboundNotFound(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "not found") || strings.Contains(s, "not exist")
}
