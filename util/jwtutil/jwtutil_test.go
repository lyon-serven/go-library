package jwtutil

import (
	"testing"
	"time"
)

func TestGenerateAndVerifyToken(t *testing.T) {
	config := NewJWTConfig("test-secret-key-123456")

	claims := &Claims{
		StandardClaims: StandardClaims{
			Subject: "user123",
		},
		CustomClaims: map[string]interface{}{
			"username": "john_doe",
			"role":     "admin",
		},
	}

	// Generate token
	token, err := config.GenerateToken(claims)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if token == "" {
		t.Fatal("Token is empty")
	}

	// Verify token
	verifiedClaims, err := config.VerifyToken(token)
	if err != nil {
		t.Fatalf("Failed to verify token: %v", err)
	}

	if verifiedClaims.Subject != "user123" {
		t.Errorf("Expected subject 'user123', got '%s'", verifiedClaims.Subject)
	}

	if verifiedClaims.CustomClaims["username"] != "john_doe" {
		t.Errorf("Expected username 'john_doe', got '%v'", verifiedClaims.CustomClaims["username"])
	}

	if verifiedClaims.CustomClaims["role"] != "admin" {
		t.Errorf("Expected role 'admin', got '%v'", verifiedClaims.CustomClaims["role"])
	}
}

func TestGenerateTokenSimple(t *testing.T) {
	config := NewJWTConfig("test-secret-key-123456")

	token, err := config.GenerateTokenSimple("user123", map[string]interface{}{
		"email": "john@example.com",
	})
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	claims, err := config.VerifyToken(token)
	if err != nil {
		t.Fatalf("Failed to verify token: %v", err)
	}

	if claims.Subject != "user123" {
		t.Errorf("Expected subject 'user123', got '%s'", claims.Subject)
	}

	if claims.CustomClaims["email"] != "john@example.com" {
		t.Errorf("Expected email 'john@example.com', got '%v'", claims.CustomClaims["email"])
	}
}

func TestTokenExpiry(t *testing.T) {
	config := NewJWTConfig("test-secret-key-123456")
	config.ExpiryDuration = 1 * time.Second

	claims := &Claims{
		StandardClaims: StandardClaims{
			Subject: "user123",
		},
	}

	token, err := config.GenerateToken(claims)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Token should be valid immediately
	_, err = config.VerifyToken(token)
	if err != nil {
		t.Errorf("Token should be valid: %v", err)
	}

	// Wait for token to expire
	time.Sleep(2 * time.Second)

	// Token should be expired now
	_, err = config.VerifyToken(token)
	if err != ErrTokenExpired {
		t.Errorf("Expected ErrTokenExpired, got %v", err)
	}
}

func TestInvalidSignature(t *testing.T) {
	config1 := NewJWTConfig("secret-key-1")
	config2 := NewJWTConfig("secret-key-2")

	claims := &Claims{
		StandardClaims: StandardClaims{
			Subject: "user123",
		},
	}

	token, err := config1.GenerateToken(claims)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Try to verify with different secret key
	_, err = config2.VerifyToken(token)
	if err != ErrInvalidSignature {
		t.Errorf("Expected ErrInvalidSignature, got %v", err)
	}
}

func TestValidateToken(t *testing.T) {
	config := NewJWTConfig("test-secret-key-123456")
	config.Issuer = "test-issuer"

	claims := &Claims{
		StandardClaims: StandardClaims{
			Subject:  "user123",
			Audience: "test-audience",
		},
	}

	token, err := config.GenerateToken(claims)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Valid issuer and audience
	_, err = config.ValidateToken(token, "test-issuer", "test-audience")
	if err != nil {
		t.Errorf("Token should be valid: %v", err)
	}

	// Invalid issuer
	_, err = config.ValidateToken(token, "wrong-issuer", "test-audience")
	if err != ErrInvalidIssuer {
		t.Errorf("Expected ErrInvalidIssuer, got %v", err)
	}

	// Invalid audience
	_, err = config.ValidateToken(token, "test-issuer", "wrong-audience")
	if err != ErrInvalidAudience {
		t.Errorf("Expected ErrInvalidAudience, got %v", err)
	}
}

func TestRefreshToken(t *testing.T) {
	config := NewJWTConfig("test-secret-key-123456")

	claims := &Claims{
		StandardClaims: StandardClaims{
			Subject: "user123",
		},
		CustomClaims: map[string]interface{}{
			"username": "john_doe",
		},
	}

	oldToken, err := config.GenerateToken(claims)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	time.Sleep(1 * time.Second)

	// Refresh token
	newToken, err := config.RefreshToken(oldToken)
	if err != nil {
		t.Fatalf("Failed to refresh token: %v", err)
	}

	if newToken == oldToken {
		t.Error("New token should be different from old token")
	}

	// Verify new token
	newClaims, err := config.VerifyToken(newToken)
	if err != nil {
		t.Fatalf("Failed to verify new token: %v", err)
	}

	if newClaims.Subject != "user123" {
		t.Errorf("Expected subject 'user123', got '%s'", newClaims.Subject)
	}

	if newClaims.CustomClaims["username"] != "john_doe" {
		t.Errorf("Expected username 'john_doe', got '%v'", newClaims.CustomClaims["username"])
	}
}

func TestHelperFunctions(t *testing.T) {
	secretKey := "test-secret-key-123456"

	// Test GenerateAccessToken and VerifyAccessToken
	token, err := GenerateAccessToken(secretKey, "user123", 1*time.Hour, map[string]interface{}{
		"role": "admin",
	})
	if err != nil {
		t.Fatalf("Failed to generate access token: %v", err)
	}

	claims, err := VerifyAccessToken(secretKey, token)
	if err != nil {
		t.Fatalf("Failed to verify access token: %v", err)
	}

	if claims.Subject != "user123" {
		t.Errorf("Expected subject 'user123', got '%s'", claims.Subject)
	}

	// Test GenerateRefreshToken
	refreshToken, err := GenerateRefreshToken(secretKey, "user123", 7*24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate refresh token: %v", err)
	}

	refreshClaims, err := VerifyAccessToken(secretKey, refreshToken)
	if err != nil {
		t.Fatalf("Failed to verify refresh token: %v", err)
	}

	if refreshClaims.CustomClaims["type"] != "refresh" {
		t.Error("Expected refresh token type")
	}
}

func TestExtractFunctions(t *testing.T) {
	secretKey := "test-secret-key-123456"

	token, err := GenerateAccessToken(secretKey, "user123", 1*time.Hour, map[string]interface{}{
		"username": "john_doe",
		"role":     "admin",
	})
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Test ExtractSubject
	subject, err := ExtractSubject(token)
	if err != nil {
		t.Fatalf("Failed to extract subject: %v", err)
	}
	if subject != "user123" {
		t.Errorf("Expected subject 'user123', got '%s'", subject)
	}

	// Test ExtractCustomClaim
	username, err := ExtractCustomClaim(token, "username")
	if err != nil {
		t.Fatalf("Failed to extract username: %v", err)
	}
	if username != "john_doe" {
		t.Errorf("Expected username 'john_doe', got '%v'", username)
	}

	role, err := ExtractCustomClaim(token, "role")
	if err != nil {
		t.Fatalf("Failed to extract role: %v", err)
	}
	if role != "admin" {
		t.Errorf("Expected role 'admin', got '%v'", role)
	}

	// Test IsTokenExpired
	expired, err := IsTokenExpired(token)
	if err != nil {
		t.Fatalf("Failed to check if token is expired: %v", err)
	}
	if expired {
		t.Error("Token should not be expired")
	}

	// Test GetTokenExpiry
	expiry, err := GetTokenExpiry(token)
	if err != nil {
		t.Fatalf("Failed to get token expiry: %v", err)
	}
	if expiry.Before(time.Now()) {
		t.Error("Token expiry should be in the future")
	}
}

func TestParseTokenWithoutVerify(t *testing.T) {
	secretKey := "test-secret-key-123456"

	token, err := GenerateAccessToken(secretKey, "user123", 1*time.Hour, map[string]interface{}{
		"username": "john_doe",
	})
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	claims, err := ParseTokenWithoutVerify(token)
	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}

	if claims.Subject != "user123" {
		t.Errorf("Expected subject 'user123', got '%s'", claims.Subject)
	}

	if claims.CustomClaims["username"] != "john_doe" {
		t.Errorf("Expected username 'john_doe', got '%v'", claims.CustomClaims["username"])
	}
}

func TestInvalidToken(t *testing.T) {
	config := NewJWTConfig("test-secret-key-123456")

	tests := []struct {
		name  string
		token string
	}{
		{"Empty token", ""},
		{"Invalid format", "invalid.token"},
		{"Too many parts", "part1.part2.part3.part4"},
		{"Invalid base64", "!!!.!!!.!!!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := config.VerifyToken(tt.token)
			if err == nil {
				t.Error("Expected error for invalid token")
			}
		})
	}
}

func BenchmarkGenerateToken(b *testing.B) {
	config := NewJWTConfig("test-secret-key-123456")
	claims := &Claims{
		StandardClaims: StandardClaims{
			Subject: "user123",
		},
		CustomClaims: map[string]interface{}{
			"username": "john_doe",
			"role":     "admin",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = config.GenerateToken(claims)
	}
}

func BenchmarkVerifyToken(b *testing.B) {
	config := NewJWTConfig("test-secret-key-123456")
	claims := &Claims{
		StandardClaims: StandardClaims{
			Subject: "user123",
		},
	}

	token, _ := config.GenerateToken(claims)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = config.VerifyToken(token)
	}
}
