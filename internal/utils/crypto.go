package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"os"
)
var encryptionKey []byte


// InitEncryption initializes encryption with a key from environment or generates one
func InitEncryption() error {
	// Try to get key from environment
	keyString := os.Getenv("LOG_ENCRYPTION_KEY")

	// If not in environment, try to read from key file
	if keyString == "" {
		keyFilePath := "/Users/saneechka/Finance/keys/log_encryption.key"
		keyData, err := os.ReadFile(keyFilePath)
		if err == nil {
			keyString = string(keyData)
		}
	}

	// If key was found, decode it
	if keyString != "" {
		var err error
		encryptionKey, err = base64.StdEncoding.DecodeString(keyString)
		if err != nil || len(encryptionKey) != 32 {
			return errors.New("invalid encryption key format")
		}
		return nil
	}

	// Generate a new key (32 bytes for AES-256)
	encryptionKey = make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, encryptionKey); err != nil {
		return err
	}

	// Save the key to file for future use
	keyDir := "/Users/saneechka/Finance/keys"
	if err := os.MkdirAll(keyDir, 0700); err != nil {
		return err
	}

	encodedKey := base64.StdEncoding.EncodeToString(encryptionKey)
	return os.WriteFile(keyDir+"/log_encryption.key", []byte(encodedKey), 0600)
}

// EncryptLogMessage encrypts a log message
func EncryptLogMessage(message interface{}) (string, error) {
	// Convert message to JSON
	jsonData, err := json.Marshal(message)
	if err != nil {
		return "", err
	}


	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Create a nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt the data
	ciphertext := gcm.Seal(nonce, nonce, jsonData, nil)

	// Encode to base64
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptLogMessage decrypts an encrypted log message
func DecryptLogMessage(encryptedMessage string) (string, error) {
	// Decode from base64
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedMessage)
	if err != nil {
		return "", err
	}

	// Create a new AES cipher
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Get nonce size
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	// Extract nonce
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
