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

	apphttp "quotesnap/internal/app/http"
	"quotesnap/internal/core/usecase"
	"quotesnap/internal/infra/config"
	"quotesnap/internal/infra/logger"
	inframongo "quotesnap/internal/infra/mongodb"
	queueasynq "quotesnap/internal/infra/queue/asynq"
	inframongorepo "quotesnap/internal/infra/repository/mongo"
)

func main() {
	cfg := config.New()
	log := logger.New(cfg.AppName)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoClient, err := inframongo.Connect(ctx, cfg.MongoURI)
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
	if _, err := inframongorepo.NewEventRepository(database); err != nil {
		log.Error("failed to initialize event repository", "error", err)
		exit(1)
	}

	queueClient := queueasynq.NewClient(cfg.RedisAddr, cfg.RedisPassword)
	defer func() {
		if err := queueClient.Close(); err != nil {
			log.Error("queue client close error", "error", err)
		}
	}()

	dispatcher := queueasynq.NewDispatcher(queueClient, cfg.AsynqQueue)
	ingestEvent := usecase.NewIngestEvent(dispatcher)
	eventHandler := apphttp.NewEventHandler(ingestEvent, cfg.RequestTimeout, log)

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

func buildRouter(log *slog.Logger, handler *apphttp.EventHandler) *gin.Engine {
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
