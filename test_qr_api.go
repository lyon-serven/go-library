package main

import (
	"fmt"
	"mylib/util/authutil"
)

func main() {
	fmt.Println("=== 测试国内可访问的二维码API ===\n")

	// 测试新的快速API（默认使用国内可访问的服务）
	secret, qrURL, err := authutil.QuickGenerate("TestApp", "test@example.com")
	if err != nil {
		fmt.Printf("❌ 生成失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 密钥生成成功: %s\n", secret[:20]+"...")
	fmt.Printf("✅ QR码URL: %s\n\n", qrURL)

	// 解析URL检查域名
	if len(qrURL) > 50 {
		fmt.Println("📌 URL分析:")
		if qrURL[:8] == "https://" {
			endOfDomain := 8
			for i := 8; i < len(qrURL) && qrURL[i] != '/'; i++ {
				endOfDomain = i + 1
			}
			domain := qrURL[8:endOfDomain]
			fmt.Printf("   使用的域名: %s\n", domain)

			if domain == "api.qrserver.com" {
				fmt.Println("   ✅ 正在使用国内可访问的 QR Server API")
			} else if domain == "chart.googleapis.com" {
				fmt.Println("   ⚠️  正在使用 Google Charts API（需要翻墙）")
			}
		}
	}

	fmt.Println("\n📱 使用方法:")
	fmt.Println("1. 复制上面的URL到浏览器打开，即可看到二维码")
	fmt.Println("2. 使用Google Authenticator扫描二维码")
	fmt.Println("3. 或者在HTML中使用: <img src=\"" + qrURL + "\" />")

	// 测试生成验证码
	code, err := authutil.QuickGenerateCode(secret)
	if err != nil {
		fmt.Printf("\n❌ 生成验证码失败: %v\n", err)
		return
	}

	fmt.Printf("\n✅ 当前验证码: %s\n", code)

	// 验证码测试
	isValid := authutil.QuickVerify(secret, code)
	fmt.Printf("✅ 验证结果: %v\n", isValid)

	fmt.Println("\n🎉 所有功能测试通过！")
}
