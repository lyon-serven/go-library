# GORM模型生成工具使用指南

## 📚 功能介绍

这是一个**GORM模型自动生成工具**，可以从现有数据库表自动生成Go结构体模型文件。

### 核心功能

✅ **自动连接数据库** - 支持MySQL、PostgreSQL、SQLite、SQL Server  
✅ **读取表结构** - 自动获取表和列信息  
✅ **生成Go结构体** - 将数据库表转换为Go模型  
✅ **生成标签** - 自动生成GORM、JSON等标签  
✅ **类型映射** - 数据库类型自动映射为Go类型  
✅ **命名转换** - 表名/列名转换为驼峰命名  
✅ **批量生成** - 支持一次生成多个表  

---

## 🚀 快速开始

### 方式1: 简单使用（推荐）

```go
package main

import (
    "log"
    "github.com/lyon-serven/go-library/util/db/gorm"
)

func main() {
    // 1. 配置数据库连接
    config := gorm.DefaultDBConfig()
    config.Host = "localhost"
    config.Port = 3306
    config.User = "root"
    config.Password = "password"
    config.DBName = "your_database"
    
    // 2. 生成模型（生成所有表）
    err := gorm.GenerateFromDatabase(config, "./models")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("✅ 模型生成完成！")
}
```

### 方式2: 生成指定表

```go
// 只生成 users 和 products 表
err := gorm.GenerateFromDatabase(
    config, 
    "./models",
    "users",      // 指定表名
    "products",
    "orders",
)
```

---

## 📖 详细使用

### 1. 数据库配置

```go
// 使用默认配置
config := gorm.DefaultDBConfig()

// 或完全自定义配置
config := &gorm.DBConfig{
    Type:            gorm.MySQL,           // 数据库类型
    Host:            "localhost",          // 主机
    Port:            3306,                 // 端口
    User:            "root",               // 用户名
    Password:        "password",           // 密码
    DBName:          "mydb",               // 数据库名
    Charset:         "utf8mb4",            // 字符集
    MaxIdleConns:    10,                   // 最大空闲连接
    MaxOpenConns:    100,                  // 最大打开连接
    ConnMaxLifetime: time.Hour,            // 连接最大生命周期
    LogLevel:        logger.Info,          // 日志级别
}
```

### 2. 高级配置

```go
package main

import (
    "github.com/lyon-serven/go-library/util/db/gorm"
)

func main() {
    // 1. 连接数据库
    dbConfig := gorm.DefaultDBConfig()
    dbConfig.Host = "localhost"
    dbConfig.User = "root"
    dbConfig.Password = "password"
    dbConfig.DBName = "mydb"
    
    db, err := gorm.Connect(dbConfig)
    if err != nil {
        panic(err)
    }
    
    // 2. 创建生成配置
    genConfig := gorm.DefaultGenConfig(db, "./models")
    
    // 3. 自定义配置
    genConfig.PackageName = "entity"              // 包名
    genConfig.Tables = []string{"users", "orders"} // 只生成指定表
    genConfig.ExcludeTables = []string{"migrations"} // 排除表
    genConfig.TablePrefix = "tbl_"                // 去除表前缀
    genConfig.ModelPrefix = ""                    // 模型前缀
    genConfig.ModelSuffix = "Model"               // 模型后缀
    genConfig.JSONTag = true                      // 生成JSON标签
    genConfig.GormTag = true                      // 生成GORM标签
    genConfig.EnableComment = true                // 生成注释
    
    // 4. 执行生成
    err = genConfig.Generate()
    if err != nil {
        panic(err)
    }
    
    fmt.Println("✅ 模型生成完成！")
}
```

---

## 📝 生成示例

### 数据库表

假设有以下MySQL表：

```sql
CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '用户ID',
    username VARCHAR(50) NOT NULL COMMENT '用户名',
    email VARCHAR(100) NOT NULL COMMENT '邮箱',
    age INT NULL COMMENT '年龄',
    balance DECIMAL(10,2) DEFAULT 0.00 COMMENT '余额',
    created_at DATETIME NOT NULL COMMENT '创建时间',
    updated_at DATETIME NULL COMMENT '更新时间'
) COMMENT='用户表';
```

### 生成的Go模型

```go
package model

import (
	"time"
)

// Users 用户表
type Users struct {
	Id        int64     `gorm:"column:id;primaryKey;autoIncrement;not null" json:"id"`         // 用户ID
	Username  string    `gorm:"column:username;not null" json:"username"`                      // 用户名
	Email     string    `gorm:"column:email;not null" json:"email"`                            // 邮箱
	Age       *int64    `gorm:"column:age" json:"age"`                                         // 年龄
	Balance   float64   `gorm:"column:balance" json:"balance"`                                 // 余额
	CreatedAt time.Time `gorm:"column:created_at;not null" json:"created_at"`                  // 创建时间
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updated_at"`                          // 更新时间
}

// TableName returns the table name
func (Users) TableName() string {
	return "users"
}
```

---

## 🎨 配置选项说明

### GenConfig 配置项

| 配置项 | 类型 | 说明 | 默认值 |
|--------|------|------|--------|
| `DB` | `*gorm.DB` | 数据库连接 | 必填 |
| `Tables` | `[]string` | 要生成的表（空=所有表） | `[]` |
| `ExcludeTables` | `[]string` | 排除的表 | `[]` |
| `OutputDir` | `string` | 输出目录 | 必填 |
| `PackageName` | `string` | 包名 | `"model"` |
| `ModelPrefix` | `string` | 模型前缀 | `""` |
| `ModelSuffix` | `string` | 模型后缀 | `""` |
| `TablePrefix` | `string` | 要去除的表前缀 | `""` |
| `JSONTag` | `bool` | 生成JSON标签 | `true` |
| `GormTag` | `bool` | 生成GORM标签 | `true` |
| `EnableComment` | `bool` | 生成注释 | `true` |
| `FileMode` | `os.FileMode` | 文件权限 | `0644` |

---

## 🔧 数据类型映射

### MySQL/PostgreSQL → Go

| 数据库类型 | Go类型 | 可空时 |
|-----------|--------|--------|
| `int`, `bigint`, `tinyint` | `int64` | `*int64` |
| `float`, `double`, `decimal` | `float64` | `*float64` |
| `bool`, `boolean` | `bool` | `*bool` |
| `datetime`, `timestamp`, `date` | `time.Time` | `*time.Time` |
| `varchar`, `text`, `char` | `string` | `string` |
| `blob`, `binary` | `[]byte` | `[]byte` |

---

## 💡 实际应用场景

### 场景1: 新项目快速生成

```go
// 连接开发数据库，生成所有模型
func main() {
    config := gorm.DefaultDBConfig()
    config.Host = "dev.mysql.com"
    config.DBName = "myapp"
    config.User = "dev"
    config.Password = "dev123"
    
    // 生成所有表到 internal/model
    err := gorm.GenerateFromDatabase(config, "./internal/model")
    if err != nil {
        log.Fatal(err)
    }
}
```

### 场景2: 只生成用户相关表

```go
// 只生成用户、订单、商品表
err := gorm.GenerateFromDatabase(
    config,
    "./internal/entity",
    "users",
    "orders",
    "products",
    "order_items",
)
```

### 场景3: 排除系统表

```go
genConfig := gorm.DefaultGenConfig(db, "./models")
genConfig.ExcludeTables = []string{
    "migrations",
    "schema_migrations",
    "ar_internal_metadata",
}
err := genConfig.Generate()
```

### 场景4: 去除表前缀

```go
// 表名: tbl_users, tbl_products
// 生成: Users, Products (去除tbl_前缀)
genConfig := gorm.DefaultGenConfig(db, "./models")
genConfig.TablePrefix = "tbl_"
err := genConfig.Generate()
```

### 场景5: 添加模型后缀

```go
// 表名: users, products
// 生成: UsersModel, ProductsModel
genConfig := gorm.DefaultGenConfig(db, "./models")
genConfig.ModelSuffix = "Model"
err := genConfig.Generate()
```

---

## 🎯 完整示例

### 示例: 从MySQL生成模型

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "github.com/lyon-serven/go-library/util/db/gorm"
    "gorm.io/gorm/logger"
)

func main() {
    // 1. 数据库配置
    dbConfig := &gorm.DBConfig{
        Type:            gorm.MySQL,
        Host:            "localhost",
        Port:            3306,
        User:            "root",
        Password:        "password",
        DBName:          "blog",
        Charset:         "utf8mb4",
        MaxIdleConns:    10,
        MaxOpenConns:    100,
        ConnMaxLifetime: time.Hour,
        LogLevel:        logger.Info,
    }
    
    // 2. 连接数据库
    db, err := gorm.Connect(dbConfig)
    if err != nil {
        log.Fatalf("连接数据库失败: %v", err)
    }
    
    // 3. 配置生成器
    genConfig := &gorm.GenConfig{
        DB:            db,
        OutputDir:     "./models",
        PackageName:   "model",
        Tables:        []string{}, // 空=生成所有表
        ExcludeTables: []string{"migrations", "sessions"},
        TablePrefix:   "blog_",
        ModelSuffix:   "",
        JSONTag:       true,
        GormTag:       true,
        EnableComment: true,
        FileMode:      0644,
    }
    
    // 4. 生成模型
    fmt.Println("开始生成模型...")
    err = genConfig.Generate()
    if err != nil {
        log.Fatalf("生成失败: %v", err)
    }
    
    fmt.Println("✅ 所有模型生成完成！")
    fmt.Printf("输出目录: %s\n", genConfig.OutputDir)
}
```

### 运行结果

```bash
$ go run main.go
开始生成模型...
Generated: ./models/users.go
Generated: ./models/posts.go
Generated: ./models/comments.go
Generated: ./models/tags.go
✅ 所有模型生成完成！
输出目录: ./models
```

---

## 🌟 支持的数据库

| 数据库 | 类型常量 | 示例 |
|--------|---------|------|
| MySQL | `gorm.MySQL` | ✅ 完全支持 |
| PostgreSQL | `gorm.PostgreSQL` | ✅ 完全支持 |
| SQLite | `gorm.SQLite` | ✅ 完全支持 |
| SQL Server | `gorm.SQLServer` | ✅ 完全支持 |

---

## ⚙️ 辅助工具函数

### 命名转换

```go
// 转换为驼峰命名（首字母大写）
modelName := gorm.ToCamelCase("user_profile")
// 结果: UserProfile

// 转换为蛇形命名
jsonTag := gorm.ToSnakeCase("UserProfile")
// 结果: user_profile
```

### 直接连接数据库

```go
// 快速连接
db, err := gorm.Connect(config)
if err != nil {
    log.Fatal(err)
}

// 使用db进行其他操作
var count int64
db.Table("users").Count(&count)
```

---

## 📦 项目结构建议

```
myproject/
├── cmd/
│   └── generator/
│       └── main.go           # 生成器主程序
├── internal/
│   └── model/                # 生成的模型
│       ├── users.go
│       ├── products.go
│       └── orders.go
├── go.mod
└── go.sum
```

---

## 🔍 常见问题

### Q1: 如何处理表前缀？

```go
// 数据库表: tbl_users, tbl_products
// 想生成: Users, Products

genConfig.TablePrefix = "tbl_"
```

### Q2: 如何自定义包名？

```go
genConfig.PackageName = "entity" // 默认是 "model"
```

### Q3: 可空字段怎么处理？

工具会自动处理：
- `NOT NULL` → `int64`, `string`
- `NULL` → `*int64`, `*string` (指针类型)

### Q4: 如何只生成某几个表？

```go
genConfig.Tables = []string{"users", "orders", "products"}
```

### Q5: 如何排除系统表？

```go
genConfig.ExcludeTables = []string{"migrations", "sessions", "cache"}
```

### Q6: 生成的文件在哪里？

在 `OutputDir` 指定的目录，文件名为表名的蛇形命名 + `.go`

例如: `users.go`, `order_items.go`

---

## ⚠️ 注意事项

1. **数据库连接**: 确保数据库可访问
2. **权限**: 需要有读取表结构的权限
3. **输出目录**: 会自动创建，但要确保有写入权限
4. **覆盖文件**: 重新生成会覆盖已存在的文件
5. **代码格式化**: 自动使用 `go fmt` 格式化生成的代码

---

## 🚀 快速命令行工具

您可以创建一个命令行工具：

```go
// cmd/genmodel/main.go
package main

import (
    "flag"
    "fmt"
    "log"
    
    "github.com/lyon-serven/go-library/util/db/gorm"
)

func main() {
    // 命令行参数
    host := flag.String("host", "localhost", "数据库主机")
    port := flag.Int("port", 3306, "数据库端口")
    user := flag.String("user", "root", "用户名")
    password := flag.String("password", "", "密码")
    dbname := flag.String("db", "", "数据库名")
    output := flag.String("out", "./models", "输出目录")
    
    flag.Parse()
    
    if *dbname == "" {
        log.Fatal("请指定数据库名 -db")
    }
    
    // 配置
    config := gorm.DefaultDBConfig()
    config.Host = *host
    config.Port = *port
    config.User = *user
    config.Password = *password
    config.DBName = *dbname
    
    // 生成
    fmt.Printf("正在从数据库 %s 生成模型...\n", *dbname)
    err := gorm.GenerateFromDatabase(config, *output)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("✅ 完成！")
}
```

使用：

```bash
# 编译
go build -o genmodel cmd/genmodel/main.go

# 使用
./genmodel -host=localhost -user=root -password=123456 -db=mydb -out=./models
```

---

## 📚 总结

这是一个功能强大的GORM模型生成工具，可以：

✅ **节省时间** - 自动生成模型，无需手写  
✅ **减少错误** - 直接从数据库读取，避免手动错误  
✅ **类型安全** - 自动映射正确的Go类型  
✅ **标准化** - 统一的命名和标签格式  
✅ **灵活配置** - 支持各种自定义需求  

**推荐用于**:
- 新项目快速搭建
- 现有数据库的Go项目迁移
- 微服务项目的模型生成
- 数据库结构变更后的模型更新

开始使用吧！🚀

