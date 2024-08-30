# dxvgef/token
适用于`Go`语言的令牌开发包，包含访问令牌和刷新令牌的生成、校验和刷新功能

使用`ULID`算法生成令牌字符串，减少了 CPU 使用率

使用兼容`Redis`协议的服务端来存储令牌，便于实现无状态的业务端

## 方法

- `New()` 创建新的 Token 客户端实例并传入配置
- `MakeAccessToken()` 生成 Access Token
- `MakeRefreshToken()` 生成 Refresh Token，需要传入 Access Token
- `ParseAccessToken()` 将字符串解析成 Access Token
- `ParseRefreshToken()` 将字符串解析成 Refresh Token
- `DestroyAccessToken()` 销毁 Access Token
- `DestroyRefreshToken()` 销毁 Refresh Token，也可以同时销毁 Access Token
- `AccessToken.Value()` 获取 Access Token 的字符串
- `AccessToken.Payload()` 获取 Access Token 的荷载内容
- `AccessToken.CreatedAt()` 获取 Access Token 的创建时间
- `AccessToken.ExpiresAt()` 获取 Access Token 的到期时间
- `AccessToken.Refresh()` 刷新 Access Token 的生命周期
- `AccessToken.Destroy()` 销毁当前 Access Token
- `RefreshToken.Value()` 获取 Refresh Token 的字符串
- `RefreshToken.AccessToken()` 获取 Refresh Token 签发的 Access Token
- `RefreshToken.ExpiresAt()` 获取 Refresh Token的到期时间
- `RefreshToken.Exchange()` 兑换新的 Access Token，旧的 Access Token 将自动销毁
- `RefreshToken.Destroy()` 销毁当前 Refresh Token，也可以同时销毁 Access Token

## 提示

- `New()`方法的入参`redisClient`使用的是`github.com/redis/go-redis`的`*Client`类型
