# XORM模型生成工具使用指南

## 📚 功能介绍

这是一个**XORM模型自动生成工具**，可以从现有数据库表自动生成符合XORM规范的Go结构体模型文件。

### 核心功能

✅ **自动连接数据库** - 支持MySQL、PostgreSQL、SQLite、SQL Server  
✅ **读取表结构** - 自动获取表和列信息  
✅ **生成Go结构体** - 将数据库表转换为Go模型  
✅ **生成XORM标签** - 自动生成pk、autoincr、created、updated等标签  
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
    "gitee.com/wangsoft/go-library/util/db/xorm"
)

func main() {
    // 1. 配置数据库连接
    config := xorm.DefaultDBConfig()
    config.Host = "localhost"
    config.Port = 3306
    config.User = "root"
    config.Password = "password"
    config.DBName = "your_database"
    
    // 2. 生成模型（生成所有表）
    err := xorm.GenerateFromDatabase(config, "./models")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("✅ 模型生成完成！")
}
```

### 方式2: 生成指定表

```go
// 只生成 users 和 products 表
err := xorm.GenerateFromDatabase(
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
config := xorm.DefaultDBConfig()

// 或完全自定义配置
config := &xorm.DBConfig{
    Type:            xorm.MySQL,           // 数据库类型
    Host:            "localhost",          // 主机
    Port:            3306,                 // 端口
    User:            "root",               // 用户名
    Password:        "password",           // 密码
    DBName:          "mydb",               // 数据库名
    Charset:         "utf8mb4",            // 字符集
    SSLMode:         "disable",            // SSL模式(PostgreSQL)
    MaxIdleConns:    10,                   // 最大空闲连接
    MaxOpenConns:    100,                  // 最大打开连接
    ConnMaxLifetime: time.Hour,            // 连接最大生命周期
    LogLevel:        log.LOG_INFO,         // 日志级别
}
```

### 2. 高级配置

```go
package main

import (
    "gitee.com/wangsoft/go-library/util/db/xorm"
)

func main() {
    // 1. 连接数据库
    dbConfig := xorm.DefaultDBConfig()
    dbConfig.Host = "localhost"
    dbConfig.User = "root"
    dbConfig.Password = "password"
    dbConfig.DBName = "mydb"
    
    engine, err := xorm.Connect(dbConfig)
    if err != nil {
        panic(err)
    }
    
    // 2. 创建生成配置
    genConfig := xorm.DefaultGenConfig(engine, "./models")
    
    // 3. 自定义配置
    genConfig.PackageName = "entity"              // 包名
    genConfig.Tables = []string{"users", "orders"} // 只生成指定表
    genConfig.ExcludeTables = []string{"migrations"} // 排除表
    genConfig.TablePrefix = "tbl_"                // 去除表前缀
    genConfig.ModelPrefix = ""                    // 模型前缀
    genConfig.ModelSuffix = "Model"               // 模型后缀
    genConfig.XormTag = true                      // 生成XORM标签
    genConfig.JSONTag = true                      // 生成JSON标签
    genConfig.EnableComment = true                // 生成注释
    genConfig.EnableCreated = true                // created标签
    genConfig.EnableUpdated = true                // updated标签
    
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
    status TINYINT DEFAULT 1 COMMENT '状态',
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
	Id        int64      `xorm:"pk autoincr bigint notnull 'id'" json:"id"`                    // 用户ID
	Username  string     `xorm:"varchar(50) notnull 'username'" json:"username"`               // 用户名
	Email     string     `xorm:"varchar(100) notnull 'email'" json:"email"`                    // 邮箱
	Age       *int64     `xorm:"int 'age'" json:"age"`                                         // 年龄
	Balance   float64    `xorm:"decimal notnull 'balance'" json:"balance"`                     // 余额
	Status    int64      `xorm:"tinyint notnull 'status'" json:"status"`                       // 状态
	CreatedAt time.Time  `xorm:"datetime notnull created 'created_at'" json:"created_at"`      // 创建时间
	UpdatedAt *time.Time `xorm:"datetime updated 'updated_at'" json:"updated_at"`              // 更新时间
}

// TableName returns the table name
func (Users) TableName() string {
	return "users"
}
```

---

## 🎨 XORM标签说明

### 标签格式

```go
`xorm:"pk autoincr bigint notnull 'column_name'"`
```

### 常用标签

| 标签 | 说明 | 示例 |
|------|------|------|
| `pk` | 主键 | `xorm:"pk"` |
| `autoincr` | 自增 | `xorm:"pk autoincr"` |
| `notnull` | 非空 | `xorm:"notnull"` |
| `created` | 创建时间（自动填充） | `xorm:"created"` |
| `updated` | 更新时间（自动更新） | `xorm:"updated"` |
| `deleted` | 软删除 | `xorm:"deleted"` |
| `version` | 乐观锁版本号 | `xorm:"version"` |
| `default(value)` | 默认值 | `xorm:"default(0)"` |
| `unique` | 唯一索引 | `xorm:"unique"` |
| `index` | 普通索引 | `xorm:"index"` |
| `varchar(50)` | 字段类型和长度 | `xorm:"varchar(50)"` |
| `'column_name'` | 数据库列名 | `xorm:"'user_name'"` |

### 组合使用

```go
// 主键自增ID
Id int64 `xorm:"pk autoincr bigint notnull 'id'" json:"id"`

// 唯一索引的用户名
Username string `xorm:"varchar(50) unique notnull 'username'" json:"username"`

// 带索引的邮箱
Email string `xorm:"varchar(100) index notnull 'email'" json:"email"`

// 自动创建时间
CreatedAt time.Time `xorm:"created 'created_at'" json:"created_at"`

// 自动更新时间
UpdatedAt time.Time `xorm:"updated 'updated_at'" json:"updated_at"`

// 软删除时间
DeletedAt *time.Time `xorm:"deleted 'deleted_at'" json:"deleted_at,omitempty"`

// 乐观锁版本号
Version int `xorm:"version 'version'" json:"version"`
```

---

## 🔧 配置选项说明

### GenConfig 配置项

| 配置项 | 类型 | 说明 | 默认值 |
|--------|------|------|--------|
| `Engine` | `*xorm.Engine` | XORM引擎 | 必填 |
| `Tables` | `[]string` | 要生成的表（空=所有表） | `[]` |
| `ExcludeTables` | `[]string` | 排除的表 | `[]` |
| `OutputDir` | `string` | 输出目录 | 必填 |
| `PackageName` | `string` | 包名 | `"model"` |
| `ModelPrefix` | `string` | 模型前缀 | `""` |
| `ModelSuffix` | `string` | 模型后缀 | `""` |
| `TablePrefix` | `string` | 要去除的表前缀 | `""` |
| `XormTag` | `bool` | 生成XORM标签 | `true` |
| `JSONTag` | `bool` | 生成JSON标签 | `true` |
| `EnableComment` | `bool` | 生成注释 | `true` |
| `EnableCreated` | `bool` | created标签 | `true` |
| `EnableUpdated` | `bool` | updated标签 | `true` |
| `FileMode` | `os.FileMode` | 文件权限 | `0644` |

---

## 🔍 数据类型映射

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
    config := xorm.DefaultDBConfig()
    config.Host = "dev.mysql.com"
    config.DBName = "myapp"
    config.User = "dev"
    config.Password = "dev123"
    
    // 生成所有表到 internal/model
    err := xorm.GenerateFromDatabase(config, "./internal/model")
    if err != nil {
        log.Fatal(err)
    }
}
```

### 场景2: 只生成用户相关表

```go
// 只生成用户、订单、商品表
err := xorm.GenerateFromDatabase(
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
genConfig := xorm.DefaultGenConfig(engine, "./models")
genConfig.ExcludeTables = []string{
    "migrations",
    "schema_migrations",
    "sessions",
}
err := genConfig.Generate()
```

### 场景4: 去除表前缀

```go
// 表名: tbl_users, tbl_products
// 生成: Users, Products (去除tbl_前缀)
genConfig := xorm.DefaultGenConfig(engine, "./models")
genConfig.TablePrefix = "tbl_"
err := genConfig.Generate()
```

### 场景5: 添加模型后缀

```go
// 表名: users, products
// 生成: UsersModel, ProductsModel
genConfig := xorm.DefaultGenConfig(engine, "./models")
genConfig.ModelSuffix = "Model"
err := genConfig.Generate()
```

### 场景6: 使用created/updated自动时间

```go
// 启用自动时间标签
genConfig := xorm.DefaultGenConfig(engine, "./models")
genConfig.EnableCreated = true  // created_at字段自动添加created标签
genConfig.EnableUpdated = true  // updated_at字段自动添加updated标签
err := genConfig.Generate()

// 使用时XORM会自动填充时间
user := &User{Username: "test"}
engine.Insert(user)  // created_at自动填充

user.Username = "updated"
engine.Update(user)  // updated_at自动更新
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
    
    "gitee.com/wangsoft/go-library/util/db/xorm"
    xormLog "xorm.io/xorm/log"
)

func main() {
    // 1. 数据库配置
    dbConfig := &xorm.DBConfig{
        Type:            xorm.MySQL,
        Host:            "localhost",
        Port:            3306,
        User:            "root",
        Password:        "password",
        DBName:          "blog",
        Charset:         "utf8mb4",
        MaxIdleConns:    10,
        MaxOpenConns:    100,
        ConnMaxLifetime: time.Hour,
        LogLevel:        xormLog.LOG_INFO,
    }
    
    // 2. 连接数据库
    engine, err := xorm.Connect(dbConfig)
    if err != nil {
        log.Fatalf("连接数据库失败: %v", err)
    }
    
    // 3. 配置生成器
    genConfig := &xorm.GenConfig{
        Engine:        engine,
        OutputDir:     "./models",
        PackageName:   "model",
        Tables:        []string{}, // 空=生成所有表
        ExcludeTables: []string{"migrations", "sessions"},
        TablePrefix:   "blog_",
        ModelSuffix:   "",
        XormTag:       true,
        JSONTag:       true,
        EnableComment: true,
        EnableCreated: true,
        EnableUpdated: true,
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
| MySQL | `xorm.MySQL` | ✅ 完全支持 |
| PostgreSQL | `xorm.PostgreSQL` | ✅ 完全支持 |
| SQLite | `xorm.SQLite` | ✅ 完全支持 |
| SQL Server | `xorm.SQLServer` | ✅ 完全支持 |

---

## ⚙️ 辅助工具函数

### 命名转换

```go
// 转换为驼峰命名（首字母大写）
modelName := xorm.ToCamelCase("user_profile")
// 结果: UserProfile

// 转换为蛇形命名
jsonTag := xorm.ToSnakeCase("UserProfile")
// 结果: user_profile
```

### 直接连接数据库

```go
// 快速连接
engine, err := xorm.Connect(config)
if err != nil {
    log.Fatal(err)
}

// 使用engine进行其他操作
var count int64
count, err = engine.Table("users").Count()
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

### Q6: created/updated标签如何工作？

```go
// 生成的模型
type User struct {
    CreatedAt time.Time `xorm:"created"` // 插入时自动填充
    UpdatedAt time.Time `xorm:"updated"` // 更新时自动更新
}

// 使用
user := &User{Username: "test"}
engine.Insert(user)  // CreatedAt自动设置为当前时间

user.Username = "new"
engine.Update(user)  // UpdatedAt自动更新为当前时间
```

### Q7: 如何使用软删除？

手动添加deleted标签：

```go
type User struct {
    DeletedAt *time.Time `xorm:"deleted"`
}

// 删除时不会真正删除，只是设置DeletedAt
engine.Delete(&User{Id: 1})

// 查询时自动过滤已删除数据
engine.Find(&users)

// 查询包括已删除数据
engine.Unscoped().Find(&users)
```

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
    
    "gitee.com/wangsoft/go-library/util/db/xorm"
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
    config := xorm.DefaultDBConfig()
    config.Host = *host
    config.Port = *port
    config.User = *user
    config.Password = *password
    config.DBName = *dbname
    
    // 生成
    fmt.Printf("正在从数据库 %s 生成模型...\n", *dbname)
    err := xorm.GenerateFromDatabase(config, *output)
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

## 🆚 与GORM版本对比

### 主要区别

| 特性 | GORM | XORM |
|------|------|------|
| 标签格式 | `gorm:"column:name"` | `xorm:"'name'"` |
| 主键 | `gorm:"primaryKey"` | `xorm:"pk"` |
| 自增 | `gorm:"autoIncrement"` | `xorm:"autoincr"` |
| 创建时间 | 需要插件 | `xorm:"created"` |
| 更新时间 | 需要插件 | `xorm:"updated"` |
| 软删除 | `gorm:"softDelete"` | `xorm:"deleted"` |

### 生成示例对比

**GORM生成：**
```go
type User struct {
    Id   int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
    Name string `gorm:"column:name;not null" json:"name"`
}
```

**XORM生成：**
```go
type User struct {
    Id   int64  `xorm:"pk autoincr bigint notnull 'id'" json:"id"`
    Name string `xorm:"varchar(50) notnull 'name'" json:"name"`
}
```

---

## 📚 总结

这是一个功能强大的XORM模型生成工具，可以：

✅ **节省时间** - 自动生成模型，无需手写  
✅ **减少错误** - 直接从数据库读取，避免手动错误  
✅ **XORM规范** - 完全符合XORM标签规范  
✅ **自动时间** - created/updated标签自动管理时间  
✅ **类型安全** - 自动映射正确的Go类型  
✅ **标准化** - 统一的命名和标签格式  
✅ **灵活配置** - 支持各种自定义需求  

**推荐用于**:
- 使用XORM的新项目快速搭建
- 现有数据库的Go项目迁移
- 微服务项目的模型生成
- 数据库结构变更后的模型更新

开始使用吧！🚀

