package infra

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics содержит все метрики приложения
type Metrics struct {
	// HTTP метрики
	HTTPRequestsTotal    *prometheus.CounterVec
	HTTPRequestDuration  *prometheus.HistogramVec
	HTTPRequestsInFlight prometheus.Gauge

	// Бизнес метрики
	SubscriptionsTotal   *prometheus.CounterVec
	SubscriptionsCreated prometheus.Counter
	SubscriptionsUpdated prometheus.Counter
	SubscriptionsDeleted prometheus.Counter
	SubscriptionsSummary *prometheus.HistogramVec

	// Кэш метрики
	CacheHits   *prometheus.CounterVec
	CacheMisses *prometheus.CounterVec

	// БД метрики
	DatabaseQueriesTotal  *prometheus.CounterVec
	DatabaseQueryDuration *prometheus.HistogramVec
	DatabaseConnections   *prometheus.GaugeVec

	// RabbitMQ метрики
	RabbitMQMessagesPublished *prometheus.CounterVec
	RabbitMQMessagesFailed    *prometheus.CounterVec
}

func NewMetrics() *Metrics {
	return &Metrics{
		// HTTP метрики
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status_code"},
		),

		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path", "status_code"},
		),

		HTTPRequestsInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "http_requests_in_flight",
				Help: "Current number of HTTP requests being processed",
			},
		),

		// Бизнес метрики
		SubscriptionsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "subscriptions_total",
				Help: "Total number of subscriptions",
			},
			[]string{"service", "status"},
		),

		SubscriptionsCreated: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "subscriptions_created_total",
				Help: "Total number of subscriptions created",
			},
		),

		SubscriptionsUpdated: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "subscriptions_updated_total",
				Help: "Total number of subscriptions updated",
			},
		),

		SubscriptionsDeleted: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "subscriptions_deleted_total",
				Help: "Total number of subscriptions deleted",
			},
		),

		SubscriptionsSummary: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "subscriptions_summary_amount",
				Help:    "Distribution of subscription summary amounts",
				Buckets: []float64{0, 100, 500, 1000, 5000, 10000, 50000, 100000},
			},
			[]string{"service"},
		),

		// Кэш метрики
		CacheHits: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cache_hits_total",
				Help: "Total number of cache hits",
			},
			[]string{"cache_type", "key_pattern"},
		),

		CacheMisses: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cache_misses_total",
				Help: "Total number of cache misses",
			},
			[]string{"cache_type", "key_pattern"},
		),

		// БД метрики
		DatabaseQueriesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "database_queries_total",
				Help: "Total number of database queries",
			},
			[]string{"operation", "table"},
		),

		DatabaseQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "database_query_duration_seconds",
				Help:    "Database query duration in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
			},
			[]string{"operation", "table"},
		),

		DatabaseConnections: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "database_connections_active",
				Help: "Number of active database connections",
			},
			[]string{"state"},
		),

		// RabbitMQ метрики
		RabbitMQMessagesPublished: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "rabbitmq_messages_published_total",
				Help: "Total number of messages published to RabbitMQ",
			},
			[]string{"exchange", "routing_key"},
		),

		RabbitMQMessagesFailed: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "rabbitmq_messages_failed_total",
				Help: "Total number of failed message publications to RabbitMQ",
			},
			[]string{"exchange", "routing_key", "error_type"},
		),
	}
}
