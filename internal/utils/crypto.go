package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// EncryptedData represents CryptoJS encrypted data structure
type EncryptedData struct {
	CT string `json:"ct"` // Ciphertext (base64)
	IV string `json:"iv"` // Initialization Vector (hex)
	S  string `json:"s"`  // Salt (hex)
}

// CryptoJSDecrypt decrypts CryptoJS AES encrypted data
// This is compatible with CryptoJS format used by IDLIX
func CryptoJSDecrypt(encryptedJSON, passphrase string) (string, error) {
	// Parse encrypted JSON
	var encData EncryptedData
	if err := json.Unmarshal([]byte(encryptedJSON), &encData); err != nil {
		return "", fmt.Errorf("failed to parse encrypted data: %w", err)
	}

	// Decode salt from hex
	salt, err := hex.DecodeString(encData.S)
	if err != nil {
		return "", fmt.Errorf("failed to decode salt: %w", err)
	}

	// Decode IV from hex
	iv, err := hex.DecodeString(encData.IV)
	if err != nil {
		return "", fmt.Errorf("failed to decode IV: %w", err)
	}

	// Validate IV length (must be 16 bytes for AES)
	if len(iv) != aes.BlockSize {
		return "", fmt.Errorf("invalid IV length: got %d, want %d", len(iv), aes.BlockSize)
	}

	// Decode ciphertext from base64
	ciphertext, err := base64.StdEncoding.DecodeString(encData.CT)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	// Validate ciphertext length (must be multiple of block size)
	if len(ciphertext)%aes.BlockSize != 0 {
		return "", fmt.Errorf("invalid ciphertext length: not a multiple of block size")
	}

	// Derive key using MD5 (EVP_BytesToKey algorithm)
	key := deriveKeyMD5([]byte(passphrase), salt, 48)

	// Key is first 32 bytes
	aesKey := key[:32]

	// Create AES cipher
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Decrypt using CBC mode
	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// Remove PKCS7 padding
	plaintext, err = unpadPKCS7(plaintext)
	if err != nil {
		return "", fmt.Errorf("failed to unpad: %w", err)
	}

	return string(plaintext), nil
}

// deriveKeyMD5 derives key using MD5 (EVP_BytesToKey compatible)
// This matches CryptoJS's key derivation
func deriveKeyMD5(passphrase, salt []byte, keyLen int) []byte {
	var result []byte
	var block []byte

	// Concatenate passphrase and salt
	data := append(passphrase, salt...)

	for len(result) < keyLen {
		// First iteration: MD5(passphrase + salt)
		// Next iterations: MD5(previous_hash + passphrase + salt)
		hash := md5.New()
		hash.Write(block)
		hash.Write(data)
		block = hash.Sum(nil)
		result = append(result, block...)
	}

	return result[:keyLen]
}

// unpadPKCS7 removes PKCS7 padding
func unpadPKCS7(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("invalid padding: empty data")
	}

	padding := int(data[len(data)-1])
	
	if padding > len(data) || padding > aes.BlockSize {
		return nil, fmt.Errorf("invalid padding size: %d", padding)
	}

	// Verify padding
	for i := len(data) - padding; i < len(data); i++ {
		if data[i] != byte(padding) {
			return nil, fmt.Errorf("invalid padding")
		}
	}

	return data[:len(data)-padding], nil
}

// Dec generates passphrase from key and m parameter
// This is the custom decryption function used by IDLIX
func Dec(key, m string) (string, error) {
	// Split key into pairs, skip first 2 characters
	if len(key) < 4 {
		return "", fmt.Errorf("key too short")
	}

	var keyList []string
	for i := 2; i < len(key); i += 4 {
		end := i + 2
		if end > len(key) {
			break
		}
		keyList = append(keyList, key[i:end])
	}

	// Reverse m string
	reversed := reverseString(m)

	// Add base64 padding
	padded := addBase64Padding(reversed)

	// Decode base64
	decoded, err := base64.StdEncoding.DecodeString(padded)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// Split by "|"
	parts := strings.Split(string(decoded), "|")

	// Build result
	var result strings.Builder
	for _, part := range parts {
		// Check if part is a digit
		if part == "" {
			continue
		}
		
		index, err := strconv.Atoi(part)
		if err != nil {
			continue
		}

		// Check bounds
		if index >= 0 && index < len(keyList) {
			result.WriteString("\\x")
			result.WriteString(keyList[index])
		}
	}

	return result.String(), nil
}

// reverseString reverses a string
func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// addBase64Padding adds padding to base64 string
func addBase64Padding(s string) string {
	remainder := len(s) % 4
	if remainder == 0 {
		return s
	}
	return s + strings.Repeat("=", 4-remainder)
}
