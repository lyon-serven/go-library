package main

import (
	"fmt"
	"time"

	"github.com/lyon-serven/go-library/util/jwtutil"
)

func main() {
	fmt.Println("=== JWT Utility 示例 ===\n")

	// 示例 1: 基本使用
	example1_BasicUsage()

	// 示例 2: 访问令牌和刷新令牌
	example2_AccessAndRefreshTokens()

	// 示例 3: 令牌验证和自定义检查
	//example3_TokenValidation()

	// 示例 4: 令牌刷新
	example4_TokenRefresh()

	// 示例 5: 提取令牌信息
	example5_ExtractTokenInfo()

	// 示例 6: 错误处理
	example6_ErrorHandling()
}

func example1_BasicUsage() {
	fmt.Println("--- 示例 1: 基本使用 ---")

	// 创建 JWT 配置
	config := jwtutil.NewJWTConfig("my-super-secret-key-12345")
	config.Issuer = "my-app"
	config.ExpiryDuration = 2 * time.Hour

	// 生成令牌
	claims := &jwtutil.Claims{
		StandardClaims: jwtutil.StandardClaims{
			Subject: "user123",
		},
		CustomClaims: map[string]interface{}{
			"username": "john_doe",
			"role":     "admin",
			"email":    "john@example.com",
		},
	}

	token, err := config.GenerateToken(claims)
	if err != nil {
		fmt.Printf("生成令牌失败: %v\n", err)
		return
	}

	fmt.Printf("生成的令牌: %s\n", token)

	// 验证令牌
	validatedClaims, err := config.VerifyToken(token)
	if err != nil {
		fmt.Printf("验证令牌失败: %v\n", err)
		return
	}

	fmt.Printf("验证成功！用户ID: %s\n", validatedClaims.Subject)
	fmt.Printf("用户名: %v\n", validatedClaims.CustomClaims["username"])
	fmt.Printf("角色: %v\n\n", validatedClaims.CustomClaims["role"])
}

func example2_AccessAndRefreshTokens() {
	fmt.Println("--- 示例 2: 访问令牌和刷新令牌 ---")

	config := jwtutil.NewJWTConfig("my-secret-key-for-access-token")
	config.ExpiryDuration = 15 * time.Minute // 访问令牌15分钟过期

	refreshConfig := jwtutil.NewJWTConfig("my-secret-key-for-refresh-token")
	refreshConfig.ExpiryDuration = 7 * 24 * time.Hour // 刷新令牌7天过期

	// 用户登录，生成访问令牌和刷新令牌
	userID := "user456"
	claims := &jwtutil.Claims{
		StandardClaims: jwtutil.StandardClaims{
			Subject: userID,
		},
		CustomClaims: map[string]interface{}{
			"username": "alice",
		},
	}

	accessToken, _ := config.GenerateToken(claims)
	refreshToken, _ := refreshConfig.GenerateToken(&jwtutil.Claims{
		StandardClaims: jwtutil.StandardClaims{
			Subject: userID,
		},
	})

	fmt.Printf("访问令牌: %s\n", accessToken)
	fmt.Printf("刷新令牌: %s\n\n", refreshToken)

	// 使用访问令牌
	validatedClaims, err := config.VerifyToken(accessToken)
	if err != nil {
		fmt.Printf("访问令牌验证失败: %v\n", err)
	} else {
		fmt.Printf("访问令牌验证成功！用户: %v\n\n", validatedClaims.CustomClaims["username"])
	}
}

//func example3_TokenValidation() {
//	fmt.Println("--- 示例 3: 令牌验证和自定义检查 ---")
//
//	config := jwtutil.NewJWTConfig("validation-secret-key")
//	config.Issuer = "trusted-issuer"
//	config.Audience = []string{"my-api"}
//
//	claims := &jwtutil.Claims{
//		StandardClaims: jwtutil.StandardClaims{
//			Subject: "user789",
//		},
//		CustomClaims: map[string]interface{}{
//			"permissions": []string{"read", "write", "delete"},
//		},
//	}
//
//	token, _ := config.GenerateToken(claims)
//
//	// 验证令牌并进行自定义检查
//	validatedClaims, err := config.ValidateTokenWithCheck(token, func(c *jwtutil.Claims) error {
//		// 自定义验证逻辑：检查权限
//		permissions, ok := c.CustomClaims["permissions"].([]interface{})
//		if !ok || len(permissions) == 0 {
//			return fmt.Errorf("缺少权限信息")
//		}
//
//		hasWritePermission := false
//		for _, p := range permissions {
//			if p == "write" {
//				hasWritePermission = true
//				break
//			}
//		}
//
//		if !hasWritePermission {
//			return fmt.Errorf("缺少写入权限")
//		}
//
//		return nil
//	})
//
//	if err != nil {
//		fmt.Printf("令牌验证失败: %v\n", err)
//	} else {
//		fmt.Printf("令牌验证成功！用户ID: %s\n", validatedClaims.Subject)
//		fmt.Printf("权限: %v\n\n", validatedClaims.CustomClaims["permissions"])
//	}
//}

func example4_TokenRefresh() {
	fmt.Println("--- 示例 4: 令牌刷新 ---")

	config := jwtutil.NewJWTConfig("refresh-example-secret-key")
	config.ExpiryDuration = 1 * time.Hour

	// 生成初始令牌
	claims := &jwtutil.Claims{
		StandardClaims: jwtutil.StandardClaims{
			Subject: "user999",
		},
		CustomClaims: map[string]interface{}{
			"username": "bob",
			"role":     "user",
		},
	}

	oldToken, _ := config.GenerateToken(claims)
	fmt.Printf("原始令牌: %s\n", oldToken)

	// 模拟一段时间后刷新令牌
	time.Sleep(1 * time.Second)

	newToken, err := config.RefreshToken(oldToken)
	if err != nil {
		fmt.Printf("刷新令牌失败: %v\n", err)
		return
	}

	fmt.Printf("新令牌: %s\n", newToken)

	// 验证新令牌
	validatedClaims, _ := config.VerifyToken(newToken)
	fmt.Printf("新令牌验证成功！用户: %v\n\n", validatedClaims.CustomClaims["username"])
}

func example5_ExtractTokenInfo() {
	fmt.Println("--- 示例 5: 提取令牌信息 ---")

	config := jwtutil.NewJWTConfig("extract-info-secret-key")

	claims := &jwtutil.Claims{
		StandardClaims: jwtutil.StandardClaims{
			Subject: "user111",
		},
		CustomClaims: map[string]interface{}{
			"username":    "charlie",
			"email":       "charlie@example.com",
			"department":  "Engineering",
			"employee_id": 12345,
		},
	}

	token, _ := config.GenerateToken(claims)

	// 提取用户ID
	userID, err := jwtutil.ExtractSubject(token)
	if err != nil {
		fmt.Printf("提取用户ID失败: %v\n", err)
		return
	}
	fmt.Printf("用户ID: %s\n", userID)

	// 提取自定义字段
	validatedClaims, _ := config.VerifyToken(token)
	fmt.Printf("用户名: %v\n", validatedClaims.CustomClaims["username"])
	fmt.Printf("邮箱: %v\n", validatedClaims.CustomClaims["email"])
	fmt.Printf("部门: %v\n", validatedClaims.CustomClaims["department"])
	fmt.Printf("员工编号: %v\n\n", validatedClaims.CustomClaims["employee_id"])
}

func example6_ErrorHandling() {
	fmt.Println("--- 示例 6: 错误处理 ---")

	config := jwtutil.NewJWTConfig("error-handling-secret-key")
	config.ExpiryDuration = 1 * time.Second // 1秒过期

	claims := &jwtutil.Claims{
		StandardClaims: jwtutil.StandardClaims{
			Subject: "user222",
		},
	}

	token, _ := config.GenerateToken(claims)
	fmt.Printf("生成的令牌: %s\n", token)

	// 立即验证（应该成功）
	_, err := config.VerifyToken(token)
	if err != nil {
		fmt.Printf("验证失败: %v\n", err)
	} else {
		fmt.Println("验证成功！")
	}

	// 等待令牌过期
	fmt.Println("等待令牌过期（2秒）...")
	time.Sleep(2 * time.Second)

	// 验证过期令牌
	_, err = config.VerifyToken(token)
	if err != nil {
		fmt.Printf("❌ 预期的错误：%v\n", err)
	}

	// 测试无效令牌
	fmt.Println("\n测试无效令牌...")
	_, err = config.VerifyToken("invalid.token.here")
	if err != nil {
		fmt.Printf("❌ 预期的错误：%v\n", err)
	}

	// 测试错误的密钥
	fmt.Println("\n测试错误的密钥...")
	wrongConfig := jwtutil.NewJWTConfig("wrong-secret-key")
	_, err = wrongConfig.VerifyToken(token)
	if err != nil {
		fmt.Printf("❌ 预期的错误：%v\n\n", err)
	}
}

/*
运行方式：
go run jwtutil_example.go

输出示例：
=== JWT Utility 示例 ===

--- 示例 1: 基本使用 ---
生成的令牌: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
验证成功！用户ID: user123
用户名: john_doe
角色: admin

--- 示例 2: 访问令牌和刷新令牌 ---
访问令牌: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
刷新令牌: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
访问令牌验证成功！用户: alice

--- 示例 3: 令牌验证和自定义检查 ---
令牌验证成功！用户ID: user789
权限: [read write delete]

--- 示例 4: 令牌刷新 ---
原始令牌: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
新令牌: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
新令牌验证成功！用户: bob

--- 示例 5: 提取令牌信息 ---
用户ID: user111
用户名: charlie
邮箱: charlie@example.com
部门: Engineering
员工编号: 12345

--- 示例 6: 错误处理 ---
生成的令牌: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
验证成功！
等待令牌过期（2秒）...
❌ 预期的错误：token is expired
测试无效令牌...
❌ 预期的错误：token contains an invalid number of segments
测试错误的密钥...
❌ 预期的错误：signature is invalid
*/
