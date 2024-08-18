package token

import (
	"context"
	"errors"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/redis/go-redis/v9"
)

// AccessToken 访问令牌
type AccessToken struct {
	value   string // 访问令牌的值
	payload string // 荷载内容
	prefix  string // 键名前缀
}

// ParseAccessToken 解析AccessToken
// 第一个error是逻辑错误，第二个error是运行时错误
func ParseAccessToken(value string) (*AccessToken, error, error) {
	var (
		err         error
		accessToken AccessToken
		ttl         time.Duration
	)
	if redisCli == nil {
		return nil, nil, errors.New(redisCliErr)
	}
	if _, err = ulid.Parse(value); err != nil {
		return nil, errors.New("invalid access token"), nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	accessToken.payload, err = redisCli.Get(ctx, accessTokenPrefix+value).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errors.New("invalid access token"), nil
		}
		return nil, nil, err
	}
	ttl, err = redisCli.TTL(ctx, accessTokenPrefix+value).Result()
	if err != nil {
		return nil, nil, err
	}
	if ttl < 0 {
		return nil, errors.New("invalid access token"), nil
	}
	accessToken.value = value
	return &accessToken, nil, nil
}

// Value 获取访问令牌的值
func (receiver *AccessToken) Value() string {
	return receiver.value
}

// Payload 获取访问令牌的荷载内容
func (receiver *AccessToken) Payload() string {
	return receiver.payload
}

// ExpiresAt 获取访问令牌的到期时间（Unix时间戳）
// 第一个参数返回-1表示已过期
func (receiver *AccessToken) ExpiresAt() (int64, error) {
	if redisCli == nil {
		return 0, errors.New(redisCliErr)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	ttl, err := redisCli.TTL(ctx, accessTokenPrefix+receiver.value).Result()
	if err != nil {
		return 0, err
	}

	switch {
	case ttl == 0:
		// key没有超时时间
		return 0, nil
	case ttl < 0:
		// key不存在或者已过期
		return -1, nil
	default:
		return time.Now().Add(ttl).Unix(), nil
	}
}
