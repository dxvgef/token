package token

import (
	"context"
	"errors"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/redis/go-redis/v9"
)

// RefreshToken 刷新令牌
type RefreshToken struct {
	token       *Token
	value       string         // 刷新令牌的值
	accessToken string         // 绑定的访问令牌
	payload     map[string]any // 绑定的访问令牌的荷载内容
	createdAt   int64          // 访问令牌的创建时间
	expiresAt   int64          // 访问令牌的到期时间
	usedCount   int            // 已使用次数
	usedAt      int64          // 上次使用时间
}

// Value 获取刷新令牌的值
func (receiver *RefreshToken) Value() string {
	return receiver.value
}

// Exchange 兑换新的访问令牌，保留当前刷新令牌，且不会更新它的TTL
// 如非必要，不建议用此方法兑换新的访问令牌，而是使用 Destroy() 方法销毁此刷新令牌，并为客户端提供一对新令牌
func (receiver *RefreshToken) Exchange() (*AccessToken, error) {
	now := time.Now().Unix()

	// 检查刷新令牌是否有效
	if receiver.expiresAt < time.Now().Unix() {
		return nil, ErrInvalidRefreshToken
	}

	// 生成新的 access token 的属性
	accessToken := &AccessToken{
		token:     receiver.token,
		value:     ulid.Make().String(),
		createdAt: now,
	}
	accessToken.expiresAt = accessToken.createdAt + receiver.token.options.AccessTokenTTL

	// 新的 access token 的 key
	key := receiver.token.options.AccessTokenPrefix + accessToken.value

	// 执行 redis 操作
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// 获取 refresh token 的 used_count 旧值
	useCount, err := receiver.token.redisClient.HGet(ctx,
		receiver.token.options.RefreshTokenPrefix+receiver.value,
		"_used_count",
	).Int()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrInvalidRefreshToken
		}
		return nil, err
	}

	_, err = receiver.token.redisClient.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		// 删除旧的 access token
		intResult := pipe.Del(ctx, receiver.token.options.AccessTokenPrefix+receiver.accessToken)
		if intResult.Err() != nil {
			return intResult.Err()
		}
		// 判断新的 access token 是否重复
		intResult = pipe.Exists(ctx, key)
		if intResult.Err() != nil {
			return intResult.Err()
		}
		if intResult.Val() == 1 {
			return errors.New("new access token already exists")
		}
		// 写入 payload
		for k := range receiver.payload {
			if k != "_access_token" && k != "_created_at" && k != "_expires_at" && k != "_used_count" && k != "_used_at" {
				if intResult = pipe.HSet(ctx, key, k, receiver.payload[k]); intResult.Err() != nil {
					return intResult.Err()
				}
			}
		}
		if intResult = pipe.HSet(ctx, key, "_created_at", accessToken.createdAt); intResult.Err() != nil {
			return intResult.Err()
		}
		if intResult = pipe.HSet(ctx, key, "_expires_at", accessToken.expiresAt); intResult.Err() != nil {
			return intResult.Err()
		}
		if intResult = pipe.HSet(ctx, key, "_refreshed_at", 0); intResult.Err() != nil {
			return intResult.Err()
		}
		if intResult = pipe.HSet(ctx, key, "_refresh_count", 0); intResult.Err() != nil {
			return intResult.Err()
		}
		// 设置 access token 的生命周期
		if boolResult := pipe.Expire(ctx, key,
			time.Duration(receiver.token.options.AccessTokenTTL)*time.Second,
		); boolResult.Err() != nil {
			return boolResult.Err()
		}

		// 更新当前刷新令牌的属性
		key = receiver.token.options.RefreshTokenPrefix + receiver.value
		if intResult = pipe.HSet(ctx, key, "_access_token", accessToken.value); intResult.Err() != nil {
			return intResult.Err()
		}
		// 更新 refresh token 的 used_at
		if intResult = pipe.HSet(ctx, key, "_used_count", useCount+1); intResult.Err() != nil {
			return intResult.Err()
		}
		// 更新 refresh token 的 used_at
		if intResult = pipe.HSet(ctx, key, "_used_at", now); intResult.Err() != nil {
			return intResult.Err()
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	receiver.usedCount = useCount + 1
	receiver.usedAt = time.Now().Unix()
	receiver.accessToken = accessToken.value
	return accessToken, nil
}

// Destroy 销毁当前 refresh token 及其 access token
func (receiver *RefreshToken) Destroy() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := receiver.token.redisClient.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		// 删除 refresh token
		intResult := receiver.token.redisClient.Del(ctx, receiver.token.options.RefreshTokenPrefix+receiver.value)
		if intResult.Err() != nil {
			return intResult.Err()
		}
		// 删除 access token
		if intResult = receiver.token.redisClient.Del(ctx,
			receiver.token.options.AccessTokenPrefix+receiver.accessToken,
		); intResult.Err() != nil {
			return intResult.Err()
		}
		return nil
	})
	return err
}

// CreatedAt 获取访问令牌的创建时间
func (receiver *RefreshToken) CreatedAt() int64 {
	return receiver.createdAt
}

// ExpiresAt 获取访问令牌的到期时间
func (receiver *RefreshToken) ExpiresAt() int64 {
	return receiver.expiresAt
}

// AccessToken 获取刷新令牌绑定的访问令牌
func (receiver *RefreshToken) AccessToken() string {
	return receiver.accessToken
}

// UsedCount 获取访问令牌的使用次数
func (receiver *RefreshToken) UsedCount() int {
	return receiver.usedCount
}

// UsedAt 获取访问令牌的最后使用时间
func (receiver *RefreshToken) UsedAt() int64 {
	return receiver.usedAt
}

// Payload 获取令牌的荷载内容
func (receiver *RefreshToken) Payload() map[string]any {
	return receiver.payload
}
