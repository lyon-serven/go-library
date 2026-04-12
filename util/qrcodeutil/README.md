# QRCodeUtil - 二维码生成工具

提供简单易用的二维码生成功能，支持自定义配置、Logo叠加等特性。

## 功能特性

- ✅ 基本二维码生成
- ✅ 自定义尺寸和容错级别
- ✅ 自定义颜色配置
- ✅ 保存到文件
- ✅ Logo叠加支持
- ✅ Logo位置和透明度自定义
- ✅ 简化API接口

## 快速开始

### 基本使用

```go
package main

import (
	"fmt"
	"github.com/lyon-serven/go-library/util/qrcodeutil"
)

func main() {
	// 生成简单二维码（字节数据）
	data, err := qrcodeutil.GenerateQRCodeSimple("https://example.com")
	if err != nil {
		panic(err)
	}
	fmt.Printf("QR code generated: %d bytes\n", len(data))

	// 保存到文件
	err = qrcodeutil.GenerateQRCodeToFileSimple("Hello World", "qrcode.png")
	if err != nil {
		panic(err)
	}
	fmt.Println("QR code saved to qrcode.png")
}
```

### 自定义配置

```go
config := qrcodeutil.QRCodeConfig{
	Size:       300,
	Level:      "H", // 最高容错级别
	Foreground: "#FF0000", // 红色前景
	Background: "#FFFF00", // 黄色背景
}

data, err := qrcodeutil.GenerateQRCode("Custom QR Code", config)
```

### 带Logo的二维码

```go
// 简单版
logoConfig := qrcodeutil.LogoConfig{
	Path:     "logo.png",
	Size:     0.2,      // 二维码尺寸的20%
	Position: "center", // 居中显示
	Opacity:  1.0,      // 完全不透明
}

data, err := qrcodeutil.GenerateQRCodeWithLogo(
	"https://example.com",
	qrcodeutil.DefaultConfig,
	logoConfig,
)

// 简化版
data, err := qrcodeutil.GenerateQRCodeWithLogoSimple("https://example.com", "logo.png")
```

## API参考

### 类型定义

#### QRCodeConfig

```go
type QRCodeConfig struct {
	Size       int    // 二维码尺寸 (像素)
	Level      string // 容错级别: L(7%), M(15%), Q(25%), H(30%)
	Foreground string // 前景色 (十六进制格式)
	Background string // 背景色 (十六进制格式)
}
```

### 类型定义

#### QRCodeConfig

```go
type QRCodeConfig struct {
	Size       int    // 二维码尺寸 (像素)
	Level      string // 容错级别: L(7%), M(15%), Q(25%), H(30%)
	Foreground string // 前景色 (十六进制格式)
	Background string // 背景色 (十六进制格式)
}
```

#### LogoConfig

```go
type LogoConfig struct {
	Path     string  // LOGO文件路径
	Size     float64 // LOGO尺寸（相对于二维码尺寸的比例，0-1）
	Position string  // LOGO位置：center, top-left, top-right, bottom-left, bottom-right
	Opacity  float64 // LOGO透明度（0-1）
}
```

### 函数列表

#### 基本功能

```go
func GenerateQRCode(content string, config QRCodeConfig) ([]byte, error)
func GenerateQRCodeToFile(content, filename string, config QRCodeConfig) error
func GenerateQRCodeSimple(content string) ([]byte, error)
func GenerateQRCodeToFileSimple(content, filename string) error
```

#### 颜色功能

```go
func GenerateQRCodeWithColor(content string, config QRCodeConfig) ([]byte, error)
func GenerateQRCodeToFileWithColor(content, filename string, config QRCodeConfig) error
```

#### LOGO功能

```go
func GenerateQRCodeWithLogo(content string, config QRCodeConfig, logoConfig LogoConfig) ([]byte, error)
func GenerateQRCodeToFileWithLogo(content, filename string, config QRCodeConfig, logoConfig LogoConfig) error
func GenerateQRCodeWithLogoSimple(content, logoPath string) ([]byte, error)
func GenerateQRCodeToFileWithLogoSimple(content, filename, logoPath string) error
```

## 默认配置

```go
var DefaultConfig = QRCodeConfig{
	Size:       256,
	Level:      "M",
	Foreground: "#000000",
	Background: "#FFFFFF",
}
```

## 错误处理

所有函数都返回error类型，建议在生产环境中进行适当的错误处理。

## 依赖

- `github.com/skip2/go-qrcode` - 二维码生成核心库
- `github.com/disintegration/imaging` - 图像处理库（用于LOGO功能）

## 示例

查看 `examples/` 目录中的示例代码了解更详细的使用方法。