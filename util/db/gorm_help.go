// Package db provides GORM utilities for database operations
package db

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// DBType 数据库类型
type DBType string

const (
	MySQL      DBType = "mysql"
	PostgreSQL DBType = "postgres"
	SQLite     DBType = "sqlite"
	SQLServer  DBType = "sqlserver"
)

// DBConfig 数据库配置
type DBConfig struct {
	Type     DBType // 数据库类型
	Host     string // 主机地址
	Port     int    // 端口
	User     string // 用户名
	Password string // 密码
	DBName   string // 数据库名
	Charset  string // 字符集（MySQL）
	SSLMode  string // SSL模式（PostgreSQL）

	// GORM 配置
	MaxIdleConns    int             // 最大空闲连接数
	MaxOpenConns    int             // 最大打开连接数
	ConnMaxLifetime time.Duration   // 连接最大生命周期
	LogLevel        logger.LogLevel // 日志级别
}

// DefaultDBConfig 返回默认数据库配置
func DefaultDBConfig() *DBConfig {
	return &DBConfig{
		Type:            MySQL,
		Host:            "localhost",
		Port:            3306,
		Charset:         "utf8mb4",
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
		LogLevel:        logger.Info,
	}
}

// Connect 连接数据库
func Connect(config *DBConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch config.Type {
	case MySQL:
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
			config.User, config.Password, config.Host, config.Port, config.DBName, config.Charset)
		dialector = mysql.Open(dsn)

	case PostgreSQL:
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)
		dialector = postgres.Open(dsn)

	case SQLite:
		dialector = sqlite.Open(config.DBName)

	case SQLServer:
		dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
			config.User, config.Password, config.Host, config.Port, config.DBName)
		dialector = sqlserver.Open(dsn)

	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: false, // 使用复数表名
		},
		Logger: logger.Default.LogMode(config.LogLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)

	return db, nil
}

// ColumnInfo 列信息
type ColumnInfo struct {
	ColumnName    string
	DataType      string
	IsNullable    string
	ColumnKey     string
	ColumnComment string
	Extra         string
}

// TableInfo 表信息
type TableInfo struct {
	TableName    string
	TableComment string
	Columns      []ColumnInfo
}

// GenConfig 代码生成配置
type GenConfig struct {
	DB            *gorm.DB    // 数据库连接
	Tables        []string    // 要生成的表名（为空则生成所有表）
	ExcludeTables []string    // 排除的表名
	OutputDir     string      // 输出目录
	PackageName   string      // 包名
	ModelPrefix   string      // 模型前缀
	ModelSuffix   string      // 模型后缀
	TagTypes      []string    // 标签类型 (gorm, json, xml, form)
	JSONTag       bool        // 是否生成 JSON 标签
	GormTag       bool        // 是否生成 GORM 标签
	TablePrefix   string      // 表前缀（去除）
	EnableComment bool        // 是否生成注释
	FileMode      os.FileMode // 文件权限
}

// DefaultGenConfig 返回默认生成配置
func DefaultGenConfig(db *gorm.DB, outputDir string) *GenConfig {
	return &GenConfig{
		DB:            db,
		OutputDir:     outputDir,
		PackageName:   "model",
		TagTypes:      []string{"gorm", "json"},
		JSONTag:       true,
		GormTag:       true,
		EnableComment: true,
		FileMode:      0644,
	}
}

// GetTables 获取所有表信息
func (gc *GenConfig) GetTables() ([]TableInfo, error) {
	var tables []TableInfo

	// 根据数据库类型查询表信息
	var tableNames []string
	var err error

	// 获取表名列表
	switch gc.DB.Dialector.Name() {
	case "mysql":
		err = gc.DB.Raw("SELECT table_name FROM information_schema.tables WHERE table_schema = DATABASE()").
			Scan(&tableNames).Error
	case "postgres":
		err = gc.DB.Raw("SELECT tablename FROM pg_tables WHERE schemaname = 'public'").
			Scan(&tableNames).Error
	case "sqlite":
		err = gc.DB.Raw("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'").
			Scan(&tableNames).Error
	default:
		return nil, fmt.Errorf("unsupported database type: %s", gc.DB.Dialector.Name())
	}

	if err != nil {
		return nil, err
	}

	// 过滤表名
	for _, tableName := range tableNames {
		// 检查是否在指定表列表中
		if len(gc.Tables) > 0 {
			found := false
			for _, t := range gc.Tables {
				if t == tableName {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// 检查是否在排除列表中
		if len(gc.ExcludeTables) > 0 {
			excluded := false
			for _, t := range gc.ExcludeTables {
				if t == tableName {
					excluded = true
					break
				}
			}
			if excluded {
				continue
			}
		}

		// 获取表信息
		tableInfo, err := gc.getTableInfo(tableName)
		if err != nil {
			return nil, err
		}
		tables = append(tables, tableInfo)
	}

	return tables, nil
}

// getTableInfo 获取表详细信息
func (gc *GenConfig) getTableInfo(tableName string) (TableInfo, error) {
	var tableInfo TableInfo
	tableInfo.TableName = tableName

	// 获取列信息
	var columns []ColumnInfo
	var err error

	switch gc.DB.Dialector.Name() {
	case "mysql":
		err = gc.DB.Raw(`
			SELECT 
				column_name, 
				data_type, 
				is_nullable, 
				column_key, 
				column_comment, 
				extra 
			FROM information_schema.columns 
			WHERE table_schema = DATABASE() AND table_name = ?
			ORDER BY ordinal_position
		`, tableName).Scan(&columns).Error

	case "postgres":
		err = gc.DB.Raw(`
			SELECT 
				column_name, 
				data_type, 
				is_nullable, 
				'' as column_key, 
				'' as column_comment, 
				'' as extra 
			FROM information_schema.columns 
			WHERE table_name = ?
			ORDER BY ordinal_position
		`, tableName).Scan(&columns).Error

	case "sqlite":
		// SQLite 使用 PRAGMA
		rows, err := gc.DB.Raw(fmt.Sprintf("PRAGMA table_info('%s')", tableName)).Rows()
		if err != nil {
			return tableInfo, err
		}
		defer rows.Close()

		for rows.Next() {
			var cid int
			var name, dataType string
			var notNull, dfltValue, pk interface{}

			if err := rows.Scan(&cid, &name, &dataType, &notNull, &dfltValue, &pk); err != nil {
				return tableInfo, err
			}

			col := ColumnInfo{
				ColumnName: name,
				DataType:   dataType,
				IsNullable: "YES",
			}
			if notNull.(int64) == 1 {
				col.IsNullable = "NO"
			}
			if pk.(int64) == 1 {
				col.ColumnKey = "PRI"
			}
			columns = append(columns, col)
		}
		err = rows.Err()
	}

	if err != nil {
		return tableInfo, err
	}

	tableInfo.Columns = columns
	return tableInfo, nil
}

// Generate 生成模型文件
func (gc *GenConfig) Generate() error {
	// 获取表信息
	tables, err := gc.GetTables()
	if err != nil {
		return fmt.Errorf("failed to get tables: %w", err)
	}

	// 创建输出目录
	if err := os.MkdirAll(gc.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// 生成每个表的模型
	for _, table := range tables {
		if err := gc.generateModel(table); err != nil {
			return fmt.Errorf("failed to generate model for table %s: %w", table.TableName, err)
		}
	}

	return nil
}

// generateModel 生成单个模型
func (gc *GenConfig) generateModel(table TableInfo) error {
	// 生成模型代码
	code, err := gc.buildModelCode(table)
	if err != nil {
		return err
	}

	// 格式化代码
	formatted, err := format.Source([]byte(code))
	if err != nil {
		// 如果格式化失败，使用原始代码
		formatted = []byte(code)
	}

	// 生成文件名
	modelName := gc.getModelName(table.TableName)
	fileName := ToSnakeCase(modelName) + ".go"
	filePath := filepath.Join(gc.OutputDir, fileName)

	// 写入文件
	if err := os.WriteFile(filePath, formatted, gc.FileMode); err != nil {
		return err
	}

	fmt.Printf("Generated: %s\n", filePath)
	return nil
}

// buildModelCode 构建模型代码
func (gc *GenConfig) buildModelCode(table TableInfo) (string, error) {
	modelName := gc.getModelName(table.TableName)

	var builder strings.Builder

	// 包声明
	builder.WriteString(fmt.Sprintf("package %s\n\n", gc.PackageName))

	// 导入
	builder.WriteString("import (\n")
	builder.WriteString("\t\"time\"\n")
	builder.WriteString(")\n\n")

	// 模型注释
	if gc.EnableComment {
		if table.TableComment != "" {
			builder.WriteString(fmt.Sprintf("// %s %s\n", modelName, table.TableComment))
		} else {
			builder.WriteString(fmt.Sprintf("// %s model\n", modelName))
		}
	}

	// 结构体定义
	builder.WriteString(fmt.Sprintf("type %s struct {\n", modelName))

	// 字段
	for _, col := range table.Columns {
		fieldName := ToCamelCase(col.ColumnName)
		fieldType := gc.mapDataType(col.DataType, col.IsNullable)

		// 字段注释
		if gc.EnableComment && col.ColumnComment != "" {
			builder.WriteString(fmt.Sprintf("\t%s %s `", fieldName, fieldType))
		} else {
			builder.WriteString(fmt.Sprintf("\t%s %s `", fieldName, fieldType))
		}

		// 标签
		tags := gc.buildTags(col)
		builder.WriteString(tags)
		builder.WriteString("`")

		if gc.EnableComment && col.ColumnComment != "" {
			builder.WriteString(fmt.Sprintf(" // %s", col.ColumnComment))
		}
		builder.WriteString("\n")
	}

	builder.WriteString("}\n\n")

	// TableName 方法
	builder.WriteString(fmt.Sprintf("// TableName returns the table name\n"))
	builder.WriteString(fmt.Sprintf("func (%s) TableName() string {\n", modelName))
	builder.WriteString(fmt.Sprintf("\treturn \"%s\"\n", table.TableName))
	builder.WriteString("}\n")

	return builder.String(), nil
}

// buildTags 构建字段标签
func (gc *GenConfig) buildTags(col ColumnInfo) string {
	var tags []string

	// GORM 标签
	if gc.GormTag {
		gormTag := fmt.Sprintf("column:%s", col.ColumnName)

		if col.ColumnKey == "PRI" {
			gormTag += ";primaryKey"
		}

		if col.Extra == "auto_increment" {
			gormTag += ";autoIncrement"
		}

		if col.IsNullable == "NO" {
			gormTag += ";not null"
		}

		tags = append(tags, fmt.Sprintf("gorm:\"%s\"", gormTag))
	}

	// JSON 标签
	if gc.JSONTag {
		jsonTag := ToSnakeCase(col.ColumnName)
		tags = append(tags, fmt.Sprintf("json:\"%s\"", jsonTag))
	}

	return strings.Join(tags, " ")
}

// mapDataType 映射数据类型
func (gc *GenConfig) mapDataType(dbType string, isNullable string) string {
	dbType = strings.ToLower(dbType)

	// 基础类型映射
	var goType string
	switch {
	case strings.Contains(dbType, "int"):
		goType = "int64"
	case strings.Contains(dbType, "float") || strings.Contains(dbType, "double") || strings.Contains(dbType, "decimal"):
		goType = "float64"
	case strings.Contains(dbType, "bool") || strings.Contains(dbType, "boolean"):
		goType = "bool"
	case strings.Contains(dbType, "time") || strings.Contains(dbType, "date"):
		goType = "time.Time"
	case strings.Contains(dbType, "char") || strings.Contains(dbType, "text"):
		goType = "string"
	case strings.Contains(dbType, "blob") || strings.Contains(dbType, "binary"):
		goType = "[]byte"
	default:
		goType = "string"
	}

	// 可空类型处理
	if isNullable == "YES" && goType != "string" && goType != "[]byte" {
		goType = "*" + goType
	}

	return goType
}

// getModelName 获取模型名称
func (gc *GenConfig) getModelName(tableName string) string {
	// 去除表前缀
	if gc.TablePrefix != "" && strings.HasPrefix(tableName, gc.TablePrefix) {
		tableName = strings.TrimPrefix(tableName, gc.TablePrefix)
	}

	// 转换为驼峰命名
	modelName := ToCamelCase(tableName)

	// 添加前缀和后缀
	if gc.ModelPrefix != "" {
		modelName = gc.ModelPrefix + modelName
	}
	if gc.ModelSuffix != "" {
		modelName = modelName + gc.ModelSuffix
	}

	return modelName
}

// ToCamelCase 转换为驼峰命名
func ToCamelCase(s string) string {
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.Title(s)
	return strings.ReplaceAll(s, " ", "")
}

// ToSnakeCase 转换为蛇形命名
func ToSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// GenerateFromDatabase 快速生成方法
func GenerateFromDatabase(config *DBConfig, outputDir string, tables ...string) error {
	// 连接数据库
	db, err := Connect(config)
	if err != nil {
		return err
	}

	// 创建生成配置
	genConfig := DefaultGenConfig(db, outputDir)
	if len(tables) > 0 {
		genConfig.Tables = tables
	}

	// 生成模型
	return genConfig.Generate()
}
