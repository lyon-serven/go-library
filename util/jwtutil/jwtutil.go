// Package jwtutil 提供JWT（JSON Web Token）生成、验证和管理的工具函数
package jwtutil

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// 常见错误
var (
	ErrInvalidToken     = errors.New("无效的令牌格式")
	ErrTokenExpired     = errors.New("令牌已过期")
	ErrTokenNotValidYet = errors.New("令牌尚未生效")
	ErrInvalidSignature = errors.New("无效的令牌签名")
	ErrInvalidIssuer    = errors.New("无效的令牌发行者")
	ErrInvalidAudience  = errors.New("无效的令牌受众")
	ErrMissingClaims    = errors.New("缺少必需的声明")
)

// 算法类型
const (
	HS256 = "HS256"
	HS384 = "HS384"
	HS512 = "HS512"
)

// 配置常量
const (
	MinSecretKeyLength = 10              // 最小密钥长度(256位)
	DefaultClockSkew   = 5 * time.Second // 默认时钟偏移容忍度
)

// StandardClaims 表示标准JWT声明
type StandardClaims struct {
	Issuer    string `json:"iss,omitempty"` // 发行者
	Subject   string `json:"sub,omitempty"` // 主题
	Audience  string `json:"aud,omitempty"` // 受众
	ExpiresAt int64  `json:"exp,omitempty"` // 过期时间（Unix时间戳）
	RefreshAt int64  `json:"ref,omitempty"` // 刷新时间（Unix时间戳）
	NotBefore int64  `json:"nbf,omitempty"` // 生效时间（Unix时间戳）
	IssuedAt  int64  `json:"iat,omitempty"` // 签发时间（Unix时间戳）
	ID        string `json:"jti,omitempty"` // JWT ID
}

// Claims 表示包含自定义数据的JWT声明
type Claims struct {
	StandardClaims
	CustomClaims map[string]interface{} `json:"custom,omitempty"`
}

// Header 表示JWT头部
type Header struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
}

// JWTConfig JWT操作的配置
type JWTConfig struct {
	SecretKey             string        // 签名密钥
	Algorithm             string        // 使用的算法（HS256, HS384, HS512）
	Issuer                string        // 默认发行者
	Audience              string        // 默认受众
	ExpiryDuration        time.Duration // 默认过期时间
	RefreshExpiryDuration time.Duration // 刷新令牌过期时间
	ClockSkew             time.Duration // 时钟偏移容忍度
}

// NewJWTConfig 创建新的JWT配置，使用默认值
func NewJWTConfig(secretKey string) *JWTConfig {
	return &JWTConfig{
		SecretKey:             secretKey,
		Algorithm:             HS256,
		ExpiryDuration:        24 * time.Hour,
		RefreshExpiryDuration: 7 * 24 * time.Hour, // 默认7天
		ClockSkew:             DefaultClockSkew,
	}
}

// validateSecretKey 验证密钥强度
func (c *JWTConfig) validateSecretKey() error {
	if c.SecretKey == "" {
		return errors.New("密钥不能为空")
	}
	if len(c.SecretKey) < MinSecretKeyLength {
		return fmt.Errorf("密钥太短：至少需要 %d 字节，实际 %d 字节", MinSecretKeyLength, len(c.SecretKey))
	}
	return nil
}

// TokenPair 表示访问令牌和刷新令牌对
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

// GenerateToken 生成JWT令牌
func (c *JWTConfig) GenerateToken(claims *Claims) (string, error) {
	if err := c.validateSecretKey(); err != nil {
		return "", err
	}

	// 设置默认值
	now := time.Now().Unix()
	if claims.IssuedAt == 0 {
		claims.IssuedAt = now
	}
	if claims.ExpiresAt == 0 {
		claims.ExpiresAt = time.Now().Add(c.ExpiryDuration).Unix()
	}
	if claims.Issuer == "" && c.Issuer != "" {
		claims.Issuer = c.Issuer
	}
	if claims.Audience == "" && c.Audience != "" {
		claims.Audience = c.Audience
	}

	// 创建头部
	header := Header{
		Algorithm: c.Algorithm,
		Type:      "JWT",
	}

	// 编码头部
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("编码头部失败: %w", err)
	}
	headerEncoded := base64.RawURLEncoding.EncodeToString(headerJSON)

	// 编码载荷
	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("编码声明失败: %w", err)
	}
	payloadEncoded := base64.RawURLEncoding.EncodeToString(payloadJSON)

	// 创建签名
	message := headerEncoded + "." + payloadEncoded
	signature, err := c.sign(message)
	if err != nil {
		return "", fmt.Errorf("签名令牌失败: %w", err)
	}

	// 组合所有部分
	token := message + "." + signature
	return token, nil
}

// GenerateTokenPair 生成访问令牌和刷新令牌对（如果配置了RefreshExpiryDuration）
func (c *JWTConfig) GenerateTokenPair(claims *Claims) (*TokenPair, error) {
	// 生成访问令牌
	accessToken, err := c.GenerateToken(claims)
	if err != nil {
		return nil, err
	}

	result := &TokenPair{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   claims.ExpiresAt - time.Now().Unix(),
	}

	// 如果配置了刷新令牌过期时间，则生成刷新令牌
	if c.RefreshExpiryDuration > 0 {
		refreshClaims := &Claims{
			StandardClaims: StandardClaims{
				Subject:   claims.Subject,
				Issuer:    claims.Issuer,
				Audience:  claims.Audience,
				IssuedAt:  claims.IssuedAt,
				ExpiresAt: time.Now().Add(c.RefreshExpiryDuration).Unix(), // 使用配置的刷新令牌过期时间
				ID:        claims.ID,
			},
			CustomClaims: map[string]interface{}{
				"type": "refresh",
			},
		}

		refreshToken, err := c.GenerateToken(refreshClaims)
		if err != nil {
			return nil, fmt.Errorf("生成刷新令牌失败: %w", err)
		}
		result.RefreshToken = refreshToken
	}

	return result, nil
}

// GenerateTokenSimple generates a JWT token with simple claims (subject and custom data)
func (c *JWTConfig) GenerateTokenSimple(subject string, customClaims map[string]interface{}) (string, error) {
	claims := &Claims{
		StandardClaims: StandardClaims{
			Subject: subject,
		},
		CustomClaims: customClaims,
	}
	return c.GenerateToken(claims)
}

// VerifyToken 验证并解析JWT令牌
func (c *JWTConfig) VerifyToken(tokenString string) (*Claims, error) {
	if err := c.validateSecretKey(); err != nil {
		return nil, err
	}

	// 分割令牌为三部分
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	headerEncoded := parts[0]
	payloadEncoded := parts[1]
	signatureEncoded := parts[2]

	// 验证签名
	message := headerEncoded + "." + payloadEncoded
	expectedSignature, err := c.sign(message)
	if err != nil {
		return nil, fmt.Errorf("生成签名失败: %w", err)
	}

	if !hmac.Equal([]byte(signatureEncoded), []byte(expectedSignature)) {
		return nil, ErrInvalidSignature
	}

	// 解码头部
	var header Header
	headerJSON, err := base64.RawURLEncoding.DecodeString(headerEncoded)
	if err != nil {
		return nil, fmt.Errorf("解码头部失败: %w", err)
	}
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return nil, fmt.Errorf("解析头部失败: %w", err)
	}

	// 解码载荷
	var claims Claims
	payloadJSON, err := base64.RawURLEncoding.DecodeString(payloadEncoded)
	if err != nil {
		return nil, fmt.Errorf("解码载荷失败: %w", err)
	}
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return nil, fmt.Errorf("解析声明失败: %w", err)
	}

	// 验证声明时间，使用时钟偏移容忍度
	now := time.Now().Unix()
	skew := int64(c.ClockSkew.Seconds())
	if skew == 0 {
		skew = int64(DefaultClockSkew.Seconds())
	}

	if claims.ExpiresAt > 0 && now > claims.ExpiresAt+skew {
		return nil, ErrTokenExpired
	}
	if claims.NotBefore > 0 && now < claims.NotBefore-skew {
		return nil, ErrTokenNotValidYet
	}

	return &claims, nil
}

// VerifyAccessToken 验证访问令牌（确保不是刷新令牌）
func (c *JWTConfig) VerifyAccessToken(tokenString string) (*Claims, error) {
	claims, err := c.VerifyToken(tokenString)
	if err != nil {
		return nil, err
	}

	// 检查是否是刷新令牌
	if claims.CustomClaims != nil {
		if tokenType, exists := claims.CustomClaims["type"]; exists && tokenType == "refresh" {
			return nil, errors.New("不能使用刷新令牌作为访问令牌")
		}
	}

	return claims, nil
}

// VerifyRefreshToken 验证刷新令牌（确保是刷新令牌）
func (c *JWTConfig) VerifyRefreshToken(tokenString string) (*Claims, error) {
	claims, err := c.VerifyToken(tokenString)
	if err != nil {
		return nil, err
	}

	// 检查是否是刷新令牌
	if claims.CustomClaims == nil {
		return nil, errors.New("无效的刷新令牌：缺少类型标识")
	}

	tokenType, exists := claims.CustomClaims["type"]
	if !exists || tokenType != "refresh" {
		return nil, errors.New("无效的刷新令牌：令牌类型不正确")
	}

	return claims, nil
}

// ValidateToken validates a token and checks specific claims
func (c *JWTConfig) ValidateToken(tokenString string, expectedIssuer, expectedAudience string) (*Claims, error) {
	claims, err := c.VerifyToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Validate issuer
	if expectedIssuer != "" && claims.Issuer != expectedIssuer {
		return nil, ErrInvalidIssuer
	}

	// Validate audience
	if expectedAudience != "" && claims.Audience != expectedAudience {
		return nil, ErrInvalidAudience
	}

	return claims, nil
}

// RefreshToken refreshes an existing token by extending its expiry time
func (c *JWTConfig) RefreshToken(tokenString string) (string, error) {
	claims, err := c.VerifyToken(tokenString)
	if err != nil {
		return "", err
	}

	// Update timestamps
	now := time.Now()
	claims.IssuedAt = now.Unix()
	claims.ExpiresAt = now.Add(c.ExpiryDuration).Unix()

	return c.GenerateToken(claims)
}

// ParseTokenWithoutVerify parses a token without verifying the signature
// ⚠️ WARNING: This function is for debugging or extracting public information only.
// NEVER use this for authentication or authorization decisions!
// Unverified token data cannot be trusted and may be tampered with.
func ParseTokenWithoutVerify(tokenString string) (*Claims, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	payloadEncoded := parts[1]
	var claims Claims
	payloadJSON, err := base64.RawURLEncoding.DecodeString(payloadEncoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode payload: %w", err)
	}
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal claims: %w", err)
	}

	return &claims, nil
}

// GetTokenExpiry extracts the expiry time from a token without full verification
func GetTokenExpiry(tokenString string) (time.Time, error) {
	claims, err := ParseTokenWithoutVerify(tokenString)
	if err != nil {
		return time.Time{}, err
	}

	if claims.ExpiresAt == 0 {
		return time.Time{}, errors.New("token has no expiry")
	}

	return time.Unix(claims.ExpiresAt, 0), nil
}

// IsTokenExpired checks if a token is expired without full verification
func IsTokenExpired(tokenString string) (bool, error) {
	expiry, err := GetTokenExpiry(tokenString)
	if err != nil {
		return false, err
	}

	return time.Now().After(expiry), nil
}

// sign creates a signature for the given message
func (c *JWTConfig) sign(message string) (string, error) {
	var hasher func() []byte

	switch c.Algorithm {
	case HS256:
		hasher = func() []byte {
			h := hmac.New(sha256.New, []byte(c.SecretKey))
			h.Write([]byte(message))
			return h.Sum(nil)
		}
	case HS384:
		hasher = func() []byte {
			h := hmac.New(sha512.New384, []byte(c.SecretKey))
			h.Write([]byte(message))
			return h.Sum(nil)
		}
	case HS512:
		hasher = func() []byte {
			h := hmac.New(sha512.New, []byte(c.SecretKey))
			h.Write([]byte(message))
			return h.Sum(nil)
		}
	default:
		return "", fmt.Errorf("unsupported algorithm: %s", c.Algorithm)
	}

	hash := hasher()
	return base64.RawURLEncoding.EncodeToString(hash), nil
}

// Helper functions for common use cases

// GenerateAccessToken generates an access token with subject and optional custom claims
func GenerateAccessToken(secretKey, subject string, expiryDuration time.Duration, customClaims map[string]interface{}) (string, error) {
	config := &JWTConfig{
		SecretKey:      secretKey,
		Algorithm:      HS256,
		ExpiryDuration: expiryDuration,
	}

	return config.GenerateTokenSimple(subject, customClaims)
}

// GenerateRefreshToken generates a refresh token with a longer expiry
func GenerateRefreshToken(secretKey, subject string, expiryDuration time.Duration) (string, error) {
	config := &JWTConfig{
		SecretKey:      secretKey,
		Algorithm:      HS256,
		ExpiryDuration: expiryDuration,
	}

	claims := &Claims{
		StandardClaims: StandardClaims{
			Subject: subject,
		},
		CustomClaims: map[string]interface{}{
			"type": "refresh",
		},
	}

	return config.GenerateToken(claims)
}

// VerifyAccessToken verifies an access token
func VerifyAccessToken(secretKey, tokenString string) (*Claims, error) {
	config := &JWTConfig{
		SecretKey: secretKey,
		Algorithm: HS256,
	}

	return config.VerifyToken(tokenString)
}

// ExtractSubject extracts the subject from a token
func ExtractSubject(tokenString string) (string, error) {
	claims, err := ParseTokenWithoutVerify(tokenString)
	if err != nil {
		return "", err
	}
	return claims.Subject, nil
}

// ExtractCustomClaim extracts a specific custom claim from a token
func ExtractCustomClaim(tokenString, key string) (interface{}, error) {
	claims, err := ParseTokenWithoutVerify(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.CustomClaims == nil {
		return nil, errors.New("no custom claims found")
	}

	value, exists := claims.CustomClaims[key]
	if !exists {
		return nil, fmt.Errorf("custom claim '%s' not found", key)
	}

	return value, nil
}
