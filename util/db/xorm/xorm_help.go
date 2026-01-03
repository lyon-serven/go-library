package xorm

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/microsoft/go-mssqldb"
	"xorm.io/xorm"
	"xorm.io/xorm/log"
)

// DatabaseType 数据库类型
type DatabaseType string

const (
	MySQL      DatabaseType = "mysql"
	PostgreSQL DatabaseType = "postgres"
	SQLite     DatabaseType = "sqlite3"
	SQLServer  DatabaseType = "mssql"
)

// DBConfig 数据库配置
type DBConfig struct {
	Type            DatabaseType  // 数据库类型
	Host            string        // 主机地址
	Port            int           // 端口号
	User            string        // 用户名
	Password        string        // 密码
	DBName          string        // 数据库名
	Charset         string        // 字符集 (MySQL)
	SSLMode         string        // SSL模式 (PostgreSQL)
	FilePath        string        // 文件路径 (SQLite)
	MaxIdleConns    int           // 最大空闲连接数
	MaxOpenConns    int           // 最大打开连接数
	ConnMaxLifetime time.Duration // 连接最大生命周期
	LogLevel        log.LogLevel  // 日志级别
}

// DefaultDBConfig 返回默认数据库配置
func DefaultDBConfig() *DBConfig {
	return &DBConfig{
		Type:            MySQL,
		Host:            "localhost",
		Port:            3306,
		User:            "root",
		Password:        "",
		DBName:          "",
		Charset:         "utf8mb4",
		SSLMode:         "disable",
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
		LogLevel:        log.LOG_INFO,
	}
}

// Connect 连接数据库
func Connect(config *DBConfig) (*xorm.Engine, error) {
	var dsn string

	switch config.Type {
	case MySQL:
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
			config.User, config.Password, config.Host, config.Port, config.DBName, config.Charset)
	case PostgreSQL:
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)
	case SQLite:
		dsn = config.FilePath
	case SQLServer:
		dsn = fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
			config.User, config.Password, config.Host, config.Port, config.DBName)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}

	engine, err := xorm.NewEngine(string(config.Type), dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 设置连接池参数
	engine.SetMaxIdleConns(config.MaxIdleConns)
	engine.SetMaxOpenConns(config.MaxOpenConns)
	engine.SetConnMaxLifetime(config.ConnMaxLifetime)

	// 设置日志级别
	engine.SetLogLevel(config.LogLevel)

	// 测试连接
	if err := engine.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return engine, nil
}

// TableInfo 表信息
type TableInfo struct {
	TableName string
	Comment   string
	Columns   []ColumnInfo
}

// ColumnInfo 列信息
type ColumnInfo struct {
	Name         string // 列名
	Type         string // 数据类型
	GoType       string // Go类型
	Comment      string // 注释
	IsPrimaryKey bool   // 是否主键
	IsAutoIncr   bool   // 是否自增
	IsNullable   bool   // 是否可空
	DefaultValue string // 默认值
	Length       int64  // 长度
}

// GenConfig 生成配置
type GenConfig struct {
	Engine        *xorm.Engine // xorm引擎
	OutputDir     string       // 输出目录
	PackageName   string       // 包名
	Tables        []string     // 要生成的表（空表示所有表）
	ExcludeTables []string     // 排除的表
	TablePrefix   string       // 表前缀（生成时会去除）
	ModelPrefix   string       // 模型前缀
	ModelSuffix   string       // 模型后缀
	XormTag       bool         // 是否生成xorm标签
	JSONTag       bool         // 是否生成json标签
	EnableComment bool         // 是否生成注释
	EnableCreated bool         // 是否使用created标签
	EnableUpdated bool         // 是否使用updated标签
	FileMode      os.FileMode  // 文件权限
}

// DefaultGenConfig 返回默认生成配置
func DefaultGenConfig(engine *xorm.Engine, outputDir string) *GenConfig {
	return &GenConfig{
		Engine:        engine,
		OutputDir:     outputDir,
		PackageName:   "model",
		XormTag:       true,
		JSONTag:       true,
		EnableComment: true,
		EnableCreated: true,
		EnableUpdated: true,
		FileMode:      0644,
	}
}

// GetTables 获取所有表名
func (c *GenConfig) GetTables() ([]string, error) {
	var tables []string

	// 获取数据库类型
	dbType := c.Engine.DriverName()

	switch dbType {
	case "mysql":
		rows, err := c.Engine.DB().DB.Query("SHOW TABLES")
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var tableName string
			if err := rows.Scan(&tableName); err != nil {
				return nil, err
			}
			tables = append(tables, tableName)
		}

	case "postgres":
		rows, err := c.Engine.DB().DB.Query(`
			SELECT tablename FROM pg_tables 
			WHERE schemaname = 'public'
		`)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var tableName string
			if err := rows.Scan(&tableName); err != nil {
				return nil, err
			}
			tables = append(tables, tableName)
		}

	case "sqlite3":
		rows, err := c.Engine.DB().DB.Query(`
			SELECT name FROM sqlite_master 
			WHERE type='table' AND name NOT LIKE 'sqlite_%'
		`)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var tableName string
			if err := rows.Scan(&tableName); err != nil {
				return nil, err
			}
			tables = append(tables, tableName)
		}

	case "mssql":
		rows, err := c.Engine.DB().DB.Query(`
			SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES 
			WHERE TABLE_TYPE = 'BASE TABLE'
		`)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var tableName string
			if err := rows.Scan(&tableName); err != nil {
				return nil, err
			}
			tables = append(tables, tableName)
		}

	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}

	return c.filterTables(tables), nil
}

// filterTables 过滤表名
func (c *GenConfig) filterTables(tables []string) []string {
	if len(c.Tables) > 0 {
		// 只生成指定的表
		tableMap := make(map[string]bool)
		for _, t := range c.Tables {
			tableMap[t] = true
		}

		var filtered []string
		for _, t := range tables {
			if tableMap[t] {
				filtered = append(filtered, t)
			}
		}
		return filtered
	}

	// 排除指定的表
	if len(c.ExcludeTables) > 0 {
		excludeMap := make(map[string]bool)
		for _, t := range c.ExcludeTables {
			excludeMap[t] = true
		}

		var filtered []string
		for _, t := range tables {
			if !excludeMap[t] {
				filtered = append(filtered, t)
			}
		}
		return filtered
	}

	return tables
}

// getTableInfo 获取表信息
func (c *GenConfig) getTableInfo(tableName string) (*TableInfo, error) {
	dbType := c.Engine.DriverName()

	switch dbType {
	case "mysql":
		return c.getTableInfoMySQL(tableName)
	case "postgres":
		return c.getTableInfoPostgreSQL(tableName)
	case "sqlite3":
		return c.getTableInfoSQLite(tableName)
	case "mssql":
		return c.getTableInfoSQLServer(tableName)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
}

// getTableInfoMySQL 获取MySQL表信息
func (c *GenConfig) getTableInfoMySQL(tableName string) (*TableInfo, error) {
	info := &TableInfo{TableName: tableName}

	// 获取表注释
	var comment sql.NullString
	err := c.Engine.DB().DB.QueryRow(`
		SELECT TABLE_COMMENT FROM information_schema.TABLES 
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?
	`, tableName).Scan(&comment)
	if err == nil && comment.Valid {
		info.Comment = comment.String
	}

	// 获取列信息
	rows, err := c.Engine.DB().DB.Query(`
		SELECT 
			COLUMN_NAME, COLUMN_TYPE, COLUMN_COMMENT, 
			IS_NULLABLE, COLUMN_KEY, EXTRA, COLUMN_DEFAULT,
			CHARACTER_MAXIMUM_LENGTH
		FROM information_schema.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
	`, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var col ColumnInfo
		var nullable, key, extra string
		var defaultVal, comment sql.NullString
		var length sql.NullInt64

		err := rows.Scan(&col.Name, &col.Type, &comment, &nullable, &key, &extra, &defaultVal, &length)
		if err != nil {
			return nil, err
		}

		col.Comment = comment.String
		col.IsNullable = nullable == "YES"
		col.IsPrimaryKey = key == "PRI"
		col.IsAutoIncr = strings.Contains(extra, "auto_increment")
		col.DefaultValue = defaultVal.String
		if length.Valid {
			col.Length = length.Int64
		}
		col.GoType = c.mapDataType(col.Type, col.IsNullable)

		info.Columns = append(info.Columns, col)
	}

	return info, nil
}

// getTableInfoPostgreSQL 获取PostgreSQL表信息
func (c *GenConfig) getTableInfoPostgreSQL(tableName string) (*TableInfo, error) {
	info := &TableInfo{TableName: tableName}

	// 获取表注释
	var comment sql.NullString
	c.Engine.DB().DB.QueryRow(`
		SELECT obj_description(oid) FROM pg_class WHERE relname = $1
	`, tableName).Scan(&comment)
	if comment.Valid {
		info.Comment = comment.String
	}

	// 获取列信息
	rows, err := c.Engine.DB().DB.Query(`
		SELECT 
			c.column_name, c.data_type, c.is_nullable,
			c.column_default, c.character_maximum_length,
			COALESCE(pgd.description, '') as comment,
			CASE WHEN pk.column_name IS NOT NULL THEN true ELSE false END as is_primary
		FROM information_schema.columns c
		LEFT JOIN pg_catalog.pg_description pgd ON pgd.objoid = (
			SELECT oid FROM pg_class WHERE relname = $1
		) AND pgd.objsubid = c.ordinal_position
		LEFT JOIN (
			SELECT ku.column_name
			FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage ku ON tc.constraint_name = ku.constraint_name
			WHERE tc.constraint_type = 'PRIMARY KEY' AND ku.table_name = $1
		) pk ON pk.column_name = c.column_name
		WHERE c.table_name = $1
		ORDER BY c.ordinal_position
	`, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var col ColumnInfo
		var nullable string
		var defaultVal, comment sql.NullString
		var length sql.NullInt64

		err := rows.Scan(&col.Name, &col.Type, &nullable, &defaultVal, &length, &comment, &col.IsPrimaryKey)
		if err != nil {
			return nil, err
		}

		col.Comment = comment.String
		col.IsNullable = nullable == "YES"
		col.IsAutoIncr = strings.Contains(defaultVal.String, "nextval")
		col.DefaultValue = defaultVal.String
		if length.Valid {
			col.Length = length.Int64
		}
		col.GoType = c.mapDataType(col.Type, col.IsNullable)

		info.Columns = append(info.Columns, col)
	}

	return info, nil
}

// getTableInfoSQLite 获取SQLite表信息
func (c *GenConfig) getTableInfoSQLite(tableName string) (*TableInfo, error) {
	info := &TableInfo{TableName: tableName}

	rows, err := c.Engine.DB().DB.Query(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var col ColumnInfo
		var cid int
		var notNull int
		var defaultVal sql.NullString
		var pk int

		err := rows.Scan(&cid, &col.Name, &col.Type, &notNull, &defaultVal, &pk)
		if err != nil {
			return nil, err
		}

		col.IsNullable = notNull == 0
		col.IsPrimaryKey = pk > 0
		col.IsAutoIncr = col.IsPrimaryKey && strings.ToUpper(col.Type) == "INTEGER"
		col.DefaultValue = defaultVal.String
		col.GoType = c.mapDataType(col.Type, col.IsNullable)

		info.Columns = append(info.Columns, col)
	}

	return info, nil
}

// getTableInfoSQLServer 获取SQL Server表信息
func (c *GenConfig) getTableInfoSQLServer(tableName string) (*TableInfo, error) {
	info := &TableInfo{TableName: tableName}

	rows, err := c.Engine.DB().DB.Query(`
		SELECT 
			c.COLUMN_NAME, c.DATA_TYPE, c.IS_NULLABLE,
			c.COLUMN_DEFAULT, c.CHARACTER_MAXIMUM_LENGTH,
			COALESCE(ep.value, '') as comment,
			CASE WHEN pk.COLUMN_NAME IS NOT NULL THEN 1 ELSE 0 END as is_primary,
			CASE WHEN c.COLUMN_DEFAULT LIKE '%IDENTITY%' THEN 1 ELSE 0 END as is_identity
		FROM INFORMATION_SCHEMA.COLUMNS c
		LEFT JOIN sys.extended_properties ep ON ep.major_id = OBJECT_ID(c.TABLE_NAME)
			AND ep.minor_id = c.ORDINAL_POSITION AND ep.name = 'MS_Description'
		LEFT JOIN (
			SELECT ku.COLUMN_NAME
			FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS tc
			JOIN INFORMATION_SCHEMA.KEY_COLUMN_USAGE ku ON tc.CONSTRAINT_NAME = ku.CONSTRAINT_NAME
			WHERE tc.CONSTRAINT_TYPE = 'PRIMARY KEY' AND ku.TABLE_NAME = ?
		) pk ON pk.COLUMN_NAME = c.COLUMN_NAME
		WHERE c.TABLE_NAME = ?
		ORDER BY c.ORDINAL_POSITION
	`, tableName, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var col ColumnInfo
		var nullable string
		var defaultVal, comment sql.NullString
		var length sql.NullInt64
		var isPrimary, isIdentity int

		err := rows.Scan(&col.Name, &col.Type, &nullable, &defaultVal, &length, &comment, &isPrimary, &isIdentity)
		if err != nil {
			return nil, err
		}

		col.Comment = comment.String
		col.IsNullable = nullable == "YES"
		col.IsPrimaryKey = isPrimary == 1
		col.IsAutoIncr = isIdentity == 1
		col.DefaultValue = defaultVal.String
		if length.Valid {
			col.Length = length.Int64
		}
		col.GoType = c.mapDataType(col.Type, col.IsNullable)

		info.Columns = append(info.Columns, col)
	}

	return info, nil
}

// mapDataType 映射数据库类型到Go类型
func (c *GenConfig) mapDataType(dbType string, nullable bool) string {
	dbType = strings.ToLower(dbType)

	// 提取基本类型（去除长度等）
	baseType := dbType
	if idx := strings.Index(dbType, "("); idx > 0 {
		baseType = dbType[:idx]
	}

	var goType string

	switch baseType {
	case "tinyint", "smallint", "mediumint", "int", "integer", "bigint":
		goType = "int64"
	case "float", "double", "decimal", "numeric", "real":
		goType = "float64"
	case "bool", "boolean":
		goType = "bool"
	case "date", "datetime", "timestamp", "time":
		goType = "time.Time"
	case "char", "varchar", "text", "tinytext", "mediumtext", "longtext", "json":
		goType = "string"
	case "blob", "tinyblob", "mediumblob", "longblob", "binary", "varbinary", "bytea":
		goType = "[]byte"
	default:
		goType = "string"
	}

	// 可空字段使用指针类型（除了string和[]byte）
	if nullable && goType != "string" && goType != "[]byte" {
		goType = "*" + goType
	}

	return goType
}

// Generate 生成模型文件
func (c *GenConfig) Generate() error {
	// 创建输出目录
	if err := os.MkdirAll(c.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// 获取表列表
	tables, err := c.GetTables()
	if err != nil {
		return fmt.Errorf("failed to get tables: %w", err)
	}

	// 生成每个表的模型
	for _, tableName := range tables {
		if err := c.generateModel(tableName); err != nil {
			return fmt.Errorf("failed to generate model for table %s: %w", tableName, err)
		}
		fmt.Printf("Generated: %s\n", filepath.Join(c.OutputDir, ToSnakeCase(tableName)+".go"))
	}

	return nil
}

// generateModel 生成单个表的模型
func (c *GenConfig) generateModel(tableName string) error {
	// 获取表信息
	tableInfo, err := c.getTableInfo(tableName)
	if err != nil {
		return err
	}

	// 生成代码
	code := c.buildModelCode(tableInfo)

	// 写入文件
	fileName := ToSnakeCase(tableName) + ".go"
	filePath := filepath.Join(c.OutputDir, fileName)

	return os.WriteFile(filePath, []byte(code), c.FileMode)
}

// buildModelCode 构建模型代码
func (c *GenConfig) buildModelCode(info *TableInfo) string {
	var sb strings.Builder

	// 包声明
	sb.WriteString(fmt.Sprintf("package %s\n\n", c.PackageName))

	// 导入
	needTime := false
	for _, col := range info.Columns {
		if strings.Contains(col.GoType, "time.Time") {
			needTime = true
			break
		}
	}
	if needTime {
		sb.WriteString("import (\n")
		sb.WriteString("\t\"time\"\n")
		sb.WriteString(")\n\n")
	}

	// 模型名称
	modelName := c.getModelName(info.TableName)

	// 注释
	if c.EnableComment && info.Comment != "" {
		sb.WriteString(fmt.Sprintf("// %s %s\n", modelName, info.Comment))
	} else if c.EnableComment {
		sb.WriteString(fmt.Sprintf("// %s 数据模型\n", modelName))
	}

	// 结构体定义
	sb.WriteString(fmt.Sprintf("type %s struct {\n", modelName))

	// 字段
	for _, col := range info.Columns {
		fieldName := ToCamelCase(col.Name)

		// 字段注释
		if c.EnableComment && col.Comment != "" {
			sb.WriteString(fmt.Sprintf("\t%s %s `%s` // %s\n",
				fieldName, col.GoType, c.buildTags(col, info.TableName), col.Comment))
		} else {
			sb.WriteString(fmt.Sprintf("\t%s %s `%s`\n",
				fieldName, col.GoType, c.buildTags(col, info.TableName)))
		}
	}

	sb.WriteString("}\n\n")

	// TableName 方法
	sb.WriteString(fmt.Sprintf("// TableName returns the table name\n"))
	sb.WriteString(fmt.Sprintf("func (%s) TableName() string {\n", modelName))
	sb.WriteString(fmt.Sprintf("\treturn \"%s\"\n", info.TableName))
	sb.WriteString("}\n")

	return sb.String()
}

// buildTags 构建字段标签
func (c *GenConfig) buildTags(col ColumnInfo, tableName string) string {
	var tags []string

	// xorm标签
	if c.XormTag {
		var xormParts []string

		// 主键
		if col.IsPrimaryKey {
			xormParts = append(xormParts, "pk")
		}

		// 自增
		if col.IsAutoIncr {
			xormParts = append(xormParts, "autoincr")
		}

		// 类型和长度
		dbType := col.Type
		if idx := strings.Index(dbType, "("); idx > 0 {
			dbType = dbType[:idx]
		}
		if col.Length > 0 {
			xormParts = append(xormParts, fmt.Sprintf("%s(%d)", dbType, col.Length))
		} else {
			xormParts = append(xormParts, dbType)
		}

		// 非空
		if !col.IsNullable {
			xormParts = append(xormParts, "notnull")
		}

		// created/updated标签
		if c.EnableCreated && (col.Name == "created_at" || col.Name == "create_time") {
			xormParts = append(xormParts, "created")
		}
		if c.EnableUpdated && (col.Name == "updated_at" || col.Name == "update_time") {
			xormParts = append(xormParts, "updated")
		}

		// 列名
		xormParts = append(xormParts, fmt.Sprintf("'%s'", col.Name))

		tags = append(tags, fmt.Sprintf("xorm:\"%s\"", strings.Join(xormParts, " ")))
	}

	// json标签
	if c.JSONTag {
		tags = append(tags, fmt.Sprintf("json:\"%s\"", ToSnakeCase(col.Name)))
	}

	return strings.Join(tags, " ")
}

// getModelName 获取模型名称
func (c *GenConfig) getModelName(tableName string) string {
	// 去除表前缀
	if c.TablePrefix != "" {
		tableName = strings.TrimPrefix(tableName, c.TablePrefix)
	}

	// 转换为驼峰命名
	modelName := ToCamelCase(tableName)

	// 添加前缀和后缀
	return c.ModelPrefix + modelName + c.ModelSuffix
}

// ToCamelCase 转换为驼峰命名（首字母大写）
func ToCamelCase(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}

	// 分割字符串
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})

	// 转换为驼峰
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}

	return strings.Join(parts, "")
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

// toLowerCamelCase 转换为小驼峰命名
func toLowerCamelCase(s string) string {
	camel := ToCamelCase(s)
	if len(camel) == 0 {
		return camel
	}
	return strings.ToLower(camel[:1]) + camel[1:]
}

// GenerateFromDatabase 从数据库生成模型（快速方法）
func GenerateFromDatabase(config *DBConfig, outputDir string, tables ...string) error {
	// 连接数据库
	engine, err := Connect(config)
	if err != nil {
		return err
	}

	// 创建生成配置
	genConfig := DefaultGenConfig(engine, outputDir)
	if len(tables) > 0 {
		genConfig.Tables = tables
	}

	// 生成模型
	return genConfig.Generate()
}
