// Package jwtutil provides utilities for JWT (JSON Web Token) generation, validation, and management.
package jwtutil

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Common errors
var (
	ErrInvalidToken     = errors.New("invalid token format")
	ErrTokenExpired     = errors.New("token has expired")
	ErrTokenNotValidYet = errors.New("token not valid yet")
	ErrInvalidSignature = errors.New("invalid token signature")
	ErrInvalidIssuer    = errors.New("invalid token issuer")
	ErrInvalidAudience  = errors.New("invalid token audience")
	ErrMissingClaims    = errors.New("missing required claims")
)

// Algorithm types
const (
	HS256 = "HS256"
	HS384 = "HS384"
	HS512 = "HS512"
)

// StandardClaims represents the standard JWT claims
type StandardClaims struct {
	Issuer    string `json:"iss,omitempty"` // Issuer
	Subject   string `json:"sub,omitempty"` // Subject
	Audience  string `json:"aud,omitempty"` // Audience
	ExpiresAt int64  `json:"exp,omitempty"` // Expiration time (Unix timestamp)
	NotBefore int64  `json:"nbf,omitempty"` // Not before (Unix timestamp)
	IssuedAt  int64  `json:"iat,omitempty"` // Issued at (Unix timestamp)
	ID        string `json:"jti,omitempty"` // JWT ID
}

// Claims represents JWT claims with custom data
type Claims struct {
	StandardClaims
	CustomClaims map[string]interface{} `json:"custom,omitempty"`
}

// Header represents the JWT header
type Header struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
}

// JWTConfig holds configuration for JWT operations
type JWTConfig struct {
	SecretKey      string        // Secret key for signing
	Algorithm      string        // Algorithm to use (HS256, HS384, HS512)
	Issuer         string        // Default issuer
	Audience       string        // Default audience
	ExpiryDuration time.Duration // Default expiry duration
}

// NewJWTConfig creates a new JWT configuration with default values
func NewJWTConfig(secretKey string) *JWTConfig {
	return &JWTConfig{
		SecretKey:      secretKey,
		Algorithm:      HS256,
		ExpiryDuration: 24 * time.Hour,
	}
}

// GenerateToken generates a JWT token with the given claims
func (c *JWTConfig) GenerateToken(claims *Claims) (string, error) {
	if c.SecretKey == "" {
		return "", errors.New("secret key is required")
	}

	// Set default values
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

	// Create header
	header := Header{
		Algorithm: c.Algorithm,
		Type:      "JWT",
	}

	// Encode header
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("failed to marshal header: %w", err)
	}
	headerEncoded := base64.RawURLEncoding.EncodeToString(headerJSON)

	// Encode payload
	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("failed to marshal claims: %w", err)
	}
	payloadEncoded := base64.RawURLEncoding.EncodeToString(payloadJSON)

	// Create signature
	message := headerEncoded + "." + payloadEncoded
	signature, err := c.sign(message)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	// Combine all parts
	token := message + "." + signature
	return token, nil
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

// VerifyToken verifies and parses a JWT token
func (c *JWTConfig) VerifyToken(tokenString string) (*Claims, error) {
	if c.SecretKey == "" {
		return nil, errors.New("secret key is required")
	}

	// Split token into parts
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	headerEncoded := parts[0]
	payloadEncoded := parts[1]
	signatureEncoded := parts[2]

	// Verify signature
	message := headerEncoded + "." + payloadEncoded
	expectedSignature, err := c.sign(message)
	if err != nil {
		return nil, fmt.Errorf("failed to generate signature: %w", err)
	}

	if !hmac.Equal([]byte(signatureEncoded), []byte(expectedSignature)) {
		return nil, ErrInvalidSignature
	}

	// Decode header
	var header Header
	headerJSON, err := base64.RawURLEncoding.DecodeString(headerEncoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode header: %w", err)
	}
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return nil, fmt.Errorf("failed to unmarshal header: %w", err)
	}

	// Decode payload
	var claims Claims
	payloadJSON, err := base64.RawURLEncoding.DecodeString(payloadEncoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode payload: %w", err)
	}
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal claims: %w", err)
	}

	// Validate claims
	now := time.Now().Unix()
	if claims.ExpiresAt > 0 && now > claims.ExpiresAt {
		return nil, ErrTokenExpired
	}
	if claims.NotBefore > 0 && now < claims.NotBefore {
		return nil, ErrTokenNotValidYet
	}

	return &claims, nil
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

// ParseTokenWithoutVerify parses a token without verifying the signature (use with caution)
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
			h := hmac.New(sha256.New224, []byte(c.SecretKey))
			h.Write([]byte(message))
			return h.Sum(nil)
		}
	case HS512:
		hasher = func() []byte {
			h := hmac.New(sha256.New, []byte(c.SecretKey))
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
