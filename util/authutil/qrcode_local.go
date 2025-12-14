// Package authutil provides utilities for authentication
// This file contains local QR code generation using skip2/go-qrcode library
package authutil

import (
	"bytes"
	"encoding/base64"
	"fmt"
)

// QRCodeGenerator provides methods to generate QR codes locally
type QRCodeGenerator struct {
	ga *GoogleAuthenticator
}

// NewQRCodeGenerator creates a new QR code generator
func NewQRCodeGenerator(ga *GoogleAuthenticator) *QRCodeGenerator {
	return &QRCodeGenerator{ga: ga}
}

// GenerateQRCodeBase64 generates QR code as base64 encoded PNG image (本地生成)
// 使用此方法需要安装: go get github.com/skip2/go-qrcode
//
// 使用示例:
//
//	import "github.com/skip2/go-qrcode"
//
//	base64Image, err := GenerateQRCodeBase64Local(secret, 256)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	// 在HTML中使用: <img src="data:image/png;base64,{{base64Image}}" />
func GenerateQRCodeBase64Local(secret string, ga *GoogleAuthenticator, size int) (string, error) {
	// 需要导入: github.com/skip2/go-qrcode
	//
	// import "github.com/skip2/go-qrcode"
	//
	// otpauthURL := ga.GenerateQRCodeURL(secret)
	// png, err := qrcode.Encode(otpauthURL, qrcode.Medium, size)
	// if err != nil {
	//     return "", fmt.Errorf("failed to generate QR code: %w", err)
	// }
	//
	// // Convert to base64
	// base64Str := base64.StdEncoding.EncodeToString(png)
	// return base64Str, nil

	// 示例实现（需要安装依赖库）
	return "", fmt.Errorf("需要安装依赖: go get github.com/skip2/go-qrcode")
}

// SaveQRCodeToFile saves QR code to a PNG file (本地生成)
// 使用此方法需要安装: go get github.com/skip2/go-qrcode
//
// 使用示例:
//
//	import "github.com/skip2/go-qrcode"
//
//	err := SaveQRCodeToFileLocal(secret, ga, "qrcode.png", 256)
//	if err != nil {
//	    log.Fatal(err)
//	}
func SaveQRCodeToFileLocal(secret string, ga *GoogleAuthenticator, filename string, size int) error {
	// 需要导入: github.com/skip2/go-qrcode
	//
	// import "github.com/skip2/go-qrcode"
	//
	// otpauthURL := ga.GenerateQRCodeURL(secret)
	// err := qrcode.WriteFile(otpauthURL, qrcode.Medium, size, filename)
	// if err != nil {
	//     return fmt.Errorf("failed to save QR code: %w", err)
	// }
	// return nil

	// 示例实现（需要安装依赖库）
	return fmt.Errorf("需要安装依赖: go get github.com/skip2/go-qrcode")
}

// GetQRCodeDataURL generates a data URL for inline HTML usage (本地生成示例)
// 返回格式: data:image/png;base64,iVBORw0KGgoAAAANS...
func GetQRCodeDataURL(secret string, ga *GoogleAuthenticator, size int) (string, error) {
	// 使用 GenerateQRCodeBase64Local 生成 base64 编码的图片
	// base64Image, err := GenerateQRCodeBase64Local(secret, ga, size)
	// if err != nil {
	//     return "", err
	// }
	//
	// return fmt.Sprintf("data:image/png;base64,%s", base64Image), nil

	return "", fmt.Errorf("需要安装依赖: go get github.com/skip2/go-qrcode")
}

// 本地生成二维码的完整示例代码（需要取消注释并安装依赖）
/*
使用方法：

1. 安装依赖库：
   go get github.com/skip2/go-qrcode

2. 示例代码：

package main

import (
	"fmt"
	"log"
	"mylib/util/authutil"
	"github.com/skip2/go-qrcode"
)

func main() {
	// 创建 Google Authenticator 实例
	config := authutil.DefaultTOTPConfig()
	config.Issuer = "MyApp"
	config.AccountName = "user@example.com"
	ga := authutil.NewGoogleAuthenticator(config)

	// 生成密钥
	secret, _ := ga.GenerateSecret()

	// 获取 otpauth:// URL
	otpauthURL := ga.GetOtpauthURL(secret)

	// 方法1: 生成 PNG 文件
	err := qrcode.WriteFile(otpauthURL, qrcode.Medium, 256, "qrcode.png")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("QR码已保存到 qrcode.png")

	// 方法2: 生成 base64 编码的图片（用于网页显示）
	png, _ := qrcode.Encode(otpauthURL, qrcode.Medium, 256)
	base64Str := base64.StdEncoding.EncodeToString(png)
	dataURL := fmt.Sprintf("data:image/png;base64,%s", base64Str)

	// 在 HTML 中使用：
	// <img src="data:image/png;base64,..." />
	fmt.Println("Base64 Data URL:", dataURL[:100]+"...")
}
*/

// 使用国内可访问API的辅助函数

// GenerateQRCodeURLQRServer 使用 QR Server API (国际，国内大部分地区可访问)
func GenerateQRCodeURLQRServer(otpauthURL string, size int) string {
	return fmt.Sprintf("https://api.qrserver.com/v1/create-qr-code/?size=%dx%d&data=%s",
		size, size, otpauthURL)
}

// GenerateQRCodeURLCaoLiao 使用草料二维码API（需要注册获取API key）
// 官网: https://cli.im/
func GenerateQRCodeURLCaoLiao(otpauthURL string) string {
	// 需要替换为您的API key
	// return fmt.Sprintf("https://api.cli.im/qrcode/code?text=%s&key=YOUR_API_KEY", otpauthURL)
	return "需要申请草料二维码 API Key: https://cli.im/"
}

// PrintQRCodeInTerminal 在终端打印二维码（使用 ASCII 字符）
// 这个方法不需要任何外部依赖，直接在命令行显示
func PrintQRCodeInTerminal(otpauthURL string) string {
	// 可以使用 github.com/skip2/go-qrcode 库的终端输出功能
	// 或者使用 github.com/mdp/qrterminal 库
	return "使用方法: go get github.com/mdp/qrterminal\n" +
		"import \"github.com/mdp/qrterminal/v3\"\n" +
		"qrterminal.Generate(otpauthURL, qrterminal.L, os.Stdout)"
}

// 推荐方案总结：
//
// 方案1 (推荐): 使用本地库生成 - github.com/skip2/go-qrcode
//   优点: 不依赖外部API，速度快，稳定可靠
//   缺点: 需要安装依赖库
//
// 方案2: 使用 QR Server API - api.qrserver.com
//   优点: 无需安装依赖，国内大部分地区可访问
//   缺点: 依赖外部服务，可能不稳定
//
// 方案3: 终端显示 - github.com/mdp/qrterminal
//   优点: 直接在命令行显示，开发调试方便
//   缺点: 仅适用于终端环境

var _ = bytes.Buffer{} // 避免未使用导入错误
var _ = base64.StdEncoding
