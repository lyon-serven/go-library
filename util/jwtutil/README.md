# JWT Utility (jwtutil)

JWT (JSON Web Token) 工具包，提供完整的 JWT 令牌生成、验证和管理功能。

## 功能特性

- ✅ JWT 令牌生成
- ✅ JWT 令牌验证
- ✅ 令牌刷新
- ✅ 自定义声明 (Claims)
- ✅ 标准声明支持 (iss, sub, aud, exp, nbf, iat, jti)
- ✅ 多种签名算法 (HS256, HS384, HS512)
- ✅ 访问令牌和刷新令牌生成
- ✅ 令牌过期检查
- ✅ 提取令牌信息

## 安装

```go
import "github.com/lyon-serven/go-library/util/jwtutil"
```

## 快速开始

### 1. 基本使用

```go
package main

import (
    "fmt"
    "time"
    "github.com/lyon-serven/go-library/util/jwtutil"
)

func main() {
    // 创建 JWT 配置
    config := jwtutil.NewJWTConfig("your-secret-key")
    config.Issuer = "your-app"
    config.ExpiryDuration = 2 * time.Hour
    
    // 生成令牌
    claims := &jwtutil.Claims{
        StandardClaims: jwtutil.StandardClaims{
            Subject: "user123",
        },
        CustomClaims: map[string]interface{}{
            "username": "john_doe",
            "role": "admin",
        },
    }
    
    token, err := config.GenerateToken(claims)
    if err != nil {
        panic(err)
    }
    
    fmt.Println("Token:", token)
    
    // 验证令牌
    verifiedClaims, err := config.VerifyToken(token)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Subject: %s\n", verifiedClaims.Subject)
    fmt.Printf("Username: %v\n", verifiedClaims.CustomClaims["username"])
    fmt.Printf("Role: %v\n", verifiedClaims.CustomClaims["role"])
}
```

### 2. 简化的令牌生成

```go
// 快速生成访问令牌
token, err := jwtutil.GenerateAccessToken(
    "your-secret-key",
    "user123",
    24*time.Hour,
    map[string]interface{}{
        "username": "john_doe",
        "role": "admin",
    },
)

// 快速验证令牌
claims, err := jwtutil.VerifyAccessToken("your-secret-key", token)
```

### 3. 访问令牌和刷新令牌

```go
// 生成访问令牌 (短期有效)
accessToken, err := jwtutil.GenerateAccessToken(
    "your-secret-key",
    "user123",
    15*time.Minute,
    map[string]interface{}{
        "username": "john_doe",
        "role": "admin",
    },
)

// 生成刷新令牌 (长期有效)
refreshToken, err := jwtutil.GenerateRefreshToken(
    "your-secret-key",
    "user123",
    7*24*time.Hour,
)
```

### 4. 令牌刷新

```go
config := jwtutil.NewJWTConfig("your-secret-key")

// 刷新令牌
newToken, err := config.RefreshToken(oldToken)
if err != nil {
    // 处理错误（可能是令牌已过期）
    panic(err)
}
```

### 5. 令牌验证和自定义检查

```go
config := jwtutil.NewJWTConfig("your-secret-key")

// 验证令牌并检查发行者和受众
claims, err := config.ValidateToken(token, "expected-issuer", "expected-audience")
if err != nil {
    // 处理验证错误
    panic(err)
}
```

### 6. 提取令牌信息（无需完全验证）

```go
// 提取主题
subject, err := jwtutil.ExtractSubject(token)

// 提取自定义声明
username, err := jwtutil.ExtractCustomClaim(token, "username")

// 检查令牌是否过期
expired, err := jwtutil.IsTokenExpired(token)

// 获取令牌过期时间
expiry, err := jwtutil.GetTokenExpiry(token)
```

## API 文档

### JWTConfig 结构体

```go
type JWTConfig struct {
    SecretKey      string        // 签名密钥
    Algorithm      string        // 算法 (HS256, HS384, HS512)
    Issuer         string        // 默认发行者
    Audience       string        // 默认受众
    ExpiryDuration time.Duration // 默认过期时间
}
```

### Claims 结构体

```go
type Claims struct {
    StandardClaims
    CustomClaims map[string]interface{} // 自定义声明
}

type StandardClaims struct {
    Issuer    string // iss - 发行者
    Subject   string // sub - 主题
    Audience  string // aud - 受众
    ExpiresAt int64  // exp - 过期时间
    NotBefore int64  // nbf - 生效时间
    IssuedAt  int64  // iat - 签发时间
    ID        string // jti - JWT ID
}
```

### 主要方法

#### NewJWTConfig
```go
func NewJWTConfig(secretKey string) *JWTConfig
```
创建新的 JWT 配置，使用默认值（HS256 算法，24 小时过期）。

#### GenerateToken
```go
func (c *JWTConfig) GenerateToken(claims *Claims) (string, error)
```
生成 JWT 令牌。

#### GenerateTokenSimple
```go
func (c *JWTConfig) GenerateTokenSimple(subject string, customClaims map[string]interface{}) (string, error)
```
使用简化参数生成 JWT 令牌。

#### VerifyToken
```go
func (c *JWTConfig) VerifyToken(tokenString string) (*Claims, error)
```
验证并解析 JWT 令牌。

#### ValidateToken
```go
func (c *JWTConfig) ValidateToken(tokenString string, expectedIssuer, expectedAudience string) (*Claims, error)
```
验证令牌并检查特定声明。

#### RefreshToken
```go
func (c *JWTConfig) RefreshToken(tokenString string) (string, error)
```
刷新现有令牌，延长过期时间。

### 辅助函数

```go
// 生成访问令牌
func GenerateAccessToken(secretKey, subject string, expiryDuration time.Duration, customClaims map[string]interface{}) (string, error)

// 生成刷新令牌
func GenerateRefreshToken(secretKey, subject string, expiryDuration time.Duration) (string, error)

// 验证访问令牌
func VerifyAccessToken(secretKey, tokenString string) (*Claims, error)

// 提取主题
func ExtractSubject(tokenString string) (string, error)

// 提取自定义声明
func ExtractCustomClaim(tokenString, key string) (interface{}, error)

// 检查令牌是否过期
func IsTokenExpired(tokenString string) (bool, error)

// 获取令牌过期时间
func GetTokenExpiry(tokenString string) (time.Time, error)

// 解析令牌（不验证签名）
func ParseTokenWithoutVerify(tokenString string) (*Claims, error)
```

## 错误处理

包提供以下预定义错误：

```go
var (
    ErrInvalidToken      = errors.New("invalid token format")
    ErrTokenExpired      = errors.New("token has expired")
    ErrTokenNotValidYet  = errors.New("token not valid yet")
    ErrInvalidSignature  = errors.New("invalid token signature")
    ErrInvalidIssuer     = errors.New("invalid token issuer")
    ErrInvalidAudience   = errors.New("invalid token audience")
    ErrMissingClaims     = errors.New("missing required claims")
)
```

使用示例：

```go
claims, err := config.VerifyToken(token)
if err != nil {
    switch err {
    case jwtutil.ErrTokenExpired:
        // 令牌已过期，可以尝试刷新
        fmt.Println("Token expired")
    case jwtutil.ErrInvalidSignature:
        // 签名无效，可能是伪造的令牌
        fmt.Println("Invalid signature")
    default:
        // 其他错误
        fmt.Println("Error:", err)
    }
}
```

## 实际应用场景

### 1. Web API 身份验证

```go
// 用户登录
func loginHandler(username, password string) (string, string, error) {
    // 验证用户名和密码...
    
    // 生成访问令牌和刷新令牌
    accessToken, err := jwtutil.GenerateAccessToken(
        secretKey,
        userID,
        15*time.Minute,
        map[string]interface{}{
            "username": username,
            "role": userRole,
        },
    )
    
    refreshToken, err := jwtutil.GenerateRefreshToken(
        secretKey,
        userID,
        7*24*time.Hour,
    )
    
    return accessToken, refreshToken, nil
}

// 验证中间件
func authMiddleware(token string) (*jwtutil.Claims, error) {
    config := jwtutil.NewJWTConfig(secretKey)
    return config.VerifyToken(token)
}
```

### 2. 令牌刷新端点

```go
func refreshTokenHandler(refreshToken string) (string, error) {
    config := jwtutil.NewJWTConfig(secretKey)
    
    // 验证刷新令牌
    claims, err := config.VerifyToken(refreshToken)
    if err != nil {
        return "", err
    }
    
    // 检查是否是刷新令牌类型
    if tokenType, ok := claims.CustomClaims["type"].(string); !ok || tokenType != "refresh" {
        return "", errors.New("invalid token type")
    }
    
    // 生成新的访问令牌
    newAccessToken, err := jwtutil.GenerateAccessToken(
        secretKey,
        claims.Subject,
        15*time.Minute,
        nil,
    )
    
    return newAccessToken, nil
}
```

### 3. 微服务间认证

```go
func generateServiceToken(serviceName string) (string, error) {
    config := jwtutil.NewJWTConfig(serviceSecretKey)
    config.Issuer = "auth-service"
    config.ExpiryDuration = 1 * time.Hour
    
    claims := &jwtutil.Claims{
        StandardClaims: jwtutil.StandardClaims{
            Subject: serviceName,
            Audience: "internal-services",
        },
        CustomClaims: map[string]interface{}{
            "service": serviceName,
            "permissions": []string{"read", "write"},
        },
    }
    
    return config.GenerateToken(claims)
}

func verifyServiceToken(token string) (*jwtutil.Claims, error) {
    config := jwtutil.NewJWTConfig(serviceSecretKey)
    return config.ValidateToken(token, "auth-service", "internal-services")
}
```

## 安全建议

1. **密钥安全**: 使用强密钥（至少 32 字节），并妥善保管
2. **密钥轮换**: 定期更换密钥
3. **HTTPS**: 始终通过 HTTPS 传输令牌
4. **短期令牌**: 使用较短的访问令牌有效期
5. **刷新令牌**: 使用刷新令牌机制避免频繁重新认证
6. **令牌撤销**: 实现令牌黑名单机制（需要额外的存储层）
7. **敏感信息**: 不要在 JWT 中存储敏感信息（JWT 可以被解码）

## 支持的算法

- **HS256**: HMAC-SHA256（推荐）
- **HS384**: HMAC-SHA384
- **HS512**: HMAC-SHA512

## 性能考虑

- JWT 验证是无状态的，不需要数据库查询
- 签名验证使用 HMAC，性能优秀
- 适合高并发场景
- 如需撤销功能，需要额外的存储机制

## 许可证

本工具包遵循项目主许可证。
