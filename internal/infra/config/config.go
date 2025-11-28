package config

import (
	"os"
	"strconv"
	"time"
)

// Config captures environment-driven runtime configuration for a service instance.
type Config struct {
	AppName          string
	HTTPAddr         string
	HTTPPort         string
	MongoURI         string
	MongoDatabase    string
	RedisAddr        string
	RedisPassword    string
	AsynqQueue       string
	AsynqConcurrency int
	ShutdownTimeout  time.Duration
	RequestTimeout   time.Duration
}

// New loads configuration from the process environment and applies sane defaults.
func New() Config {
	return Config{
		AppName:          getEnv("APP_NAME", "tracking-service"),
		HTTPAddr:         getEnv("HTTP_ADDR", "0.0.0.0"),
		HTTPPort:         getEnv("HTTP_PORT", "8080"),
		MongoURI:         getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDatabase:    getEnv("MONGO_DATABASE", "tracking"),
		RedisAddr:        getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:    os.Getenv("REDIS_PASSWORD"),
		AsynqQueue:       getEnv("ASYNQ_QUEUE", "tracking_events"),
		AsynqConcurrency: getEnvInt("ASYNQ_CONCURRENCY", 50),
		ShutdownTimeout:  getEnvDuration("SHUTDOWN_TIMEOUT", 15*time.Second),
		RequestTimeout:   getEnvDuration("REQUEST_TIMEOUT", 3*time.Second),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return fallback
}
