package infra

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// Tracer настройки трейсинга
type Tracer struct {
	tracer trace.Tracer
}

func NewTracer(serviceName, jaegerEndpoint string) (*Tracer, error) {
	// Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerEndpoint)))
	if err != nil {
		return nil, fmt.Errorf("failed to create jaeger exporter: %w", err)
	}

	// ресурс с информацией о сервисе
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String("1.0.0"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// создаем TracerProvider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(1.0)),
	)

	// глобальный TracerProvider
	otel.SetTracerProvider(tp)

	// глобальный propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return &Tracer{
		tracer: tp.Tracer(serviceName),
	}, nil
}

// GetTracer возвращает tracer
func (t *Tracer) GetTracer() trace.Tracer {
	return t.tracer
}

// StartSpan создает новый span
func (t *Tracer) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, name, opts...)
}

// AddSpanAttributes добавляет атрибуты к span
// не используется, но может понадобиться
func AddSpanAttributes(span trace.Span, attrs map[string]interface{}) {
	for key, value := range attrs {
		span.SetAttributes(attribute.String(key, fmt.Sprintf("%v", value)))
	}
}

// Close закрывает tracer
func (t *Tracer) Close(ctx context.Context) error {
	if tp, ok := otel.GetTracerProvider().(*sdktrace.TracerProvider); ok {
		return tp.Shutdown(ctx)
	}
	return nil
}
