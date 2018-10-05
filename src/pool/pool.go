package pool

import (
	"os"
	"time"

	"github.com/gomodule/redigo/redis"
)

var RedisPool *redis.Pool

func NewRedisPool() *redis.Pool {
	// idle connection limit:3    active connection limit:1000
	return &redis.Pool{
		MaxIdle:     3,
		MaxActive:   1000,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", os.Getenv("REDIS_URL")) },
	}
}
