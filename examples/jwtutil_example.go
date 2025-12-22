package main

import (
	"fmt"
	"time"

	"gitee.com/wangsoft/go-library/util/jwtutil"
)

func main() {
	fmt.Println("=== JWT Utility 示例 ===\n")

	// 示例 1: 基本使用
	example1_BasicUsage()

	// 示例 2: 访问令牌和刷新令牌
	example2_AccessAndRefreshTokens()

	// 示例 3: 令牌验证和自定义检查
	example3_TokenValidation()

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

	fmt.Printf("生成的令牌: %s...\n", token[:50])

	// 验证令牌
	verifiedClaims, err := config.VerifyToken(token)
	if err != nil {
		fmt.Printf("验证令牌失败: %v\n", err)
		return
	}

	fmt.Printf("主题 (Subject): %s\n", verifiedClaims.Subject)
	fmt.Printf("发行者 (Issuer): %s\n", verifiedClaims.Issuer)
	fmt.Printf("用户名: %v\n", verifiedClaims.CustomClaims["username"])
	fmt.Printf("角色: %v\n", verifiedClaims.CustomClaims["role"])
	fmt.Printf("邮箱: %v\n", verifiedClaims.CustomClaims["email"])
	fmt.Println()
}

func example2_AccessAndRefreshTokens() {
	fmt.Println("--- 示例 2: 访问令牌和刷新令牌 ---")

	secretKey := "my-super-secret-key-12345"
	userID := "user123"

	// 生成访问令牌 (15 分钟有效期)
	accessToken, err := jwtutil.GenerateAccessToken(
		secretKey,
		userID,
		15*time.Minute,
		map[string]interface{}{
			"username": "john_doe",
			"role":     "admin",
		},
	)
	if err != nil {
		fmt.Printf("生成访问令牌失败: %v\n", err)
		return
	}

	fmt.Printf("访问令牌: %s...\n", accessToken[:50])

	// 生成刷新令牌 (7 天有效期)
	refreshToken, err := jwtutil.GenerateRefreshToken(
		secretKey,
		userID,
		7*24*time.Hour,
	)
	if err != nil {
		fmt.Printf("生成刷新令牌失败: %v\n", err)
		return
	}

	fmt.Printf("刷新令牌: %s...\n", refreshToken[:50])

	// 验证访问令牌
	claims, err := jwtutil.VerifyAccessToken(secretKey, accessToken)
	if err != nil {
		fmt.Printf("验证访问令牌失败: %v\n", err)
		return
	}

	fmt.Printf("访问令牌主题: %s\n", claims.Subject)
	fmt.Printf("访问令牌角色: %v\n", claims.CustomClaims["role"])

	// 验证刷新令牌
	refreshClaims, err := jwtutil.VerifyAccessToken(secretKey, refreshToken)
	if err != nil {
		fmt.Printf("验证刷新令牌失败: %v\n", err)
		return
	}

	fmt.Printf("刷新令牌主题: %s\n", refreshClaims.Subject)
	fmt.Printf("刷新令牌类型: %v\n", refreshClaims.CustomClaims["type"])
	fmt.Println()
}

func example3_TokenValidation() {
	fmt.Println("--- 示例 3: 令牌验证和自定义检查 ---")

	config := jwtutil.NewJWTConfig("my-super-secret-key-12345")
	config.Issuer = "auth-service"

	claims := &jwtutil.Claims{
		StandardClaims: jwtutil.StandardClaims{
			Subject:  "user123",
			Audience: "api-service",
		},
		CustomClaims: map[string]interface{}{
			"username": "john_doe",
		},
	}

	token, err := config.GenerateToken(claims)
	if err != nil {
		fmt.Printf("生成令牌失败: %v\n", err)
		return
	}

	// 验证令牌并检查发行者和受众
	validatedClaims, err := config.ValidateToken(token, "auth-service", "api-service")
	if err != nil {
		fmt.Printf("验证令牌失败: %v\n", err)
		return
	}

	fmt.Printf("验证成功!\n")
	fmt.Printf("主题: %s\n", validatedClaims.Subject)
	fmt.Printf("发行者: %s\n", validatedClaims.Issuer)
	fmt.Printf("受众: %s\n", validatedClaims.Audience)
	fmt.Println()
}

func example4_TokenRefresh() {
	fmt.Println("--- 示例 4: 令牌刷新 ---")

	config := jwtutil.NewJWTConfig("my-super-secret-key-12345")
	config.ExpiryDuration = 1 * time.Hour

	claims := &jwtutil.Claims{
		StandardClaims: jwtutil.StandardClaims{
			Subject: "user123",
		},
		CustomClaims: map[string]interface{}{
			"username": "john_doe",
		},
	}

	oldToken, err := config.GenerateToken(claims)
	if err != nil {
		fmt.Printf("生成原始令牌失败: %v\n", err)
		return
	}

	oldClaims, _ := config.VerifyToken(oldToken)
	fmt.Printf("原始令牌过期时间: %s\n", time.Unix(oldClaims.ExpiresAt, 0).Format("2006-01-02 15:04:05"))

	time.Sleep(1 * time.Second)

	// 刷新令牌
	newToken, err := config.RefreshToken(oldToken)
	if err != nil {
		fmt.Printf("刷新令牌失败: %v\n", err)
		return
	}

	newClaims, _ := config.VerifyToken(newToken)
	fmt.Printf("新令牌过期时间: %s\n", time.Unix(newClaims.ExpiresAt, 0).Format("2006-01-02 15:04:05"))
	fmt.Printf("用户数据保持不变: username=%v\n", newClaims.CustomClaims["username"])
	fmt.Println()
}

func example5_ExtractTokenInfo() {
	fmt.Println("--- 示例 5: 提取令牌信息 ---")

	secretKey := "my-super-secret-key-12345"

	token, _ := jwtutil.GenerateAccessToken(
		secretKey,
		"user123",
		24*time.Hour,
		map[string]interface{}{
			"username": "john_doe",
			"role":     "admin",
			"email":    "john@example.com",
		},
	)

	// 提取主题（无需完全验证）
	subject, err := jwtutil.ExtractSubject(token)
	if err != nil {
		fmt.Printf("提取主题失败: %v\n", err)
		return
	}
	fmt.Printf("主题: %s\n", subject)

	// 提取自定义声明
	username, _ := jwtutil.ExtractCustomClaim(token, "username")
	fmt.Printf("用户名: %v\n", username)

	role, _ := jwtutil.ExtractCustomClaim(token, "role")
	fmt.Printf("角色: %v\n", role)

	email, _ := jwtutil.ExtractCustomClaim(token, "email")
	fmt.Printf("邮箱: %v\n", email)

	// 检查令牌是否过期
	expired, _ := jwtutil.IsTokenExpired(token)
	fmt.Printf("令牌是否过期: %v\n", expired)

	// 获取令牌过期时间
	expiry, _ := jwtutil.GetTokenExpiry(token)
	fmt.Printf("令牌过期时间: %s\n", expiry.Format("2006-01-02 15:04:05"))
	fmt.Println()
}

func example6_ErrorHandling() {
	fmt.Println("--- 示例 6: 错误处理 ---")

	config := jwtutil.NewJWTConfig("my-super-secret-key-12345")
	config.ExpiryDuration = 1 * time.Second

	claims := &jwtutil.Claims{
		StandardClaims: jwtutil.StandardClaims{
			Subject: "user123",
		},
	}

	token, _ := config.GenerateToken(claims)

	// 1. 正常验证
	_, err := config.VerifyToken(token)
	if err != nil {
		fmt.Printf("验证错误: %v\n", err)
	} else {
		fmt.Println("令牌有效 ✓")
	}

	// 2. 等待令牌过期
	time.Sleep(2 * time.Second)
	_, err = config.VerifyToken(token)
	if err != nil {
		switch err {
		case jwtutil.ErrTokenExpired:
			fmt.Println("错误: 令牌已过期 (可以尝试刷新)")
		case jwtutil.ErrInvalidSignature:
			fmt.Println("错误: 签名无效")
		default:
			fmt.Printf("错误: %v\n", err)
		}
	}

	// 3. 使用错误的密钥
	wrongConfig := jwtutil.NewJWTConfig("wrong-secret-key")
	validToken, _ := config.GenerateToken(&jwtutil.Claims{
		StandardClaims: jwtutil.StandardClaims{
			Subject: "user123",
		},
	})

	_, err = wrongConfig.VerifyToken(validToken)
	if err != nil {
		switch err {
		case jwtutil.ErrInvalidSignature:
			fmt.Println("错误: 签名无效 (可能使用了错误的密钥)")
		default:
			fmt.Printf("错误: %v\n", err)
		}
	}

	// 4. 无效的令牌格式
	_, err = config.VerifyToken("invalid.token.format")
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	}

	fmt.Println()
}
