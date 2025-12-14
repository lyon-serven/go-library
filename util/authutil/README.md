# AuthUtil - Google Authenticator (TOTP) 工具包

AuthUtil 提供了完整的 Google Authenticator 兼容的时间基础一次性密码（TOTP）功能，支持两步验证（2FA）实现。

## 🔐 主要功能

### 核心特性
- **TOTP生成与验证**: 基于RFC 6238标准的时间基础一次性密码
- **Google Authenticator兼容**: 完全兼容Google Authenticator应用
- **QR码生成**: 自动生成设置QR码，方便用户扫描
- **多用户管理**: 支持多用户TOTP管理
- **备份恢复码**: 生成和管理备份恢复码
- **自定义配置**: 支持自定义数字位数、时间间隔等
- **时间容差**: 支持时间偏移容差验证

### 技术规格
- **算法**: HMAC-SHA1 (可配置SHA256, SHA512)
- **码长度**: 6位数字 (可配置4-8位)
- **时间步长**: 30秒 (可配置)
- **时间容差**: 支持前后时间窗口验证
- **安全性**: 使用crypto/rand生成密钥

## 🚀 快速开始

### 基础使用

```go
import "mylib/util/authutil"

// 快速生成密钥和QR码
secret, qrCodeURL, err := authutil.QuickGenerate("MyApp", "user@example.com")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("密钥: %s\n", secret)
fmt.Printf("QR码: %s\n", qrCodeURL)

// 生成当前TOTP码
code, err := authutil.QuickGenerateCode(secret)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("当前码: %s\n", code)

// 验证TOTP码
isValid := authutil.QuickVerify(secret, code)
fmt.Printf("验证结果: %v\n", isValid)
```

### 高级使用

```go
// 创建自定义配置
config := &authutil.TOTPConfig{
    Issuer:      "MyCompany",
    AccountName: "user@example.com",
    Digits:      6,
    Period:      30,
    Algorithm:   "sha1",
}

// 创建Google Authenticator实例
ga := authutil.NewGoogleAuthenticator(config)

// 生成密钥
secret, err := ga.GenerateSecret()
if err != nil {
    log.Fatal(err)
}

// 生成QR码URL
qrURL := ga.GenerateQRCodeImageURL(secret)

// 验证码（支持时间容差）
isValid := ga.VerifyCodeWithTolerance(secret, userInputCode, 1)
```

## 📱 多用户管理

```go
// 创建TOTP管理器
manager := authutil.NewTOTPManager("MyApp")

// 为用户设置TOTP
secret, qrURL, err := manager.SetupUser("user123", "john@example.com")
if err != nil {
    log.Fatal(err)
}

// 验证用户码
isValid := manager.VerifyUserCode("user123", userInputCode)

// 获取用户密钥（用于备份）
secret, exists := manager.GetUserSecret("user123")

// 移除用户TOTP
manager.RemoveUser("user123")
```

## � 二维码生成方案（国内可用）

Google Charts API 在国内需要翻墙，提供以下替代方案：

### 方案1: 使用本地库生成（推荐 ⭐）

```go
// 1. 安装依赖
// go get github.com/skip2/go-qrcode

import "github.com/skip2/go-qrcode"

// 获取 otpauth:// URL
otpauthURL := ga.GetOtpauthURL(secret)

// 方式A: 生成PNG文件
qrcode.WriteFile(otpauthURL, qrcode.Medium, 256, "qrcode.png")

// 方式B: 生成base64编码（用于Web显示）
png, _ := qrcode.Encode(otpauthURL, qrcode.Medium, 256)
base64Str := base64.StdEncoding.EncodeToString(png)
dataURL := fmt.Sprintf("data:image/png;base64,%s", base64Str)
// 在HTML中: <img src="data:image/png;base64,..." />
```

**优点**: 不依赖外部API，速度快，稳定可靠

### 方案2: 使用国内可访问的在线API

```go
// 使用 QR Server API (国内大部分地区可访问)
qrURL := ga.GenerateQRCodeImageURLCN(secret)
// 返回: https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=...

// 直接在浏览器打开或在 <img> 标签中使用
fmt.Printf("<img src=\"%s\" />", qrURL)
```

**优点**: 无需安装依赖，使用简单

### 方案3: 获取原始URL手动生成

```go
// 获取原始 otpauth:// URL
otpauthURL := ga.GetOtpauthURL(secret)
// 返回: otpauth://totp/MyApp:user@example.com?secret=...

// 复制URL到以下网站生成二维码：
// - https://cli.im/ (草料二维码)
// - https://www.wwei.cn/qrcode.html (微微二维码)
```

### 方案4: 终端显示（开发调试用）

```go
// 安装依赖: go get github.com/mdp/qrterminal/v3

import qrterminal "github.com/mdp/qrterminal/v3"

otpauthURL := ga.GetOtpauthURL(secret)
qrterminal.Generate(otpauthURL, qrterminal.L, os.Stdout)
// 直接在命令行显示二维码
```

### 完整示例

参见 `examples/qrcode_solutions.go` 文件，包含所有方案的详细示例。

```bash
go run examples/qrcode_solutions.go
```

## �🛡️ 备份恢复码

```go
// 生成备份码
backupCodes, err := authutil.GenerateBackupCodes(10)
if err != nil {
    log.Fatal(err)
}

// 显示备份码给用户
for i, code := range backupCodes.Codes {
    fmt.Printf("%d. %s\n", i+1, code)
}

// 使用备份码
used := backupCodes.UseBackupCode(userInputBackupCode)
if used {
    fmt.Println("备份码验证成功")
} else {
    fmt.Println("备份码无效或已使用")
}

// 获取未使用的备份码
unusedCodes := backupCodes.GetUnusedCodes()
```

## 🔧 API 参考

### TOTPConfig 配置

```go
type TOTPConfig struct {
    Secret      string // Base32编码的密钥
    Issuer      string // 发行方名称
    AccountName string // 账户名称
    Digits      int    // 验证码位数 (4-8)
    Period      int    // 时间步长 (秒)
    Algorithm   string // 哈希算法 (sha1, sha256, sha512)
}
```

### GoogleAuthenticator 方法

| 方法 | 描述 |
|------|------|
| `GenerateSecret()` | 生成新的密钥 |
| `GenerateCode(secret)` | 生成当前TOTP码 |
| `GenerateCodeAtTime(secret, time)` | 生成指定时间的TOTP码 |
| `VerifyCode(secret, code)` | 验证TOTP码 |
| `VerifyCodeWithTolerance(secret, code, tolerance)` | 带容差验证TOTP码 |
| `GenerateQRCodeURL(secret)` | 生成QR码URL |
| `GenerateQRCodeImageURL(secret)` | 生成QR码图片URL |
| `GetRemainingTime()` | 获取当前码剩余时间 |

### TOTPManager 方法

| 方法 | 描述 |
|------|------|
| `SetupUser(userID, accountName)` | 为用户设置TOTP |
| `VerifyUserCode(userID, code)` | 验证用户TOTP码 |
| `GetUserSecret(userID)` | 获取用户密钥 |
| `RemoveUser(userID)` | 移除用户TOTP设置 |

### BackupCodes 方法

| 方法 | 描述 |
|------|------|
| `UseBackupCode(code)` | 使用备份码 |
| `GetUnusedCodes()` | 获取未使用的备份码 |

## 💡 使用场景

### 1. 网站两步验证

```go
// 用户启用2FA
manager := authutil.NewTOTPManager("MyWebsite")
secret, qrURL, err := manager.SetupUser(userID, userEmail)

// 显示QR码给用户扫描
fmt.Printf("请用Google Authenticator扫描: %s\n", qrURL)

// 用户登录时验证
isValid := manager.VerifyUserCode(userID, userInputCode)
if isValid {
    // 登录成功
} else {
    // 需要重新输入验证码
}
```

### 2. API访问令牌

```go
// 为API访问生成TOTP密钥
config := authutil.DefaultTOTPConfig()
config.Issuer = "MyAPI"
config.AccountName = apiKeyID

ga := authutil.NewGoogleAuthenticator(config)
secret, _ := ga.GenerateSecret()

// API调用时验证
if ga.VerifyCode(secret, requestTOTPCode) {
    // 允许API访问
}
```

### 3. 企业内部系统

```go
// 员工TOTP管理
manager := authutil.NewTOTPManager("CompanySystem")

// 批量为员工设置
employees := getEmployeeList()
for _, emp := range employees {
    secret, qrURL, _ := manager.SetupUser(emp.ID, emp.Email)
    sendQRCodeToEmployee(emp.Email, qrURL)
}

// 登录验证
func authenticateEmployee(empID, totpCode string) bool {
    return manager.VerifyUserCode(empID, totpCode)
}
```

## 🔒 安全最佳实践

### 1. 密钥存储
```go
// ❌ 不要明文存储密钥
// userSecrets["user1"] = "JBSWY3DPEHPK3PXP"

// ✅ 加密存储密钥
encryptedSecret := encryptSecret(secret, userPassword)
store.Save(userID, encryptedSecret)
```

### 2. 时间同步
```go
// 确保服务器时间准确
// 使用NTP同步服务器时间
// 提供时间容差以应对客户端时间偏差

isValid := ga.VerifyCodeWithTolerance(secret, code, 1) // 允许±30秒
```

### 3. 防重放攻击
```go
// 记录已使用的TOTP码（在有效期内）
usedCodes := make(map[string]time.Time)

func verifyCodeOnce(secret, code string) bool {
    if lastUsed, exists := usedCodes[code]; exists {
        if time.Since(lastUsed) < 30*time.Second {
            return false // 码已被使用
        }
    }
    
    if ga.VerifyCode(secret, code) {
        usedCodes[code] = time.Now()
        return true
    }
    return false
}
```

### 4. 备份码安全
```go
// 安全生成和存储备份码
backupCodes, _ := authutil.GenerateBackupCodes(10)

// 哈希存储备份码
hashedCodes := make([]string, len(backupCodes.Codes))
for i, code := range backupCodes.Codes {
    hashedCodes[i] = hashPassword(code) // 使用bcrypt等
}

// 验证时比较哈希
func verifyBackupCode(inputCode string, hashedCodes []string) bool {
    for _, hash := range hashedCodes {
        if checkPassword(inputCode, hash) {
            return true
        }
    }
    return false
}
```

## 🧪 测试示例

```go
func TestTOTP(t *testing.T) {
    ga := authutil.NewGoogleAuthenticator(authutil.DefaultTOTPConfig())
    
    // 生成密钥
    secret, err := ga.GenerateSecret()
    assert.NoError(t, err)
    assert.NotEmpty(t, secret)
    
    // 生成验证码
    code, err := ga.GenerateCode(secret)
    assert.NoError(t, err)
    assert.Len(t, code, 6)
    
    // 验证码应该有效
    assert.True(t, ga.VerifyCode(secret, code))
    
    // 错误码应该无效
    assert.False(t, ga.VerifyCode(secret, "000000"))
}
```

## 🔗 相关标准

- **RFC 4226**: HOTP: An HMAC-Based One-Time Password Algorithm
- **RFC 6238**: TOTP: Time-Based One-Time Password Algorithm  
- **Google Authenticator**: Key URI Format

## 📝 注意事项

1. **时间同步**: 确保服务器时间准确，客户端时间偏差会影响验证
2. **密钥安全**: 密钥应加密存储，避免泄露
3. **容差设置**: 合理设置时间容差，平衡安全性和用户体验
4. **备份方案**: 提供备份码以防用户丢失设备
5. **防重放**: 在有效期内防止同一验证码被重复使用
6. **用户引导**: 提供清晰的设置和使用指南

## 🔄 与现有系统集成

### Web应用集成
```go
// Gin框架中间件示例
func TOTPMiddleware(manager *authutil.TOTPManager) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := getCurrentUserID(c)
        totpCode := c.GetHeader("X-TOTP-Code")
        
        if !manager.VerifyUserCode(userID, totpCode) {
            c.JSON(401, gin.H{"error": "Invalid TOTP code"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

### 数据库存储
```go
type User struct {
    ID          string    `json:"id"`
    Email       string    `json:"email"`
    TOTPSecret  string    `json:"totp_secret,omitempty"`  // 加密存储
    TOTPEnabled bool      `json:"totp_enabled"`
    BackupCodes []string  `json:"backup_codes,omitempty"` // 哈希存储
    CreatedAt   time.Time `json:"created_at"`
}
```

这个Google Authenticator工具包提供了完整的两步验证功能，可以轻松集成到任何Go应用中！