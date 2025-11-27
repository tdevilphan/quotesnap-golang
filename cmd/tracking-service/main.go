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

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/tdevilphan/quote-snap-golang/internal/config"
	"github.com/tdevilphan/quote-snap-golang/internal/platform/logger"
	"github.com/tdevilphan/quote-snap-golang/internal/platform/mongodb"
	"github.com/tdevilphan/quote-snap-golang/internal/platform/queue"
	redispkg "github.com/tdevilphan/quote-snap-golang/internal/platform/redis"
	"github.com/tdevilphan/quote-snap-golang/internal/tracking/repository"
	"github.com/tdevilphan/quote-snap-golang/internal/tracking/service"
	httptransport "github.com/tdevilphan/quote-snap-golang/internal/tracking/transport/http"
)

func main() {
	cfg := config.New()
	log := logger.New(cfg.AppName)

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
	if _, err := repository.NewMongoEventRepository(database); err != nil {
		log.Error("failed to initialize event repository", "error", err)
		exit(1)
	}

	redisClient := redispkg.NewClient(cfg.RedisAddr, cfg.RedisPassword)
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Error("redis close error", "error", err)
		}
	}()

	queueClient := queue.NewClient(cfg.RedisAddr, cfg.RedisPassword)
	defer func() {
		if err := queueClient.Close(); err != nil {
			log.Error("queue client close error", "error", err)
		}
	}()

	eventService := service.NewEventService(queueClient)
	eventHandler := httptransport.NewEventHandler(eventService, cfg.AsynqQueue, cfg.RequestTimeout, log)

	router := buildRouter(log, eventHandler)

	srv := &http.Server{
		Addr:         cfg.HTTPAddr + ":" + cfg.HTTPPort,
		Handler:      router,
		ReadTimeout:  cfg.RequestTimeout + time.Second,
		WriteTimeout: cfg.RequestTimeout + 2*time.Second,
	}

	errorCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errorCh <- err
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		log.Info("shutdown signal received", "signal", sig.String())
	case err := <-errorCh:
		log.Error("http server error", "error", err)
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancelShutdown()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("http server shutdown error", "error", err)
	}
}

func buildRouter(log *slog.Logger, handler *httptransport.EventHandler) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(cors.Default())

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")
	handler.Register(api)

	return r
}

func exit(code int) {
	os.Exit(code)
}
