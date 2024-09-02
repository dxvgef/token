package token

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/redis/go-redis/v9"
)

var (
	ErrInvalidAccessToken  = errors.New("invalid access token")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
)

// Token 令牌实例
type Token struct {
	options     *Options
	redisClient *redis.Client
}

// Options 创建 token 的配置
type Options struct {
	AccessTokenTTL     int64  // access token 的TTL（秒）
	AccessTokenPrefix  string // access token 的键名前缀
	RefreshTokenTTL    int64  // refresh token 的TTL（秒），必须大于 AccessTokenTTL
	RefreshTokenPrefix string // refresh token 的键名前缀
	Timeout            int    // 每次操作 redis 的超时时间（秒）
}

// New 新建实例
func New(redisClient *redis.Client, opts *Options) (token *Token, err error) {
	if redisClient == nil {
		return nil, errors.New("redis client is nil")
	}
	if opts == nil {
		opts = &Options{
			AccessTokenTTL:     600,
			AccessTokenPrefix:  "access_token:",
			RefreshTokenTTL:    86400,
			RefreshTokenPrefix: "refresh_token:",
			Timeout:            10,
		}
	} else if opts.Timeout < 1 {
		return nil, errors.New("Timeout value must be > 1")
	} else if opts.AccessTokenTTL < 1 {
		return nil, errors.New("AccessTokenTTL value must be > 1")
	} else if opts.RefreshTokenTTL <= opts.AccessTokenTTL {
		return nil, errors.New("RefreshTokenTTL value must be > AccessTokenTTL")
	}
	token = &Token{
		redisClient: redisClient,
		options:     opts,
	}
	return
}

// MakeAccessToken 创建一个新的访问令牌
func (receiver *Token) MakeAccessToken(payload map[string]any) (*AccessToken, error) {
	now := time.Now().Unix()
	// 生成 access token
	accessToken := &AccessToken{
		token:     receiver,
		value:     ulid.Make().String(),
		createdAt: now,
		expiresAt: now + receiver.options.AccessTokenTTL,
	}

	key := receiver.options.AccessTokenPrefix + accessToken.value

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(receiver.options.Timeout)*time.Second)
	defer cancel()

	_, err := receiver.redisClient.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		// 判断 access token 是否重复
		intResult := pipe.Exists(ctx, key)
		if intResult.Err() != nil {
			return intResult.Err()
		}
		if intResult.Val() == 1 {
			return errors.New("new access token already exists")
		}

		// 写入 payload
		for k := range payload {
			if intResult = pipe.HSet(ctx, key, k, payload[k]); intResult.Err() != nil {
				return intResult.Err()
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
			time.Duration(receiver.options.AccessTokenTTL)*time.Second,
		); boolResult.Err() != nil {
			return boolResult.Err()
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return accessToken, nil
}

// ParseAccessToken 解析AccessToken
func (receiver *Token) ParseAccessToken(value string) (*AccessToken, error) {
	var err error

	if _, err = ulid.Parse(value); err != nil {
		return nil, ErrInvalidAccessToken
	}

	accessToken := &AccessToken{
		token: receiver,
		value: value,
	}

	key := receiver.options.AccessTokenPrefix + accessToken.value

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 获得 payload
	payloadResult := receiver.redisClient.HMGet(ctx, key, "_created_at", "_expires_at", "_refreshed_at", "_refresh_count")
	if payloadResult.Err() != nil {
		return nil, payloadResult.Err()
	}
	if payloadResult.Val()[0] != nil {
		if str, ok := payloadResult.Val()[0].(string); ok {
			if accessToken.createdAt, err = strconv.ParseInt(str, 10, 64); err != nil {
				return nil, err
			}
			if accessToken.createdAt < 1 {
				return nil, ErrInvalidAccessToken
			}
		} else {
			return nil, ErrInvalidAccessToken
		}
	}
	if payloadResult.Val()[1] != nil {
		if str, ok := payloadResult.Val()[1].(string); ok {
			if accessToken.expiresAt, err = strconv.ParseInt(str, 10, 64); err != nil {
				return nil, err
			}
			if accessToken.expiresAt < 1 {
				return nil, ErrInvalidAccessToken
			}
		} else {
			return nil, ErrInvalidAccessToken
		}
	}
	if payloadResult.Val()[2] != nil {
		if str, ok := payloadResult.Val()[2].(string); ok {
			if accessToken.refreshedAt, err = strconv.ParseInt(str, 10, 64); err != nil {
				return nil, err
			}
		} else {
			return nil, ErrInvalidAccessToken
		}
	}
	if payloadResult.Val()[3] != nil {
		if str, ok := payloadResult.Val()[3].(string); ok {
			if accessToken.refreshCount, err = strconv.Atoi(str); err != nil {
				return nil, err
			}
		} else {
			return nil, ErrInvalidAccessToken
		}
	}
	return accessToken, nil
}

// MakeRefreshToken 创建一个新的刷新令牌，需传入兑换 access token 时的 payload
func (receiver *Token) MakeRefreshToken(payload map[string]any) (*RefreshToken, error) {
	refreshToken := &RefreshToken{
		token:     receiver,
		value:     ulid.Make().String(),
		createdAt: time.Now().Unix(),
		payload:   payload,
	}
	refreshToken.expiresAt = refreshToken.createdAt + receiver.options.RefreshTokenTTL

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 判断新的刷新令牌值是否重复
	key := receiver.options.RefreshTokenPrefix + refreshToken.value
	intResult := receiver.redisClient.Exists(ctx, key)
	if intResult.Err() != nil {
		return nil, intResult.Err()
	}
	if intResult.Val() == 1 {
		return nil, errors.New("new refresh token already exists")
	}

	_, err := receiver.redisClient.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		for k := range payload {
			if intResult = pipe.HSet(ctx, key, k, payload[k]); intResult.Err() != nil {
				return intResult.Err()
			}
		}
		if intResult = pipe.HSet(ctx, key, "_created_at", refreshToken.createdAt); intResult.Err() != nil {
			return intResult.Err()
		}
		if intResult = pipe.HSet(ctx, key, "_expires_at", refreshToken.expiresAt); intResult.Err() != nil {
			return intResult.Err()
		}
		if intResult = pipe.HSet(ctx, key, "_used_count", refreshToken.usedCount); intResult.Err() != nil {
			return intResult.Err()
		}
		if intResult = pipe.HSet(ctx, key, "_used_at", refreshToken.usedAt); intResult.Err() != nil {
			return intResult.Err()
		}
		// 设置生命周期
		if boolResult := pipe.Expire(ctx, key,
			time.Duration(receiver.options.RefreshTokenTTL)*time.Second,
		); boolResult.Err() != nil {
			return boolResult.Err()
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return refreshToken, nil
}

// ParseRefreshToken 解析 refresh token
func (receiver *Token) ParseRefreshToken(value string) (*RefreshToken, error) {
	var (
		err     error
		payload map[string]string
	)
	if _, err = ulid.Parse(value); err != nil {
		return nil, ErrInvalidRefreshToken
	}

	refreshToken := &RefreshToken{
		token: receiver,
		value: value,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	payload, err = receiver.redisClient.HGetAll(ctx, receiver.options.RefreshTokenPrefix+refreshToken.value).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrInvalidRefreshToken
		}
		return nil, err
	}
	for k := range payload {
		switch k {
		case "_created_at":
			if refreshToken.createdAt, err = strconv.ParseInt(payload[k], 10, 64); err != nil {
				return nil, ErrInvalidRefreshToken
			}
		case "_expires_at":
			if refreshToken.expiresAt, err = strconv.ParseInt(payload[k], 10, 64); err != nil {
				return nil, ErrInvalidRefreshToken
			}
		case "_access_token":
			refreshToken.accessToken = payload[k]
		case "_used_count":
			if refreshToken.usedCount, err = strconv.Atoi(payload[k]); err != nil {
				return nil, ErrInvalidRefreshToken
			}
		case "_used_at":
			if refreshToken.usedAt, err = strconv.ParseInt(payload[k], 10, 64); err != nil {
				return nil, ErrInvalidRefreshToken
			}
		}
	}
	return refreshToken, nil
}

// DestroyAccessToken 销毁 access token
func (receiver *Token) DestroyAccessToken(accessToken string) error {
	// 执行 redis 操作
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := receiver.redisClient.Del(ctx, receiver.options.AccessTokenPrefix+accessToken).Result()
	return err
}

// DestroyRefreshToken 销毁 refresh token，同时自动销毁它生成的 access token
func (receiver *Token) DestroyRefreshToken(refreshToken string) error {
	var (
		err         error
		accessToken string
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 获取 access token
	if accessToken, err = receiver.redisClient.HGet(ctx,
		receiver.options.RefreshTokenPrefix+refreshToken,
		"_access_token",
	).Result(); err != nil {
		return nil
	}

	_, err = receiver.redisClient.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		// 删除 refresh token
		result := pipe.Del(ctx, receiver.options.RefreshTokenPrefix+refreshToken)
		if result.Err() != nil {
			return result.Err()
		}
		// 删除 access token
		if result = pipe.Del(ctx, receiver.options.AccessTokenPrefix+accessToken); result.Err() != nil {
			return result.Err()
		}
		return nil
	})

	return err
}
