// QR码生成方案示例
// 本文件展示多种生成Google Authenticator二维码的方法
package main

import (
	"fmt"
	"mylib/util/authutil"
)

func main() {
	fmt.Println("=== Google Authenticator 二维码生成方案 ===\n")

	// 创建配置
	config := authutil.DefaultTOTPConfig()
	config.Issuer = "MyApp"
	config.AccountName = "user@example.com"
	ga := authutil.NewGoogleAuthenticator(config)

	// 生成密钥
	secret, err := ga.GenerateSecret()
	if err != nil {
		fmt.Printf("生成密钥失败: %v\n", err)
		return
	}

	fmt.Printf("生成的密钥: %s\n\n", secret)

	// 方案1: 获取原始 otpauth:// URL（推荐）
	fmt.Println("【方案1】原始 otpauth:// URL (推荐)")
	otpauthURL := ga.GetOtpauthURL(secret)
	fmt.Printf("URL: %s\n", otpauthURL)
	fmt.Println("说明: 复制此URL，使用任何二维码生成工具生成即可")
	fmt.Println("在线工具推荐:")
	fmt.Println("  - https://cli.im/ (草料二维码)")
	fmt.Println("  - https://www.wwei.cn/qrcode.html (微微二维码)")
	fmt.Println()

	// 方案2: 使用国内可访问的API
	fmt.Println("【方案2】使用国内可访问的在线API")
	qrURLCN := ga.GenerateQRCodeImageURLCN(secret)
	fmt.Printf("QR码图片URL (QR Server API): %s\n", qrURLCN)
	fmt.Println("说明: 此URL可直接在浏览器打开，或在<img>标签中使用")
	fmt.Println("提示: api.qrserver.com 在国内大部分地区可访问")
	fmt.Println()

	// 方案3: Google Charts API（需要翻墙）
	fmt.Println("【方案3】Google Charts API（需要VPN）")
	qrURLGoogle := ga.GenerateQRCodeImageURL(secret)
	fmt.Printf("QR码图片URL (Google): %s\n", qrURLGoogle)
	fmt.Println("说明: 此API在国内需要翻墙访问")
	fmt.Println()

	// 方案4: 本地生成（需要安装依赖库）
	fmt.Println("【方案4】本地生成二维码（需要安装依赖）")
	fmt.Println("步骤:")
	fmt.Println("1. 安装依赖库:")
	fmt.Println("   go get github.com/skip2/go-qrcode")
	fmt.Println()
	fmt.Println("2. 使用示例代码:")
	fmt.Println(`
   import "github.com/skip2/go-qrcode"
   
   // 生成PNG文件
   otpauthURL := ga.GetOtpauthURL(secret)
   qrcode.WriteFile(otpauthURL, qrcode.Medium, 256, "qrcode.png")
   
   // 生成base64编码（用于网页）
   png, _ := qrcode.Encode(otpauthURL, qrcode.Medium, 256)
   base64Str := base64.StdEncoding.EncodeToString(png)
`)
	fmt.Println("优点: 不依赖外部API，速度快，稳定")
	fmt.Println()

	// 方案5: 终端显示
	fmt.Println("【方案5】在终端显示二维码（开发调试用）")
	fmt.Println("步骤:")
	fmt.Println("1. 安装依赖库:")
	fmt.Println("   go get github.com/mdp/qrterminal/v3")
	fmt.Println()
	fmt.Println("2. 使用示例代码:")
	fmt.Println(`
   import qrterminal "github.com/mdp/qrterminal/v3"
   
   otpauthURL := ga.GetOtpauthURL(secret)
   qrterminal.Generate(otpauthURL, qrterminal.L, os.Stdout)
`)
	fmt.Println("优点: 直接在命令行显示，开发测试方便")
	fmt.Println()

	// 推荐方案总结
	fmt.Println("=== 推荐方案 ===")
	fmt.Println()
	fmt.Println("🌟 生产环境推荐:")
	fmt.Println("   【方案4】本地生成 - 使用 github.com/skip2/go-qrcode")
	fmt.Println("   理由: 不依赖外部服务，快速稳定，无网络问题")
	fmt.Println()
	fmt.Println("🚀 快速开发/测试:")
	fmt.Println("   【方案2】在线API - api.qrserver.com")
	fmt.Println("   理由: 无需安装依赖，国内可访问")
	fmt.Println()
	fmt.Println("🔧 命令行工具:")
	fmt.Println("   【方案5】终端显示 - github.com/mdp/qrterminal")
	fmt.Println("   理由: 开发调试时直接在终端查看")
	fmt.Println()

	// 实际应用示例
	fmt.Println("=== 实际应用示例 ===")
	fmt.Println()
	fmt.Println("在Web应用中显示二维码:")
	fmt.Println(`
// 后端Go代码
func setupTOTP(w http.ResponseWriter, r *http.Request) {
    secret, qrURL, _ := authutil.QuickGenerate("MyApp", "user@example.com")
    
    // 方法A: 使用在线API (简单快速)
    json.NewEncoder(w).Encode(map[string]string{
        "secret": secret,
        "qrUrl": qrURL,  // 使用 GenerateQRCodeImageURLCN
    })
    
    // 方法B: 使用本地生成 (推荐)
    // otpauthURL := ga.GetOtpauthURL(secret)
    // png, _ := qrcode.Encode(otpauthURL, qrcode.Medium, 256)
    // base64Str := base64.StdEncoding.EncodeToString(png)
    // json.NewEncoder(w).Encode(map[string]string{
    //     "secret": secret,
    //     "qrDataUrl": "data:image/png;base64," + base64Str,
    // })
}

// 前端HTML代码
<div class="setup-totp">
    <h3>设置Google身份验证器</h3>
    <img src="${qrUrl}" alt="扫描二维码" />
    <!-- 或使用 base64 data URL -->
    <!-- <img src="${qrDataUrl}" alt="扫描二维码" /> -->
    <p>密钥: ${secret}</p>
    <p>请使用Google Authenticator扫描二维码</p>
</div>
`)

	fmt.Println("\n=== 示例完成 ===")
	fmt.Println("\n当前使用快速API测试:")

	// 使用快速API测试
	testSecret, testQRURL, _ := authutil.QuickGenerate("TestApp", "test@example.com")
	fmt.Printf("测试密钥: %s\n", testSecret)
	fmt.Printf("测试QR码URL: %s\n", testQRURL)
	fmt.Println("\n提示: 复制QR码URL到浏览器打开，即可看到二维码")
}
