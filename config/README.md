# Config Package

配置管理包，提供YAML配置文件的加载、解析和管理功能。

## 主要功能

### 1. 基础配置加载
- YAML文件解析
- 结构体映射
- 类型安全转换
- 环境变量替换

### 2. 配置管理器
- 缓存机制
- 文件监听
- 热重载
- 多配置管理

### 3. 高级特性
- 默认值设置
- 配置验证
- 配置分段加载
- 多环境配置

## 快速开始

```go
package main

import (
    "mylib/config"
)

type DatabaseConfig struct {
    Host     string `yaml:"host"`
    Port     int    `yaml:"port"`
    Username string `yaml:"username"`
    Password string `yaml:"password"`
}

func main() {
    var dbConfig DatabaseConfig
    
    // 基础加载
    err := config.LoadYAMLConfig("config.yaml", &dbConfig)
    if err != nil {
        panic(err)
    }
    
    // 使用配置
    fmt.Printf("Database: %s:%d\n", dbConfig.Host, dbConfig.Port)
}
```

## 核心组件

### ConfigManager
配置管理器，提供完整的配置生命周期管理：

```go
manager := config.NewConfigManager()
defer manager.Close()

// 加载配置
err := manager.LoadConfig("app.yaml", &appConfig)

// 监听配置变化
err = manager.WatchConfig("app.yaml", &appConfig, func() {
    fmt.Println("配置已更新")
})

// 保存配置
err = manager.SaveConfig("new_config.yaml", &appConfig)
```

### 配置验证
实现 Validator 接口进行自动验证：

```go
type Config struct {
    Host string `yaml:"host"`
    Port int    `yaml:"port"`
}

func (c *Config) Validate() error {
    if c.Host == "" {
        return fmt.Errorf("host不能为空")
    }
    if c.Port <= 0 {
        return fmt.Errorf("端口必须大于0")
    }
    return nil
}

// 加载并验证
err := config.LoadYAMLConfigWithValidation("config.yaml", &cfg)
```

### 默认值支持
为配置字段设置默认值：

```go
configWithDefaults := config.NewConfigWithDefaults(&appConfig).
    SetDefault("Database.Port", 5432).
    SetDefault("Server.Timeout", "30s").
    SetDefault("Cache.TTL", time.Hour)

err := configWithDefaults.Load("config.yaml")
```

### 环境变量替换
支持在YAML中使用环境变量：

```yaml
database:
  host: "${DB_HOST}"
  port: ${DB_PORT:-5432}
  password: "${DB_PASSWORD}"
```

```go
err := config.LoadYAMLConfigWithEnv("config.yaml", &cfg)
```

### 多环境配置
基于profile的配置管理：

```go
// 加载开发环境配置
profileConfig := config.NewProfiledConfig("./configs", "development")
err := profileConfig.LoadConfig(&appConfig)

// 加载生产环境配置
profileConfig := config.NewProfiledConfig("./configs", "production")
err := profileConfig.LoadConfig(&appConfig)
```

### 配置分段
加载配置文件的特定段落：

```go
// 只加载数据库配置段
dbSection := config.NewConfigSection("database")
err := dbSection.LoadSection("app.yaml", &dbConfig)

// 加载嵌套段落
tlsSection := config.NewConfigSection("server.tls")
err := tlsSection.LoadSection("app.yaml", &tlsConfig)
```

### 文件监听
自动监听配置文件变化：

```go
watcher := config.NewFileWatcher()
defer watcher.Close()

err := watcher.Watch("config.yaml", func() {
    fmt.Println("配置文件已变化，需要重新加载")
    // 重新加载配置逻辑
})
```

## 配置文件示例

### 基础配置 (app.yaml)
```yaml
app:
  name: "MyApplication"
  version: "1.0.0"
  environment: "development"
  debug: true

database:
  host: "localhost"
  port: 5432
  username: "admin"
  password: "password123"
  database: "myapp"
  max_open_conns: 10
  max_idle_conns: 5
  timeout: "30s"

server:
  address: "0.0.0.0"
  port: 8080
  tls:
    enabled: false
    cert_file: ""
    key_file: ""
  timeouts:
    read: "10s"
    write: "10s"
    idle: "60s"

logging:
  level: "info"
  format: "json"
  output: "stdout"
  max_size: 100
  max_backups: 3
  max_age: 28
```

### 环境变量配置
```yaml
database:
  host: "${DB_HOST}"
  port: ${DB_PORT:-5432}
  username: "${DB_USER:-admin}"
  password: "${DB_PASSWORD}"

server:
  port: ${SERVER_PORT:-8080}
  debug: ${DEBUG_MODE:-false}
```

## API 参考

### 基础函数
- `LoadYAMLConfig(filePath, config)` - 加载YAML配置
- `LoadYAMLFromBytes(data, config)` - 从字节数据加载
- `SaveYAMLConfig(filePath, config)` - 保存配置到文件
- `LoadYAMLConfigWithEnv(filePath, config)` - 加载并替换环境变量
- `LoadYAMLConfigWithValidation(filePath, config)` - 加载并验证

### ConfigManager 方法
- `NewConfigManager()` - 创建配置管理器
- `LoadConfig(filePath, config)` - 加载配置
- `ReloadConfig(filePath, config)` - 重新加载配置
- `WatchConfig(filePath, config, callback)` - 监听配置变化
- `SaveConfig(filePath, config)` - 保存配置
- `GetCachedConfig(filePath)` - 获取缓存的配置
- `Close()` - 关闭管理器

### 高级功能
- `NewConfigWithDefaults(config)` - 创建带默认值的配置
- `NewProfiledConfig(profilePath, profile)` - 创建多环境配置
- `NewConfigSection(sectionPath)` - 创建配置分段加载器
- `NewFileWatcher()` - 创建文件监听器

## 最佳实践

### 1. 配置结构设计
```go
// 使用嵌套结构组织配置
type AppConfig struct {
    App      AppInfo        `yaml:"app"`
    Database DatabaseConfig `yaml:"database"`
    Server   ServerConfig   `yaml:"server"`
    Cache    CacheConfig    `yaml:"cache"`
}

// 每个配置段都有自己的结构
type DatabaseConfig struct {
    Host         string        `yaml:"host"`
    Port         int           `yaml:"port"`
    Timeout      time.Duration `yaml:"timeout"`
    MaxOpenConns int           `yaml:"max_open_conns"`
}
```

### 2. 配置验证
```go
func (c *DatabaseConfig) Validate() error {
    if c.Host == "" {
        return fmt.Errorf("数据库主机不能为空")
    }
    if c.Port <= 0 || c.Port > 65535 {
        return fmt.Errorf("数据库端口必须在1-65535之间")
    }
    if c.Timeout <= 0 {
        return fmt.Errorf("超时时间必须大于0")
    }
    return nil
}
```

### 3. 环境配置管理
```
configs/
├── default.yaml      # 默认配置
├── development.yaml  # 开发环境
├── testing.yaml      # 测试环境
├── staging.yaml      # 预发布环境
└── production.yaml   # 生产环境
```

### 4. 错误处理
```go
func loadConfig() (*AppConfig, error) {
    var config AppConfig
    
    // 尝试加载配置
    if err := config.LoadYAMLConfigWithValidation("app.yaml", &config); err != nil {
        return nil, fmt.Errorf("配置加载失败: %w", err)
    }
    
    return &config, nil
}
```

### 5. 热重载
```go
func setupConfigWatcher(configFile string, config *AppConfig) error {
    manager := config.NewConfigManager()
    
    return manager.WatchConfig(configFile, config, func() {
        log.Println("检测到配置变化，正在重新加载...")
        
        // 执行配置重载后的业务逻辑
        if err := updateServices(config); err != nil {
            log.Printf("服务更新失败: %v", err)
        }
    })
}
```

## 注意事项

1. **线程安全**: ConfigManager 是线程安全的，可以在多goroutine中使用
2. **内存管理**: 长期运行的应用应该调用 `Close()` 方法释放资源
3. **文件权限**: 确保应用有读取配置文件的权限
4. **配置敏感信息**: 生产环境中避免在配置文件中明文存储敏感信息
5. **默认值**: 为可选配置项设置合理的默认值
6. **验证规则**: 实现完整的配置验证逻辑，确保配置的正确性

## 依赖

- `gopkg.in/yaml.v3` - YAML解析库
- `github.com/fsnotify/fsnotify` - 文件系统监听库（可选）

确保在 go.mod 中添加相应依赖：

```bash
go get gopkg.in/yaml.v3
go get github.com/fsnotify/fsnotify
```