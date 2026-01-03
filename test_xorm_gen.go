package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gitee.com/wangsoft/go-library/util/db/xorm"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	fmt.Println("=== XORM 模型生成工具验证测试 ===\n")

	// 测试1: 创建SQLite测试数据库
	dbPath := "./test_xorm.db"
	if err := createTestDatabase(dbPath); err != nil {
		log.Fatalf("❌ 创建测试数据库失败: %v", err)
	}
	defer os.Remove(dbPath) // 清理测试文件

	fmt.Println("✅ 测试数据库创建成功\n")

	// 测试2: 使用快速方法生成模型
	fmt.Println("【测试1】快速生成方法")
	fmt.Println("-------------------------------")

	outputDir := "./test_models"
	defer os.RemoveAll(outputDir) // 清理生成的文件

	config := &xorm.DBConfig{
		Type:     xorm.SQLite,
		FilePath: dbPath,
	}

	err := xorm.GenerateFromDatabase(config, outputDir)
	if err != nil {
		log.Fatalf("❌ 生成模型失败: %v", err)
	}

	fmt.Println("✅ 模型生成成功！")
	fmt.Printf("输出目录: %s\n\n", outputDir)

	// 测试3: 验证生成的文件
	fmt.Println("【测试2】验证生成的文件")
	fmt.Println("-------------------------------")

	files, err := os.ReadDir(outputDir)
	if err != nil {
		log.Fatalf("❌ 读取目录失败: %v", err)
	}

	if len(files) == 0 {
		log.Fatal("❌ 没有生成任何文件")
	}

	fmt.Printf("生成了 %d 个文件:\n", len(files))
	for _, file := range files {
		fmt.Printf("  - %s\n", file.Name())

		// 读取并显示文件内容
		content, err := os.ReadFile(filepath.Join(outputDir, file.Name()))
		if err != nil {
			log.Printf("  ⚠️  读取文件失败: %v", err)
			continue
		}

		fmt.Println("\n生成的代码:")
		fmt.Println("```go")
		fmt.Println(string(content))
		fmt.Println("```\n")
	}

	// 测试4: 高级配置测试
	fmt.Println("【测试3】高级配置测试")
	fmt.Println("-------------------------------")

	engine, err := xorm.Connect(config)
	if err != nil {
		log.Fatalf("❌ 连接数据库失败: %v", err)
	}

	genConfig := &xorm.GenConfig{
		Engine:        engine,
		OutputDir:     outputDir + "_advanced",
		PackageName:   "entity",
		Tables:        []string{"users"}, // 只生成users表
		XormTag:       true,
		JSONTag:       true,
		EnableComment: true,
		EnableCreated: true,
		EnableUpdated: true,
		FileMode:      0644,
	}
	defer os.RemoveAll(outputDir + "_advanced")

	err = genConfig.Generate()
	if err != nil {
		log.Fatalf("❌ 高级配置生成失败: %v", err)
	}

	fmt.Println("✅ 高级配置生成成功！")
	fmt.Printf("包名: %s\n", genConfig.PackageName)
	fmt.Printf("只生成表: %v\n\n", genConfig.Tables)

	// 测试5: 获取表列表
	fmt.Println("【测试4】获取表列表")
	fmt.Println("-------------------------------")

	genConfig2 := xorm.DefaultGenConfig(engine, outputDir)
	tables, err := genConfig2.GetTables()
	if err != nil {
		log.Fatalf("❌ 获取表列表失败: %v", err)
	}

	fmt.Printf("数据库中的表 (%d个):\n", len(tables))
	for i, table := range tables {
		fmt.Printf("  %d. %s\n", i+1, table)
	}

	fmt.Println("\n=== ✅ 所有测试通过！===")
}

// createTestDatabase 创建SQLite测试数据库
func createTestDatabase(dbPath string) error {
	// 删除已存在的数据库
	os.Remove(dbPath)

	// 创建数据库
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	// 创建测试表
	sqls := []string{
		// 用户表
		`CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username VARCHAR(50) NOT NULL,
			email VARCHAR(100) NOT NULL,
			age INTEGER,
			balance REAL DEFAULT 0.0,
			status INTEGER DEFAULT 1,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME
		)`,

		// 订单表
		`CREATE TABLE orders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			order_no VARCHAR(50) NOT NULL,
			amount REAL NOT NULL,
			status INTEGER DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME
		)`,

		// 产品表
		`CREATE TABLE products (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name VARCHAR(100) NOT NULL,
			description TEXT,
			price REAL NOT NULL,
			stock INTEGER DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME
		)`,
	}

	for _, sqlStr := range sqls {
		if _, err := db.Exec(sqlStr); err != nil {
			return fmt.Errorf("执行SQL失败: %w", err)
		}
	}

	return nil
}
