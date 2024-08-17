package token

import (
	"github.com/redis/go-redis/v9"
)

const redisCliErr = "redis client is nil"

var redisCli *redis.Client

// UseRedisClient 使用传入Redis客户端的连接
func UseRedisClient(cli *redis.Client) {
	if cli == nil {
		redisCli = redis.NewClient(&redis.Options{})
	} else {
		redisCli = cli
	}
}

// NewRedisClient 使用新的Redis客户端连接
func NewRedisClient(opts *redis.Options) {
	if opts == nil {
		opts = &redis.Options{}
	}
	redisCli = redis.NewClient(opts)
}
