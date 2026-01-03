package main

import (
	"fmt"
	"log"

	"gitee.com/wangsoft/go-library/util/authutil"
)

func main() {
	// 初始化配置
	BaseDemo()
}
func BaseDemo() {
	// 初始化配置
	// 快速生成密钥和QR码
	secret, qrCodeURL, err := authutil.QuickGenerate("MyApp", "user@example.com")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("密钥: %s\n", secret)
	fmt.Printf("QR码: %s\n", qrCodeURL)

	// 生成当前TOTP码
	code, err := authutil.QuickGenerateCode(secret)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("当前码: %s\n", code)

	// 验证TOTP码
	isValid := authutil.QuickVerify(secret, code)
	fmt.Printf("验证结果: %v\n", isValid)
}

func AdvancedDemo() {
	// 创建自定义配置
	config := &authutil.TOTPConfig{
		Issuer:      "MyCompany",
		AccountName: "user@example.com",
		Digits:      6,
		Period:      30,
		Algorithm:   "sha1",
	}

	// 创建Google Authenticator实例
	ga := authutil.NewGoogleAuthenticator(config)

	// 生成密钥
	secret, err := ga.GenerateSecret()
	if err != nil {
		log.Fatal(err)
	}

	// 生成QR码URL
	qrURL := ga.GenerateQRCodeImageURL(secret)

	// 验证码（支持时间容差）
	// 假设用户输入的验证码为 "123456"，实际使用时应从用户输入获取
	userInputCode := "123456"
	isValid := ga.VerifyCodeWithTolerance(secret, userInputCode, 1)
	fmt.Printf("验证结果: %v\n", isValid)
	fmt.Printf("QR码URL: %s\n", qrURL)
}

func MultiDemo() {
	// 创建TOTP管理器
	manager := authutil.NewTOTPManager("MyApp")

	// 为用户设置TOTP
	secret, qrURL, err := manager.SetupUser("user123", "john@example.com")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("密钥: %s\n", secret)
	fmt.Printf("QR码URL: %s\n", qrURL)
	// 假设用户输入的验证码为 "123456"，实际使用时应从用户输入获取
	userInputCode := "123456"
	// 验证用户码
	isValid := manager.VerifyUserCode("user123", userInputCode)
	fmt.Printf("验证结果: %v\n", isValid)
	// 获取用户密钥（用于备份）
	secret, exists := manager.GetUserSecret("user123")
	if exists {
		fmt.Printf("用户密钥: %s\n", secret)
	}
	// 移除用户TOTP
	manager.RemoveUser("user123")
}
