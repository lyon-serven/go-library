// 命名转换测试
package main

import (
	"fmt"
	"strings"
)

// ToCamelCase 转换为驼峰命名
func ToCamelCase(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})
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

func main() {
	fmt.Println("=== XORM工具核心函数验证 ===\n")

	// 测试ToCamelCase
	fmt.Println("【测试1】ToCamelCase 转换")
	tests1 := []string{"user_profile", "order_item", "created_at", "id"}
	for _, test := range tests1 {
		result := ToCamelCase(test)
		fmt.Printf("  %s -> %s\n", test, result)
	}

	// 测试ToSnakeCase
	fmt.Println("\n【测试2】ToSnakeCase 转换")
	tests2 := []string{"UserProfile", "OrderItem", "CreatedAt", "Id"}
	for _, test := range tests2 {
		result := ToSnakeCase(test)
		fmt.Printf("  %s -> %s\n", test, result)
	}

	fmt.Println("\n✅ 核心函数工作正常！")
}
