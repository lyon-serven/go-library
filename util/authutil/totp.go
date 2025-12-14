// Package authutil provides utilities for authentication, including Google Authenticator (TOTP) functionality.
package authutil

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// TOTPConfig holds configuration for TOTP generation
type TOTPConfig struct {
	// Secret key (base32 encoded)
	Secret string
	// Issuer name (e.g., "MyCompany")
	Issuer string
	// Account name (e.g., user email)
	AccountName string
	// Number of digits in the code (usually 6)
	Digits int
	// Time step in seconds (usually 30)
	Period int
	// Hash algorithm (sha1, sha256, sha512)
	Algorithm string
}

// DefaultTOTPConfig returns default TOTP configuration
func DefaultTOTPConfig() *TOTPConfig {
	return &TOTPConfig{
		Digits:    6,
		Period:    30,
		Algorithm: "sha1",
	}
}

// GoogleAuthenticator represents a Google Authenticator helper
type GoogleAuthenticator struct {
	config *TOTPConfig
}

// NewGoogleAuthenticator creates a new Google Authenticator instance
func NewGoogleAuthenticator(config *TOTPConfig) *GoogleAuthenticator {
	if config == nil {
		config = DefaultTOTPConfig()
	}
	return &GoogleAuthenticator{config: config}
}

// GenerateSecret generates a new secret key for TOTP
func (ga *GoogleAuthenticator) GenerateSecret() (string, error) {
	// Generate 20 random bytes (160 bits)
	secret := make([]byte, 20)
	_, err := rand.Read(secret)
	if err != nil {
		return "", fmt.Errorf("failed to generate random secret: %w", err)
	}

	// Encode to base32 (without padding for compatibility)
	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secret)
	return encoded, nil
}

// GenerateCode generates a TOTP code for the current time
func (ga *GoogleAuthenticator) GenerateCode(secret string) (string, error) {
	return ga.GenerateCodeAtTime(secret, time.Now())
}

// GenerateCodeAtTime generates a TOTP code for a specific time
func (ga *GoogleAuthenticator) GenerateCodeAtTime(secret string, timestamp time.Time) (string, error) {
	// Decode base32 secret
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil {
		return "", fmt.Errorf("failed to decode secret: %w", err)
	}

	// Calculate time counter (number of time steps since Unix epoch)
	counter := uint64(timestamp.Unix()) / uint64(ga.config.Period)

	// Generate HOTP value
	code := ga.generateHOTP(key, counter)

	// Format with leading zeros
	format := "%0" + strconv.Itoa(ga.config.Digits) + "d"
	return fmt.Sprintf(format, code), nil
}

// VerifyCode verifies a TOTP code against the current time (with tolerance)
func (ga *GoogleAuthenticator) VerifyCode(secret, code string) bool {
	return ga.VerifyCodeWithTolerance(secret, code, 1)
}

// VerifyCodeWithTolerance verifies a TOTP code with time tolerance
// tolerance: number of time steps to check before and after current time
func (ga *GoogleAuthenticator) VerifyCodeWithTolerance(secret, code string, tolerance int) bool {
	now := time.Now()

	// Check current time and tolerance range
	for i := -tolerance; i <= tolerance; i++ {
		checkTime := now.Add(time.Duration(i) * time.Duration(ga.config.Period) * time.Second)
		expectedCode, err := ga.GenerateCodeAtTime(secret, checkTime)
		if err != nil {
			continue
		}
		if code == expectedCode {
			return true
		}
	}

	return false
}

// GenerateQRCodeURL generates a QR code URL for Google Authenticator
func (ga *GoogleAuthenticator) GenerateQRCodeURL(secret string) string {
	// Build otpauth URL
	params := url.Values{}
	params.Set("secret", secret)
	params.Set("issuer", ga.config.Issuer)
	params.Set("algorithm", strings.ToUpper(ga.config.Algorithm))
	params.Set("digits", strconv.Itoa(ga.config.Digits))
	params.Set("period", strconv.Itoa(ga.config.Period))

	// Build the otpauth URL
	otpauthURL := fmt.Sprintf("otpauth://totp/%s:%s?%s",
		url.QueryEscape(ga.config.Issuer),
		url.QueryEscape(ga.config.AccountName),
		params.Encode())

	return otpauthURL
}

// GenerateQRCodeImageURL generates a URL to display QR code image via Google Charts API
// Note: Google Charts API requires VPN access in China
func (ga *GoogleAuthenticator) GenerateQRCodeImageURL(secret string) string {
	otpauthURL := ga.GenerateQRCodeURL(secret)

	// Use Google Charts API to generate QR code image (需要翻墙)
	qrURL := fmt.Sprintf("https://chart.googleapis.com/chart?chs=200x200&chld=M|0&cht=qr&chl=%s",
		url.QueryEscape(otpauthURL))

	return qrURL
}

// GenerateQRCodeImageURLCN generates a QR code image URL using Chinese accessible APIs
// 使用国内可访问的二维码生成API（多个备选方案）
func (ga *GoogleAuthenticator) GenerateQRCodeImageURLCN(secret string) string {
	otpauthURL := ga.GenerateQRCodeURL(secret)

	// 方案1: 使用QR Server API (国内可访问，免费)
	// qrURL := fmt.Sprintf("https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=%s",
	// 	url.QueryEscape(otpauthURL))

	// 方案2: 使用草料二维码API (需要申请API key)
	// qrURL := fmt.Sprintf("https://api.cli.im/qrcode/code?text=%s",
	// 	url.QueryEscape(otpauthURL))

	// 方案3: 使用联图网二维码API (免费，国内稳定)
	qrURL := fmt.Sprintf("https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=%s",
		url.QueryEscape(otpauthURL))

	return qrURL
}

// GetOtpauthURL returns the raw otpauth:// URL for manual QR code generation
// 返回原始的 otpauth:// URL，可用于本地生成二维码
func (ga *GoogleAuthenticator) GetOtpauthURL(secret string) string {
	return ga.GenerateQRCodeURL(secret)
}

// generateHOTP generates HOTP value based on RFC 4226
func (ga *GoogleAuthenticator) generateHOTP(key []byte, counter uint64) uint32 {
	// Convert counter to big-endian byte array
	counterBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(counterBytes, counter)

	// Generate HMAC-SHA1 hash
	h := hmac.New(sha1.New, key)
	h.Write(counterBytes)
	hash := h.Sum(nil)

	// Dynamic truncation
	offset := hash[19] & 0x0f
	truncated := binary.BigEndian.Uint32(hash[offset:offset+4]) & 0x7fffffff

	// Generate code with specified number of digits
	modulo := uint32(1)
	for i := 0; i < ga.config.Digits; i++ {
		modulo *= 10
	}

	return truncated % modulo
}

// GetRemainingTime returns the remaining time in seconds until the next code change
func (ga *GoogleAuthenticator) GetRemainingTime() int {
	now := time.Now()
	elapsed := int(now.Unix()) % ga.config.Period
	return ga.config.Period - elapsed
}

// BackupCodes represents backup recovery codes
type BackupCodes struct {
	Codes     []string  `json:"codes"`
	Generated time.Time `json:"generated"`
	Used      []bool    `json:"used"`
}

// GenerateBackupCodes generates backup recovery codes
func GenerateBackupCodes(count int) (*BackupCodes, error) {
	if count <= 0 {
		count = 10 // Default to 10 backup codes
	}

	codes := make([]string, count)
	used := make([]bool, count)

	for i := 0; i < count; i++ {
		// Generate 8-digit backup code
		codeBytes := make([]byte, 4)
		_, err := rand.Read(codeBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to generate backup code: %w", err)
		}

		// Convert to 8-digit code
		code := binary.BigEndian.Uint32(codeBytes) % 100000000
		codes[i] = fmt.Sprintf("%08d", code)
		used[i] = false
	}

	return &BackupCodes{
		Codes:     codes,
		Generated: time.Now(),
		Used:      used,
	}, nil
}

// UseBackupCode marks a backup code as used
func (bc *BackupCodes) UseBackupCode(code string) bool {
	for i, backupCode := range bc.Codes {
		if backupCode == code && !bc.Used[i] {
			bc.Used[i] = true
			return true
		}
	}
	return false
}

// GetUnusedCodes returns all unused backup codes
func (bc *BackupCodes) GetUnusedCodes() []string {
	var unused []string
	for i, code := range bc.Codes {
		if !bc.Used[i] {
			unused = append(unused, code)
		}
	}
	return unused
}

// TOTPManager manages TOTP for multiple users
type TOTPManager struct {
	secrets map[string]string // userID -> secret
	ga      *GoogleAuthenticator
}

// NewTOTPManager creates a new TOTP manager
func NewTOTPManager(issuer string) *TOTPManager {
	config := DefaultTOTPConfig()
	config.Issuer = issuer

	return &TOTPManager{
		secrets: make(map[string]string),
		ga:      NewGoogleAuthenticator(config),
	}
}

// SetupUser sets up TOTP for a user
func (tm *TOTPManager) SetupUser(userID, accountName string) (string, string, error) {
	// Generate secret
	secret, err := tm.ga.GenerateSecret()
	if err != nil {
		return "", "", err
	}

	// Store secret
	tm.secrets[userID] = secret

	// Update account name
	tm.ga.config.AccountName = accountName

	// Generate QR code URL
	qrURL := tm.ga.GenerateQRCodeImageURL(secret)

	return secret, qrURL, nil
}

// VerifyUserCode verifies TOTP code for a user
func (tm *TOTPManager) VerifyUserCode(userID, code string) bool {
	secret, exists := tm.secrets[userID]
	if !exists {
		return false
	}

	return tm.ga.VerifyCode(secret, code)
}

// GetUserSecret gets the secret for a user (for backup purposes)
func (tm *TOTPManager) GetUserSecret(userID string) (string, bool) {
	secret, exists := tm.secrets[userID]
	return secret, exists
}

// RemoveUser removes TOTP setup for a user
func (tm *TOTPManager) RemoveUser(userID string) {
	delete(tm.secrets, userID)
}

// Simple utility functions for quick usage

// QuickGenerate generates a secret and QR code URL with default settings
// 默认使用国内可访问的QR Server API生成二维码
func QuickGenerate(issuer, accountName string) (secret, qrCodeURL string, err error) {
	config := DefaultTOTPConfig()
	config.Issuer = issuer
	config.AccountName = accountName

	ga := NewGoogleAuthenticator(config)

	secret, err = ga.GenerateSecret()
	if err != nil {
		return "", "", err
	}

	qrCodeURL = ga.GenerateQRCodeImageURLCN(secret)
	return secret, qrCodeURL, nil
}

// QuickGenerateWithGoogle generates a secret and QR code URL using Google Charts API
// 使用Google Charts API（需要翻墙）
func QuickGenerateWithGoogle(issuer, accountName string) (secret, qrCodeURL string, err error) {
	config := DefaultTOTPConfig()
	config.Issuer = issuer
	config.AccountName = accountName

	ga := NewGoogleAuthenticator(config)

	secret, err = ga.GenerateSecret()
	if err != nil {
		return "", "", err
	}

	qrCodeURL = ga.GenerateQRCodeImageURL(secret)
	return secret, qrCodeURL, nil
}

// QuickVerify verifies a TOTP code with default settings
func QuickVerify(secret, code string) bool {
	ga := NewGoogleAuthenticator(DefaultTOTPConfig())
	return ga.VerifyCode(secret, code)
}

// QuickGenerateCode generates a TOTP code with default settings
func QuickGenerateCode(secret string) (string, error) {
	ga := NewGoogleAuthenticator(DefaultTOTPConfig())
	return ga.GenerateCode(secret)
}
