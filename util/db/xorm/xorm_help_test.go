package xorm

import (
	"os"
	"testing"
)

// TestToCamelCase 测试驼峰命名转换
func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"user_profile", "UserProfile"},
		{"order_item", "OrderItem"},
		{"created_at", "CreatedAt"},
		{"id", "Id"},
		{"user_name", "UserName"},
	}

	for _, tt := range tests {
		result := ToCamelCase(tt.input)
		if result != tt.expected {
			t.Errorf("ToCamelCase(%s) = %s; want %s", tt.input, result, tt.expected)
		}
	}
}

// TestToSnakeCase 测试蛇形命名转换
func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"UserProfile", "user_profile"},
		{"OrderItem", "order_item"},
		{"CreatedAt", "created_at"},
		{"Id", "id"},
		{"UserName", "user_name"},
	}

	for _, tt := range tests {
		result := ToSnakeCase(tt.input)
		if result != tt.expected {
			t.Errorf("ToSnakeCase(%s) = %s; want %s", tt.input, result, tt.expected)
		}
	}
}

// TestMapDataType 测试数据类型映射
func TestMapDataType(t *testing.T) {
	config := DefaultGenConfig(nil, "")

	tests := []struct {
		dbType   string
		nullable bool
		expected string
	}{
		{"int", false, "int64"},
		{"int", true, "*int64"},
		{"varchar(50)", false, "string"},
		{"varchar(50)", true, "string"}, // string类型可空也不用指针
		{"datetime", false, "time.Time"},
		{"datetime", true, "*time.Time"},
		{"float", false, "float64"},
		{"float", true, "*float64"},
		{"bool", false, "bool"},
		{"bool", true, "*bool"},
	}

	for _, tt := range tests {
		result := config.mapDataType(tt.dbType, tt.nullable)
		if result != tt.expected {
			t.Errorf("mapDataType(%s, %v) = %s; want %s", tt.dbType, tt.nullable, result, tt.expected)
		}
	}
}

// TestDefaultDBConfig 测试默认配置
func TestDefaultDBConfig(t *testing.T) {
	config := DefaultDBConfig()

	if config.Type != MySQL {
		t.Errorf("Default Type = %s; want %s", config.Type, MySQL)
	}

	if config.Host != "localhost" {
		t.Errorf("Default Host = %s; want localhost", config.Host)
	}

	if config.Port != 3306 {
		t.Errorf("Default Port = %d; want 3306", config.Port)
	}

	if config.Charset != "utf8mb4" {
		t.Errorf("Default Charset = %s; want utf8mb4", config.Charset)
	}
}

// TestFilterTables 测试表过滤
func TestFilterTables(t *testing.T) {
	config := DefaultGenConfig(nil, "")
	allTables := []string{"users", "orders", "products", "migrations", "sessions"}

	// 测试1: 指定要生成的表
	config.Tables = []string{"users", "orders"}
	filtered := config.filterTables(allTables)
	if len(filtered) != 2 {
		t.Errorf("Filter with Tables: got %d tables, want 2", len(filtered))
	}

	// 测试2: 排除某些表
	config.Tables = []string{}
	config.ExcludeTables = []string{"migrations", "sessions"}
	filtered = config.filterTables(allTables)
	if len(filtered) != 3 {
		t.Errorf("Filter with ExcludeTables: got %d tables, want 3", len(filtered))
	}

	// 测试3: 不过滤
	config.Tables = []string{}
	config.ExcludeTables = []string{}
	filtered = config.filterTables(allTables)
	if len(filtered) != 5 {
		t.Errorf("No filter: got %d tables, want 5", len(filtered))
	}
}

// TestGetModelName 测试模型名称生成
func TestGetModelName(t *testing.T) {
	config := DefaultGenConfig(nil, "")

	tests := []struct {
		tableName   string
		prefix      string
		modelPrefix string
		modelSuffix string
		expected    string
	}{
		{"users", "", "", "", "Users"},
		{"tbl_users", "tbl_", "", "", "Users"},
		{"users", "", "Model", "", "ModelUsers"},
		{"users", "", "", "Model", "UsersModel"},
		{"tbl_users", "tbl_", "My", "Entity", "MyUsersEntity"},
	}

	for _, tt := range tests {
		config.TablePrefix = tt.prefix
		config.ModelPrefix = tt.modelPrefix
		config.ModelSuffix = tt.modelSuffix
		result := config.getModelName(tt.tableName)
		if result != tt.expected {
			t.Errorf("getModelName(%s) with prefix=%s, modelPrefix=%s, modelSuffix=%s = %s; want %s",
				tt.tableName, tt.prefix, tt.modelPrefix, tt.modelSuffix, result, tt.expected)
		}
	}
}

// TestBuildTags 测试标签生成
func TestBuildTags(t *testing.T) {
	config := DefaultGenConfig(nil, "")
	config.XormTag = true
	config.JSONTag = true
	config.EnableCreated = true
	config.EnableUpdated = true

	// 测试主键自增字段
	col := ColumnInfo{
		Name:         "id",
		Type:         "bigint",
		IsPrimaryKey: true,
		IsAutoIncr:   true,
		IsNullable:   false,
		Length:       0,
	}

	tags := config.buildTags(col, "users")

	if tags == "" {
		t.Error("buildTags returned empty string")
	}

	// 检查是否包含关键标签
	if !contains(tags, "pk") {
		t.Error("Tags should contain 'pk'")
	}
	if !contains(tags, "autoincr") {
		t.Error("Tags should contain 'autoincr'")
	}
	if !contains(tags, "notnull") {
		t.Error("Tags should contain 'notnull'")
	}
	if !contains(tags, "json:") {
		t.Error("Tags should contain 'json:'")
	}

	// 测试created_at字段
	col2 := ColumnInfo{
		Name:       "created_at",
		Type:       "datetime",
		IsNullable: false,
	}

	tags2 := config.buildTags(col2, "users")
	if !contains(tags2, "created") {
		t.Error("created_at field should contain 'created' tag")
	}

	// 测试updated_at字段
	col3 := ColumnInfo{
		Name:       "updated_at",
		Type:       "datetime",
		IsNullable: true,
	}

	tags3 := config.buildTags(col3, "users")
	if !contains(tags3, "updated") {
		t.Error("updated_at field should contain 'updated' tag")
	}
}

// 辅助函数：检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestBuildModelCode 测试模型代码生成
func TestBuildModelCode(t *testing.T) {
	config := DefaultGenConfig(nil, "")
	config.PackageName = "model"
	config.XormTag = true
	config.JSONTag = true
	config.EnableComment = true

	tableInfo := &TableInfo{
		TableName: "users",
		Comment:   "用户表",
		Columns: []ColumnInfo{
			{
				Name:         "id",
				Type:         "bigint",
				GoType:       "int64",
				Comment:      "用户ID",
				IsPrimaryKey: true,
				IsAutoIncr:   true,
				IsNullable:   false,
			},
			{
				Name:       "username",
				Type:       "varchar(50)",
				GoType:     "string",
				Comment:    "用户名",
				IsNullable: false,
				Length:     50,
			},
		},
	}

	code := config.buildModelCode(tableInfo)

	// 检查生成的代码
	if !contains(code, "package model") {
		t.Error("Generated code should contain 'package model'")
	}
	if !contains(code, "type Users struct") {
		t.Error("Generated code should contain 'type Users struct'")
	}
	if !contains(code, "func (Users) TableName()") {
		t.Error("Generated code should contain TableName method")
	}
	if !contains(code, "Id") {
		t.Error("Generated code should contain 'Id' field")
	}
	if !contains(code, "Username") {
		t.Error("Generated code should contain 'Username' field")
	}

	t.Logf("Generated code:\n%s", code)
}

// TestDefaultGenConfig 测试默认生成配置
func TestDefaultGenConfig(t *testing.T) {
	config := DefaultGenConfig(nil, "./models")

	if config.OutputDir != "./models" {
		t.Errorf("OutputDir = %s; want ./models", config.OutputDir)
	}

	if config.PackageName != "model" {
		t.Errorf("PackageName = %s; want model", config.PackageName)
	}

	if !config.XormTag {
		t.Error("XormTag should be true")
	}

	if !config.JSONTag {
		t.Error("JSONTag should be true")
	}

	if !config.EnableComment {
		t.Error("EnableComment should be true")
	}

	if !config.EnableCreated {
		t.Error("EnableCreated should be true")
	}

	if !config.EnableUpdated {
		t.Error("EnableUpdated should be true")
	}

	if config.FileMode != 0644 {
		t.Errorf("FileMode = %o; want 0644", config.FileMode)
	}
}

// 清理测试文件
func cleanup() {
	os.RemoveAll("./test_models")
	os.RemoveAll("./models")
	os.Remove("./test.db")
}
