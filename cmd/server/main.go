package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/iokiris/efm-subscription-api/internal/config"
	"github.com/iokiris/efm-subscription-api/internal/handler"
	"github.com/iokiris/efm-subscription-api/internal/infra"
	"github.com/iokiris/efm-subscription-api/internal/logger"
	"github.com/iokiris/efm-subscription-api/internal/middleware"
	"github.com/iokiris/efm-subscription-api/internal/repo"
	"github.com/iokiris/efm-subscription-api/internal/service"

	_ "github.com/iokiris/efm-subscription-api/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// @title           Subscriptions API
// @version         1.0
// @description     API для управления подписками и получения агрегированных сумм.
// @BasePath        /
func main() {
	logger.InitGlobal()
	defer func(L *zap.Logger) {
		_ = L.Sync()
	}(logger.L)

	// CONFIG
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, _ := config.Load()

	// POSTGRES
	dbPool, err := repo.NewPostgresPool(ctx, cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)
	if err != nil {
		logger.L.Fatal("DB Pool error:", zap.Error(err))
	}
	subRepo := repo.NewSubscriptionRepo(dbPool)

	// RABBITMQ
	conn, ch, err := infra.NewRabbitMQ(
		&infra.RabbitConfig{
			User:     cfg.RabbitUser,
			Password: cfg.RabbitPass,
			Host:     cfg.RabbitHost,
			Port:     cfg.RabbitPort,
		})
	if err != nil {
		logger.L.Fatal("rabbit connection error", zap.Error(err))
	}
	publisher := service.NewRabbitPublisher(ch, 100)
	defer conn.Close()
	defer ch.Close()

	// REDIS
	rcli, err := infra.NewRedis(ctx, infra.RedisConfig{
		Addr:         cfg.RedisAddr,
		Password:     cfg.RedisPassword,
		DB:           cfg.RedisDB,
		PoolSize:     cfg.RedisPoolSize,
		MinIdleConns: cfg.RedisMinIdleConns,
		DialTimeout:  cfg.RedisDialTimeout,
		ReadTimeout:  cfg.RedisReadTimeout,
		WriteTimeout: cfg.RedisWriteTimeout,
		PoolTimeout:  cfg.RedisPoolTimeout,
	})
	if err != nil {
		logger.L.Fatal("failed to connect to Redis:", zap.Error(err))
	}
	defer rcli.Close()

	// МОНИТОРИНГ
	var metrics *infra.Metrics
	var tracer *infra.Tracer

	if cfg.MetricsEnabled {
		metrics = infra.NewMetrics()
		logger.L.Info("Prometheus metrics enabled")
	}

	if cfg.TracingEnabled {
		var err error
		tracer, err = infra.NewTracer("github.com/iokiris/efm-subscription-api", cfg.JaegerEndpoint)
		if err != nil {
			logger.L.Fatal("failed to initialize tracer", zap.Error(err))
		}
		defer tracer.Close(ctx)
		logger.L.Info("Jaeger tracing enabled", zap.String("endpoint", cfg.JaegerEndpoint))
	}

	// SUBSCRIPTION SERVICE
	subService := service.NewSubscriptionService(subRepo, rcli, publisher, cfg.CacheTTL)
	if metrics != nil {
		subService.SetMetrics(metrics)
	}

	// GIN ROUTES INIT
	r := gin.New()
	r.Use(gin.Recovery())

	// Middleware для мониторинга
	if cfg.MetricsEnabled && metrics != nil {
		r.Use(middleware.MetricsMiddleware(metrics))
	}

	if cfg.TracingEnabled {
		r.Use(middleware.TracingMiddleware())
	}

	// healthcheck
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	// Prometheus metrics endpoint
	if cfg.MetricsEnabled && metrics != nil {
		r.GET("/metrics", middleware.PrometheusHandler())
		logger.L.Info("Prometheus metrics endpoint available at /metrics")
	}

	// swagger
	url := ginSwagger.URL("/swagger/doc.json")
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	// subscriptions
	// NOTE:
	// authRequired = true → требует авторизацию с middleware.JWTMiddleware() (т.к. заглушка, будет выдавать 401)
	h := handler.NewSubscriptionHandler(subService)
	h.RegisterRoutes(r, false)

	// HTTP
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      r,
		ReadTimeout:  cfg.HTTPReadTimeout,
		WriteTimeout: cfg.HTTPWriteTimeout,
	}

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.L.Fatal("server failed", zap.Error(err))
	}

	// graceful shutdown
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()
	_ = srv.Shutdown(shutdownCtx)
}
