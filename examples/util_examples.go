package main

import (
	"fmt"
	"log"
	"time"

	"mylib/util/cryptoutil"
	"mylib/util/httputil"
	"mylib/util/timeutil"
)

func main() {
	fmt.Println("=== MyLib Utility Classes Examples ===\n")

	// Time utilities examples
	timeUtilExamples()

	// Crypto utilities examples
	cryptoUtilExamples()

	// HTTP utilities examples
	httpUtilExamples()
}

func timeUtilExamples() {
	fmt.Println("⏰ Time Utilities Examples:")

	// Current time operations
	now := timeutil.Now()
	fmt.Printf("Current time: %s\n", timeutil.Format(now, timeutil.DateTimeFormat))
	fmt.Printf("Current time in Shanghai: %s\n", timeutil.Format(timeutil.NowInZone(timeutil.Shanghai), timeutil.DateTimeFormat))

	// Date calculations
	today := timeutil.Today()
	fmt.Printf("Today: %s\n", timeutil.Format(today, timeutil.DateFormat))

	startOfWeek := timeutil.StartOfWeek(now)
	endOfWeek := timeutil.EndOfWeek(now)
	fmt.Printf("This week: %s to %s\n",
		timeutil.Format(startOfWeek, timeutil.DateFormat),
		timeutil.Format(endOfWeek, timeutil.DateFormat))

	startOfMonth := timeutil.StartOfMonth(now)
	endOfMonth := timeutil.EndOfMonth(now)
	fmt.Printf("This month: %s to %s\n",
		timeutil.Format(startOfMonth, timeutil.DateFormat),
		timeutil.Format(endOfMonth, timeutil.DateFormat))

	// Business days
	fmt.Printf("Is today a business day: %v\n", timeutil.IsBusinessDay(now))
	nextBusinessDay := timeutil.AddBusinessDays(now, 5)
	fmt.Printf("5 business days from now: %s\n", timeutil.Format(nextBusinessDay, timeutil.DateFormat))

	// Age calculation
	birthdate, _ := timeutil.Parse("1990-05-15", timeutil.DateFormat)
	age := timeutil.Age(birthdate)
	fmt.Printf("Age for someone born on 1990-05-15: %d years\n", age)

	// Time range
	range1 := timeutil.NewTimeRange(startOfMonth, endOfMonth)
	fmt.Printf("Current month duration: %s\n", timeutil.FormatDuration(range1.Duration()))

	// Performance measurement
	duration := timeutil.Benchmark(func() {
		time.Sleep(10 * time.Millisecond)
		fmt.Printf("  → Executed expensive operation\n")
	})
	fmt.Printf("Operation took: %s\n", timeutil.FormatDuration(duration))

	// Quarter calculations
	quarter := timeutil.Quarter(now)
	fmt.Printf("Current quarter: Q%d\n", quarter)

	fmt.Println()
}

func cryptoUtilExamples() {
	fmt.Println("🔐 Crypto Utilities Examples:")

	// Hash functions
	data := "Hello, Crypto World!"
	fmt.Printf("Original data: %s\n", data)
	fmt.Printf("MD5: %s\n", cryptoutil.MD5String(data))
	fmt.Printf("SHA1: %s\n", cryptoutil.SHA1String(data))
	fmt.Printf("SHA256: %s\n", cryptoutil.SHA256String(data))
	fmt.Printf("SHA512: %s\n", cryptoutil.SHA512String(data))

	// HMAC
	secret := "my-secret-key"
	hmacSHA256, _ := cryptoutil.HMACString(data, secret, cryptoutil.SHA256)
	fmt.Printf("HMAC-SHA256: %s\n", hmacSHA256)

	// Verify HMAC
	valid, _ := cryptoutil.VerifyHMACString(data, secret, hmacSHA256, cryptoutil.SHA256)
	fmt.Printf("HMAC verification: %v\n", valid)

	// Base64 encoding
	encoded := cryptoutil.Base64Encode([]byte(data))
	fmt.Printf("Base64 encoded: %s\n", encoded)

	decoded, _ := cryptoutil.Base64Decode(encoded)
	fmt.Printf("Base64 decoded: %s\n", string(decoded))

	// Random generation
	randomBytes, _ := cryptoutil.GenerateRandomBytes(16)
	fmt.Printf("Random bytes (hex): %s\n", cryptoutil.HexEncode(randomBytes))

	randomString, _ := cryptoutil.GenerateRandomString(16, cryptoutil.HexEncoding)
	fmt.Printf("Random string: %s\n", randomString)

	// AES encryption/decryption
	key := "my-32-byte-secret-key-for-aes!!"
	plaintext := "This is a secret message that needs to be encrypted!"

	encrypted, err := cryptoutil.AESEncryptString(plaintext, key)
	if err != nil {
		log.Printf("AES encryption error: %v", err)
	} else {
		fmt.Printf("AES encrypted: %s\n", encrypted[:50]+"...")

		decrypted, err := cryptoutil.AESDecryptString(encrypted, key)
		if err != nil {
			log.Printf("AES decryption error: %v", err)
		} else {
			fmt.Printf("AES decrypted: %s\n", decrypted)
		}
	}

	// Simple encryption (password-based)
	password := "my-password"
	simpleEncrypted, err := cryptoutil.SimpleEncrypt(plaintext, password)
	if err != nil {
		log.Printf("Simple encryption error: %v", err)
	} else {
		fmt.Printf("Simple encrypted: %s\n", simpleEncrypted[:50]+"...")

		simpleDecrypted, err := cryptoutil.SimpleDecrypt(simpleEncrypted, password)
		if err != nil {
			log.Printf("Simple decryption error: %v", err)
		} else {
			fmt.Printf("Simple decrypted: %s\n", simpleDecrypted)
		}
	}

	// RSA key generation
	fmt.Println("Generating RSA key pair...")
	keyPair, err := cryptoutil.GenerateRSAKeyPair(2048)
	if err != nil {
		log.Printf("RSA key generation error: %v", err)
	} else {
		fmt.Printf("RSA key pair generated successfully\n")

		// RSA encryption/decryption
		message := "RSA encrypted message"
		rsaEncrypted, err := cryptoutil.RSAEncrypt([]byte(message), keyPair.PublicKey)
		if err != nil {
			log.Printf("RSA encryption error: %v", err)
		} else {
			fmt.Printf("RSA encrypted message length: %d bytes\n", len(rsaEncrypted))

			rsaDecrypted, err := cryptoutil.RSADecrypt(rsaEncrypted, keyPair.PrivateKey)
			if err != nil {
				log.Printf("RSA decryption error: %v", err)
			} else {
				fmt.Printf("RSA decrypted: %s\n", string(rsaDecrypted))
			}
		}

		// RSA signature
		signature, err := cryptoutil.RSASign([]byte(message), keyPair.PrivateKey)
		if err != nil {
			log.Printf("RSA signing error: %v", err)
		} else {
			fmt.Printf("RSA signature length: %d bytes\n", len(signature))

			err = cryptoutil.RSAVerify([]byte(message), signature, keyPair.PublicKey)
			if err != nil {
				log.Printf("RSA verification failed: %v", err)
			} else {
				fmt.Printf("RSA signature verified successfully\n")
			}
		}
	}

	fmt.Println()
}

func httpUtilExamples() {
	fmt.Println("🌐 HTTP Utilities Examples:")

	// Create HTTP client with configuration
	client := httputil.NewClient(
		httputil.WithTimeout(10*time.Second),
		httputil.WithUserAgent("MyLib/1.0"),
		httputil.WithHeader("Accept", "application/json"),
	)
	fmt.Printf("✓ Created HTTP client with 10s timeout and custom headers\n")
	_ = client // Client ready for HTTP requests

	// URL utilities
	baseURL := "https://api.example.com/users"
	params := map[string]string{
		"page": "1",
		"size": "10",
		"sort": "name",
	}

	fullURL, err := httputil.BuildURL(baseURL, params)
	if err != nil {
		log.Printf("URL building error: %v", err)
	} else {
		fmt.Printf("Built URL: %s\n", fullURL)
	}

	// URL validation
	fmt.Printf("Is valid URL: %v\n", httputil.IsValidURL(fullURL))
	fmt.Printf("Is valid URL (invalid): %v\n", httputil.IsValidURL("not-a-url"))

	// URL encoding/decoding
	originalText := "Hello World & Special Characters!"
	encoded := httputil.URLEncode(originalText)
	fmt.Printf("URL encoded: %s\n", encoded)

	decoded, err := httputil.URLDecode(encoded)
	if err != nil {
		log.Printf("URL decode error: %v", err)
	} else {
		fmt.Printf("URL decoded: %s\n", decoded)
	}

	// Query parameter manipulation
	testURL := "https://api.example.com/search?q=golang"
	param, _ := httputil.GetQueryParam(testURL, "q")
	fmt.Printf("Query param 'q': %s\n", param)

	newURL, _ := httputil.SetQueryParam(testURL, "limit", "50")
	fmt.Printf("URL with new param: %s\n", newURL)

	// HTTP request examples (using mock/test endpoints)
	fmt.Println("\n--- HTTP Request Examples ---")
	fmt.Println("Note: These examples use hypothetical endpoints")

	// Simple GET request
	fmt.Println("Example GET request structure:")
	req := httputil.NewRequest("GET", "/api/users").
		WithQuery("page", "1").
		WithQuery("limit", "10").
		WithHeader("Accept", "application/json")

	fmt.Printf("  Method: %s\n", req.Method)
	fmt.Printf("  URL: %s\n", req.URL)
	fmt.Printf("  Query params: %v\n", req.QueryParams)
	fmt.Printf("  Headers: %v\n", req.Headers)

	// POST request with JSON body
	fmt.Println("\nExample POST request structure:")
	userData := map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
	}

	postReq := httputil.NewRequest("POST", "/api/users").
		WithJSON(userData).
		WithHeader("Content-Type", "application/json")

	fmt.Printf("  Method: %s\n", postReq.Method)
	fmt.Printf("  URL: %s\n", postReq.URL)
	fmt.Printf("  Body: %+v\n", postReq.Body)
	fmt.Printf("  Content-Type: %s\n", postReq.ContentType)

	// Form data request
	fmt.Println("\nExample form data request structure:")
	formData := map[string]string{
		"username": "johndoe",
		"password": "secret123",
	}

	formReq := httputil.NewRequest("POST", "/api/login").
		WithForm(formData)

	fmt.Printf("  Method: %s\n", formReq.Method)
	fmt.Printf("  URL: %s\n", formReq.URL)
	fmt.Printf("  Content-Type: %s\n", formReq.ContentType)

	// Client with different configurations
	fmt.Println("\n--- HTTP Client Configurations ---")

	// Client with base URL
	_ = httputil.NewClient(
		httputil.WithBaseURL("https://api.myservice.com"),
		httputil.WithTimeout(15*time.Second),
	)
	fmt.Printf("✓ API client configured with base URL and 15s timeout\n")

	// Client with authentication
	_ = httputil.NewClient(
		httputil.WithBearerToken("your-jwt-token-here"),
		httputil.WithUserAgent("MyApp/2.0"),
	)
	fmt.Printf("✓ Auth client configured with Bearer token\n")

	// Client with basic auth
	_ = httputil.NewClient(
		httputil.WithBasicAuth("username", "password"),
	)
	fmt.Printf("✓ Basic auth client configured\n")

	// Retry configuration
	fmt.Println("\n--- Retry Configuration ---")
	retryConfig := httputil.DefaultRetryConfig()
	fmt.Printf("Default retry config: MaxRetries=%d, Delay=%v\n",
		retryConfig.MaxRetries, retryConfig.Delay)

	// Custom retry config
	customRetryConfig := &httputil.RetryConfig{
		MaxRetries: 5,
		Delay:      2 * time.Second,
		BackoffFn: func(attempt int) time.Duration {
			return time.Duration(attempt*attempt) * time.Second // Exponential backoff
		},
	}
	fmt.Printf("Custom retry config: MaxRetries=%d, Initial Delay=%v\n",
		customRetryConfig.MaxRetries, customRetryConfig.Delay)

	fmt.Println("\n✅ HTTP utilities examples completed!")
	fmt.Println("Note: Actual HTTP requests were not made in this demo.")
}
