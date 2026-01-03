# Examples 示例目录

本目录包含了 go-library 各个功能模块的使用示例。每个示例都在独立的包（文件夹）中，可以单独运行。

## 📁 目录结构

```
examples/
├── cache/                          # 缓存系统示例
│   ├── cache_basic_example.go      # 基础缓存使用
│   └── multilevel_cache_example.go # 多级缓存示例
├── db/                             # 数据库工具示例
│   ├── gorm/                       # GORM 示例
│   │   └── gorm_gen_example.go     # GORM 模型生成工具
│   └── xorm/                       # XORM 示例
│       └── xorm_gen_example.go     # XORM 模型生成工具
├── jwtutil/                        # JWT 工具示例
│   └── jwtutil_example.go          # JWT 认证示例
└── README.md                       # 本文件
```

## 🚀 运行示例

每个示例都可以独立运行，进入对应目录后执行 `go run` 命令：

### 1. 缓存系统示例

#### 基础缓存示例
```bash
cd examples/cache
go run cache_basic_example.go
```

**功能展示：**
- Set/Get 基本操作
- Exists 检查键是否存在
- GetOrSet 模式（缓存未命中时执行加载函数）
- 不同的序列化器（JSON、String）
- 多个缓存实例管理
- Delete 删除操作

#### 多级缓存示例
```bash
cd examples/cache
go run multilevel_cache_example.go
```

**功能展示：**
- L1（内存）→ L2（Redis）→ L3（数据库）三级缓存
- 自动缓存降级和提升
- 异步写入和写回策略
- 性能指标统计
- 缓存命中率分析

---

### 2. 数据库模型生成示例

#### GORM 模型生成
```bash
cd examples/db/gorm
go run gorm_gen_example.go
```

**功能展示：**
- 快速生成所有表的模型
- 高级配置（自定义包名、表过滤、前缀处理）
- 生成指定表
- 支持 MySQL/PostgreSQL/SQLite/SQL Server
- 自动生成 GORM 标签

**示例配置：**
```go
// 修改数据库连接信息
config.Host = "localhost"
config.Port = 3306
config.User = "root"
config.Password = "password"
config.DBName = "your_database"
```

#### XORM 模型生成
```bash
cd examples/db/xorm
go run xorm_gen_example.go
```

**功能展示：**
- 快速生成所有表的模型
- 高级配置（自定义包名、表过滤、前缀处理）
- 生成指定表
- 支持 MySQL/PostgreSQL/SQLite/SQL Server
- 自动生成 XORM 标签（pk、autoincr、created、updated）

**XORM 特色：**
- `created` 标签：插入时自动填充创建时间
- `updated` 标签：更新时自动更新时间

---

### 3. JWT 认证示例

```bash
cd examples/jwtutil
go run jwtutil_example.go
```

**功能展示：**
- 基本令牌生成和验证
- 访问令牌（Access Token）和刷新令牌（Refresh Token）
- 自定义声明（Custom Claims）
- 令牌验证和自定义检查
- 令牌刷新
- 提取令牌信息
- 错误处理（过期、无效、签名错误）

**支持的算法：**
- HS256（默认）
- HS384
- HS512

---

## 📝 示例说明

### 为什么每个示例在独立的包中？

之前所有示例都在 `examples` 根目录下，导致多个 `package main` 和 `main()` 函数冲突，无法编译运行。

**现在的结构：**
```
examples/
├── cache/          # 独立包
│   └── main.go
├── jwtutil/        # 独立包
│   └── main.go
└── db/
    ├── gorm/       # 独立包
    │   └── main.go
    └── xorm/       # 独立包
        └── main.go
```

每个文件夹是一个独立的 `package main`，可以单独编译和运行。

---

## 🎯 快速测试所有示例

### 测试脚本（Linux/macOS）

```bash
#!/bin/bash

echo "测试 JWT 工具..."
cd examples/jwtutil && go run jwtutil_example.go

echo -e "\n测试基础缓存..."
cd ../cache && go run cache_basic_example.go

echo -e "\n测试多级缓存..."
go run multilevel_cache_example.go

echo -e "\n✅ 所有示例测试完成！"
```

### 测试脚本（Windows PowerShell）

```powershell
# test-examples.ps1

Write-Host "测试 JWT 工具..." -ForegroundColor Green
Set-Location examples/jwtutil
go run jwtutil_example.go

Write-Host "`n测试基础缓存..." -ForegroundColor Green
Set-Location ../cache
go run cache_basic_example.go

Write-Host "`n测试多级缓存..." -ForegroundColor Green
go run multilevel_cache_example.go

Write-Host "`n✅ 所有示例测试完成！" -ForegroundColor Green
```

---

## 📚 相关文档

- **JWT 工具文档**: `util/jwtutil/README.md`
- **缓存系统文档**: `cache/README.md`
- **多级缓存设计**: `cache/MULTILEVEL_CACHE.md`
- **GORM 工具文档**: `util/db/gorm/README.md`
- **XORM 工具文档**: `util/db/xorm/README.md`

---

## 💡 使用建议

### 1. 从简单示例开始

推荐顺序：
1. JWT 工具示例（最简单）
2. 基础缓存示例
3. 数据库模型生成示例
4. 多级缓存示例（最复杂）

### 2. 修改配置

运行前请修改示例中的配置：
- 数据库连接信息（host、user、password、dbname）
- JWT 密钥
- Redis 连接信息（如果使用）

### 3. 查看输出

每个示例都有详细的输出说明，帮助理解执行过程。

### 4. 复制到项目

示例代码可以直接复制到您的项目中使用，只需调整配置即可。

---

## 🔧 故障排除

### 问题1: 导入包失败

```bash
go: cannot find module providing package gitee.com/wangsoft/go-library
```

**解决方案：**
```bash
# 在项目根目录执行
go mod tidy
```

### 问题2: 数据库连接失败

```
连接数据库失败: dial tcp 127.0.0.1:3306: connect: connection refused
```

**解决方案：**
1. 确保数据库服务已启动
2. 检查数据库连接信息是否正确
3. 对于 SQLite，确保文件路径正确

### 问题3: Go 版本不兼容

```
compile: version "go1.22.6" does not match go tool version "go1.23.9"
```

**解决方案：**
```bash
go clean -cache
go mod tidy
```

---

## 📧 反馈

如果您有任何问题或建议，请提交 Issue 或 Pull Request。

---

## 📄 许可证

本项目采用 MIT 许可证。

