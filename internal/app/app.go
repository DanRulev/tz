package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"tz/internal/config"
	"tz/internal/db"
	"tz/internal/handler"
	"tz/internal/repository"
	"tz/internal/server"
	"tz/internal/service"
	"tz/pkg/logger"

	"go.uber.org/zap"
)

func Start() {
	cfg, err := config.New()
	if err != nil {
		panic(err)
	}

	log, err := logger.New(cfg.LoggerConfig)
	if err != nil {
		panic(err)
	}

	log.Info("starting app")

	db, err := db.New(cfg.DatabaseConfig, log)
	if err != nil {
		log.Fatal("failed to connect to database", zap.Error(err))
		return
	}
	defer func() {
		err := db.Close()
		if err != nil {
			log.Error("failed to close database connection", zap.Error(err))
		}
	}()
	log.Info("connected to database")

	repository := repository.NewSubscriptionRepository(db)
	service := service.NewSubscriptionService(repository, log)
	handler := handler.NewHandler(service, log)
	log.Info("dependencies initialized")

	server := server.New(cfg.Server, handler.Init())
	log.Info("server initialized")

	go func() {
		if err := server.Run(); err != nil && err != http.ErrServerClosed {
			log.Error("server failed", zap.Error(err))
		} else {
			log.Info("server stopped gracefully")
		}
	}()

	log.Info("application is running. Waiting for termination signal...")
	quite := make(chan os.Signal, 1)
	signal.Notify(quite, syscall.SIGINT, syscall.SIGTERM)
	<-quite

	log.Info("termination signal received. Shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("server shutdown error", zap.Error(err))
	}
}
