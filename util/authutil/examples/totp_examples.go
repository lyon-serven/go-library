// Package main demonstrates usage of Google Authenticator (TOTP) utilities.
package main

import (
	"fmt"
	"log"
	"time"

	"mylib/util/authutil"
)

func main() {
	fmt.Println("=== Google Authenticator (TOTP) 示例 ===\n")

	// 示例1: 基础TOTP功能
	fmt.Println("1. 基础TOTP功能演示")
	basicTOTPExample()

	// 示例2: 用户管理
	fmt.Println("\n2. 多用户TOTP管理")
	userManagementExample()

	// 示例3: 备份码功能
	fmt.Println("\n3. 备份恢复码功能")
	backupCodesExample()

	// 示例4: 快速API使用
	fmt.Println("\n4. 快速API使用")
	quickAPIExample()

	// 示例5: 自定义配置
	fmt.Println("\n5. 自定义配置示例")
	customConfigExample()

	fmt.Println("\n=== 所有示例完成 ===")
}

// basicTOTPExample 演示基础TOTP功能
func basicTOTPExample() {
	// 创建Google Authenticator实例
	config := authutil.DefaultTOTPConfig()
	config.Issuer = "MyCompany"
	config.AccountName = "user@example.com"

	ga := authutil.NewGoogleAuthenticator(config)

	// 生成密钥
	secret, err := ga.GenerateSecret()
	if err != nil {
		log.Printf("生成密钥失败: %v", err)
		return
	}

	fmt.Printf("生成的密钥: %s\n", secret)

	// 生成当前TOTP码
	currentCode, err := ga.GenerateCode(secret)
	if err != nil {
		log.Printf("生成TOTP码失败: %v", err)
		return
	}

	fmt.Printf("当前TOTP码: %s\n", currentCode)

	// 验证TOTP码
	isValid := ga.VerifyCode(secret, currentCode)
	fmt.Printf("验证结果: %v\n", isValid)

	// 生成QR码URL
	qrURL := ga.GenerateQRCodeURL(secret)
	fmt.Printf("QR码URL: %s\n", qrURL)

	// 生成QR码图片URL
	qrImageURL := ga.GenerateQRCodeImageURL(secret)
	fmt.Printf("QR码图片URL: %s\n", qrImageURL)

	// 显示剩余时间
	remainingTime := ga.GetRemainingTime()
	fmt.Printf("当前码剩余时间: %d秒\n", remainingTime)

	// 演示时间容差验证
	fmt.Println("\n演示时间容差验证:")

	// 生成前一个时间段的码
	prevTime := time.Now().Add(-30 * time.Second)
	prevCode, _ := ga.GenerateCodeAtTime(secret, prevTime)
	fmt.Printf("前一个时间段的码: %s\n", prevCode)

	// 使用容差验证
	isValidWithTolerance := ga.VerifyCodeWithTolerance(secret, prevCode, 1)
	fmt.Printf("使用容差验证前一个码: %v\n", isValidWithTolerance)
}

// userManagementExample 演示多用户TOTP管理
func userManagementExample() {
	// 创建TOTP管理器
	manager := authutil.NewTOTPManager("MyApp")

	// 为用户设置TOTP
	users := []struct {
		ID    string
		Email string
	}{
		{"user1", "alice@example.com"},
		{"user2", "bob@example.com"},
		{"user3", "charlie@example.com"},
	}

	userSecrets := make(map[string]string)

	for _, user := range users {
		secret, qrURL, err := manager.SetupUser(user.ID, user.Email)
		if err != nil {
			log.Printf("设置用户 %s TOTP失败: %v", user.ID, err)
			continue
		}

		userSecrets[user.ID] = secret
		fmt.Printf("用户 %s (%s):\n", user.ID, user.Email)
		fmt.Printf("  密钥: %s\n", secret)
		fmt.Printf("  QR码: %s\n", qrURL)

		// 生成测试码
		ga := authutil.NewGoogleAuthenticator(authutil.DefaultTOTPConfig())
		testCode, _ := ga.GenerateCode(secret)
		fmt.Printf("  测试码: %s\n", testCode)
		fmt.Println()
	}

	// 验证用户码
	fmt.Println("验证用户码:")
	for userID, secret := range userSecrets {
		ga := authutil.NewGoogleAuthenticator(authutil.DefaultTOTPConfig())
		testCode, _ := ga.GenerateCode(secret)

		isValid := manager.VerifyUserCode(userID, testCode)
		fmt.Printf("用户 %s 的码 %s: %v\n", userID, testCode, isValid)
	}

	// 测试错误的码
	fmt.Println("\n测试错误的码:")
	isValid := manager.VerifyUserCode("user1", "000000")
	fmt.Printf("用户 user1 的错误码 000000: %v\n", isValid)

	// 移除用户
	manager.RemoveUser("user3")
	fmt.Println("已移除用户 user3")
}

// backupCodesExample 演示备份恢复码功能
func backupCodesExample() {
	// 生成备份码
	backupCodes, err := authutil.GenerateBackupCodes(10)
	if err != nil {
		log.Printf("生成备份码失败: %v", err)
		return
	}

	fmt.Printf("生成时间: %s\n", backupCodes.Generated.Format("2006-01-02 15:04:05"))
	fmt.Println("备份恢复码:")
	for i, code := range backupCodes.Codes {
		fmt.Printf("  %d. %s\n", i+1, code)
	}

	// 使用备份码
	testCode := backupCodes.Codes[0]
	fmt.Printf("\n使用备份码: %s\n", testCode)

	used := backupCodes.UseBackupCode(testCode)
	fmt.Printf("使用结果: %v\n", used)

	// 再次尝试使用同一个码
	used = backupCodes.UseBackupCode(testCode)
	fmt.Printf("重复使用结果: %v\n", used)

	// 获取未使用的码
	unusedCodes := backupCodes.GetUnusedCodes()
	fmt.Printf("剩余未使用码数量: %d\n", len(unusedCodes))
	fmt.Println("未使用的码:")
	for i, code := range unusedCodes {
		if i < 5 { // 只显示前5个
			fmt.Printf("  %s\n", code)
		}
	}
	if len(unusedCodes) > 5 {
		fmt.Printf("  ... 还有 %d 个\n", len(unusedCodes)-5)
	}
}

// quickAPIExample 演示快速API使用
func quickAPIExample() {
	// 快速生成
	secret, qrURL, err := authutil.QuickGenerate("QuickApp", "quick@example.com")
	if err != nil {
		log.Printf("快速生成失败: %v", err)
		return
	}

	fmt.Printf("快速生成结果:\n")
	fmt.Printf("  密钥: %s\n", secret)
	fmt.Printf("  QR码: %s\n", qrURL)

	// 快速生成码
	code, err := authutil.QuickGenerateCode(secret)
	if err != nil {
		log.Printf("快速生成码失败: %v", err)
		return
	}

	fmt.Printf("  当前码: %s\n", code)

	// 快速验证
	isValid := authutil.QuickVerify(secret, code)
	fmt.Printf("  验证结果: %v\n", isValid)

	// 测试错误码
	isValid = authutil.QuickVerify(secret, "000000")
	fmt.Printf("  错误码验证: %v\n", isValid)
}

// customConfigExample 演示自定义配置
func customConfigExample() {
	// 自定义配置：8位数码，60秒周期
	config := &authutil.TOTPConfig{
		Issuer:      "CustomApp",
		AccountName: "custom@example.com",
		Digits:      8,
		Period:      60,
		Algorithm:   "sha1",
	}

	ga := authutil.NewGoogleAuthenticator(config)

	// 生成密钥
	secret, err := ga.GenerateSecret()
	if err != nil {
		log.Printf("生成密钥失败: %v", err)
		return
	}

	fmt.Printf("自定义配置 (8位数码, 60秒周期):\n")
	fmt.Printf("  密钥: %s\n", secret)

	// 生成码
	code, err := ga.GenerateCode(secret)
	if err != nil {
		log.Printf("生成码失败: %v", err)
		return
	}

	fmt.Printf("  当前码: %s (8位数)\n", code)

	// 验证
	isValid := ga.VerifyCode(secret, code)
	fmt.Printf("  验证结果: %v\n", isValid)

	// 剩余时间
	remainingTime := ga.GetRemainingTime()
	fmt.Printf("  剩余时间: %d秒 (60秒周期)\n", remainingTime)

	// QR码
	qrURL := ga.GenerateQRCodeImageURL(secret)
	fmt.Printf("  QR码: %s\n", qrURL)
}

// 演示实际应用场景
func demonstrateRealWorldScenario() {
	fmt.Println("\n=== 实际应用场景演示 ===")

	// 场景：用户注册并启用2FA
	fmt.Println("场景：用户启用两步验证")

	// 1. 用户请求启用2FA
	manager := authutil.NewTOTPManager("MyWebsite")
	userID := "12345"
	userEmail := "john@example.com"

	// 2. 系统生成密钥和QR码
	secret, qrURL, err := manager.SetupUser(userID, userEmail)
	if err != nil {
		log.Printf("设置失败: %v", err)
		return
	}

	fmt.Printf("步骤1 - 生成QR码供用户扫描:\n")
	fmt.Printf("  用户: %s\n", userEmail)
	fmt.Printf("  密钥: %s\n", secret)
	fmt.Printf("  QR码: %s\n", qrURL)

	// 3. 用户扫描QR码后，输入验证码确认
	fmt.Printf("\n步骤2 - 用户确认设置:\n")
	ga := authutil.NewGoogleAuthenticator(authutil.DefaultTOTPConfig())
	confirmCode, _ := ga.GenerateCode(secret)
	fmt.Printf("  用户输入验证码: %s\n", confirmCode)

	isSetupValid := manager.VerifyUserCode(userID, confirmCode)
	fmt.Printf("  设置验证结果: %v\n", isSetupValid)

	if isSetupValid {
		fmt.Println("  ✓ 两步验证设置成功！")

		// 4. 生成备份码
		backupCodes, _ := authutil.GenerateBackupCodes(8)
		fmt.Printf("\n步骤3 - 生成备份恢复码:\n")
		for i, code := range backupCodes.Codes {
			fmt.Printf("  %d. %s\n", i+1, code)
		}
		fmt.Println("  请妥善保存这些备份码！")

		// 5. 模拟用户登录验证
		fmt.Printf("\n步骤4 - 用户登录验证:\n")
		time.Sleep(1 * time.Second) // 确保生成不同的码
		loginCode, _ := ga.GenerateCode(secret)
		fmt.Printf("  用户登录时输入: %s\n", loginCode)

		isLoginValid := manager.VerifyUserCode(userID, loginCode)
		fmt.Printf("  登录验证结果: %v\n", isLoginValid)

		if isLoginValid {
			fmt.Println("  ✓ 登录成功！")
		} else {
			fmt.Println("  ✗ 登录失败！")
		}
	}
}
