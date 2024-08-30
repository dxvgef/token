package token

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/redis/go-redis/v9"
)

const redisCliErr = "redis client is nil"

// Token 令牌实例
type Token struct {
	options     *Options
	redisClient *redis.Client
}

// Options 创建令牌的配置
type Options struct {
	AccessTokenTTL     int64  // 访问令牌的TTL（秒）
	AccessTokenPrefix  string // 访问令牌的键名前缀
	CookieSessionMode  bool   // cookie-session模式
	RefreshTokenTTL    int64  // 刷新令牌的TTL（秒），值小于0表示不使用刷新令牌
	RefreshTokenPrefix string // 刷新令牌的键名前缀
}

// New 新建实例
func New(redisClient *redis.Client, opts *Options) (token *Token, err error) {
	if redisClient == nil {
		return nil, errors.New(redisCliErr)
	}
	if opts == nil {
		opts = &Options{
			AccessTokenTTL:     600,
			AccessTokenPrefix:  "access_token:",
			CookieSessionMode:  false,
			RefreshTokenTTL:    1800,
			RefreshTokenPrefix: "refresh_token:",
		}
	}
	token = &Token{
		redisClient: redisClient,
		options:     opts,
	}
	return
}

// MakeAccessToken 创建一个新的访问令牌
func (receiver *Token) MakeAccessToken(payload map[string]string) (*AccessToken, error) {
	var (
		accessToken AccessToken
		key         string
		now         = time.Now().Unix()
	)

	// 生成 access token
	accessToken.token = receiver
	accessToken.value = ulid.Make().String()
	accessToken.createdAt = now
	accessToken.expiresAt = now + receiver.options.AccessTokenTTL
	accessToken.payload = payload
	if accessToken.payload == nil {
		accessToken.payload = make(map[string]string)
	}
	key = receiver.options.AccessTokenPrefix + accessToken.value

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := receiver.redisClient.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		// 判断 access token 是否重复
		intResult := pipe.Exists(ctx, key)
		if intResult.Err() != nil {
			return intResult.Err()
		}
		if intResult.Val() == 1 {
			return errors.New("access token [" + key + "] already exists")
		}

		// 写入 payload
		for k := range accessToken.payload {
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
			if intResult = pipe.HSet(ctx, key, k, accessToken.payload[k]); intResult.Err() != nil {
				return intResult.Err()
			}
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

	return &accessToken, nil
}

// ParseAccessToken 解析AccessToken
// 第一个error是逻辑错误，第二个error是运行时错误
func (receiver *Token) ParseAccessToken(value string) (*AccessToken, error, error) {
	var (
		err         error
		accessToken AccessToken
		payload     map[string]string
	)

	if _, err = ulid.Parse(value); err != nil {
		return nil, errors.New("invalid access token"), nil
	}

	accessToken.value = value

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 获得 payload
	if payload, err = receiver.redisClient.HGetAll(ctx, receiver.options.AccessTokenPrefix+value).Result(); err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errors.New("invalid access token"), nil
		}
		return nil, nil, err
	}

	accessToken.payload = make(map[string]string)

	for k := range payload {
		switch k {
		case "_created_at":
			if accessToken.createdAt, err = strconv.ParseInt(payload[k], 10, 64); err != nil {
				return nil, errors.New("invalid access token"), nil
			}
		case "_expires_at":
			if accessToken.expiresAt, err = strconv.ParseInt(payload[k], 10, 64); err != nil {
				return nil, errors.New("invalid access token"), nil
			}
		case "_refreshed_at":
			if accessToken.refreshedAt, err = strconv.ParseInt(payload[k], 10, 64); err != nil {
				return nil, errors.New("invalid access token"), nil
			}
		case "_refresh_count":
			if accessToken.refreshCount, err = strconv.Atoi(payload[k]); err != nil {
				return nil, errors.New("invalid access token"), nil
			}
		default:
			accessToken.payload[k] = payload[k]
		}
	}

	accessToken.token = receiver
	return &accessToken, nil, nil
}

// MakeRefreshToken 创建一个新的刷新令牌
func (receiver *Token) MakeRefreshToken(accessToken string) (*RefreshToken, error) {
	var (
		refreshToken RefreshToken
	)
	if accessToken == "" {
		return nil, errors.New("access token is empty")
	}

	refreshToken.accessToken = accessToken
	refreshToken.value = ulid.Make().String()
	refreshToken.createdAt = time.Now().Unix()
	refreshToken.expiresAt = refreshToken.createdAt + receiver.options.RefreshTokenTTL
	refreshToken.token = receiver
	refreshToken.useCount = 0
	refreshToken.usedAt = 0

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	intResult := receiver.redisClient.Exists(ctx, receiver.options.AccessTokenPrefix+accessToken)
	if intResult.Err() != nil {
		return nil, intResult.Err()
	}
	if intResult.Val() < 1 {
		return nil, errors.New("invalid access token")
	}
	_, err := receiver.redisClient.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		key := receiver.options.RefreshTokenPrefix + refreshToken.value
		if intResult = pipe.HSet(ctx, key, "_created_at", refreshToken.createdAt); intResult.Err() != nil {
			return intResult.Err()
		}
		if intResult = pipe.HSet(ctx, key, "_expires_at", refreshToken.expiresAt); intResult.Err() != nil {
			return intResult.Err()
		}
		if intResult = pipe.HSet(ctx, key, "_access_token", refreshToken.accessToken); intResult.Err() != nil {
			return intResult.Err()
		}
		if intResult = pipe.HSet(ctx, key, "_use_count", refreshToken.useCount); intResult.Err() != nil {
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

	return &refreshToken, nil
}

// ParseRefreshToken 解析RefreshToken
// 第一个error是逻辑错误，第二个error是运行时错误
func (receiver *Token) ParseRefreshToken(value string) (*RefreshToken, error, error) {
	var (
		err          error
		refreshToken RefreshToken
		payload      map[string]string
	)
	if _, err = ulid.Parse(value); err != nil {
		return nil, errors.New("invalid refresh token"), nil
	}
	refreshToken.value = value

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	payload, err = receiver.redisClient.HGetAll(ctx, receiver.options.RefreshTokenPrefix+refreshToken.value).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errors.New("invalid refresh token"), nil
		}
		return nil, nil, err
	}
	for k := range payload {
		switch k {
		case "_created_at":
			if refreshToken.createdAt, err = strconv.ParseInt(payload[k], 10, 64); err != nil {
				return nil, errors.New("invalid refresh token"), nil
			}
		case "_expires_at":
			if refreshToken.expiresAt, err = strconv.ParseInt(payload[k], 10, 64); err != nil {
				return nil, errors.New("invalid refresh token"), nil
			}
		case "_access_token":
			refreshToken.accessToken = payload[k]
		case "_use_count":
			if refreshToken.useCount, err = strconv.Atoi(payload[k]); err != nil {
				return nil, errors.New("invalid refresh token"), nil
			}
		case "_used_at":
			if refreshToken.usedAt, err = strconv.ParseInt(payload[k], 10, 64); err != nil {
				return nil, errors.New("invalid refresh token"), nil
			}
		}
	}
	refreshToken.token = receiver
	return &refreshToken, nil, nil
}

// DestroyAccessToken 销毁 access token
func (receiver *Token) DestroyAccessToken(accessToken string) error {
	// 执行 redis 操作
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := receiver.redisClient.Del(ctx, receiver.options.AccessTokenPrefix+accessToken).Result()
	return err
}

// DestroyRefreshToken 销毁 refresh token
func (receiver *Token) DestroyRefreshToken(refreshToken string, delAccessToken bool) error {
	var (
		err         error
		accessToken string
	)
	// 执行 redis 操作
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// 获取 access token
	if delAccessToken {
		if accessToken, err = receiver.redisClient.HGet(ctx,
			receiver.options.RefreshTokenPrefix+refreshToken,
			"_access_token",
		).Result(); err != nil {
			return nil
		}
	}
	_, err = receiver.redisClient.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		// 删除 refresh token
		result := pipe.Del(ctx, receiver.options.RefreshTokenPrefix+refreshToken)
		if result.Err() != nil {
			return result.Err()
		}
		if delAccessToken {
			// 删除 access token
			if result = pipe.Del(ctx, receiver.options.AccessTokenPrefix+accessToken); result.Err() != nil {
				return result.Err()
			}
		}
		return nil
	})
	return err
}
