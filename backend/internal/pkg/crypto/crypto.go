package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

// Ключ AES-256.
func deriveKey(key string) []byte {
	sum := sha256.Sum256([]byte(key))
	return sum[:]
}

// AES-GCM encrypt.
func Encrypt(plaintext, key string) (string, error) {
	block, err := aes.NewCipher(deriveKey(key))
	if err != nil {
		return "", fmt.Errorf("crypto: ініціалізація шифру: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("crypto: ініціалізація GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("crypto: генерація nonce: %w", err)
	}

	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// AES-GCM decrypt.
func Decrypt(encoded, key string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("crypto: декодування base64: %w", err)
	}

	block, err := aes.NewCipher(deriveKey(key))
	if err != nil {
		return "", fmt.Errorf("crypto: ініціалізація шифру: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("crypto: ініціалізація GCM: %w", err)
	}

	if len(data) < gcm.NonceSize() {
		return "", errors.New("crypto: пошкоджені дані")
	}
	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("crypto: розшифрування: %w", err)
	}
	return string(plaintext), nil
}
