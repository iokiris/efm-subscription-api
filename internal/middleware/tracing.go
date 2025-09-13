package middleware

import (
	"context"

	"github.com/iokiris/efm-subscription-api/internal/logger"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
)

type contextKey string

const (
	traceIDKey contextKey = "trace_id"
	spanIDKey  contextKey = "span_id"
)

// TracingMiddleware создает middleware для трейсинга HTTP запросов
func TracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Извлекаем trace context из заголовков
		ctx := otel.GetTextMapPropagator().Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

		// Создаем span для запроса
		tracer := otel.Tracer("github.com/iokiris/efm-subscription-api")
		ctx, span := tracer.Start(ctx, c.Request.Method+" "+c.FullPath())
		defer span.End()

		// Добавляем атрибуты к span
		span.SetAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
			attribute.String("http.user_agent", c.Request.UserAgent()),
			attribute.String("http.scheme", c.Request.URL.Scheme),
			attribute.String("http.host", c.Request.Host),
		)

		// Добавляем trace_id в контекст для логирования
		if spanCtx := span.SpanContext(); spanCtx.IsValid() {
			ctx = context.WithValue(ctx, traceIDKey, spanCtx.TraceID().String())
			ctx = context.WithValue(ctx, spanIDKey, spanCtx.SpanID().String())
		}

		c.Request = c.Request.WithContext(ctx)

		c.Next()

		// Добавляем атрибуты после обработки
		span.SetAttributes(
			attribute.Int("http.status_code", c.Writer.Status()),
			attribute.Int("http.response_size", c.Writer.Size()),
		)

		// Логируем с trace_id
		if traceID, ok := ctx.Value("trace_id").(string); ok {
			logger.L.Info("request completed",
				zap.String("trace_id", traceID),
				zap.String("method", c.Request.Method),
				zap.String("path", c.FullPath()),
				zap.Int("status", c.Writer.Status()),
			)
		}

	}
}

// DatabaseTracingMiddleware создает middleware для трейсинга БД операций
func DatabaseTracingMiddleware(operation, table string) func(context.Context) (context.Context, func()) {
	return func(ctx context.Context) (context.Context, func()) {
		tracer := otel.Tracer("github.com/iokiris/efm-subscription-api")
		ctx, span := tracer.Start(ctx, "db."+operation)

		span.SetAttributes(
			attribute.String("db.operation", operation),
			attribute.String("db.table", table),
		)

		return ctx, func() {
			span.End()
		}
	}
}

// CacheTracingMiddleware создает middleware для трейсинга кэш операций
func CacheTracingMiddleware(operation, key string) func(context.Context) (context.Context, func()) {
	return func(ctx context.Context) (context.Context, func()) {
		tracer := otel.Tracer("github.com/iokiris/efm-subscription-api")
		ctx, span := tracer.Start(ctx, "cache."+operation)

		span.SetAttributes(
			attribute.String("cache.operation", operation),
			attribute.String("cache.key", key),
		)

		return ctx, func() {
			span.End()
		}
	}
}

// RabbitMQTracingMiddleware создает middleware для трейсинга RabbitMQ операций
func RabbitMQTracingMiddleware(operation, exchange, routingKey string) func(context.Context) (context.Context, func()) {
	return func(ctx context.Context) (context.Context, func()) {
		tracer := otel.Tracer("github.com/iokiris/efm-subscription-api")
		ctx, span := tracer.Start(ctx, "rabbitmq."+operation)

		span.SetAttributes(
			attribute.String("messaging.operation", operation),
			attribute.String("messaging.destination", exchange),
			attribute.String("messaging.destination_kind", "queue"),
			attribute.String("messaging.rabbitmq.routing_key", routingKey),
		)

		return ctx, func() {
			span.End()
		}
	}
}
