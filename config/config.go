package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// Config structure to hold secrets from AWS Secrets Manager
type Config struct {
	BotEmail          string `json:"BOT_EMAIL"`
	GoogleProjectID   string `json:"GOOGLE_PROJECT_ID"`
	WebexAccessToken  string `json:"WEBEX_ACCESS_TOKEN"`
	ServiceAccountKey string `json:"SERVICE_ACCOUNT_KEY"`
}

// LoadConfig fetches secrets from AWS Secrets Manager and loads the Google service account key
func LoadConfig() (*Config, error) {
	secretName := "webex_bot"

	// Load Google Service Account credentials from file or AWS Secrets Manager
	serviceAccountKey := "gifted-fragment-436605-u0-d3cc86aed1ab.json" // Replace with the actual path

	file, err := os.ReadFile(serviceAccountKey)
	if err != nil {
		return nil, fmt.Errorf("failed to read Google service account file: %w", err)
	}

	var config Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal Google service account file: %w", err)
	}

	// Retrieve AWS Secrets Manager values
	svc, err := loadAWSSecretsManager()
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	result, err := svc.GetSecretValue(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve secret: %w", err)
	}

	secretString := *result.SecretString
	err = json.Unmarshal([]byte(secretString), &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal secret string: %w", err)
	}

	return &config, nil
}

// loadAWSSecretsManager is used to initialize AWS Secrets Manager
func loadAWSSecretsManager() (*secretsmanager.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}
	return secretsmanager.NewFromConfig(cfg), nil
}
