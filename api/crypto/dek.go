package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

// EncryptSecret encrypts a plaintext secret value using a freshly generated DEK,
// then wraps the DEK with the customer's KEK.
// Returns an EncryptedSecret ready to be stored in MongoDB.
func EncryptSecret(plaintextValue string, kek []byte) (*EncryptedSecret, error) {
	// 1. Generate a random DEK (one per secret)
	dek := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, dek); err != nil {
		return nil, fmt.Errorf("failed to generate DEK: %w", err)
	}
	defer zeroize(dek) // wipe DEK from memory after we're done

	// 2. Encrypt the secret value with the DEK
	encryptedValue, iv, authTag, err := aesGCMEncrypt(dek, []byte(plaintextValue))
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt secret value: %w", err)
	}

	// 3. Encrypt the DEK with the customer KEK
	encryptedDEK, _, _, err := aesGCMEncrypt(kek, dek)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt DEK: %w", err)
	}

	return &EncryptedSecret{
		EncryptedValue: encryptedValue,
		EncryptedDEK:   encryptedDEK,
		IV:             iv,
		AuthTag:        authTag,
		Algorithm:      "AES-256-GCM",
	}, nil
}

// DecryptSecret reverses the process: unwraps the DEK using the KEK, then
// decrypts the secret value.
func DecryptSecret(es *EncryptedSecret, kek []byte) (string, error) {
	// 1. Decrypt the DEK using the KEK
	dek, err := aesGCMDecrypt(kek, es.EncryptedDEK, "", "")
	if err != nil {
		return "", fmt.Errorf("failed to decrypt DEK: %w", err)
	}
	defer zeroize(dek)

	// 2. Decrypt the secret value using the DEK
	plaintext, err := aesGCMDecrypt(dek, es.EncryptedValue, es.IV, es.AuthTag)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt secret value: %w", err)
	}

	return string(plaintext), nil
}

// aesGCMEncrypt encrypts plaintext with AES-256-GCM.
// Returns: base64(ciphertext), base64(iv), base64(authTag), error.
// The auth tag is embedded in Go's GCM Seal output (last 16 bytes),
// so we store them separately for clarity.
func aesGCMEncrypt(key, plaintext []byte) (ciphertextB64, ivB64, authTagB64 string, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", "", "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", "", err
	}

	iv := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return "", "", "", err
	}

	// gcm.Seal output: ciphertext || authTag (16 bytes at end)
	sealed := gcm.Seal(nil, iv, plaintext, nil)
	tagSize := 16
	ciphertext := sealed[:len(sealed)-tagSize]
	authTag := sealed[len(sealed)-tagSize:]

	return base64.StdEncoding.EncodeToString(ciphertext),
		base64.StdEncoding.EncodeToString(iv),
		base64.StdEncoding.EncodeToString(authTag),
		nil
}

func aesGCMDecrypt(key []byte, ciphertextB64, ivB64, authTagB64 string) ([]byte, error) {
	// When decrypting DEK, IV and authTag are embedded in EncryptedDEK field
	// (same format as master key encryption). Handle both cases:
	if ivB64 == "" {
		// DEK case: ciphertextB64 contains iv+ciphertext+tag concatenated
		data, err := base64.StdEncoding.DecodeString(ciphertextB64)
		if err != nil {
			return nil, err
		}

		block, err := aes.NewCipher(key)
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
		return gcm.Open(nil, data[:nonceSize], data[nonceSize:], nil)
	}

	// Secret value case: IV and authTag stored separately
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return nil, err
	}
	iv, err := base64.StdEncoding.DecodeString(ivB64)
	if err != nil {
		return nil, err
	}
	authTag, err := base64.StdEncoding.DecodeString(authTagB64)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Reassemble ciphertext+tag for gcm.Open
	combined := append(ciphertext, authTag...)
	return gcm.Open(nil, iv, combined, nil)
}

// zeroize wipes a byte slice from memory after use
func zeroize(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
