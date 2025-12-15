# MyLib - 企业级 Go 工具库

MyLib 是一个功能完备、设计精良的 Go 工具库，提供了从基础工具到企业级缓存、配置管理的全套解决方案。采用模块化设计，支持依赖注入，满足不同规模应用的需求。

## � 安装

**注意**: 发布到仓库前，请先按照 [快速开始.md](./快速开始.md) 或 [发布演示.md](./发布演示.md) 完成模块配置。

### 发布后使用

```bash
# 安装最新版本
go get github.com/yourname/mylib@latest

# 安装指定版本
go get github.com/yourname/mylib@v1.0.0
```

### 发布前准备

如果您是项目维护者，准备发布此库，请：

1. 运行配置脚本更新模块路径:
   ```powershell
   .\update-module-path.ps1 -RepoURL "github.com/yourname/mylib"
   ```

2. 推送到仓库并创建版本标签:
   ```bash
   git push origin main
   git tag v1.0.0
   git push origin v1.0.0
   ```

3. 详细步骤请查看: [快速开始.md](./快速开始.md)

## �🚀 核心特性

- **🎯 企业级架构**: 基于依赖注入的模块化设计
- **⚡ 高性能实现**: 针对 Go 语言特性优化
- **🔒 安全可靠**: 完整的加密和配置管理
- **🛠️ 功能丰富**: 覆盖常见开发场景
- **📚 完善文档**: 详细的API文档和示例

## 📦 功能模块

### 🔧 Config System - 配置管理系统 ⭐ NEW
企业级 YAML 配置管理解决方案：
- **多种加载方式**: 文件、字节、读取器多种配置源
- **环境变量替换**: 支持 `${VAR_NAME}` 格式的环境变量
- **配置验证**: 自动验证配置完整性和正确性
- **默认值支持**: 灵活的配置默认值设置
- **多环境配置**: Profile-based 环境配置管理
- **配置分段**: 支持加载配置文件的特定段落
- **热重载**: 文件监听和自动重载功能
- **缓存机制**: 配置缓存和管理器模式

**核心组件**:
- `ConfigManager` - 配置管理器
- `FileWatcher` - 文件监听器
- `ProfiledConfig` - 多环境配置
- `ConfigSection` - 配置分段加载器

### 🗄️ Cache System - 缓存管理系统
基于依赖注入的缓存管理系统，参考 ABP vNext 设计理念：
- **缓存提供程序**: Memory Cache, Redis Cache, Mock Redis (测试用)
- **序列化器**: JSON, Gob, String, Binary, Compressed JSON
- **依赖注入**: 灵活配置不同缓存实例使用不同提供程序和序列化器
- **高级特性**: TTL支持、LRU淘汰、GetOrSet模式、异步操作
- **监控支持**: 缓存统计、性能监控

**核心接口**:
- `ICacheManager` - 缓存管理器
- `ICache` - 缓存操作接口  
- `ICacheProvider` - 缓存提供程序接口
- `ICacheSerializer` - 序列化器接口

### 🛠️ Util Package - 通用工具包

#### TimeUtil - 时间工具
全面的时间处理工具集：
- **时间格式化**: 支持多种常用格式 (ISO8601, RFC3339, 自定义格式)
- **日期计算**: 今天、本周、本月、本季度、本年的开始/结束时间
- **业务日历**: 工作日计算、节假日判断、业务日期加减
- **时区转换**: 支持多时区操作和转换
- **年龄计算**: 精确的年龄和日期差计算
- **时间范围**: 时间区间操作和重叠判断
- **性能测量**: 函数执行时间测量和格式化

#### CryptoUtil - 加解密工具
企业级加密解密工具集：
- **哈希算法**: MD5, SHA1, SHA256, SHA512
- **HMAC**: 支持多种哈希算法的HMAC生成和验证
- **对称加密**: AES-256 加密解密，支持CBC模式
- **非对称加密**: RSA 密钥生成、加密解密、数字签名
- **编码工具**: Base64, Hex, URL 编码解码
- **随机生成**: 安全的随机字节和字符串生成
- **简单接口**: 基于密码的简单加解密接口

#### HttpUtil - HTTP 工具
功能丰富的HTTP客户端工具：
- **配置化客户端**: 支持超时、认证、自定义头部等配置
- **多种认证**: Bearer Token, Basic Auth 支持
- **请求构建**: 链式API构建复杂HTTP请求
- **自动重试**: 可配置的重试策略和退避算法
- **文件操作**: 文件上传下载支持
- **URL工具**: URL构建、参数操作、编码解码
- **响应处理**: 自动JSON解析、状态码判断

#### AuthUtil - 身份认证工具 ⭐ NEW
Google Authenticator兼容的两步验证（2FA）工具集：
- **TOTP生成与验证**: 基于RFC 6238标准的时间基础一次性密码
- **Google Authenticator兼容**: 完全兼容Google Authenticator应用
- **QR码生成**: 自动生成设置QR码，方便用户扫描
- **多用户管理**: 支持多用户TOTP管理和配置
- **备份恢复码**: 生成和管理备份恢复码
- **自定义配置**: 支持自定义数字位数、时间间隔等
- **时间容差**: 支持时间偏移容差验证
- **企业级安全**: 适用于网站登录、API访问、企业系统等场景

### 1. StringUtil - 字符串工具
提供字符串处理的常用函数：
- `IsEmpty(s string) bool` - 检查字符串是否为空或只包含空格
- `Reverse(s string) string` - 反转字符串
- `Contains(s, substr string) bool` - 检查是否包含子字符串
- `ContainsIgnoreCase(s, substr string) bool` - 忽略大小写的子字符串检查
- `Capitalize(s string) string` - 首字母大写
- `TrimAndLower(s string) string` - 去除空格并转小写
- `SplitAndTrim(s, delimiter string) []string` - 分割并去除空格
- `PadLeft/PadRight(s string, length int, padChar rune) string` - 填充字符串

### 2. SliceUtil - 切片工具
提供泛型切片操作函数：
- `Contains[T comparable](slice []T, element T) bool` - 检查切片是否包含元素
- `Remove[T comparable](slice []T, element T) []T` - 移除第一个匹配的元素
- `Unique[T comparable](slice []T) []T` - 去重
- `Reverse[T any](slice []T) []T` - 反转切片
- `Filter[T any](slice []T, predicate func(T) bool) []T` - 过滤
- `Map[T, R any](slice []T, mapper func(T) R) []R` - 映射转换
- `Find[T any](slice []T, predicate func(T) bool) (T, bool)` - 查找元素
- `Chunk[T any](slice []T, size int) [][]T` - 分块

### 3. FileUtil - 文件工具
提供文件操作的便捷函数：
- `Exists(path string) bool` - 检查文件/目录是否存在
- `IsFile/IsDir(path string) bool` - 检查是否为文件/目录
- `ReadJSON/WriteJSON(path string, v interface{}) error` - JSON 文件读写
- `CopyFile(src, dst string) error` - 复制文件
- `ReadLines/WriteLines(path string, lines []string) error` - 按行读写
- `WalkFiles(root string, walkFn func(path string, info fs.FileInfo) error)` - 遍历文件

### 4. Validator - 数据验证
提供各种数据验证函数：
- `IsEmail(email string) bool` - 验证邮箱格式
- `IsURL(s string) bool` - 验证 URL 格式
- `IsIP/IsIPv4/IsIPv6(s string) bool` - 验证 IP 地址
- `IsNumeric/IsInteger/IsFloat(s string) bool` - 验证数字格式
- `IsAlpha/IsAlphanumeric(s string) bool` - 验证字符类型
- `IsPhone(phone string) bool` - 验证电话号码
- `IsCreditCard(cardNumber string) bool` - 验证信用卡号（Luhn 算法）
- `IsPasswordStrong(password string) bool` - 验证密码强度
- `IsHexColor/IsUUID/IsBase64(s string) bool` - 验证特定格式

## 安装使用

### 1. 初始化项目
```bash
go mod init your-project
go get ./mylib
```

### 2. 导入使用
```go
import (
    "mylib/stringutil"
    "mylib/sliceutil" 
    "mylib/fileutil"
    "mylib/validator"
)
```

## 使用示例

### 缓存系统示例
```go
// 创建缓存管理器
manager := cache.NewCacheManager()
defer manager.Close()

// 注册组件
manager.RegisterProvider("memory", providers.NewMemoryCache(nil))
manager.RegisterSerializer("json", serializers.NewJSONSerializer())

// 配置缓存
manager.Configure("user-cache", "memory", "json")

// 使用缓存
cache := manager.GetCache("user-cache")
key := cache.NewCacheKey("user:123", "users")

// GetOrSet 模式 - 缓存穿透保护
user, err := cache.GetOrSet(ctx, key, func() (interface{}, error) {
    return fetchUserFromDatabase(userID), nil
}, options)
```

查看 `examples/` 目录获取详细的使用示例。

## 运行示例

```bash
# 运行缓存系统测试
cd examples
go run cache_quick_test.go

# 运行完整的缓存示例
go run cache_example.go

# 运行通用工具类示例
go run util_examples.go

# 运行配置管理示例
cd config/examples
go run config_examples.go

# 运行Google Authenticator示例
cd util/authutil/examples
go run totp_examples.go

# 运行综合功能演示
go run integration_demo.go

# 运行基础工具示例  
cd pkg
go run main.go
```

## 🎯 快速开始

### 配置管理快速使用
```go
import "mylib/config"

type AppConfig struct {
    Database struct {
        Host string `yaml:"host"`
        Port int    `yaml:"port"`
    } `yaml:"database"`
}

var cfg AppConfig
// 基础加载
config.LoadYAMLConfig("app.yaml", &cfg)

// 环境变量替换
config.LoadYAMLConfigWithEnv("app.yaml", &cfg)

// 配置验证
config.LoadYAMLConfigWithValidation("app.yaml", &cfg)
```

### 缓存系统快速使用
```go
import "mylib/cache"

// 创建管理器并配置缓存
manager := cache.NewCacheManager()
manager.Configure("app-cache", "memory", "json")

// 使用缓存
cache := manager.GetCache("app-cache")
cache.Set(ctx, key, data, ttl)
cache.Get(ctx, key, &result)
```

### 工具包快速使用
```go
import (
    "mylib/util/timeutil"
    "mylib/util/cryptoutil"
    "mylib/util/httputil"
    "mylib/util/authutil"
)

// 时间工具
startOfWeek := timeutil.StartOfWeek(time.Now())
age := timeutil.Age(birthday)

// 加密工具
hash := cryptoutil.SHA256Hash([]byte("data"))
encrypted, _ := cryptoutil.AESEncrypt(data, key)

// HTTP工具
client := httputil.NewClient()
resp, _ := client.Get("https://api.example.com")

// Google Authenticator (2FA)
secret, qrURL, _ := authutil.QuickGenerate("MyApp", "user@example.com")
code, _ := authutil.QuickGenerateCode(secret)
isValid := authutil.QuickVerify(secret, code)
```

## 📋 依赖管理

```bash
# 安装依赖
go get gopkg.in/yaml.v3              # YAML解析
go get github.com/fsnotify/fsnotify   # 文件监听
go get github.com/redis/go-redis/v9  # Redis客户端
```

## 🏗️ 项目结构

```
mylib/
├── cache/                    # 缓存管理系统
├── config/                   # 配置管理系统 ⭐ NEW
├── util/                     # 工具包集合
├── examples/                 # 综合示例
├── integration_demo.go       # 综合功能演示
├── go.mod                    # 模块定义
├── README.md                 # 项目说明
└── PROJECT_SUMMARY.md        # 详细项目总结
```

## ✨ 核心特性

- **🎛️ 企业级配置管理**: YAML配置、环境变量、热重载、多环境支持 ⭐ NEW
- **🔄 依赖注入架构**: 参考 ABP vNext 设计，支持灵活的组件配置
- **🚀 高性能缓存**: 内存缓存纳秒级访问，支持 TTL 和 LRU 淘汰策略  
- **🔧 多种提供程序**: Memory、Redis、Mock 等，支持分布式缓存
- **📦 多种序列化器**: JSON、Gob、String、Binary 等，满足不同场景需求
- **⚡ 泛型支持**: SliceUtil 使用 Go 1.18+ 泛型，类型安全且高效
- **🛠️ 丰富的功能**: 覆盖日常开发中的常用工具函数
- **📖 完善的文档**: 每个模块都有详细的文档和示例
- **🎯 易于使用**: 简洁的 API 设计，开箱即用
- **✅ 高质量代码**: 遵循 Go 最佳实践，提供完整测试示例

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request 来改进这个库！

1. Fork 本项目
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 📄 许可证

MIT License

---

**MyLib** - 让 Go 开发更简单、更高效！ 🚀