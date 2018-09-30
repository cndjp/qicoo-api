package db

import (
	"os"
	"time"

	"github.com/gomodule/redigo/redis"
)

// RedisPool
var RedisPool *redis.Pool

// initRedisPool RedisConnectionPoolからconnectionを取り出す
func InitRedisPool() {
	url := os.Getenv("REDIS_URL")

	// idle connection limit:3    active connection limit:1000
	pool := &redis.Pool{
		MaxIdle:     3,
		MaxActive:   1000,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", url) },
	}

	RedisPool = pool
}

// GetRedisConnection
func GetRedisConnection() (conn redis.Conn) {
	return RedisPool.Get()
}
