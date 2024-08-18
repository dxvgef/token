package token

import (
	"github.com/redis/go-redis/v9"
)

const redisCliErr = "redis client is nil"

var redisCli *redis.Client
var accessTokenPrefix = ""
var refreshTokenPrefix = ""

// UseRedisClient 使用传入redis客户端的连接
func UseRedisClient(cli *redis.Client) {
	if cli == nil {
		redisCli = redis.NewClient(&redis.Options{})
	} else {
		redisCli = cli
	}
}

// NewRedisClient 使用新的redis客户端连接
func NewRedisClient(opts *redis.Options) {
	if opts == nil {
		opts = &redis.Options{}
	}
	redisCli = redis.NewClient(opts)
}

// SetAccessTokenPrefix 设置access token的redis键名前缀
func SetAccessTokenPrefix(value string) {
	accessTokenPrefix = value
}

// SetRefreshTokenPrefix 设置refresh token的redis键名前缀
func SetRefreshTokenPrefix(value string) {
	refreshTokenPrefix = value
}
