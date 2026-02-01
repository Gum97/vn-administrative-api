package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"vn-admin-api/internal/api"
	"vn-admin-api/internal/cache"
	"vn-admin-api/internal/config"
	"vn-admin-api/internal/database"
	"vn-admin-api/internal/logger"
)

func main() {
	// 1. Load Config
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 2. Init Logger
	appLog := logger.New("logs/server.log", false)
	appLog.Info("Starting API Server", "port", cfg.ServerPort)

	// 3. Connect DB
	repo, err := database.Connect(cfg)
	if err != nil {
		appLog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer repo.Close()

	// 4. Initialize Cache (Redis or Memory fallback)
	var appCache cache.Cache
	if cfg.RedisURL != "" {
		redisCache, err := cache.NewRedisCache(cfg.RedisURL, cfg.CacheTTL)
		if err != nil {
			appLog.Warn("Redis connection failed, falling back to memory cache", "error", err)
			appCache = cache.NewMemoryCache(cfg.CacheTTL)
		} else {
			appLog.Info("Redis cache connected", "url", cfg.RedisURL)
			appCache = redisCache
		}
	} else {
		appLog.Info("Using in-memory cache (no REDIS_URL configured)")
		appCache = cache.NewMemoryCache(cfg.CacheTTL)
	}

	// 5. Create Router
	router := api.NewRouterWithCache(repo, appLog, appCache)

	// 6. Configure Server with Production Timeouts
	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Run server in goroutine
	go func() {
		appLog.Info("Server listening", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLog.Error("Server failed", "error", err)
			os.Exit(1)
		}
	}()

	// 7. Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	appLog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		appLog.Error("Server forced to shutdown", "error", err)
	}

	appLog.Info("Server exited properly")
}
