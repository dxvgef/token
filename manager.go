package token

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrMakeToken    = errors.New("unable to make manager")
	ErrInvalidToken = errors.New("invalid token")
)

// Manager 管理器实例
type Manager struct {
	options     *ManagerOptions
	redisClient *redis.Client
}

// ManagerOptions 管理器配置
type ManagerOptions struct {
	Timeout        int               // 每次操作 redis 的超时时间（秒）
	KeyPrefix      string            // redis 的键名前缀
	MakeTokenFunc  func() string     // 生成 Manager 的函数
	CheckTokenFunc func(string) bool // 检查 Manager 格式是否正确
}

// NewManager 新建管理器实例
func NewManager(redisClient *redis.Client, opts *ManagerOptions) (token *Manager, err error) {
	if redisClient == nil {
		return nil, errors.New("redis client is nil")
	}
	if opts == nil {
		opts = &ManagerOptions{
			Timeout: 10,
		}
	} else if opts.Timeout < 1 {
		return nil, errors.New("timeout value must be > 1")
	} else if opts.MakeTokenFunc == nil {
		return nil, errors.New("MakeTokenFunc undefined")
	} else if opts.CheckTokenFunc == nil {
		return nil, errors.New("CheckTokenFunc undefined")
	}
	token = &Manager{
		redisClient: redisClient,
		options:     opts,
	}
	return
}

// MakeToken 创建一个新的访问令牌
func (manager *Manager) MakeToken(meta *MetaData, payload map[string]any) (*Token, error) {
	now := time.Now().Unix()
	if payload == nil {
		payload = make(map[string]any)
	}
	if meta == nil {
		meta = &MetaData{}
	}

	// 创建令牌字符串
	tokenStr := manager.options.MakeTokenFunc()
	if tokenStr == "" {
		return nil, ErrMakeToken
	}

	token := Token{
		manager:      manager,
		value:        tokenStr,
		createdAt:    now,
		ttl:          meta.TTL,
		expiresAt:    now + meta.TTL,
		refreshLimit: meta.RefreshLimit,
		ip:           meta.IP,
		fingerprint:  meta.Fingerprint,
		childToken:   "",
	}

	// 写入保留 payload
	payload["_created_at"] = token.createdAt
	payload["_ttl"] = token.ttl
	payload["_expires_at"] = token.expiresAt
	payload["_refreshed_at"] = 0
	payload["_refreshed_count"] = 0
	payload["_refresh_limit"] = meta.RefreshLimit
	payload["_ip"] = meta.IP
	payload["_fingerprint"] = meta.Fingerprint
	payload["_child_token"] = ""

	key := manager.options.KeyPrefix + token.value

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(manager.options.Timeout)*time.Second)
	defer cancel()

	// 判断 token 是否存在
	if result := manager.redisClient.Exists(ctx, key); result.Err() != nil {
		return nil, result.Err()
	} else if result.Val() == 1 {
		return nil, errors.New("token already exists")
	}

	// 启用事务
	_, err := manager.redisClient.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		// 写入payload
		if err := pipe.HSet(ctx, key, payload).Err(); err != nil {
			return err
		}
		if token.ttl > 1 {
			// 设置 token 的生命周期
			if result := pipe.Expire(ctx, key, time.Duration(meta.TTL)*time.Second); result.Err() != nil {
				return result.Err()
			}
		}
		// 执行事务
		_, err := pipe.Exec(ctx)
		return err
	})
	if err != nil {
		return nil, err
	}

	return &token, nil
}

// ParseToken 解析令牌
func (manager *Manager) ParseToken(value string) (*Token, error) {
	var err error

	if manager.options.CheckTokenFunc != nil {
		if !manager.options.CheckTokenFunc(value) {
			return nil, ErrInvalidToken
		}
	}

	token := Token{
		manager: manager,
		value:   value,
	}

	key := manager.options.KeyPrefix + token.value

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 获得 payload
	payloadResult := manager.redisClient.HGetAll(ctx, key)
	if payloadResult.Err() != nil {
		return nil, payloadResult.Err()
	}
	payload := payloadResult.Val()

	token.ip = payload["_ip"]
	token.fingerprint = payload["_fingerprint"]
	token.childToken = payload["_child_token"]
	if token.createdAt, err = strconv.ParseInt(payload["_created_at"], 10, 64); err != nil {
		return nil, ErrInvalidToken
	}
	if token.expiresAt, err = strconv.ParseInt(payload["_expires_at"], 10, 64); err != nil {
		return nil, ErrInvalidToken
	}
	if token.ttl, err = strconv.ParseInt(payload["_ttl"], 10, 64); err != nil {
		return nil, ErrInvalidToken
	}
	if token.refreshedAt, err = strconv.ParseInt(payload["_refreshed_at"], 10, 64); err != nil {
		return nil, ErrInvalidToken
	}
	if token.refreshedCount, err = strconv.Atoi(payload["_refreshed_count"]); err != nil {
		return nil, ErrInvalidToken
	}
	if token.refreshLimit, err = strconv.Atoi(payload["_refresh_limit"]); err != nil {
		return nil, ErrInvalidToken
	}
	return &token, nil
}

// DestroyToken 销毁令牌，可决定是否同时销毁子令牌
func (manager *Manager) DestroyToken(tokenStr string, destroyChildToken bool) error {
	if !manager.options.CheckTokenFunc(tokenStr) {
		return ErrInvalidToken
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	token, err := manager.ParseToken(tokenStr)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil
		}
		return err
	}
	if !destroyChildToken || token.childToken == "" {
		return manager.redisClient.Del(ctx, manager.options.KeyPrefix+token.value).Err()
	}

	currTokenKey := manager.options.KeyPrefix + token.value
	childTokenKey := manager.options.KeyPrefix + token.childToken

	return manager.redisClient.Del(ctx, currTokenKey, childTokenKey).Err()
}
