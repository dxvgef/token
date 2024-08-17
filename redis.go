package token

import (
	"github.com/redis/go-redis/v9"
)

var redisCli *redis.Client

// UseRedisClient 使用传入Redis客户端的连接
func UseRedisClient(cli *redis.Client) {
	redisCli = cli
}

// NewRedisClient 使用新的Redis客户端连接
func NewRedisClient(opts *redis.Options) {
	redisCli = redis.NewClient(opts)
}
