package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"

	"github.com/tdevilphan/quote-snap-golang/internal/config"
	"github.com/tdevilphan/quote-snap-golang/internal/platform/logger"
	"github.com/tdevilphan/quote-snap-golang/internal/platform/mongodb"
	"github.com/tdevilphan/quote-snap-golang/internal/platform/queue"
	trackingqueue "github.com/tdevilphan/quote-snap-golang/internal/tracking/queue"
	"github.com/tdevilphan/quote-snap-golang/internal/tracking/repository"
	"github.com/tdevilphan/quote-snap-golang/internal/tracking/worker"
)

func main() {
	cfg := config.New()
	log := logger.New(cfg.AppName + "-worker")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoClient, err := mongodb.Connect(ctx, cfg.MongoURI)
	if err != nil {
		log.Error("failed to connect to mongodb", "error", err)
		exit(1)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := mongoClient.Disconnect(shutdownCtx); err != nil {
			log.Error("mongodb disconnect error", "error", err)
		}
	}()

	database := mongoClient.Database(cfg.MongoDatabase)
	eventRepo, err := repository.NewMongoEventRepository(database)
	if err != nil {
		log.Error("failed to initialize event repository", "error", err)
		exit(1)
	}

	processor := worker.NewEventProcessor(eventRepo, log)

	mux := asynq.NewServeMux()
	mux.Handle(trackingqueue.EventIngestTaskType, processor.Handler())

	server := queue.NewServer(cfg.RedisAddr, cfg.RedisPassword, cfg.AsynqQueue, cfg.AsynqConcurrency, log)

	errorCh := make(chan error, 1)
	go func() {
		if err := server.Run(mux); err != nil && !errors.Is(err, asynq.ErrServerClosed) {
			errorCh <- err
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		log.Info("shutdown signal received", "signal", sig.String())
	case err := <-errorCh:
		log.Error("asynq server error", "error", err)
	}

	server.Shutdown()
}

func exit(code int) {
	os.Exit(code)
}
