# dxvgef/token
适用于`Go`语言的令牌开发包，包含访问令牌和刷新令牌的生成、校验和刷新功能

使用`ULID`算法生成令牌字符串，减少了 CPU 使用率

使用兼容`Redis`协议的服务端来存储令牌，便于实现无状态的业务端

## 运作流程：

1. 生成令牌对`Token Pair`（访问令牌`Access Token`和刷新令牌`Refresh Token`） 
2. 客户端每次请求资源时，都要带上`Access Token`，供资源服务端鉴权 
3. 可使用`Refresh Token`及其对应的`Access Token`来兑换新的`Token Pair`
4. 兑换新的`Token Pair`时，可以保留旧的`Refresh Token`反复使用直到其过期 
5. 当`Refresh Token`也失效时，需要客户端重新校验账号密码，并生成新的`Token Pair`

## 方法

- `NewRedisClient()` 创建新的Redis客户端实例（新建连接）
- `UseRedisClient()` 使用传入的Redis客户端实例（复用连接）
- `Make()` 生成新的`Token Pair`
- `ParseAccessToken()` 将字符串解析成`AccessToken`
- `ParseRefreshToken()` 将字符串解析成`RefreshToken`
- `AccessToken.Value()` 获取`Access Token`的字符串
- `AccessToken.Payload()` 获取`Access Token`的荷载内容字符串
- `AccessToken.ExpiresAt()` 获取`Access Token`的到期时间
- `RefreshToken.Value()` 获取`Refresh Token`的字符串
- `RefreshToken.AccessToken()` 获取`Refresh Token`签发的`Access Token`
- `RefreshToken.ExpiresAt()` 获取`Refresh Token`的到期时间
- `RefreshToken.Exchange()` 兑换新的`Token Pair`

## 注意事项

- `NewRedisClient()`和`UseRedisClient()`这两个方法的入参使用的是`github.com/redis/go-redis/v9`中的相关变量或参数，分别是：
