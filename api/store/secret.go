package store

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"yourmodule/crypto"
)

type SecretDocument struct {
	ID            primitive.ObjectID     `bson:"_id,omitempty"`
	CustomerID    string                 `bson:"customer_id"`
	SecretKey     string                 `bson:"secret_key"` // consider encrypting this too
	EncryptedData crypto.EncryptedSecret `bson:"encrypted_data"`
	KEKVersion    int                    `bson:"kek_version"`
	CreatedAt     time.Time              `bson:"created_at"`
	UpdatedAt     time.Time              `bson:"updated_at"`
}

// CustomerKEKDocument stores the encrypted KEK per customer in a separate collection
type CustomerKEKDocument struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	CustomerID   string             `bson:"customer_id"`
	EncryptedKEK string             `bson:"encrypted_kek"` // KEK encrypted by master key
	Version      int                `bson:"version"`
	CreatedAt    time.Time          `bson:"created_at"`
}

type SecretStore struct {
	secrets *mongo.Collection
	keks    *mongo.Collection
}

func NewSecretStore(db *mongo.Database) *SecretStore {
	return &SecretStore{
		secrets: db.Collection("secrets"),
		keks:    db.Collection("customer_keks"),
	}
}

func (s *SecretStore) SaveSecret(ctx context.Context, customerID, secretKey string, encrypted *crypto.EncryptedSecret, kekVersion int) error {
	filter := bson.M{"customer_id": customerID, "secret_key": secretKey}
	update := bson.M{
		"$set": bson.M{
			"encrypted_data": encrypted,
			"kek_version":    kekVersion,
			"updated_at":     time.Now(),
		},
		"$setOnInsert": bson.M{
			"customer_id": customerID,
			"secret_key":  secretKey,
			"created_at":  time.Now(),
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err := s.secrets.UpdateOne(ctx, filter, update, opts)
	return err
}

func (s *SecretStore) GetSecret(ctx context.Context, customerID, secretKey string) (*SecretDocument, error) {
	var doc SecretDocument
	err := s.secrets.FindOne(ctx, bson.M{
		"customer_id": customerID,
		"secret_key":  secretKey,
	}).Decode(&doc)
	if err != nil {
		return nil, fmt.Errorf("secret not found: %w", err)
	}
	return &doc, nil
}

func (s *SecretStore) SaveCustomerKEK(ctx context.Context, customerID, encryptedKEK string, version int) error {
	doc := CustomerKEKDocument{
		CustomerID:   customerID,
		EncryptedKEK: encryptedKEK,
		Version:      version,
		CreatedAt:    time.Now(),
	}
	_, err := s.keks.InsertOne(ctx, doc)
	return err
}

func (s *SecretStore) GetCustomerKEK(ctx context.Context, customerID string, version int) (*CustomerKEKDocument, error) {
	var doc CustomerKEKDocument
	err := s.keks.FindOne(ctx, bson.M{
		"customer_id": customerID,
		"version":     version,
	}).Decode(&doc)
	if err != nil {
		return nil, fmt.Errorf("KEK not found for customer %s version %d: %w", customerID, version, err)
	}
	return &doc, nil
}
