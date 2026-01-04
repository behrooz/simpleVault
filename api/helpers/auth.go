package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

var AuthServiceURL = getEnv("AUTH_SERVICE_URL", "http://192.168.1.4:8083")

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

type TokenValidationRequest struct {
	Token string `json:"token"`
}

type TokenValidationResponse struct {
	Valid    bool   `json:"valid"`
	Username string `json:"username,omitempty"`
	Message  string `json:"message,omitempty"`
}

// ValidateTokenWithAuthService validates a JWT token with the auth service
func ValidateTokenWithAuthService(token string) (*TokenValidationResponse, error) {
	// Strip "Bearer " prefix if present
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	reqBody := TokenValidationRequest{
		Token: token,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest("POST", AuthServiceURL+"/validate", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call auth service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("auth service returned error: %s - %s", resp.Status, string(body))
	}

	var validationResp TokenValidationResponse
	if err := json.NewDecoder(resp.Body).Decode(&validationResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &validationResp, nil
}

// ValidateToken validates a token and returns the username if valid
func ValidateToken(token string) (string, error) {
	resp, err := ValidateTokenWithAuthService(token)
	if err != nil {
		return "", err
	}

	if !resp.Valid {
		return "", fmt.Errorf("token validation failed: %s", resp.Message)
	}

	return resp.Username, nil
}

type AccessKeyAuthRequest struct {
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
}

type AccessKeyAuthResponse struct {
	Token    string `json:"token"`
	UserID   string `json:"userID"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Message  string `json:"message"`
}

// ValidateAccessKey validates access key and secret key with the auth service
func ValidateAccessKey(accessKey, secretKey string) (*AccessKeyAuthResponse, error) {
	reqBody := AccessKeyAuthRequest{
		AccessKey: accessKey,
		SecretKey: secretKey,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest("POST", AuthServiceURL+"/auth/apikey", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call auth service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("auth service returned error: %s - %s", resp.Status, string(body))
	}

	var authResp AccessKeyAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &authResp, nil
}
