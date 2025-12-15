package main

import (
	"fmt"

	"gitee.com/wangsoft/go-library/util/cryptoutil"
)

func main() {
	fmt.Println("=== DES 兼容性测试 ===\n")

	key := "12345678" // 8字节密钥
	originalText := "Hello, World!"

	fmt.Printf("原始文本: %s\n", originalText)
	fmt.Printf("密钥: %s\n\n", key)

	// 测试 1: 使用固定 IV 的 Hex 编码版本（兼容你的代码）
	fmt.Println("【方式1】DES + 固定IV + Hex编码（兼容模式）")
	encrypted1, err := cryptoutil.DESEncryptHex(originalText, key)
	if err != nil {
		fmt.Printf("加密失败: %v\n", err)
		return
	}
	fmt.Printf("加密结果 (Hex): %s\n", encrypted1)

	decrypted1, err := cryptoutil.DESDecryptHex(encrypted1, key)
	if err != nil {
		fmt.Printf("解密失败: %v\n", err)
		return
	}
	fmt.Printf("解密结果: %s\n", decrypted1)

	if originalText == decrypted1 {
		fmt.Println("✓ 加密解密成功!\n")
	} else {
		fmt.Println("✗ 加密解密失败!\n")
	}

	// 测试 2: 使用随机 IV 的 Base64 编码版本（安全模式）
	fmt.Println("【方式2】DES + 随机IV + Base64编码（安全模式）")
	encrypted2, err := cryptoutil.DESEncryptString(originalText, key)
	if err != nil {
		fmt.Printf("加密失败: %v\n", err)
		return
	}
	fmt.Printf("加密结果 (Base64): %s\n", encrypted2)

	decrypted2, err := cryptoutil.DESDecryptString(encrypted2, key)
	if err != nil {
		fmt.Printf("解密失败: %v\n", err)
		return
	}
	fmt.Printf("解密结果: %s\n", decrypted2)

	if originalText == decrypted2 {
		fmt.Println("✓ 加密解密成功!\n")
	} else {
		fmt.Println("✗ 加密解密失败!\n")
	}

	// 测试 3: 空字符串处理
	fmt.Println("【测试】空字符串处理")
	emptyEncrypted, err := cryptoutil.DESEncryptHex("", key)
	if err != nil {
		fmt.Printf("加密失败: %v\n", err)
		return
	}
	fmt.Printf("空字符串加密结果: '%s'\n", emptyEncrypted)

	emptyDecrypted, err := cryptoutil.DESDecryptHex("", key)
	if err != nil {
		fmt.Printf("解密失败: %v\n", err)
		return
	}
	fmt.Printf("空字符串解密结果: '%s'\n", emptyDecrypted)
	fmt.Println("✓ 空字符串处理正确!\n")

	// 测试 4: 多次加密，固定IV每次结果相同
	fmt.Println("【对比】固定IV vs 随机IV")
	enc1a, _ := cryptoutil.DESEncryptHex(originalText, key)
	enc1b, _ := cryptoutil.DESEncryptHex(originalText, key)
	fmt.Printf("固定IV加密两次:\n  第一次: %s\n  第二次: %s\n", enc1a, enc1b)
	if enc1a == enc1b {
		fmt.Println("  ✓ 固定IV: 相同输入产生相同密文")
	}

	enc2a, _ := cryptoutil.DESEncryptString(originalText, key)
	enc2b, _ := cryptoutil.DESEncryptString(originalText, key)
	fmt.Printf("随机IV加密两次:\n  第一次: %s\n  第二次: %s\n", enc2a, enc2b)
	if enc2a != enc2b {
		fmt.Println("  ✓ 随机IV: 相同输入产生不同密文（更安全）")
	}

	fmt.Println("\n=== 总结 ===")
	fmt.Println("✓ DESEncryptHex/DESDecryptHex - 兼容你的旧代码（固定IV + Hex）")
	fmt.Println("✓ DESEncryptString/DESDecryptString - 新的安全实现（随机IV + Base64）")
	fmt.Println("✓ 建议: 新项目使用随机IV版本，旧数据使用固定IV版本解密")
}
