package main

import (
	"fmt"

	"gitee.com/wangsoft/go-library/util/cryptoutil"
)

func main() {
	fmt.Println("=== DES 加密测试 ===")

	// 测试 DES 加密解密
	desKey := "12345678" // DES 密钥必须是 8 字节
	originalText := "Hello, DES Encryption!"

	fmt.Printf("原始文本: %s\n", originalText)
	fmt.Printf("DES 密钥: %s (8字节)\n\n", desKey)

	// 加密
	encrypted, err := cryptoutil.DESEncryptString(originalText, desKey)
	if err != nil {
		fmt.Printf("加密失败: %v\n", err)
		return
	}
	fmt.Printf("加密结果 (Base64): %s\n", encrypted)

	// 解密
	decrypted, err := cryptoutil.DESDecryptString(encrypted, desKey)
	if err != nil {
		fmt.Printf("解密失败: %v\n", err)
		return
	}
	fmt.Printf("解密结果: %s\n\n", decrypted)

	if originalText == decrypted {
		fmt.Println("✓ DES 加密解密测试通过!")
	} else {
		fmt.Println("✗ DES 加密解密测试失败!")
	}

	fmt.Println("\n=== 3DES 加密测试 ===")

	// 测试 3DES 加密解密
	tripleDesKey := "123456789012345678901234" // 3DES 密钥必须是 24 字节
	originalText2 := "Hello, 3DES Encryption!"

	fmt.Printf("原始文本: %s\n", originalText2)
	fmt.Printf("3DES 密钥: %s (24字节)\n\n", tripleDesKey)

	// 加密
	encrypted2, err := cryptoutil.TripleDESEncryptString(originalText2, tripleDesKey)
	if err != nil {
		fmt.Printf("加密失败: %v\n", err)
		return
	}
	fmt.Printf("加密结果 (Base64): %s\n", encrypted2)

	// 解密
	decrypted2, err := cryptoutil.TripleDESDecryptString(encrypted2, tripleDesKey)
	if err != nil {
		fmt.Printf("解密失败: %v\n", err)
		return
	}
	fmt.Printf("解密结果: %s\n\n", decrypted2)

	if originalText2 == decrypted2 {
		fmt.Println("✓ 3DES 加密解密测试通过!")
	} else {
		fmt.Println("✗ 3DES 加密解密测试失败!")
	}

	fmt.Println("\n=== 错误处理测试 ===")

	// 测试错误密钥长度
	_, err = cryptoutil.DESEncryptString("test", "short")
	if err != nil {
		fmt.Printf("✓ 正确捕获 DES 密钥长度错误: %v\n", err)
	}

	_, err = cryptoutil.TripleDESEncryptString("test", "short")
	if err != nil {
		fmt.Printf("✓ 正确捕获 3DES 密钥长度错误: %v\n", err)
	}
}
