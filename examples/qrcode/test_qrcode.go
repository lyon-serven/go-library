package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"

	"github.com/lyon-serven/go-library/util/qrcodeutil"
)

func main() {
	fmt.Println("Testing QR Code Generation...")

	// 测试简单二维码生成
	fmt.Println("1. Testing simple QR code generation...")
	data, err := qrcodeutil.GenerateQRCodeSimple("https://github.com/lyon-serven/go-library")
	if err != nil {
		fmt.Printf("Error generating simple QR code: %v\n", err)
		return
	}
	fmt.Printf("✓ Simple QR code generated: %d bytes\n", len(data))

	// 测试保存到文件
	fmt.Println("2. Testing QR code file generation...")
	err = qrcodeutil.GenerateQRCodeToFileSimple("Hello from QR Code Util!", "./examples/files/test_qrcode.png")
	if err != nil {
		fmt.Printf("Error saving QR code to file: %v\n", err)
		return
	}
	fmt.Println("✓ QR code saved to test_qrcode.png")

	// 测试自定义配置
	config := qrcodeutil.QRCodeConfig{
		Size:       300,
		Level:      "H",
		Foreground: "#FF5733",
		Background: "#F0F8FF",
	}

	customData, err := qrcodeutil.GenerateQRCode("Custom QR Code", config)
	if err != nil {
		fmt.Printf("Error generating custom QR code: %v\n", err)
		return
	}
	fmt.Printf("✓ Custom QR code generated: %d bytes\n", len(customData))

	// 测试保存自定义配置的二维码
	err = qrcodeutil.GenerateQRCodeToFile("Custom Configuration Test", "./examples/files/custom_qrcode.png", config)
	if err != nil {
		fmt.Printf("Error saving custom QR code: %v\n", err)
		return
	}
	fmt.Println("✓ Custom QR code saved to custom_qrcode.png")

	// 测试带LOGO的二维码生成
	fmt.Println("4. Testing QR code with logo...")

	// 创建简单的LOGO图片用于测试
	err = createTestLogo("./examples/files/test_logo.png")
	if err != nil {
		fmt.Printf("Error creating test logo: %v\n", err)
		return
	}
	fmt.Println("✓ Test logo created: test_logo.png")

	// 测试简化版带LOGO二维码
	logoData, err := qrcodeutil.GenerateQRCodeWithLogoSimple("QR Code with Logo", "./examples/files/test_logo.png")
	if err != nil {
		fmt.Printf("Error generating QR code with logo: %v\n", err)
		return
	}
	fmt.Printf("✓ QR code with logo generated: %d bytes\n", len(logoData))

	// 测试保存带LOGO的二维码
	err = qrcodeutil.GenerateQRCodeToFileWithLogoSimple("QR Code with Logo Test", "./examples/files/logo_qrcode.png", "./examples/files/test_logo.png")
	if err != nil {
		fmt.Printf("Error saving QR code with logo: %v\n", err)
		return
	}
	fmt.Println("✓ QR code with logo saved to logo_qrcode.png")

	// 测试自定义LOGO配置
	fmt.Println("5. Testing custom logo configuration...")
	logoConfig := qrcodeutil.LogoConfig{
		Path:     "./examples/files/test_logo.png",
		Size:     0.15,     // 二维码尺寸的15%
		Position: "center", // 居中显示
		Opacity:  0.9,      // 90%透明度
	}

	customLogoData, err := qrcodeutil.GenerateQRCodeWithLogo("Custom Logo QR Code", config, logoConfig)
	if err != nil {
		fmt.Printf("Error generating custom logo QR code: %v\n", err)
		return
	}
	fmt.Printf("✓ Custom logo QR code generated: %d bytes\n", len(customLogoData))

	// 测试不同位置的LOGO
	positions := []string{"top-left", "top-right", "bottom-left", "bottom-right"}
	for _, pos := range positions {
		logoConfig.Position = pos
		filename := fmt.Sprintf("./examples/files/logo_%s.png", pos)
		err = qrcodeutil.GenerateQRCodeToFileWithLogo("Position Test: "+pos, filename, config, logoConfig)
		if err != nil {
			fmt.Printf("Error saving QR code with %s logo: %v\n", pos, err)
			return
		}
		fmt.Printf("✓ QR code with %s logo saved to %s\n", pos, filename)
	}

	fmt.Println("\n✅ All QR code tests passed successfully!")
	fmt.Println("Generated files:")
	fmt.Println("  - test_qrcode.png (simple QR code)")
	fmt.Println("  - custom_qrcode.png (custom configuration)")
	fmt.Println("  - test_logo.png (test logo image)")
	fmt.Println("  - logo_qrcode.png (QR code with logo)")
	fmt.Println("  - logo_top-left.png (logo at top-left)")
	fmt.Println("  - logo_top-right.png (logo at top-right)")
	fmt.Println("  - logo_bottom-left.png (logo at bottom-left)")
	fmt.Println("  - logo_bottom-right.png (logo at bottom-right)")
}

// createTestLogo 创建测试用的LOGO图片
func createTestLogo(filename string) error {
	// 创建一个简单的PNG图片作为LOGO
	// 这里我们创建一个简单的红色圆形作为测试LOGO
	width, height := 100, 100
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// 设置背景为透明
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{0, 0, 0, 0}) // 透明背景
		}
	}

	// 绘制一个红色圆形
	centerX, centerY := width/2, height/2
	radius := 40

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			dx := x - centerX
			dy := y - centerY
			distance := dx*dx + dy*dy

			if distance <= radius*radius {
				img.Set(x, y, color.RGBA{255, 0, 0, 255}) // 红色圆形
			}
		}
	}

	// 确保目录存在
	err := os.MkdirAll(filepath.Dir(filename), 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 保存PNG文件
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		return fmt.Errorf("failed to encode PNG: %w", err)
	}

	return nil
}
