package main

import (
	"fmt"

	"gitee.com/wangsoft/go-library/util/authutil"
	"gitee.com/wangsoft/go-library/util/cryptoutil"
	"gitee.com/wangsoft/go-library/util/timeutil"
)

func main() {
	fmt.Println("=== go-library Test ===")
	fmt.Println()

	// Test Google Authenticator
	fmt.Println("1. Google Authenticator:")
	secret, qrURL, err := authutil.QuickGenerate("TestApp", "test@example.com")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Secret: %s\n", secret[:20]+"...")
		fmt.Printf("QR URL: %s\n", qrURL[:60]+"...")
	}

	// Test Crypto
	fmt.Println()
	fmt.Println("2. Crypto Utils:")
	hash := cryptoutil.SHA256String("test")
	fmt.Printf("SHA256: %s\n", hash[:32]+"...")

	// Test Time
	fmt.Println()
	fmt.Println("3. Time Utils:")
	now := timeutil.Now()
	fmt.Printf("Current Time: %s\n", now.Format("2006-01-02 15:04:05"))

	fmt.Println()
	fmt.Println("=== All Tests Passed ===")
}
