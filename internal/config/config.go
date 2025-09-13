package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string

	DBUser     string
	DBPass     string
	DBHost     string
	DBPort     string
	DBName     string
	DBMaxConns int
	DBMinConns int

	RedisAddr         string
	RedisPassword     string
	RedisDB           int
	RedisPoolSize     int
	RedisMinIdleConns int
	RedisDialTimeout  time.Duration
	RedisReadTimeout  time.Duration
	RedisWriteTimeout time.Duration
	RedisPoolTimeout  time.Duration
	CacheTTL          time.Duration

	RabbitUser string
	RabbitPass string
	RabbitHost string
	RabbitPort int

	HTTPReadTimeout  time.Duration
	HTTPWriteTimeout time.Duration
	WorkerTick       time.Duration

	LogLevel string

	// Мониторинг TODO
	JaegerEndpoint string
	MetricsEnabled bool
	TracingEnabled bool
}

// Load инициализация конфига
func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	c := &Config{
		Port:      getEnv("PORT", "8080"),
		DBHost:    getEnv("DB_HOST", "localhost"),
		DBPort:    getEnv("DB_PORT", "5432"),
		DBUser:    getEnv("DB_USER", "postgres"),
		DBPass:    getEnv("DB_PASSWORD", "postgres"),
		DBName:    getEnv("DB_NAME", "subscriptions"),
		RedisAddr: getEnv("REDIS_ADDR", "localhost:6379"),
		LogLevel:  getEnv("LOG_LEVEL", "info"),
	}

	// int поля
	c.DBMaxConns = getEnvAsInt("DB_MAX_CONNS", 50)
	c.DBMinConns = getEnvAsInt("DB_MIN_CONNS", 5)

	// NOTE: тип этих полей - time.Duration
	c.CacheTTL = getEnvAsDuration("CACHE_TTL", 10*time.Minute)
	c.HTTPReadTimeout = getEnvAsDuration("HTTP_READ_TIMEOUT", 15*time.Second)
	c.HTTPWriteTimeout = getEnvAsDuration("HTTP_WRITE_TIMEOUT", 15*time.Second)
	c.WorkerTick = getEnvAsDuration("WORKER_TICK", 5*time.Minute)

	c.RedisPassword = getEnv("REDIS_PASSWORD", "")
	c.RedisDB = getEnvAsInt("REDIS_DB", 0)
	c.RedisPoolSize = getEnvAsInt("REDIS_POOL_SIZE", 50)
	c.RedisMinIdleConns = getEnvAsInt("REDIS_MIN_IDLE_CONNS", 10)

	c.RedisDialTimeout = getEnvAsDuration("REDIS_DIAL_TIMEOUT", 3*time.Second)
	c.RedisReadTimeout = getEnvAsDuration("REDIS_READ_TIMEOUT", 2*time.Second)
	c.RedisWriteTimeout = getEnvAsDuration("REDIS_WRITE_TIMEOUT", 2*time.Second)
	c.RedisPoolTimeout = getEnvAsDuration("REDIS_POOL_TIMEOUT", 4*time.Second)

	c.RabbitUser = getEnv("RABBIT_USER", "guest")
	c.RabbitPass = getEnv("RABBIT_PASS", "guest")
	c.RabbitHost = getEnv("RABBIT_HOST", "rabbitmq")
	c.RabbitPort = getEnvAsInt("RABBIT_PORT", 5672)

	// Мониторинг
	c.JaegerEndpoint = getEnv("JAEGER_ENDPOINT", "http://jaeger:14268/api/traces")
	c.MetricsEnabled = getEnvAsBool("METRICS_ENABLED", true)
	c.TracingEnabled = getEnvAsBool("TRACING_ENABLED", true)

	return c, nil
}

// NOTE:
// Вспомогательные функции, используются для преобразования типов из строки в те, что указаны в кфг
//

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvAsDuration(key string, fallback time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return fallback
}

func getEnvAsBool(key string, fallback bool) bool {
	if val := os.Getenv(key); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	}
	return fallback
}
