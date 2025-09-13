package middleware

import (
	"strconv"
	"time"

	"github.com/iokiris/efm-subscription-api/internal/infra"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsMiddleware создает middleware для сбора HTTP метрик
func MetricsMiddleware(metrics *infra.Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// счетчик активных запросов
		metrics.HTTPRequestsInFlight.Inc()
		defer metrics.HTTPRequestsInFlight.Dec()

		c.Next()

		// сбор метрик после обработки
		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(c.Writer.Status())

		metrics.HTTPRequestsTotal.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			statusCode,
		).Inc()

		metrics.HTTPRequestDuration.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			statusCode,
		).Observe(duration)
	}
}

// PrometheusHandler эндпоинт /metrics
func PrometheusHandler() gin.HandlerFunc {
	return gin.WrapH(promhttp.Handler())
}
