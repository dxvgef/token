# dxvgef/token

`Go`语言的访问令牌开发包，可应用于 Session Cookie 或 OAuth 等多种场景 

## 特点

- 令牌可存储荷载内容
- 令牌可定义客户端的IP和指纹信息，用于在校验令牌的同时验证客户端有效性 
- 用自定义回调函数生成和校验令牌字符串，方便使用各种算法
- 可限制每个令牌的生命周期的刷新次数
- 使用`Redis`协议兼容的服务端存储令牌数据

## 应用场景

### Session Cookie 模式

创建一个旧短生命周期的令牌作为会话ID，写入到客户端 Cookie。

服务端使用`Token.Valid()`验证令牌有效性。

可用以下两种方案保持会话存活：
1. 使用`Token.Refresh()`方法延长令牌的生命周期
2. 销毁当前令牌，并创建一个同样荷载内容的新令牌，发给客户端替换原来的令牌

### OAuth 模式

先创建一个生命周期较长的令牌，作为刷新令牌（refreshToken）。

然后使用刷新令牌的`MakeChildToken()`方法创建一个生命周期较短的子令牌，作为访问令牌（accessToken）

当访问令牌到期，客户端将访问和刷新令牌都发到服务端进行校验。

服务端校验刷新令牌有效，并且校验其子令牌与提交的访问令牌是否匹配

校验都通过后，使用`Manager.DestroyToken(refreshToken, true)`或者`refreshToken.Destroy(true)`将当前的两个令牌都销毁

最后重新创建一对新令牌，发给客户端替换原来的令牌

## 方法列表

### 管理器 (Manager)

- `NewManager()` 创建令牌管理器实例，必须传入`github.com/redis/go-redis`的`*Client`类型的变量
- `ManagerOptions.MakeTokenFunc()` 管理器创建令牌字符串的回调函数
- `ManagerOptions.CheckTokenFunc()` 管理器校验令牌字符串的回调函数
- `Manager.MakeToken()` 创建令牌
- `Manager.ParseToken()` 将字符串解析成令牌
- `Manager.DestroyToken()` 销毁指定的令牌，可同时销毁子令牌

### 令牌 (Token)

- `MakeChildToken()` 创建子令牌
- `Destroy()` 销毁当前令牌，可同时销毁子令牌
- `Refresh()` 刷新生命周期
- `Set()` 设置荷载中的键值
- `Value()` 获取令牌的字符串
- `Get()` 获取荷载的指定键的数据
- `GetFields()` 获取荷载的多个键的数据
- `GetAll()` 获取荷载的所有数据，可包含令牌的元信息(_开头的字段)
- `CreatedAt()` 获取创建时间
- `ExpiresAt()` 获取到期时间
- `TTL()` 获取TTL值
- `RefreshedAt()` 获取最后刷新时间
- `RefreshedCount()` 获取刷新次数
- `RefreshLimit()` 获取刷新上限制
- `IP()` 获取绑定的客户端IP
- `Fingerprint()` 获取绑定的客户端指纹信息
- `ChildToken()` 获取子令牌
- `ValidateIP()` 验证IP是否匹配
- `ValidateFingerprint` 验证指纹是否匹配
- `IsUnexpired()` 未过期
