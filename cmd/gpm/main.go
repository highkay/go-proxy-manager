package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"go-proxy-manager/internal/config"
	"go-proxy-manager/internal/manager"
	"go-proxy-manager/internal/server"
	"go-proxy-manager/internal/store"
	"go-proxy-manager/pkg/logger"
)

func main() {
	configPath := flag.String("config", "configs/config.yaml", "path to config file")
	flag.Parse()

	// 1. Load Config
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// 2. Init Logger
	logger.InitLogger(cfg.App.LogLevel)
	slog.Info("starting go-proxy-manager", "port", cfg.App.Port)

	// 3. Init Store & Manager
	s := store.NewStore()
	m := manager.NewManager(cfg, s)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 4. Start Manager
	m.Start(ctx)

	// 5. Start Server
	srv := server.NewServer(cfg.App.Port, s)
	go func() {
		if err := srv.Start(); err != nil {
			slog.Error("server failed", "error", err)
			cancel()
		}
	}()

	// 6. Graceful Shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		slog.Info("received signal, shutting down", "signal", sig)
	case <-ctx.Done():
		slog.Info("context cancelled, shutting down")
	}

	m.Stop()
	slog.Info("shutdown complete")
}
