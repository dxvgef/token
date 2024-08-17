package token

import (
	"context"
	"errors"
	"time"

	"github.com/oklog/ulid/v2"
)

// Options 创建令牌的配置
type Options struct {
	AccessTokenTTL     time.Duration // 访问令牌的TTL,不能 <1
	AccessTokenPayload string        // 访问令牌荷载内容
	RefreshTokenTTL    time.Duration // 刷新令牌的TTL
}

// Make 创建一个新的令牌
func Make(opts *Options) (*AccessToken, *RefreshToken, error) {
	var (
		accessToken  AccessToken
		refreshToken RefreshToken
	)
	if redisCli == nil {
		return nil, nil, errors.New(redisCliErr)
	}
	if opts == nil {
		return nil, nil, errors.New("invalid options")
	}
	if opts.AccessTokenPayload == "" {
		opts.AccessTokenPayload = " "
	}
	if opts.AccessTokenTTL < 1 {
		return nil, nil, errors.New("invalid AccessTokenTTL value")
	}
	accessToken.payload = opts.AccessTokenPayload
	accessToken.value = ulid.Make().String()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result := redisCli.SetNX(
		ctx,
		"access_token:"+accessToken.value,
		accessToken.payload,
		opts.AccessTokenTTL,
	)
	if result.Err() != nil {
		return nil, nil, result.Err()
	}
	if !result.Val() {
		return nil, nil, errors.New("failed to generate access token")
	}
	// RefreshTokenTTL <1 表示不生成刷新令牌
	if opts.RefreshTokenTTL > 1 {
		refreshToken.value = ulid.Make().String()
		refreshToken.accessToken = accessToken.value
		result = redisCli.SetNX(
			ctx,
			"refresh_token:"+refreshToken.value,
			accessToken.value,
			opts.RefreshTokenTTL,
		)
		if result.Err() != nil {
			return nil, nil, result.Err()
		}
		if !result.Val() {
			return nil, nil, errors.New("failed to generate refresh token")
		}
	}

	return &accessToken, &refreshToken, nil
}
