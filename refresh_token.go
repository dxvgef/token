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
	value       string // 刷新令牌的值
	accessToken string // 绑定的访问令牌
	createdAt   int64  // 访问令牌的创建时间
	expiresAt   int64  // 访问令牌的到期时间
	useCount    int    // 使用次数
	usedAt      int64  // 上次使用时间
}

// Value 获取刷新令牌的值
func (receiver *RefreshToken) Value() string {
	return receiver.value
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

// UseCount 获取访问令牌的使用次数
func (receiver *RefreshToken) UseCount() int {
	return receiver.useCount
}

// UsedAt 获取访问令牌的最后使用时间
func (receiver *RefreshToken) UsedAt() int64 {
	return receiver.usedAt
}

// Exchange 兑换新的令牌对，保留当前刷新令牌，且不会更新旧刷新令牌的TTL
// 第一个error是逻辑错误，第二个error是运行时错误
func (receiver *RefreshToken) Exchange(payload map[string]string) (*AccessToken, error, error) {
	var accessToken AccessToken
	now := time.Now().Unix()

	// 检查刷新令牌是否有效
	if receiver.expiresAt < time.Now().Unix() {
		return nil, errors.New("refresh token invalid"), nil
	}

	// 生成新的 access token 的属性
	accessToken.value = ulid.Make().String()
	accessToken.payload = payload
	if accessToken.payload == nil {
		accessToken.payload = make(map[string]string)
	}
	accessToken.createdAt = now
	accessToken.expiresAt = accessToken.createdAt + receiver.token.options.AccessTokenTTL
	accessToken.token = receiver.token

	// 执行 redis 操作
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// 获取 refresh token 的 use_count
	useCount, err := receiver.token.redisClient.HGet(ctx,
		receiver.token.options.RefreshTokenPrefix+receiver.value,
		"_use_count",
	).Int()
	if err != nil {
		return nil, nil, err
	}
	_, err = receiver.token.redisClient.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		key := receiver.token.options.AccessTokenPrefix + accessToken.value
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
			return errors.New("access token [" + key + "] already exists")
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
		// 写入 payload
		for k := range accessToken.payload {
			if intResult = pipe.HSet(ctx, key, k, accessToken.payload[k]); intResult.Err() != nil {
				return intResult.Err()
			}
		}
		// 设置 access token 的生命周期
		if boolResult := pipe.Expire(ctx, key,
			time.Duration(receiver.token.options.AccessTokenTTL)*time.Second,
		); boolResult.Err() != nil {
			return boolResult.Err()
		}

		key = receiver.token.options.RefreshTokenPrefix + receiver.value

		if intResult = pipe.HSet(ctx, key, "_access_token", accessToken.value); intResult.Err() != nil {
			return intResult.Err()
		}
		// 更新 refresh token 的 used_at
		if intResult = pipe.HSet(ctx, key, "_use_count", useCount+1); intResult.Err() != nil {
			return intResult.Err()
		}
		// 更新 refresh token 的 used_at
		if intResult = pipe.HSet(ctx, key, "_used_at", now); intResult.Err() != nil {
			return intResult.Err()
		}

		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	receiver.useCount++
	receiver.usedAt = time.Now().Unix()
	receiver.accessToken = accessToken.value
	return &accessToken, nil, nil
}

// Destroy 销毁当前 refresh token，或及其 access token
func (receiver *RefreshToken) Destroy(delAccessToken bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := receiver.token.redisClient.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		// 删除 refresh token
		intResult := receiver.token.redisClient.Del(ctx, receiver.token.options.RefreshTokenPrefix+receiver.value)
		if intResult.Err() != nil {
			return intResult.Err()
		}
		if delAccessToken {
			// 删除 access token
			if intResult = receiver.token.redisClient.Del(ctx,
				receiver.token.options.AccessTokenPrefix+receiver.accessToken,
			); intResult.Err() != nil {
				return intResult.Err()
			}
		}
		return nil
	})
	return err
}
