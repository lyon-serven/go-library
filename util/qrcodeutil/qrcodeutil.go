package qrcodeutil

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"

	"github.com/disintegration/imaging"
	qrcode "github.com/skip2/go-qrcode"
)

// QRCodeConfig 二维码配置
type QRCodeConfig struct {
	Size       int    // 二维码尺寸 (像素)
	Level      string // 容错级别: L(7%), M(15%), Q(25%), H(30%)
	Foreground string // 前景色 (十六进制格式)
	Background string // 背景色 (十六进制格式)
}

// DefaultConfig 默认配置
var DefaultConfig = QRCodeConfig{
	Size:       256,
	Level:      "M",
	Foreground: "#000000",
	Background: "#FFFFFF",
}

// GenerateQRCode 生成二维码并返回PNG字节数据
func GenerateQRCode(content string, config QRCodeConfig) ([]byte, error) {
	level := getRecoveryLevel(config.Level)

	// 创建QRCode对象
	qr, err := qrcode.New(content, level)
	if err != nil {
		return nil, err
	}

	// 设置颜色（总是设置，简化逻辑）
	fgColor, err := parseHexColor(config.Foreground)
	if err != nil {
		return nil, fmt.Errorf("invalid foreground color: %w", err)
	}

	bgColor, err := parseHexColor(config.Background)
	if err != nil {
		return nil, fmt.Errorf("invalid background color: %w", err)
	}

	qr.ForegroundColor = fgColor
	qr.BackgroundColor = bgColor

	// 直接生成PNG字节数据
	return qr.PNG(config.Size)
}

// GenerateQRCodeToFile 生成二维码并保存到文件
func GenerateQRCodeToFile(content, filename string, config QRCodeConfig) error {
	level := getRecoveryLevel(config.Level)

	// 创建QRCode对象
	qr, err := qrcode.New(content, level)
	if err != nil {
		return err
	}

	// 设置颜色（总是设置，简化逻辑）
	fgColor, err := parseHexColor(config.Foreground)
	if err != nil {
		return fmt.Errorf("invalid foreground color: %w", err)
	}

	bgColor, err := parseHexColor(config.Background)
	if err != nil {
		return fmt.Errorf("invalid background color: %w", err)
	}

	qr.ForegroundColor = fgColor
	qr.BackgroundColor = bgColor

	// 直接保存到文件
	return qr.WriteFile(config.Size, filename)
}

// GenerateQRCodeSimple 简化版二维码生成（使用默认配置）
func GenerateQRCodeSimple(content string) ([]byte, error) {
	return GenerateQRCode(content, DefaultConfig)
}

// GenerateQRCodeToFileSimple 简化版二维码生成并保存到文件
func GenerateQRCodeToFileSimple(content, filename string) error {
	return GenerateQRCodeToFile(content, filename, DefaultConfig)
}

// getRecoveryLevel 将字符串转换为RecoveryLevel
func getRecoveryLevel(level string) qrcode.RecoveryLevel {
	switch level {
	case "L":
		return qrcode.Low
	case "M":
		return qrcode.Medium
	case "Q":
		return qrcode.High
	case "H":
		return qrcode.Highest
	default:
		return qrcode.Medium
	}
}

// parseHexColor 解析十六进制颜色字符串
func parseHexColor(s string) (color.RGBA, error) {
	if len(s) != 7 || s[0] != '#' {
		return color.RGBA{}, fmt.Errorf("invalid color format: %s", s)
	}

	hexToByte := func(b byte) byte {
		switch {
		case b >= '0' && b <= '9':
			return b - '0'
		case b >= 'a' && b <= 'f':
			return b - 'a' + 10
		case b >= 'A' && b <= 'F':
			return b - 'A' + 10
		default:
			return 0
		}
	}

	r := hexToByte(s[1])<<4 + hexToByte(s[2])
	g := hexToByte(s[3])<<4 + hexToByte(s[4])
	b := hexToByte(s[5])<<4 + hexToByte(s[6])

	return color.RGBA{R: r, G: g, B: b, A: 255}, nil
}

// LogoConfig LOGO配置
type LogoConfig struct {
	Path     string  // LOGO文件路径
	Size     float64 // LOGO尺寸（相对于二维码尺寸的比例，0-1）
	Position string  // LOGO位置：center, top-left, top-right, bottom-left, bottom-right
	Opacity  float64 // LOGO透明度（0-1）
}

// DefaultLogoConfig 默认LOGO配置
var DefaultLogoConfig = LogoConfig{
	Size:     0.2, // 二维码尺寸的20%
	Position: "center",
	Opacity:  1.0, // 完全不透明
}

// GenerateQRCodeWithLogo 生成带LOGO的二维码
func GenerateQRCodeWithLogo(content string, config QRCodeConfig, logoConfig LogoConfig) ([]byte, error) {
	// 先生成基础二维码
	qrData, err := GenerateQRCode(content, config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate base QR code: %w", err)
	}

	// 如果没有提供LOGO路径，直接返回基础二维码
	if logoConfig.Path == "" {
		return qrData, nil
	}

	// 解码二维码图像
	qrcodeImg, _, err := image.Decode(bytes.NewReader(qrData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode QR code image: %w", err)
	}

	// 加载LOGO图像
	logoFile, err := os.Open(logoConfig.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open logo file: %w", err)
	}
	defer logoFile.Close()

	logoImg, _, err := image.Decode(logoFile)
	if err != nil {
		return nil, fmt.Errorf("failed to decode logo image: %w", err)
	}

	// 处理LOGO图像
	logoImg = processLogoImage(logoImg, qrcodeImg.Bounds().Dx(), logoConfig)

	// 合并二维码和LOGO
	resultImg := mergeQRCodeWithLogo(qrcodeImg, logoImg, logoConfig)

	// 编码为PNG
	var buf bytes.Buffer
	err = png.Encode(&buf, resultImg)
	if err != nil {
		return nil, fmt.Errorf("failed to encode PNG with logo: %w", err)
	}

	return buf.Bytes(), nil
}

// GenerateQRCodeToFileWithLogo 生成带LOGO的二维码并保存到文件
func GenerateQRCodeToFileWithLogo(content, filename string, config QRCodeConfig, logoConfig LogoConfig) error {
	data, err := GenerateQRCodeWithLogo(content, config, logoConfig)
	if err != nil {
		return err
	}

	// 确保目录存在
	err = os.MkdirAll(filepath.Dir(filename), 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 保存文件
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// processLogoImage 处理LOGO图像（缩放、调整透明度等）
func processLogoImage(logoImg image.Image, qrSize int, logoConfig LogoConfig) image.Image {
	// 计算LOGO尺寸
	logoSize := int(float64(qrSize) * logoConfig.Size)
	if logoSize < 10 {
		logoSize = 10 // 最小尺寸限制
	}

	// 缩放LOGO
	logoImg = imaging.Resize(logoImg, logoSize, logoSize, imaging.Lanczos)

	return logoImg
}

// mergeQRCodeWithLogo 合并二维码和LOGO
func mergeQRCodeWithLogo(qrcodeImg, logoImg image.Image, logoConfig LogoConfig) image.Image {
	qrBounds := qrcodeImg.Bounds()
	logoBounds := logoImg.Bounds()

	// 计算LOGO位置
	logoPos := calculateLogoPosition(qrBounds, logoBounds, logoConfig.Position)

	// 使用imaging.Overlay合并图像，支持透明度
	result := imaging.Overlay(qrcodeImg, logoImg, logoPos, logoConfig.Opacity)

	return result
}

// calculateLogoPosition 计算LOGO位置
func calculateLogoPosition(qrBounds, logoBounds image.Rectangle, position string) image.Point {
	qrWidth := qrBounds.Dx()
	qrHeight := qrBounds.Dy()
	logoWidth := logoBounds.Dx()
	logoHeight := logoBounds.Dy()

	switch position {
	case "top-left":
		return image.Point{X: 10, Y: 10}
	case "top-right":
		return image.Point{X: qrWidth - logoWidth - 10, Y: 10}
	case "bottom-left":
		return image.Point{X: 10, Y: qrHeight - logoHeight - 10}
	case "bottom-right":
		return image.Point{X: qrWidth - logoWidth - 10, Y: qrHeight - logoHeight - 10}
	default: // center
		return image.Point{
			X: (qrWidth - logoWidth) / 2,
			Y: (qrHeight - logoHeight) / 2,
		}
	}
}

// GenerateQRCodeWithLogoSimple 简化版带LOGO二维码生成
func GenerateQRCodeWithLogoSimple(content, logoPath string) ([]byte, error) {
	logoConfig := DefaultLogoConfig
	logoConfig.Path = logoPath
	return GenerateQRCodeWithLogo(content, DefaultConfig, logoConfig)
}

// GenerateQRCodeToFileWithLogoSimple 简化版带LOGO二维码生成并保存到文件
func GenerateQRCodeToFileWithLogoSimple(content, filename, logoPath string) error {
	logoConfig := DefaultLogoConfig
	logoConfig.Path = logoPath
	return GenerateQRCodeToFileWithLogo(content, filename, DefaultConfig, logoConfig)
}
