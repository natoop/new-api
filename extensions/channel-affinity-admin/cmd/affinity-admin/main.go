package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/QuantumNous/new-api/extensions/channel-affinity-admin/internal/affinity"
	"github.com/QuantumNous/new-api/extensions/channel-affinity-admin/internal/auth"
	"github.com/QuantumNous/new-api/extensions/channel-affinity-admin/internal/config"
	"github.com/QuantumNous/new-api/extensions/channel-affinity-admin/internal/httpapi"
	"github.com/go-redis/redis/v8"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	appConfig, err := config.Load()
	if err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	rdb := redis.NewClient(appConfig.RedisOptions)
	pingContext, cancel := context.WithTimeout(context.Background(), appConfig.RedisTimeout)
	err = rdb.Ping(pingContext).Err()
	cancel()
	if err != nil {
		logger.Error("Redis is unavailable", "error", err)
		os.Exit(1)
	}
	defer rdb.Close()

	store := affinity.NewStore(rdb, appConfig.RuleName, appConfig.TTL, appConfig.AuditStream)
	identityClient := &http.Client{Timeout: appConfig.AuthTimeout}
	authenticator := auth.NewNewAPIAuthenticator(appConfig.NewAPIBaseURL, identityClient)
	api := httpapi.New(appConfig, store, authenticator, logger)
	server := &http.Server{
		Addr:              appConfig.ListenAddr,
		Handler:           api.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		logger.Info("affinity admin service listening", "addr", appConfig.ListenAddr, "rule_name", appConfig.RuleName, "ttl_seconds", int(appConfig.TTL.Seconds()))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("HTTP server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	shutdownContext, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownContext); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}
}
