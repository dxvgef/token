package token

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// Token 令牌
type Token struct {
	manager        *Manager
	value          string // 令牌的值
	createdAt      int64  // 令牌的创建时间
	refreshLimit   int    // 刷新次数限制（0:不限制,-1一次性）
	refreshedCount int    // 令牌刷新次数
	refreshedAt    int64  // 令牌的刷新时间
	ttl            int64  // 令牌的TTL
	expiresAt      int64  // 令牌的到期时间
	ip             string // 绑定ip
	fingerprint    string // 绑定指纹
	childToken     string // 生成的子令牌
}

// MetaData 令牌的元数据
type MetaData struct {
	TTL          int64  // 生命周期（秒）
	IP           string // 绑定客户端IP
	Fingerprint  string // 绑定客户端指纹
	RefreshLimit int    // manager 刷新次数限制，-1 为一次性,0为无限制
	ChildToken   string // 子令牌
}

// Refresh 刷新令牌的TTL
func (token *Token) Refresh() error {
	if token.refreshLimit < 0 {
		return errors.New("this token cannot be refreshed")
	}

	key := token.manager.options.KeyPrefix + token.value
	refreshedCount := 0
	now := time.Now().Unix()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 获取 _refreshed_count，同时也可判断key是否存在
	strResult, err := token.manager.redisClient.HGet(ctx, key, "_refreshed_count").Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrInvalidToken
		}
		return err
	}
	if strResult == "" {
		return ErrInvalidToken
	}
	if refreshedCount, err = strconv.Atoi(strResult); err != nil {
		return err
	}
	// 判断刷新次数是否超过限制
	if token.refreshLimit > 0 && token.refreshedCount >= token.refreshLimit {
		return errors.New("refresh limit reached")
	}
	fields := map[string]any{
		"_refreshed_at":    now,
		"_refreshed_count": refreshedCount + 1,
		"_expires_at":      now + token.ttl,
	}

	// 启用事务
	_, err = token.manager.redisClient.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		if intResult := pipe.HSet(ctx, key, fields); intResult.Err() != nil {
			return intResult.Err()
		}
		// 更新 TTL
		boolResult := pipe.Expire(ctx, key, time.Duration(token.ttl)*time.Second)
		return boolResult.Err()
	})
	if err != nil {
		return err
	}

	token.refreshedCount = refreshedCount + 1
	token.refreshedAt = now
	token.expiresAt = now + token.ttl
	return nil
}

// Get 获取令牌 payload 中的某个字段值
func (token *Token) Get(field string) (string, error) {
	key := token.manager.options.KeyPrefix + token.value

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	value, err := token.manager.redisClient.HGet(ctx, key, field).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrInvalidToken
		}
	}
	return value, err
}

// GetFields 获取令牌 payload 中的多个字段值
func (token *Token) GetFields(fields ...string) (map[string]any, error) {
	if len(fields) == 0 {
		return nil, errors.New("fields is empty")
	}
	key := token.manager.options.KeyPrefix + token.value
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	values := token.manager.redisClient.HMGet(ctx, key, fields...)
	if values.Err() != nil {
		if errors.Is(values.Err(), redis.Nil) {
			return nil, ErrInvalidToken
		}
	}
	valueArr := values.Val()
	result := make(map[string]any)
	for k := range valueArr {
		// 如果字段不存在，值会是nil
		result[fields[k]] = valueArr[k]
	}
	return result, values.Err()
}

// GetAll 获取令牌所有的荷载内容
func (token *Token) GetAll(includeMeta bool) (map[string]string, error) {
	key := token.manager.options.KeyPrefix + token.value

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	value, err := token.manager.redisClient.HGetAll(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrInvalidToken
		}
	}
	if !includeMeta {
		// 删除掉令牌的属性字段
		delete(value, "_created_at")
		delete(value, "_expires_at")
		delete(value, "_refreshed_at")
		delete(value, "_refreshed_count")
		delete(value, "_refresh_limit")
		delete(value, "_ip")
		delete(value, "_fingerprint")
		delete(value, "_ttl")
		delete(value, "_child_token")
	}
	return value, err
}

// Set 设置令牌 payload 中的某个字段值
func (token *Token) Set(field string, value any) error {
	key := token.manager.options.KeyPrefix + token.value

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return token.manager.redisClient.HSet(ctx, key, field, value).Err()
}

// MakeChildToken 创建子令牌
func (token *Token) MakeChildToken(meta *MetaData, payload map[string]any) (*Token, error) {
	// 如果当前令牌已经有子令牌
	if token.childToken != "" {
		return nil, errors.New("current token has child token")
	}

	if payload == nil {
		payload = make(map[string]any)
	}
	if meta == nil {
		meta = &MetaData{}
	}

	// 创建令牌字符串
	tokenStr := token.manager.options.MakeTokenFunc()
	if tokenStr == "" {
		return nil, ErrMakeToken
	}

	now := time.Now().Unix()
	newToken := Token{
		manager:      token.manager,
		value:        tokenStr,
		createdAt:    now,
		expiresAt:    now + meta.TTL,
		refreshLimit: meta.RefreshLimit,
		ip:           meta.IP,
		fingerprint:  meta.Fingerprint,
		childToken:   "",
	}
	payload["_created_at"] = newToken.createdAt
	payload["_ttl"] = newToken.ttl
	payload["_expires_at"] = newToken.expiresAt
	payload["_refreshed_at"] = newToken.refreshedAt
	payload["_refreshed_count"] = 0
	payload["_refresh_limit"] = newToken.refreshLimit
	payload["_ip"] = newToken.ip
	payload["_fingerprint"] = newToken.fingerprint
	payload["_child_token"] = ""

	key := token.manager.options.KeyPrefix + newToken.value

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(token.manager.options.Timeout)*time.Second)
	defer cancel()

	// 判断新令牌是否存在
	if result := token.manager.redisClient.Exists(ctx, key); result.Err() != nil {
		return nil, result.Err()
	} else if result.Val() == 1 {
		return nil, errors.New("child token already exists")
	}

	// 启用事务
	_, err := token.manager.redisClient.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		// 写入新令牌的payload
		if err := pipe.HSet(ctx, key, payload).Err(); err != nil {
			return err
		}
		if meta.TTL > 1 {
			// 设置新令牌的生命周期
			if boolResult := pipe.Expire(ctx, key, time.Duration(meta.TTL)*time.Second); boolResult.Err() != nil {
				return boolResult.Err()
			}
		}
		// 在父令牌中记录子令牌信息
		if err := pipe.HSet(ctx, token.manager.options.KeyPrefix+token.value, "_child_token", newToken.value).Err(); err != nil {
			return err
		}
		// 执行事务
		_, err := pipe.Exec(ctx)
		return err
	})
	if err != nil {
		return nil, err
	}
	token.childToken = newToken.value
	return &newToken, nil
}

// Destroy 销毁当前令牌，同时销毁子令牌
func (token *Token) Destroy(destroyChildToken bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 如果是子令牌，则销毁完自己就退出
	if !destroyChildToken || token.childToken == "" {
		return token.manager.redisClient.Del(ctx, token.manager.options.KeyPrefix+token.value).Err()
	}

	currTokenKey := token.manager.options.KeyPrefix + token.value
	childTokenKey := token.manager.options.KeyPrefix + token.childToken

	return token.manager.redisClient.Del(ctx, currTokenKey, childTokenKey).Err()
}

// Value 获取访问令牌的值
func (token *Token) Value() string {
	return token.value
}

// CreatedAt 获取令牌的创建时间（Unix时间戳）
func (token *Token) CreatedAt() int64 {
	return token.createdAt
}

// TTL 获取令牌的TTL（秒）
func (token *Token) TTL() int64 {
	return token.ttl
}

// ExpiresAt 获取令牌的到期时间（Unix时间戳）
func (token *Token) ExpiresAt() int64 {
	return token.expiresAt
}

// RefreshedAt 获取令牌的最后刷新的时间（Unix时间戳）
func (token *Token) RefreshedAt() int64 {
	return token.refreshedAt
}

// RefreshedCount 获取令牌的刷新次数
func (token *Token) RefreshedCount() int {
	return token.refreshedCount
}

// RefreshLimit 获取令牌的刷新限制次数
func (token *Token) RefreshLimit() int {
	return token.refreshLimit
}

// IP 获取令牌的绑定IP
func (token *Token) IP() string {
	return token.ip
}

// Fingerprint 获取令牌的绑定指纹
func (token *Token) Fingerprint() string {
	return token.fingerprint
}

// ChildToken 获取子令牌
func (token *Token) ChildToken() string {
	return token.childToken
}

// ValidateIP 验证IP
func (token *Token) ValidateIP(clientIP string) bool {
	return token.ip == "" || clientIP == token.ip
}

// ValidateFingerprint 验证指纹
func (token *Token) ValidateFingerprint(fingerprint string) bool {
	return token.fingerprint == "" || fingerprint == token.fingerprint
}

// IsUnexpired 未过期
func (token *Token) IsUnexpired() bool {
	return token.expiresAt == 0 || token.expiresAt > time.Now().Unix()
}
