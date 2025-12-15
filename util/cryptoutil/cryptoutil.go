// Package cryptoutil provides utility functions for encryption, decryption, and hashing.
package cryptoutil

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"hash"
	"io"
)

// Hash algorithms
const (
	MD5    = "md5"
	SHA1   = "sha1"
	SHA256 = "sha256"
	SHA512 = "sha512"
)

// Encoding types
const (
	HexEncoding    = "hex"
	Base64Encoding = "base64"
)

// MD5Hash returns MD5 hash of the input data
func MD5Hash(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

// MD5String returns MD5 hash of the input string
func MD5String(s string) string {
	return MD5Hash([]byte(s))
}

// SHA1Hash returns SHA1 hash of the input data
func SHA1Hash(data []byte) string {
	hash := sha1.Sum(data)
	return hex.EncodeToString(hash[:])
}

// SHA1String returns SHA1 hash of the input string
func SHA1String(s string) string {
	return SHA1Hash([]byte(s))
}

// SHA256Hash returns SHA256 hash of the input data
func SHA256Hash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// SHA256String returns SHA256 hash of the input string
func SHA256String(s string) string {
	return SHA256Hash([]byte(s))
}

// SHA512Hash returns SHA512 hash of the input data
func SHA512Hash(data []byte) string {
	hash := sha512.Sum512(data)
	return hex.EncodeToString(hash[:])
}

// SHA512String returns SHA512 hash of the input string
func SHA512String(s string) string {
	return SHA512Hash([]byte(s))
}

// Hash returns hash of the input data using the specified algorithm
func Hash(data []byte, algorithm string) (string, error) {
	var h hash.Hash

	switch algorithm {
	case MD5:
		h = md5.New()
	case SHA1:
		h = sha1.New()
	case SHA256:
		h = sha256.New()
	case SHA512:
		h = sha512.New()
	default:
		return "", fmt.Errorf("unsupported hash algorithm: %s", algorithm)
	}

	h.Write(data)
	return hex.EncodeToString(h.Sum(nil)), nil
}

// HashString returns hash of the input string using the specified algorithm
func HashString(s, algorithm string) (string, error) {
	return Hash([]byte(s), algorithm)
}

// HMAC generates HMAC using the specified hash algorithm
func HMAC(data, key []byte, algorithm string) (string, error) {
	var h func() hash.Hash

	switch algorithm {
	case MD5:
		h = md5.New
	case SHA1:
		h = sha1.New
	case SHA256:
		h = sha256.New
	case SHA512:
		h = sha512.New
	default:
		return "", fmt.Errorf("unsupported hash algorithm: %s", algorithm)
	}

	mac := hmac.New(h, key)
	mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil)), nil
}

// HMACString generates HMAC for string data
func HMACString(data, key, algorithm string) (string, error) {
	return HMAC([]byte(data), []byte(key), algorithm)
}

// VerifyHMAC verifies HMAC signature
func VerifyHMAC(data, key []byte, signature, algorithm string) (bool, error) {
	expectedSignature, err := HMAC(data, key, algorithm)
	if err != nil {
		return false, err
	}
	return hmac.Equal([]byte(signature), []byte(expectedSignature)), nil
}

// VerifyHMACString verifies HMAC signature for string data
func VerifyHMACString(data, key, signature, algorithm string) (bool, error) {
	return VerifyHMAC([]byte(data), []byte(key), signature, algorithm)
}

// Base64Encode encodes data to base64 string
func Base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// Base64Decode decodes base64 string to data
func Base64Decode(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

// Base64URLEncode encodes data to URL-safe base64 string
func Base64URLEncode(data []byte) string {
	return base64.URLEncoding.EncodeToString(data)
}

// Base64URLDecode decodes URL-safe base64 string to data
func Base64URLDecode(s string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(s)
}

// HexEncode encodes data to hex string
func HexEncode(data []byte) string {
	return hex.EncodeToString(data)
}

// HexDecode decodes hex string to data
func HexDecode(s string) ([]byte, error) {
	return hex.DecodeString(s)
}

// GenerateRandomBytes generates random bytes of specified length
func GenerateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// GenerateRandomString generates random string of specified length
func GenerateRandomString(length int, encoding string) (string, error) {
	bytes, err := GenerateRandomBytes(length)
	if err != nil {
		return "", err
	}

	switch encoding {
	case HexEncoding:
		return HexEncode(bytes), nil
	case Base64Encoding:
		return Base64Encode(bytes), nil
	default:
		return "", fmt.Errorf("unsupported encoding: %s", encoding)
	}
}

// AESEncrypt encrypts data using AES with the given key
func AESEncrypt(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Generate a random IV
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	// Encrypt using CBC mode
	mode := cipher.NewCBCEncrypter(block, iv)

	// Pad the data to be multiple of block size
	paddedData := PKCS7Pad(data, aes.BlockSize)

	encrypted := make([]byte, len(paddedData))
	mode.CryptBlocks(encrypted, paddedData)

	// Prepend IV to encrypted data
	result := make([]byte, len(iv)+len(encrypted))
	copy(result[:len(iv)], iv)
	copy(result[len(iv):], encrypted)

	return result, nil
}

// AESDecrypt decrypts data using AES with the given key
func AESDecrypt(encryptedData, key []byte) ([]byte, error) {
	if len(encryptedData) < aes.BlockSize {
		return nil, errors.New("encrypted data too short")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Extract IV from the beginning
	iv := encryptedData[:aes.BlockSize]
	encrypted := encryptedData[aes.BlockSize:]

	if len(encrypted)%aes.BlockSize != 0 {
		return nil, errors.New("encrypted data is not a multiple of block size")
	}

	// Decrypt using CBC mode
	mode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(encrypted))
	mode.CryptBlocks(decrypted, encrypted)

	// Remove padding
	return PKCS7Unpad(decrypted)
}

// AESEncryptString encrypts string using AES
func AESEncryptString(data, key string) (string, error) {
	encrypted, err := AESEncrypt([]byte(data), []byte(key))
	if err != nil {
		return "", err
	}
	return Base64Encode(encrypted), nil
}

// AESDecryptString decrypts base64 encoded string using AES
func AESDecryptString(encryptedData, key string) (string, error) {
	encrypted, err := Base64Decode(encryptedData)
	if err != nil {
		return "", err
	}

	decrypted, err := AESDecrypt(encrypted, []byte(key))
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}

// PKCS7Pad applies PKCS7 padding to data
func PKCS7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := make([]byte, padding)
	for i := 0; i < padding; i++ {
		padtext[i] = byte(padding)
	}
	return append(data, padtext...)
}

// PKCS7Unpad removes PKCS7 padding from data
func PKCS7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("invalid padding")
	}

	padding := int(data[len(data)-1])
	if padding > len(data) {
		return nil, errors.New("invalid padding")
	}

	return data[:len(data)-padding], nil
}

// DESEncrypt encrypts data using DES with the given key (8 bytes)
func DESEncrypt(data, key []byte) ([]byte, error) {
	if len(key) != 8 {
		return nil, errors.New("DES key must be 8 bytes")
	}

	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Generate a random IV
	iv := make([]byte, des.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	// Encrypt using CBC mode
	mode := cipher.NewCBCEncrypter(block, iv)

	// Pad the data to be multiple of block size
	paddedData := PKCS7Pad(data, des.BlockSize)

	encrypted := make([]byte, len(paddedData))
	mode.CryptBlocks(encrypted, paddedData)

	// Prepend IV to encrypted data
	result := make([]byte, len(iv)+len(encrypted))
	copy(result[:len(iv)], iv)
	copy(result[len(iv):], encrypted)

	return result, nil
}

// DESDecrypt decrypts data using DES with the given key (8 bytes)
func DESDecrypt(encryptedData, key []byte) ([]byte, error) {
	if len(key) != 8 {
		return nil, errors.New("DES key must be 8 bytes")
	}

	if len(encryptedData) < des.BlockSize {
		return nil, errors.New("encrypted data too short")
	}

	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Extract IV from the beginning
	iv := encryptedData[:des.BlockSize]
	encrypted := encryptedData[des.BlockSize:]

	if len(encrypted)%des.BlockSize != 0 {
		return nil, errors.New("encrypted data is not a multiple of block size")
	}

	// Decrypt using CBC mode
	mode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(encrypted))
	mode.CryptBlocks(decrypted, encrypted)

	// Remove padding
	return PKCS7Unpad(decrypted)
}

// DESEncryptString encrypts string using DES
func DESEncryptString(data, key string) (string, error) {
	if len(key) != 8 {
		return "", errors.New("DES key must be 8 characters")
	}

	encrypted, err := DESEncrypt([]byte(data), []byte(key))
	if err != nil {
		return "", err
	}
	return Base64Encode(encrypted), nil
}

// DESDecryptString decrypts base64 encoded string using DES
func DESDecryptString(encryptedData, key string) (string, error) {
	if len(key) != 8 {
		return "", errors.New("DES key must be 8 characters")
	}

	encrypted, err := Base64Decode(encryptedData)
	if err != nil {
		return "", err
	}

	decrypted, err := DESDecrypt(encrypted, []byte(key))
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}

// DESEncryptWithFixedIV encrypts data using DES with fixed IV (key as IV) - compatible mode
func DESEncryptWithFixedIV(data, key []byte) ([]byte, error) {
	if len(key) != 8 {
		return nil, errors.New("DES key must be 8 bytes")
	}

	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Use key as IV (fixed IV mode)
	iv := key

	// Encrypt using CBC mode
	mode := cipher.NewCBCEncrypter(block, iv)

	// Pad the data to be multiple of block size (PKCS5 is same as PKCS7 for 8-byte blocks)
	paddedData := PKCS7Pad(data, des.BlockSize)

	encrypted := make([]byte, len(paddedData))
	mode.CryptBlocks(encrypted, paddedData)

	return encrypted, nil
}

// DESDecryptWithFixedIV decrypts data using DES with fixed IV (key as IV) - compatible mode
func DESDecryptWithFixedIV(encryptedData, key []byte) ([]byte, error) {
	if len(key) != 8 {
		return nil, errors.New("DES key must be 8 bytes")
	}

	if len(encryptedData)%des.BlockSize != 0 {
		return nil, errors.New("encrypted data is not a multiple of block size")
	}

	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Use key as IV (fixed IV mode)
	iv := key

	// Decrypt using CBC mode
	mode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(encryptedData))
	mode.CryptBlocks(decrypted, encryptedData)

	// Remove padding
	return PKCS7Unpad(decrypted)
}

// DESEncryptHex encrypts string using DES with fixed IV and returns hex string - compatible with your code
func DESEncryptHex(data, key string) (string, error) {
	if len(data) == 0 {
		return "", nil
	}

	if len(key) != 8 {
		return "", errors.New("DES key must be 8 characters")
	}

	encrypted, err := DESEncryptWithFixedIV([]byte(data), []byte(key))
	if err != nil {
		return "", err
	}

	return HexEncode(encrypted), nil
}

// DESDecryptHex decrypts hex encoded string using DES with fixed IV - compatible with your code
func DESDecryptHex(encryptedHex, key string) (string, error) {
	if len(encryptedHex) == 0 {
		return "", nil
	}

	if len(key) != 8 {
		return "", errors.New("DES key must be 8 characters")
	}

	// Decode hex string to bytes
	encrypted, err := HexDecode(encryptedHex)
	if err != nil {
		return "", err
	}

	decrypted, err := DESDecryptWithFixedIV(encrypted, []byte(key))
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}

// TripleDESEncrypt encrypts data using 3DES with the given key (24 bytes)
func TripleDESEncrypt(data, key []byte) ([]byte, error) {
	if len(key) != 24 {
		return nil, errors.New("3DES key must be 24 bytes")
	}

	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return nil, err
	}

	// Generate a random IV
	iv := make([]byte, des.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	// Encrypt using CBC mode
	mode := cipher.NewCBCEncrypter(block, iv)

	// Pad the data to be multiple of block size
	paddedData := PKCS7Pad(data, des.BlockSize)

	encrypted := make([]byte, len(paddedData))
	mode.CryptBlocks(encrypted, paddedData)

	// Prepend IV to encrypted data
	result := make([]byte, len(iv)+len(encrypted))
	copy(result[:len(iv)], iv)
	copy(result[len(iv):], encrypted)

	return result, nil
}

// TripleDESDecrypt decrypts data using 3DES with the given key (24 bytes)
func TripleDESDecrypt(encryptedData, key []byte) ([]byte, error) {
	if len(key) != 24 {
		return nil, errors.New("3DES key must be 24 bytes")
	}

	if len(encryptedData) < des.BlockSize {
		return nil, errors.New("encrypted data too short")
	}

	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return nil, err
	}

	// Extract IV from the beginning
	iv := encryptedData[:des.BlockSize]
	encrypted := encryptedData[des.BlockSize:]

	if len(encrypted)%des.BlockSize != 0 {
		return nil, errors.New("encrypted data is not a multiple of block size")
	}

	// Decrypt using CBC mode
	mode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(encrypted))
	mode.CryptBlocks(decrypted, encrypted)

	// Remove padding
	return PKCS7Unpad(decrypted)
}

// TripleDESEncryptString encrypts string using 3DES
func TripleDESEncryptString(data, key string) (string, error) {
	if len(key) != 24 {
		return "", errors.New("3DES key must be 24 characters")
	}

	encrypted, err := TripleDESEncrypt([]byte(data), []byte(key))
	if err != nil {
		return "", err
	}
	return Base64Encode(encrypted), nil
}

// TripleDESDecryptString decrypts base64 encoded string using 3DES
func TripleDESDecryptString(encryptedData, key string) (string, error) {
	if len(key) != 24 {
		return "", errors.New("3DES key must be 24 characters")
	}

	encrypted, err := Base64Decode(encryptedData)
	if err != nil {
		return "", err
	}

	decrypted, err := TripleDESDecrypt(encrypted, []byte(key))
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}

// RSAKeyPair represents an RSA key pair
type RSAKeyPair struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
}

// GenerateRSAKeyPair generates a new RSA key pair
func GenerateRSAKeyPair(bits int) (*RSAKeyPair, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, err
	}

	return &RSAKeyPair{
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
	}, nil
}

// PrivateKeyToPEM converts private key to PEM format
func (kp *RSAKeyPair) PrivateKeyToPEM() ([]byte, error) {
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(kp.PrivateKey)
	if err != nil {
		return nil, err
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	return privateKeyPEM, nil
}

// PublicKeyToPEM converts public key to PEM format
func (kp *RSAKeyPair) PublicKeyToPEM() ([]byte, error) {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(kp.PublicKey)
	if err != nil {
		return nil, err
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return publicKeyPEM, nil
}

// RSAEncrypt encrypts data using RSA public key
func RSAEncrypt(data []byte, publicKey *rsa.PublicKey) ([]byte, error) {
	return rsa.EncryptPKCS1v15(rand.Reader, publicKey, data)
}

// RSADecrypt decrypts data using RSA private key
func RSADecrypt(encryptedData []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	return rsa.DecryptPKCS1v15(rand.Reader, privateKey, encryptedData)
}

// RSASign signs data using RSA private key with SHA256
func RSASign(data []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	hash := sha256.Sum256(data)
	return rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash[:])
}

// RSAVerify verifies signature using RSA public key with SHA256
func RSAVerify(data, signature []byte, publicKey *rsa.PublicKey) error {
	hash := sha256.Sum256(data)
	return rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hash[:], signature)
}

// ParsePrivateKeyFromPEM parses private key from PEM format
func ParsePrivateKeyFromPEM(pemData []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("not an RSA private key")
	}

	return rsaPrivateKey, nil
}

// ParsePublicKeyFromPEM parses public key from PEM format
func ParsePublicKeyFromPEM(pemData []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}

	return rsaPublicKey, nil
}

// SimpleEncrypt provides a simple encryption interface using AES-256
func SimpleEncrypt(data, password string) (string, error) {
	// Generate a 32-byte key from password using SHA256
	key := SHA256Hash([]byte(password))
	keyBytes, _ := HexDecode(key)

	encrypted, err := AESEncrypt([]byte(data), keyBytes)
	if err != nil {
		return "", err
	}

	return Base64Encode(encrypted), nil
}

// SimpleDecrypt provides a simple decryption interface using AES-256
func SimpleDecrypt(encryptedData, password string) (string, error) {
	// Generate a 32-byte key from password using SHA256
	key := SHA256Hash([]byte(password))
	keyBytes, _ := HexDecode(key)

	encrypted, err := Base64Decode(encryptedData)
	if err != nil {
		return "", err
	}

	decrypted, err := AESDecrypt(encrypted, keyBytes)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}
