// Package encrypt provides field-level encryption using AES-256-GCM.
package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

// Encryptor defines the interface for encrypting and decrypting strings.
type Encryptor interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

// AESEncryptor implements Encryptor using AES-256-GCM.
type AESEncryptor struct {
	gcm cipher.AEAD
}

// NewAESEncryptor creates a new AES-256-GCM encryptor with the given key.
// Key must be exactly 32 bytes for AES-256.
func NewAESEncryptor(key []byte) (*AESEncryptor, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("key must be 32 bytes for AES-256, got %d", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}

	return &AESEncryptor{gcm: gcm}, nil
}

// Encrypt encrypts plaintext and returns base64-encoded ciphertext.
// Returns empty string for empty input.
func (e *AESEncryptor) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	nonce := make([]byte, e.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext := e.gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts base64-encoded ciphertext and returns plaintext.
// Returns empty string for empty input.
func (e *AESEncryptor) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("decode base64: %w", err)
	}

	nonceSize := e.gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := e.gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}

	return string(plaintext), nil
}

// NoOpEncryptor is a no-op encryptor that passes data through unchanged.
// Useful for testing or when encryption is disabled.
type NoOpEncryptor struct{}

// NewNoOpEncryptor creates a new no-op encryptor.
func NewNoOpEncryptor() *NoOpEncryptor {
	return &NoOpEncryptor{}
}

// Encrypt returns the plaintext unchanged.
func (e *NoOpEncryptor) Encrypt(plaintext string) (string, error) {
	return plaintext, nil
}

// Decrypt returns the ciphertext unchanged.
func (e *NoOpEncryptor) Decrypt(ciphertext string) (string, error) {
	return ciphertext, nil
}
