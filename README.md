# dxvgef/token

`Go`语言的访问令牌开发包，可用于 session 或 access token 两种模式的应用场景 

使用`ULID`算法生成令牌字符串，和`Redis`协议兼容的服务端存储令牌数据

access token 支持 HASH 类型的荷载内容

## 应用场景

### session 模式

此场景只需使用 access token，如果 access token 到期，可选择以下两种方案：

- 销毁已过期的，并生成新的 access token
- 刷新已过期的 access token

### token 模式

此场景需要先生成 refresh token，然后使用它的`Exchange()`方法来兑换一个新的 access token
 
`Exchange()`方法会自动销毁它上次生成的 access token

如果 access token 到期，可以用以下两种方案获得新的 access token：
1. 使用 refresh token 兑换新的 access token
2. 销毁旧的 refresh token，然后生成新的 refresh token 并兑换新的 access token

推荐使用第 2 种方案，防止 refresh token 泄露造成的滥用

## 方法列表

### Token.

- `New()` 创建 token 实例，使用了`github.com/redis/go-redis`的`*Client`类型
- `MakeAccessToken()` 生成 access token
- `ParseAccessToken()` 将字符串解析成 access token
- `DestroyAccessToken()` 销毁指定的 access token
- `MakeRefreshToken()` 生成 refresh token
- `ParseRefreshToken()` 将字符串解析成 refresh token
- `DestroyRefreshToken()` 销毁指定的 refresh token，并自动销毁其生成的 access token

### AccessToken.

- `Value()` 获取 token 的字符串
- `Get()` 获取荷载内容中指定键名的值
- `GetAll()` 获取所有荷载内容
- `Set()` 设置荷载内容中指定键名的值
- `CreatedAt()` 获取创建时间
- `ExpiresAt()` 获取到期时间
- `Refresh()` 刷新生命周期
- `Destroy()` 销毁

### RefreshToken.

- `Value()` 获取 token 的字符串
- `AccessToken()` 获取其生成的 access token 字符串
- `ExpiresAt()` 获取到期时间
- `Exchange()` 兑换新的 access token，旧的将自动销毁
- `Destroy()` 销毁，并自动销毁其生成的 access token
