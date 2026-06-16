package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
)

// RootKeyManager holds the master key loaded from env/K8s secret.
// This key encrypts all customer KEKs — never use it to encrypt data directly.
type RootKeyManager struct {
	masterKey []byte // 32 bytes, AES-256
}

func NewRootKeyManager() (*RootKeyManager, error) {
	raw := os.Getenv("VAULT_MASTER_KEY") // base64-encoded 32-byte key
	if raw == "" {
		return nil, errors.New("VAULT_MASTER_KEY env var is not set")
	}

	key, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid VAULT_MASTER_KEY encoding: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("VAULT_MASTER_KEY must be 32 bytes, got %d", len(key))
	}

	return &RootKeyManager{masterKey: key}, nil
}

// GenerateCustomerKEK creates a new random 32-byte KEK for a customer,
// encrypts it with the master key, and returns the ciphertext for storage in MongoDB.
func (r *RootKeyManager) GenerateCustomerKEK() (plaintextKEK []byte, encryptedKEK string, err error) {
	kek := make([]byte, 32)
	if _, err = io.ReadFull(rand.Reader, kek); err != nil {
		return nil, "", fmt.Errorf("failed to generate KEK: %w", err)
	}

	encrypted, err := r.encryptWithMaster(kek)
	if err != nil {
		return nil, "", fmt.Errorf("failed to encrypt KEK: %w", err)
	}

	return kek, encrypted, nil
}

// DecryptCustomerKEK takes the stored encrypted KEK from MongoDB and returns the plaintext KEK.
func (r *RootKeyManager) DecryptCustomerKEK(encryptedKEK string) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedKEK)
	if err != nil {
		return nil, fmt.Errorf("invalid encrypted KEK encoding: %w", err)
	}
	return r.decryptWithMaster(ciphertext)
}

// encryptWithMaster encrypts data using AES-256-GCM with the master key.
// Output format: [12-byte IV][ciphertext+tag] — all concatenated, then base64 encoded.
func (r *RootKeyManager) encryptWithMaster(plaintext []byte) (string, error) {
	block, err := aes.NewCipher(r.masterKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize()) // 12 bytes
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Seal appends ciphertext+tag to nonce
	sealed := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

func (r *RootKeyManager) decryptWithMaster(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(r.masterKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
