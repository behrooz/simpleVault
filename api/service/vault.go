package service

import (
	"context"
	"fmt"

	"simple-vault/api/crypto"
	"simple-vault/api/store"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type VaultService struct {
	rootKey    *crypto.RootKeyManager
	store      *store.SecretStore
	kekVersion int // current active KEK version
}

func NewVaultService(rootKey *crypto.RootKeyManager, store *store.SecretStore) *VaultService {
	return &VaultService{
		rootKey:    rootKey,
		store:      store,
		kekVersion: 1,
	}
}

// ProvisionCustomer generates and stores a KEK for a new customer.
// Call this when a customer account is created.
func (v *VaultService) ProvisionCustomer(ctx context.Context, customerID string) error {
	_, encryptedKEK, err := v.rootKey.GenerateCustomerKEK()
	if err != nil {
		return fmt.Errorf("failed to generate KEK for customer %s: %w", customerID, err)
	}

	return v.store.SaveCustomerKEK(ctx, customerID, encryptedKEK, v.kekVersion)
}

// SaveSecret encrypts and stores a customer's secret.
func (v *VaultService) SaveSecret(ctx context.Context, customerID, name, description string, data map[string]string) error {
	// 1. Fetch and decrypt the customer's KEK
	kek, err := v.getCustomerKEK(ctx, customerID, v.kekVersion)
	if err != nil {
		return err
	}
	defer crypto.Zeroize(kek) // wipe from memory after use

	// 2. Encrypt the secret value
	encryptedFields := make(map[string]*crypto.EncryptedSecret, len(data))
	for fieldKey, filedvalue := range data {
		encrypted, err := crypto.EncryptSecret(filedvalue, kek)
		if err != nil {
			return fmt.Errorf("encryption failed: %w", fieldKey, err)
		}
		encryptedFields[fieldKey] = encrypted
	}

	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	// 3. Persist to MongoDB — only ciphertext is stored
	return v.store.SaveSecret(ctx, customerID, name, description, encryptedFields, v.kekVersion)
}

// GetSecret decrypts and returns all fields of a secret group
func (v *VaultService) GetSecret(ctx context.Context, customerID, name string) (map[string]string, error) {
	doc, err := v.store.GetSecret(ctx, customerID, name)
	if err != nil {
		return nil, err
	}

	kek, err := v.getCustomerKEK(ctx, customerID, doc.KEKVersion)
	if err != nil {
		return nil, err
	}
	defer crypto.Zeroize(kek)

	// Decrypt each field
	result := make(map[string]string, len(doc.EncryptedFields))
	for fieldKey, encryptedField := range doc.EncryptedFields {
		plaintext, err := crypto.DecryptSecret(encryptedField, kek)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt field %q: %w", fieldKey, err)
		}
		result[fieldKey] = plaintext
	}

	return result, nil
}

func (v *VaultService) getCustomerKEK(ctx context.Context, customerID string, version int) ([]byte, error) {
	kekDoc, err := v.store.GetCustomerKEK(ctx, customerID, version)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch KEK: %w", err)
	}

	kek, err := v.rootKey.DecryptCustomerKEK(kekDoc.EncryptedKEK)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt KEK: %w", err)
	}

	return kek, nil
}

func (v *VaultService) UpdateSecret(ctx context.Context, customerID string, id primitive.ObjectID, name, description string, data map[string]string) error {
	// Fetch and decrypt KEK once
	kek, err := v.getCustomerKEK(ctx, customerID, v.kekVersion)
	if err != nil {
		return err
	}
	defer crypto.Zeroize(kek)

	// Re-encrypt every field fresh — new DEK per field
	encryptedFields := make(map[string]*crypto.EncryptedSecret, len(data))
	for fieldKey, fieldValue := range data {
		encrypted, err := crypto.EncryptSecret(fieldValue, kek)
		if err != nil {
			return fmt.Errorf("failed to encrypt field %q: %w", fieldKey, err)
		}
		encryptedFields[fieldKey] = encrypted
	}

	return v.store.UpdateSecret(ctx, customerID, id, name, description, encryptedFields, v.kekVersion)
}
