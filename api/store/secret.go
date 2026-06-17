package store

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"simple-vault/api/crypto"
)

type SecretDocument struct {
	ID              primitive.ObjectID                 `bson:"_id,omitempty"`
	CustomerID      string                             `bson:"customer_id"`
	Name            string                             `bson:"name"`
	Description     string                             `bson:"description"`
	EncryptedFields map[string]*crypto.EncryptedSecret `bson:"encrypted_fields"` // 👈 this was missing
	KEKVersion      int                                `bson:"kek_version"`
	CreatedAt       time.Time                          `bson:"created_at"`
	UpdatedAt       time.Time                          `bson:"updated_at"`
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

func (s *SecretStore) SaveSecret(ctx context.Context, customerID, name, description string, fields map[string]*crypto.EncryptedSecret, kekVersion int) error {
	filter := bson.M{"customer_id": customerID, "name": name}
	update := bson.M{
		"$set": bson.M{
			"encrypted_fields": fields,
			"description":      description,
			"kek_version":      kekVersion,
			"updated_at":       time.Now(),
		},
		"$setOnInsert": bson.M{
			"customer_id": customerID,
			"name":        name,
			"created_at":  time.Now(),
		},
	}
	_, err := s.secrets.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	return err
}

func (s *SecretStore) GetSecret(ctx context.Context, customerID, name string) (*SecretDocument, error) {
	var doc SecretDocument
	filter := bson.M{
		"customer_id": customerID,
	}

	if name != "" {
		filter["name"] = name
	}
	err := s.secrets.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch secret: %w", err)
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

func (s *SecretStore) UpdateSecret(ctx context.Context, customerID string, id primitive.ObjectID, name, description string, fields map[string]*crypto.EncryptedSecret, kekVersion int) error {
	result, err := s.secrets.UpdateOne(ctx,
		bson.M{
			"_id":         id,
			"customer_id": customerID, // ownership check at DB level too
		},
		bson.M{"$set": bson.M{
			"name":             name,
			"description":      description,
			"encrypted_fields": fields,
			"kek_version":      kekVersion,
			"updated_at":       time.Now(),
		}},
	)
	if err != nil {
		return fmt.Errorf("failed to update secret: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("secret not found or access denied")
	}
	return nil
}
