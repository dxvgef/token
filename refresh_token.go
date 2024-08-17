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
	value       string // 刷新令牌的值
	accessToken string // 绑定的访问令牌
}

// ParseRefreshToken 解析RefreshToken
func ParseRefreshToken(value string) (*RefreshToken, error) {
	var (
		err          error
		refreshToken RefreshToken
		ttl          time.Duration
	)
	if _, err = ulid.Parse(value); err != nil {
		return nil, errors.New("invalid refresh token")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	refreshToken.accessToken, err = redisCli.Get(ctx, "refresh_token:"+value).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errors.New("invalid refresh token")
		}
		return nil, err
	}
	ttl, err = redisCli.TTL(ctx, "refresh_token:"+value).Result()
	if err != nil {
		return nil, err
	}
	if ttl < 1 {
		return nil, errors.New("invalid refresh token")
	}
	refreshToken.value = value
	return &refreshToken, nil
}

// Value 获取刷新令牌的值
func (receiver *RefreshToken) Value() string {
	return receiver.value
}

// ExpiresAt 获取访问令牌的到期时间
func (receiver *RefreshToken) ExpiresAt() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	ttl, err := redisCli.TTL(ctx, "refresh_token:"+receiver.value).Result()
	if err != nil {
		return 0, err
	}
	if ttl < 1 {
		return 0, nil
	}
	return time.Now().Add(ttl).Unix(), nil
}

// AccessToken 获取刷新令牌绑定的访问令牌
func (receiver *RefreshToken) AccessToken() string {
	return receiver.accessToken
}

// Exchange 兑换新的令牌对，可保留旧的刷新令牌，且不会更新旧刷新令牌的TTL
func (receiver *RefreshToken) Exchange(oldAccessToken string, opts *Options, makeNewRefreshToken bool) (*AccessToken, error) {
	var (
		accessToken  AccessToken
		redisBoolCmd *redis.BoolCmd
		redisIntCmd  *redis.IntCmd
	)
	if oldAccessToken == "" || receiver.accessToken != oldAccessToken {
		return nil, errors.New("invalid access token")
	}
	if opts.AccessTokenTTL < 1 {
		return nil, errors.New("invalid AccessTokenTTL value")
	}
	if makeNewRefreshToken && opts.RefreshTokenTTL < 1 {
		return nil, errors.New("invalid RefreshTokenTTL value")
	}

	// 生成新的访问令牌
	if opts.AccessTokenPayload == "" {
		opts.AccessTokenPayload = " "
	}
	accessToken.payload = opts.AccessTokenPayload
	accessToken.value = ulid.Make().String()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	redisBoolCmd = redisCli.SetNX(
		ctx,
		"access_token:"+accessToken.value,
		accessToken.payload,
		opts.AccessTokenTTL,
	)
	if redisBoolCmd.Err() != nil {
		return nil, redisBoolCmd.Err()
	}
	if !redisBoolCmd.Val() {
		return nil, errors.New("failed to generate access token")
	}
	receiver.accessToken = accessToken.value

	// 删除旧的访问令牌，使旧的会话失效
	if redisIntCmd = redisCli.Del(ctx, "access_token:"+oldAccessToken); redisIntCmd.Err() != nil {
		return nil, redisIntCmd.Err()
	}

	if !makeNewRefreshToken {
		// 更新旧的刷新令牌
		redisBoolCmd = redisCli.SetXX(
			ctx,
			"refresh_token:"+receiver.value,
			receiver.accessToken,
			-1, // 不修改TTL
		)
		if redisBoolCmd.Err() != nil {
			return nil, redisBoolCmd.Err()
		}
		if !redisBoolCmd.Val() {
			return nil, errors.New("failed to update refresh token")
		}
	} else {
		// 删除旧的刷新令牌
		if redisIntCmd = redisCli.Del(ctx, "refresh_token:"+receiver.value); redisIntCmd.Err() != nil {
			return nil, redisIntCmd.Err()
		}
		receiver.value = ulid.Make().String()
		redisBoolCmd = redisCli.SetNX(
			ctx,
			"refresh_token:"+receiver.value,
			receiver.accessToken,
			opts.RefreshTokenTTL,
		)
		if redisBoolCmd.Err() != nil {
			return nil, redisBoolCmd.Err()
		}
		if !redisBoolCmd.Val() {
			return nil, errors.New("failed to generate refresh token")
		}
	}
	return &accessToken, nil
}
