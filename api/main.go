package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"simple-vault/api/helpers"
)

type User struct {
	ID       string `json:"id,omitempty"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password,omitempty"`
}

type Secret struct {
	ID          string            `bson:"_id" json:"id"`
	UserID      string            `bson:"userId" json:"userId"`
	Name        string            `bson:"name" json:"name"`
	Description string            `bson:"description" json:"description"`
	Data        map[string]string `bson:"data" json:"data"`
	CreatedAt   time.Time         `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time         `bson:"updatedAt" json:"updatedAt"`
}

var mongoClient *mongo.Client
var db *mongo.Database
var secretsCollection *mongo.Collection
var usersCollection *mongo.Collection

func initDB() error {
	// Get MongoDB connection string from environment
	uri := os.Getenv("MONGODB_URI")
	uri = "mongodb://root:secret123@212.64.215.155:32169/vault?authSource=admin"
	if uri == "" {
		// Build URI from individual components
		host := getEnv("DB_HOST", "localhost")
		port := getEnv("DB_PORT", "27017")
		user := getEnv("DB_USER", "")
		password := getEnv("DB_PASSWORD", "")
		dbname := getEnv("DB_NAME", "vault")

		if user != "" && password != "" {
			uri = "mongodb://" + user + ":" + password + "@" + host + ":" + port + "/" + dbname + "?authSource=admin"
		} else {
			uri = "mongodb://" + host + ":" + port + "/" + dbname
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return err
	}

	// Store client globally for health checks
	mongoClient = client

	// Ping the database to verify connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return err
	}

	// Get database name from URI or use default
	dbname := getEnv("DB_NAME", "vault")
	db = client.Database(dbname)
	secretsCollection = db.Collection("secrets")

	// Users collection is in vcluster database
	vclusterDB := client.Database("vcluster")
	usersCollection = vclusterDB.Collection("users")

	// Create indexes for faster lookups
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{bson.E{Key: "name", Value: 1}},
			Options: options.Index().SetUnique(false),
		},
		{
			Keys:    bson.D{bson.E{Key: "userId", Value: 1}},
			Options: options.Index().SetUnique(false),
		},
		{
			Keys:    bson.D{bson.E{Key: "userId", Value: 1}, bson.E{Key: "name", Value: 1}},
			Options: options.Index().SetUnique(false),
		},
	}
	_, err = secretsCollection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		log.Printf("Warning: Failed to create indexes: %v", err)
	}

	log.Println("MongoDB connected successfully")
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getUserIDFromToken extracts userID from JWT token by:
// 1. Getting Authorization header
// 2. Validating token with auth service to get username
// 3. Querying users collection to get user's _id
func getUserIDFromToken(c *gin.Context) (string, error) {
	// Get Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", gin.Error{Err: nil, Type: gin.ErrorTypePublic, Meta: "Authorization header is required"}
	}

	// Validate token with auth service and get username
	username, err := helpers.ValidateToken(authHeader)
	if err != nil {
		return "", gin.Error{Err: err, Type: gin.ErrorTypePublic, Meta: err.Error()}
	}

	// Query users collection to get user's _id from MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var userDoc bson.M
	err = usersCollection.FindOne(ctx, bson.M{"username": username}).Decode(&userDoc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", gin.Error{Err: nil, Type: gin.ErrorTypePublic, Meta: "User not found"}
		}
		return "", err
	}

	// Extract the _id (ObjectID) from the MongoDB document
	userIDObj, ok := userDoc["_id"].(primitive.ObjectID)
	if !ok {
		return "", gin.Error{Err: nil, Type: gin.ErrorTypePublic, Meta: "Invalid user ID format"}
	}

	// Convert ObjectID to string
	userID := userIDObj.Hex()
	return userID, nil
}

// authMiddleware extracts userID from JWT and stores it in context
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := getUserIDFromToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized User"})
			c.Abort()
			return
		}

		// Store userID in context for use in handlers
		c.Set("userID", userID)
		c.Next()
	}
}

func main() {
	// Initialize database
	if err := initDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	r := gin.Default()

	// Custom CORS middleware to ensure headers are always set
	r.Use(func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			origin = "*"
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "false")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		c.Writer.Header().Set("Access-Control-Max-Age", "1728000")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// Health check endpoint (no authentication required)
	r.GET("/health", func(c *gin.Context) {
		// Check database connection
		if mongoClient == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  "database client not initialized",
			})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := mongoClient.Ping(ctx, nil); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  "database connection failed",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
		})
	})

	// API key authentication endpoint (no JWT middleware)
	r.POST("/api/v1/secrets/access", getSecretByAccessKey)

	// API routes with authentication middleware
	api := r.Group("/api/v1")
	api.Use(authMiddleware())
	{
		// Get all secrets
		api.GET("/secrets", getSecrets)
		// Get a specific secret
		api.GET("/secrets/:id", getSecret)
		// Create a new secret
		api.POST("/secrets", createSecret)
		// Update a secret
		api.PUT("/secrets/:id", updateSecret)
		// Delete a secret
		api.DELETE("/secrets/:id", deleteSecret)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

func getSecrets(c *gin.Context) {
	// Get userID from context (set by authMiddleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	userIDStr := userID.(string)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Filter secrets by userID
	filter := bson.M{"userId": userIDStr}
	cursor, err := secretsCollection.Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch secrets: " + err.Error()})
		return
	}
	defer cursor.Close(ctx)

	var secrets []Secret
	if err = cursor.All(ctx, &secrets); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode secrets: " + err.Error()})
		return
	}

	// Convert to API format
	result := make([]map[string]interface{}, len(secrets))
	for i, secret := range secrets {
		result[i] = map[string]interface{}{
			"id":          secret.ID,
			"userId":      secret.UserID,
			"name":        secret.Name,
			"description": secret.Description,
			"data":        secret.Data,
			"createdAt":   secret.CreatedAt.Format(time.RFC3339),
			"updatedAt":   secret.UpdatedAt.Format(time.RFC3339),
		}
	}

	c.JSON(http.StatusOK, gin.H{"secrets": result})
}

func getSecret(c *gin.Context) {
	id := c.Param("id")

	// Get userID from context (set by authMiddleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	userIDStr := userID.(string)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find secret by ID and userID to ensure ownership
	var secret Secret
	err := secretsCollection.FindOne(ctx, bson.M{"_id": id, "userId": userIDStr}).Decode(&secret)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Secret not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch secret: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"id":          secret.ID,
		"userId":      secret.UserID,
		"name":        secret.Name,
		"description": secret.Description,
		"data":        secret.Data,
		"createdAt":   secret.CreatedAt.Format(time.RFC3339),
		"updatedAt":   secret.UpdatedAt.Format(time.RFC3339),
	})
}

func createSecret(c *gin.Context) {
	var req struct {
		Name        string            `json:"name" binding:"required"`
		Description string            `json:"description"`
		Data        map[string]string `json:"data" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get userID from context (set by authMiddleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	userIDStr := userID.(string)
	now := time.Now()
	secret := Secret{
		ID:          uuid.New().String(),
		UserID:      userIDStr,
		Name:        req.Name,
		Description: req.Description,
		Data:        req.Data,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := secretsCollection.InsertOne(ctx, secret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create secret: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, map[string]interface{}{
		"id":          secret.ID,
		"userId":      secret.UserID,
		"name":        secret.Name,
		"description": secret.Description,
		"data":        secret.Data,
		"createdAt":   secret.CreatedAt.Format(time.RFC3339),
		"updatedAt":   secret.UpdatedAt.Format(time.RFC3339),
	})
}

func updateSecret(c *gin.Context) {
	id := c.Param("id")

	// Get userID from context (set by authMiddleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	userIDStr := userID.(string)

	var req struct {
		Name        string            `json:"name" binding:"required"`
		Description string            `json:"description"`
		Data        map[string]string `json:"data" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if secret exists and belongs to user
	var existing Secret
	err := secretsCollection.FindOne(ctx, bson.M{"_id": id, "userId": userIDStr}).Decode(&existing)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Secret not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch secret: " + err.Error()})
		return
	}

	// Update secret (only if it belongs to the user)
	update := bson.M{
		"$set": bson.M{
			"name":        req.Name,
			"description": req.Description,
			"data":        req.Data,
			"updatedAt":   time.Now(),
		},
	}

	result, err := secretsCollection.UpdateOne(ctx, bson.M{"_id": id, "userId": userIDStr}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update secret: " + err.Error()})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Secret not found"})
		return
	}

	// Fetch updated secret
	var updated Secret
	secretsCollection.FindOne(ctx, bson.M{"_id": id, "userId": userIDStr}).Decode(&updated)

	c.JSON(http.StatusOK, map[string]interface{}{
		"id":          updated.ID,
		"userId":      updated.UserID,
		"name":        updated.Name,
		"description": updated.Description,
		"data":        updated.Data,
		"createdAt":   updated.CreatedAt.Format(time.RFC3339),
		"updatedAt":   updated.UpdatedAt.Format(time.RFC3339),
	})
}

func deleteSecret(c *gin.Context) {
	id := c.Param("id")

	// Get userID from context (set by authMiddleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	userIDStr := userID.(string)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Delete secret only if it belongs to the user
	result, err := secretsCollection.DeleteOne(ctx, bson.M{"_id": id, "userId": userIDStr})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete secret: " + err.Error()})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Secret not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Secret deleted successfully"})
}

func getSecretByAccessKey(c *gin.Context) {
	var req struct {
		AccessKey string `json:"accessKey" binding:"required"`
		SecretKey string `json:"secretKey" binding:"required"`
		Name      string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid JSON body", "error": err.Error()})
		return
	}

	if req.AccessKey == "" || req.SecretKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Access key and secret key are required"})
		return
	}

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Secret name is required"})
		return
	}

	// Validate access key and secret key with auth service
	authResp, err := helpers.ValidateAccessKey(req.AccessKey, req.SecretKey)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Authentication failed", "error": err.Error()})
		return
	}

	// Get userID from auth response
	userID := authResp.UserID
	if userID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "User ID not found in auth response"})
		return
	}

	// Find secret by name and userID
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var secret Secret
	err = secretsCollection.FindOne(ctx, bson.M{"name": req.Name, "userId": userID}).Decode(&secret)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"message": "Secret not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch secret", "error": err.Error()})
		return
	}

	// Return secret in the requested format
	c.JSON(http.StatusOK, map[string]interface{}{
		"name":        secret.Name,
		"description": secret.Description,
		"data":        secret.Data,
	})
}
