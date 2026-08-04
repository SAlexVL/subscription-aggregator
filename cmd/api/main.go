package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/programmerpark/subscription-aggregator/docs"
	"github.com/programmerpark/subscription-aggregator/internal/config"
	"github.com/programmerpark/subscription-aggregator/internal/handler"
	"github.com/programmerpark/subscription-aggregator/internal/migrate"
	"github.com/programmerpark/subscription-aggregator/internal/middleware"
	"github.com/programmerpark/subscription-aggregator/internal/repository"
	"github.com/programmerpark/subscription-aggregator/internal/service"
)

// @title           Subscription Aggregator API
// @version         1.0
// @description     REST API for aggregating online subscription data
// @host            localhost:8080
// @BasePath        /api/v1
func main() {
	cfgPath := envOr("CONFIG_PATH", "config.yaml")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	log := newLogger(cfg.Logging.Level)
	slog.SetDefault(log)

	migrationsDir := envOr("MIGRATIONS_DIR", "migrations")
	if err := migrate.Up(cfg.Database.DSN(), migrationsDir, log); err != nil {
		log.Error("migrations failed", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.Database.DSN())
	if err != nil {
		log.Error("connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Error("ping database", "error", err)
		os.Exit(1)
	}
	log.Info("connected to database")

	repo := repository.NewSubscriptionRepository(pool)
	svc := service.NewSubscriptionService(repo, log)
	h := handler.NewSubscriptionHandler(svc, log)

	router := setupRouter(h, log)

	srv := &http.Server{
		Addr:              cfg.Server.Addr(),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Info("starting HTTP server", "addr", cfg.Server.Addr())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Info("shutting down")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	_ = srv.Shutdown(shutdownCtx)
}

func setupRouter(h *handler.SubscriptionHandler, log *slog.Logger) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger(log))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api/v1")
	{
		api.POST("/subscriptions", h.Create)
		api.GET("/subscriptions/sum", h.Sum)
		api.GET("/subscriptions", h.List)
		api.GET("/subscriptions/:id", h.Get)
		api.PUT("/subscriptions/:id", h.Update)
		api.DELETE("/subscriptions/:id", h.Delete)
	}

	return r
}

func newLogger(level string) *slog.Logger {
	var lvl slog.Level
	switch strings.ToLower(level) {
	case "debug":
		lvl = slog.LevelDebug
	case "warn", "warning":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl}))
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
