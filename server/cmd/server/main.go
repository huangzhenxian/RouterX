package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/routex/routex/internal/api"
	"github.com/routex/routex/internal/auth"
	"github.com/routex/routex/internal/config"
	"github.com/routex/routex/internal/db"
	"github.com/routex/routex/internal/logger"
	"github.com/routex/routex/internal/scheduler"
	"github.com/routex/routex/internal/service"
	"github.com/routex/routex/internal/xray"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	zlog := logger.New(cfg.AppEnv)
	defer zlog.Sync()

	gormDB, err := db.Connect(cfg)
	if err != nil {
		zlog.Sugar().Fatalf("connect db: %v", err)
	}
	if err := db.AutoMigrate(gormDB); err != nil {
		zlog.Sugar().Fatalf("automigrate: %v", err)
	}

	// 首启时无 admin 则自动创建（密码随机，打到日志）
	if err := auth.EnsureDefaultAdmin(gormDB, zlog); err != nil {
		zlog.Sugar().Fatalf("bootstrap admin: %v", err)
	}

	xrayClient, err := xray.New(
		net.JoinHostPort(cfg.XrayAPIHost, cfg.XrayAPIPort),
		cfg.XrayInboundTag,
	)
	if err != nil {
		zlog.Sugar().Fatalf("xray client: %v", err)
	}
	defer xrayClient.Close()

	// service 层
	userSvc := service.NewUserService(gormDB, xrayClient)
	adminSvc := service.NewAdminService(gormDB, cfg.JWTSecret)
	nodeSvc := service.NewNodeService(gormDB)
	subSvc := service.NewSubscriptionService(gormDB, cfg)
	providerSvc := service.NewProviderService(gormDB)

	// 后台调度：流量轮询 + Xray 健康/自愈
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	collector := scheduler.NewTrafficCollector(gormDB, xrayClient, userSvc, zlog, cfg.TrafficPollInterval)
	go collector.Run(ctx)

	watcher := scheduler.NewXrayWatcher(xrayClient, userSvc, zlog, 15*time.Second)
	watcher.SyncNow(ctx) // 启动时立刻同步一次
	go watcher.Run(ctx)

	providerChecker := scheduler.NewProviderHealthChecker(providerSvc, zlog, cfg.ProviderHealthInterval)
	go providerChecker.Run(ctx)

	// HTTP 路由
	router := api.NewRouter(api.Deps{
		Cfg:       cfg,
		Log:       zlog,
		XC:        xrayClient,
		Users:     userSvc,
		Admins:    adminSvc,
		Nodes:     nodeSvc,
		Subs:      subSvc,
		Providers: providerSvc,
	})

	srv := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		zlog.Sugar().Infof("listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			zlog.Sugar().Fatalf("serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zlog.Info("shutting down")

	cancel() // 通知 scheduler 退出
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		zlog.Sugar().Errorf("shutdown: %v", err)
	}
}
