package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/routex/routex/internal/api"
	"github.com/routex/routex/internal/config"
	"github.com/routex/routex/internal/db"
	"github.com/routex/routex/internal/logger"
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

	router := api.NewRouter(cfg, zlog, gormDB)

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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		zlog.Sugar().Errorf("shutdown: %v", err)
	}
}
