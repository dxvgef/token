# dxvgef/token
适用于`Go`语言的令牌开发包，包含访问令牌和刷新令牌的生成、校验和刷新功能

使用`ULID`算法生成令牌字符串，减少了 CPU 使用率

使用兼容`Redis`协议的服务端来存储令牌，便于实现无状态的业务端

## 使用流程：

1. 使用`NewRedisClient()`或`UseRedisClient()`设置Redis客户端实例，用于存取Token。同时可根据需要使用`SetAccessTokenPrefix()`和`SetRefreshTokenPrefix()`来设置 Redis 的键名前缀
2. 使用`Make()`生成令牌对 Token Pair（访问令牌 Access Token 和刷新令牌 Refresh Token）
3. 将 Token Pair 都发送给客户端进行安全存储 
4. 客户端向后端发送请求时，都要带上 Access Token 
5. 服务端使用`ParseAccessToken()`来判断客是否有效 
6. 当 Access Token 过期后，客户端可使用 Refresh Token 及其对应的 Access Token 来向资源服务端兑换新的 Token Pair：
   1. 使用`ParseRefreshToken()`判断 Refresh Token 是否有效，并将字符串转为`RefreshToken`对象
   2. 使用`RefreshToken.Exchange()`来换取新的 Token Pair  
7. 当 Refresh Token 也失效时，需要客户端重新校验账号密码，并生成新的 Token Pair

## 方法

- `NewRedisClient()` 创建新的 Redis 客户端实例（新建连接）
- `UseRedisClient()` 使用传入的 Redis 客户端实例（复用连接）
- `SetAccessTokenPrefix()` 设置 AccessToken 的 Redis 键名前缀
- `SetRefreshTokenPrefix()` 设置 RefreshToken 的 Redis 键名前缀
- `Make()` 生成新的 Token Pair
- `ParseAccessToken()` 将字符串解析成 Access Token
- `ParseRefreshToken()` 将字符串解析成 Refresh Token
- `AccessToken.Value()` 获取 Access Token 的字符串
- `AccessToken.Payload()` 获取 Access Token 的荷载内容字符串
- `AccessToken.ExpiresAt()` 获取 Access Token 的到期时间
- `RefreshToken.Value()` 获取 Refresh Token 的字符串
- `RefreshToken.AccessToken()` 获取 Refresh Token 签发的 Access Token
- `RefreshToken.ExpiresAt()` 获取 Refresh Token的到期时间
- `RefreshToken.Exchange()` 兑换新的 Token Pair

## 注意事项

- `NewRedisClient()`和`UseRedisClient()`这两个方法的入参使用的是`github.com/redis/go-redis/v9`中的相关变量或参数
