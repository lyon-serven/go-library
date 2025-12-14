# Google Authenticator 二维码生成 - 国内解决方案

## 问题说明

Google Charts API (`https://chart.googleapis.com/chart`) 在中国大陆需要翻墙才能访问，导致无法直接生成二维码。

## 解决方案汇总

### 🌟 方案1: 本地生成（强烈推荐用于生产环境）

**安装依赖:**
```bash
go get github.com/skip2/go-qrcode
```

**使用方法:**
```go
import "github.com/skip2/go-qrcode"

// 获取 otpauth:// URL
otpauthURL := ga.GetOtpauthURL(secret)

// 方式A: 保存为PNG文件
err := qrcode.WriteFile(otpauthURL, qrcode.Medium, 256, "qrcode.png")

// 方式B: 生成base64编码（用于Web）
png, err := qrcode.Encode(otpauthURL, qrcode.Medium, 256)
base64Str := base64.StdEncoding.EncodeToString(png)
dataURL := fmt.Sprintf("data:image/png;base64,%s", base64Str)

// 在HTML中使用:
// <img src="data:image/png;base64,..." alt="扫描二维码" />
```

**优点:**
- ✅ 不依赖外部API，完全本地化
- ✅ 速度快，稳定可靠
- ✅ 无网络限制问题
- ✅ 适合生产环境

**缺点:**
- ❌ 需要安装第三方库

---

### 🚀 方案2: QR Server API（快速开发推荐）

**使用方法:**
```go
// 已在代码中实现
qrURL := ga.GenerateQRCodeImageURLCN(secret)
// 返回: https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=...

// 在HTML中直接使用:
// <img src="{{qrURL}}" alt="扫描二维码" />
```

**优点:**
- ✅ 无需安装依赖
- ✅ 代码简单，开箱即用
- ✅ 国内大部分地区可访问

**缺点:**
- ❌ 依赖外部服务稳定性
- ❌ 可能存在网络延迟

---

### 🔧 方案3: 获取原始URL手动生成

**使用方法:**
```go
// 获取原始 otpauth:// URL
otpauthURL := ga.GetOtpauthURL(secret)
// 返回: otpauth://totp/MyApp:user@example.com?secret=...&issuer=MyApp...

// 复制URL到以下网站生成二维码:
```

**国内可用的在线二维码生成工具:**
- [草料二维码](https://cli.im/) - 功能强大，支持API
- [微微二维码](https://www.wwei.cn/qrcode.html) - 简单快速
- [二维工坊](https://www.2weima.com/) - 稳定可靠

**优点:**
- ✅ 完全免费
- ✅ 国内访问稳定
- ✅ 适合快速测试

**缺点:**
- ❌ 需要手动操作
- ❌ 不适合自动化场景

---

### 💻 方案4: 终端显示（开发调试推荐）

**安装依赖:**
```bash
go get github.com/mdp/qrterminal/v3
```

**使用方法:**
```go
import qrterminal "github.com/mdp/qrterminal/v3"

otpauthURL := ga.GetOtpauthURL(secret)
qrterminal.Generate(otpauthURL, qrterminal.L, os.Stdout)
// 直接在终端显示ASCII二维码
```

**优点:**
- ✅ 开发调试超级方便
- ✅ 无需浏览器查看
- ✅ 命令行工具友好

**缺点:**
- ❌ 仅适用于终端环境
- ❌ 不适合Web应用

---

## 实际应用建议

### 🏢 生产环境 (Web应用)

**推荐: 方案1 - 本地生成**

```go
// 后端API
func setupTwoFactorAuth(w http.ResponseWriter, r *http.Request) {
    // 创建Google Authenticator
    config := authutil.DefaultTOTPConfig()
    config.Issuer = "MyApp"
    config.AccountName = userEmail
    ga := authutil.NewGoogleAuthenticator(config)
    
    // 生成密钥
    secret, _ := ga.GenerateSecret()
    
    // 获取otpauth URL
    otpauthURL := ga.GetOtpauthURL(secret)
    
    // 使用本地库生成二维码
    png, _ := qrcode.Encode(otpauthURL, qrcode.Medium, 256)
    base64Str := base64.StdEncoding.EncodeToString(png)
    
    // 返回给前端
    json.NewEncoder(w).Encode(map[string]string{
        "secret": secret,
        "qrDataUrl": "data:image/png;base64," + base64Str,
    })
}
```

```html
<!-- 前端HTML -->
<div class="totp-setup">
    <h3>绑定Google身份验证器</h3>
    <img src="{{.qrDataUrl}}" alt="扫描二维码" />
    <p>密钥: <code>{{.secret}}</code></p>
    <p>1. 打开Google Authenticator应用</p>
    <p>2. 扫描上方二维码或手动输入密钥</p>
    <p>3. 输入6位验证码完成绑定</p>
</div>
```

---

### 🧪 快速开发/测试

**推荐: 方案2 - QR Server API**

```go
// 使用内置的国内API
secret, qrURL, _ := authutil.QuickGenerate("MyApp", "user@example.com")

// qrURL 可直接在浏览器打开或在<img>中使用
fmt.Printf("QR码: %s\n", qrURL)
```

---

### 🔧 命令行工具

**推荐: 方案4 - 终端显示**

```go
import qrterminal "github.com/mdp/qrterminal/v3"

// 直接在终端显示二维码
otpauthURL := ga.GetOtpauthURL(secret)
qrterminal.Generate(otpauthURL, qrterminal.L, os.Stdout)

// 用手机Google Authenticator扫描终端二维码即可
```

---

## 示例代码

查看完整示例:
```bash
# 查看所有方案的详细示例
go run util/authutil/examples/qrcode_solutions.go

# 查看TOTP基础功能示例
go run util/authutil/examples/totp_examples.go
```

---

## 性能对比

| 方案 | 速度 | 稳定性 | 网络依赖 | 复杂度 | 推荐场景 |
|------|------|--------|----------|--------|----------|
| 本地生成 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | 无 | 中 | 生产环境 |
| QR Server API | ⭐⭐⭐ | ⭐⭐⭐ | 有 | 低 | 快速开发 |
| 手动生成 | ⭐⭐ | ⭐⭐⭐⭐ | 有 | 高 | 测试验证 |
| 终端显示 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | 无 | 低 | 命令行工具 |

---

## 总结

✅ **生产环境必选**: 方案1 - 本地生成（github.com/skip2/go-qrcode）
✅ **快速开发首选**: 方案2 - QR Server API  
✅ **命令行工具**: 方案4 - 终端显示（github.com/mdp/qrterminal）

---

## 相关资源

- [skip2/go-qrcode](https://github.com/skip2/go-qrcode) - Go QR码生成库
- [mdp/qrterminal](https://github.com/mdp/qrterminal) - 终端QR码显示
- [草料二维码](https://cli.im/) - 在线二维码工具
- [Google Authenticator](https://play.google.com/store/apps/details?id=com.google.android.apps.authenticator2) - 官方应用

---

**更新日期**: 2025-12-12  
**作者**: MyLib Team
