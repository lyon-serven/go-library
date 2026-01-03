package main

import (
	"fmt"
	"log"
	"time"

	"gitee.com/wangsoft/go-library/util/db/xorm"
	xormLog "xorm.io/xorm/log"
)

func main() {
	fmt.Println("=== XORM 模型生成工具示例 ===\n")

	// 示例1: 快速生成（推荐）
	example1()

	// 示例2: 高级配置
	// example2()

	// 示例3: 生成指定表
	// example3()
}

// 示例1: 快速生成所有表
func example1() {
	fmt.Println("【示例1】快速生成所有表")
	fmt.Println("-------------------------------")

	// 配置数据库连接
	config := xorm.DefaultDBConfig()
	config.Host = "172.24.140.239"
	config.Port = 3306
	config.User = "root"
	config.Password = "zhifu.99.com!@#"
	config.DBName = "186_comconfigdb" // 改为你的数据库名

	// 一行代码生成所有表的模型
	err := xorm.GenerateFromDatabase(config, "./models")
	if err != nil {
		log.Printf("❌ 生成失败: %v\n", err)
		return
	}

	fmt.Println("✅ 所有模型生成完成！")
	fmt.Println("输出目录: ./models\n")
}

// 示例2: 高级配置生成
func example2() {
	fmt.Println("【示例2】高级配置生成")
	fmt.Println("-------------------------------")

	// 1. 配置数据库连接
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
		log.Printf("❌ 连接数据库失败: %v\n", err)
		return
	}
	fmt.Println("✅ 数据库连接成功")

	// 3. 配置生成器
	genConfig := &xorm.GenConfig{
		Engine:        engine,
		OutputDir:     "./models",
		PackageName:   "entity",                                    // 自定义包名
		Tables:        []string{},                                  // 空=生成所有表
		ExcludeTables: []string{"migrations", "sessions", "cache"}, // 排除表
		TablePrefix:   "blog_",                                     // 去除表前缀
		ModelPrefix:   "",                                          // 模型前缀
		ModelSuffix:   "",                                          // 模型后缀
		XormTag:       true,                                        // XORM标签
		JSONTag:       true,                                        // JSON标签
		EnableComment: true,                                        // 注释
		EnableCreated: true,                                        // created标签
		EnableUpdated: true,                                        // updated标签
		FileMode:      0644,
	}

	// 4. 生成模型
	fmt.Println("开始生成模型...")
	err = genConfig.Generate()
	if err != nil {
		log.Printf("❌ 生成失败: %v\n", err)
		return
	}

	fmt.Println("✅ 所有模型生成完成！")
	fmt.Printf("输出目录: %s\n", genConfig.OutputDir)
	fmt.Printf("包名: %s\n\n", genConfig.PackageName)
}

// 示例3: 只生成指定表
func example3() {
	fmt.Println("【示例3】只生成指定表")
	fmt.Println("-------------------------------")

	config := xorm.DefaultDBConfig()
	config.Host = "localhost"
	config.User = "root"
	config.Password = "password"
	config.DBName = "mydb"

	// 只生成 users, orders, products 三个表
	err := xorm.GenerateFromDatabase(
		config,
		"./models",
		"users",
		"orders",
		"products",
	)
	if err != nil {
		log.Printf("❌ 生成失败: %v\n", err)
		return
	}

	fmt.Println("✅ 指定表模型生成完成！")
	fmt.Println("已生成: users.go, orders.go, products.go\n")
}

// 示例4: PostgreSQL数据库
func example4() {
	fmt.Println("【示例4】PostgreSQL数据库")
	fmt.Println("-------------------------------")

	config := &xorm.DBConfig{
		Type:     xorm.PostgreSQL,
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "password",
		DBName:   "mydb",
		SSLMode:  "disable",
	}

	err := xorm.GenerateFromDatabase(config, "./models")
	if err != nil {
		log.Printf("❌ 生成失败: %v\n", err)
		return
	}

	fmt.Println("✅ PostgreSQL模型生成完成！\n")
}

// 示例5: SQLite数据库
func example5() {
	fmt.Println("【示例5】SQLite数据库")
	fmt.Println("-------------------------------")

	config := &xorm.DBConfig{
		Type:     xorm.SQLite,
		FilePath: "./data.db",
	}

	err := xorm.GenerateFromDatabase(config, "./models")
	if err != nil {
		log.Printf("❌ 生成失败: %v\n", err)
		return
	}

	fmt.Println("✅ SQLite模型生成完成！\n")
}

// 示例6: 去除表前缀
func example6() {
	fmt.Println("【示例6】去除表前缀")
	fmt.Println("-------------------------------")

	config := xorm.DefaultDBConfig()
	config.DBName = "mydb"

	engine, err := xorm.Connect(config)
	if err != nil {
		log.Fatal(err)
	}

	genConfig := xorm.DefaultGenConfig(engine, "./models")
	genConfig.TablePrefix = "tbl_" // 去除 tbl_ 前缀
	// 表: tbl_users, tbl_products
	// 生成: Users, Products

	err = genConfig.Generate()
	if err != nil {
		log.Printf("❌ 生成失败: %v\n", err)
		return
	}

	fmt.Println("✅ 已去除表前缀！")
	fmt.Println("tbl_users -> Users")
	fmt.Println("tbl_products -> Products\n")
}

// 示例7: 添加模型后缀
func example7() {
	fmt.Println("【示例7】添加模型后缀")
	fmt.Println("-------------------------------")

	config := xorm.DefaultDBConfig()
	config.DBName = "mydb"

	engine, err := xorm.Connect(config)
	if err != nil {
		log.Fatal(err)
	}

	genConfig := xorm.DefaultGenConfig(engine, "./models")
	genConfig.ModelSuffix = "Model"
	// 表: users, products
	// 生成: UsersModel, ProductsModel

	err = genConfig.Generate()
	if err != nil {
		log.Printf("❌ 生成失败: %v\n", err)
		return
	}

	fmt.Println("✅ 已添加模型后缀！")
	fmt.Println("users -> UsersModel")
	fmt.Println("products -> ProductsModel\n")
}

/*
使用说明：
1. 修改数据库配置信息（Host, User, Password, DBName）
2. 选择要运行的示例（取消注释）
3. 运行程序：go run xorm_gen_example.go
4. 查看生成的模型文件（默认在 ./models 目录）

生成的模型示例：
// users 表 -> users.go
type Users struct {
    Id        int64      `xorm:"pk autoincr bigint notnull 'id'" json:"id"`
    Username  string     `xorm:"varchar(50) notnull 'username'" json:"username"`
    Email     string     `xorm:"varchar(100) notnull 'email'" json:"email"`
    CreatedAt time.Time  `xorm:"created 'created_at'" json:"created_at"`
    UpdatedAt time.Time  `xorm:"updated 'updated_at'" json:"updated_at"`
}

func (Users) TableName() string {
    return "users"
}

XORM标签说明：
- pk: 主键
- autoincr: 自增
- notnull: 非空
- created: 创建时间（自动填充）
- updated: 更新时间（自动更新）
- 'column_name': 数据库列名
*/
