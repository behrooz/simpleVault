package service

import (
	"context"
	"fmt"

	"simple-vault/api/crypto"
	"simple-vault/api/store"
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
func (v *VaultService) SaveSecret(ctx context.Context, customerID, key, value string) error {
	// 1. Fetch and decrypt the customer's KEK
	kek, err := v.getCustomerKEK(ctx, customerID, v.kekVersion)
	if err != nil {
		return err
	}
	defer crypto.Zeroize(kek) // wipe from memory after use

	// 2. Encrypt the secret value
	encrypted, err := crypto.EncryptSecret(value, kek)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	// 3. Persist to MongoDB — only ciphertext is stored
	return v.store.SaveSecret(ctx, customerID, key, encrypted, v.kekVersion)
}

// GetSecret fetches and decrypts a customer's secret.
func (v *VaultService) GetSecret(ctx context.Context, customerID, key string) (string, error) {
	// 1. Fetch the encrypted record
	doc, err := v.store.GetSecret(ctx, customerID, key)
	if err != nil {
		return "", err
	}

	// 2. Fetch and decrypt the KEK used at the time of encryption
	kek, err := v.getCustomerKEK(ctx, customerID, doc.KEKVersion)
	if err != nil {
		return "", err
	}
	defer crypto.Zeroize(kek)

	// 3. Decrypt and return — plaintext never touches MongoDB
	return crypto.DecryptSecret(&doc.EncryptedData, kek)
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
