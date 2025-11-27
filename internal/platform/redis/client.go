package redis

import "github.com/redis/go-redis/v9"

// NewClient provides a Redis client configured for high-throughput workloads.
func NewClient(addr, password string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		PoolSize:     200,
		MinIdleConns: 10,
	})
}
