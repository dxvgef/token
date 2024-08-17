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
}

// ParseAccessToken 解析AccessToken
func ParseAccessToken(value string) (*AccessToken, error) {
	var (
		err         error
		accessToken AccessToken
		ttl         time.Duration
	)
	if _, err = ulid.Parse(value); err != nil {
		return nil, errors.New("invalid access token")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	accessToken.payload, err = redisCli.Get(ctx, "access_token:"+value).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errors.New("invalid access token")
		}
		return nil, err
	}
	ttl, err = redisCli.TTL(ctx, "access_token:"+value).Result()
	if err != nil {
		return nil, err
	}
	if ttl < 0 {
		return nil, errors.New("invalid access token")
	}
	accessToken.value = value
	return &accessToken, nil
}

// Value 获取访问令牌的值
func (receiver *AccessToken) Value() string {
	return receiver.value
}

// Payload 获取访问令牌的荷载内容
func (receiver *AccessToken) Payload() string {
	return receiver.payload
}

// ExpiresAt 获取访问令牌的到期时间
func (receiver *AccessToken) ExpiresAt() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	ttl, err := redisCli.TTL(ctx, "access_token:"+receiver.value).Result()
	if err != nil {
		return 0, err
	}
	if ttl < 1 {
		return 0, nil
	}
	return time.Now().Add(ttl).Unix(), nil
}
