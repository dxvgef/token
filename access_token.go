package token

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// AccessToken 访问令牌
type AccessToken struct {
	token        *Token
	value        string // 访问令牌的值
	createdAt    int64  // 访问令牌的创建时间
	refreshCount int    // 访问令牌刷新次数
	refreshedAt  int64  // 访问令牌的刷新时间
	expiresAt    int64  // 访问令牌的到期时间
}

// Refresh 刷新访问令牌的TTL
func (receiver *AccessToken) Refresh() error {
	key := receiver.token.options.AccessTokenPrefix + receiver.value

	now := time.Now().Unix()

	// 执行 redis 操作
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// 获取 _refresh_count
	refreshCount, err := receiver.token.redisClient.HGet(ctx, key, "_refresh_count").Int()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrInvalidAccessToken
		}
		return err
	}
	_, err = receiver.token.redisClient.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		// 更新 TTL
		if boolResult := pipe.Expire(ctx, key,
			time.Duration(receiver.token.options.AccessTokenTTL)*time.Second,
		); boolResult.Err() != nil {
			return boolResult.Err()
		}
		// 更新 _refreshed_at
		intResult := pipe.HSet(ctx, key, "_refreshed_at", now)
		if intResult.Err() != nil {
			return intResult.Err()
		}
		// 更新 _refresh_count
		if intResult = pipe.HSet(ctx, key, "_refresh_count", refreshCount+1); intResult.Err() != nil {
			return intResult.Err()
		}
		// 更新 _expires_at
		if intResult = pipe.HSet(ctx, key, "_expires_at", now+receiver.token.options.AccessTokenTTL); intResult.Err() != nil {
			return intResult.Err()
		}
		return nil
	})
	if err != nil {
		return err
	}

	receiver.refreshCount = refreshCount + 1
	receiver.refreshedAt = now
	receiver.expiresAt = now + receiver.token.options.AccessTokenTTL
	return nil
}

// Get 获取访问令牌的 payload 中的某个字段值
func (receiver *AccessToken) Get(field string) (string, error) {
	key := receiver.token.options.AccessTokenPrefix + receiver.value

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	value, err := receiver.token.redisClient.HGet(ctx, key, field).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrInvalidAccessToken
		}
	}
	return value, err
}

// Set 修改访问令牌的 payload 中的某个字段值
func (receiver *AccessToken) Set(field string, value any) error {
	key := receiver.token.options.AccessTokenPrefix + receiver.value

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := receiver.token.redisClient.HSet(ctx, key, field, value).Result()
	return err
}

// Destroy 销毁当前 access token
func (receiver *AccessToken) Destroy() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := receiver.token.redisClient.Del(ctx, receiver.token.options.AccessTokenPrefix+receiver.value).Result()
	return err
}

// Value 获取访问令牌的值
func (receiver *AccessToken) Value() string {
	return receiver.value
}

// GetAll 获取访问令牌所有的荷载内容
func (receiver *AccessToken) GetAll() (map[string]string, error) {
	key := receiver.token.options.AccessTokenPrefix + receiver.value

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	value, err := receiver.token.redisClient.HGetAll(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrInvalidAccessToken
		}
	}
	delete(value, "_created_at")
	delete(value, "_expires_at")
	delete(value, "_refreshed_at")
	delete(value, "_refresh_count")
	return value, err
}

// CreatedAt 获取访问令牌的创建时间（Unix时间戳）
func (receiver *AccessToken) CreatedAt() int64 {
	return receiver.createdAt
}

// ExpiresAt 获取访问令牌的到期时间（Unix时间戳）
func (receiver *AccessToken) ExpiresAt() int64 {
	return receiver.expiresAt
}

// RefreshedAt 获取访问令牌的最后刷新的时间（Unix时间戳）
func (receiver *AccessToken) RefreshedAt() int64 {
	return receiver.refreshedAt
}

// RefreshCount 获取访问令牌的刷新次数
func (receiver *AccessToken) RefreshCount() int {
	return receiver.refreshCount
}
