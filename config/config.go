package config

import (
    "context"
    "encoding/json"
    "fmt"
    "narubot-backend/models" 

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// SecretsManagerInterface defines the interface for Secrets Manager client methods used in our code.
type SecretsManagerInterface interface {
    GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

// SecretManagerFunc allows for injecting a custom Secrets Manager function for testing.
var SecretManagerFunc = func() (SecretsManagerInterface, error) {
    cfg, err := loadAWSConfig(context.TODO())
    if err != nil {
        return nil, err
    }
    return secretsmanager.NewFromConfig(cfg), nil
}

// loadAWSConfig is a variable that points to the function that loads AWS config.
var loadAWSConfig = config.LoadDefaultConfig

// LoadConfig fetches the secrets from AWS Secrets Manager and returns a models.Config struct
func LoadConfig() (*models.Config, error) {
    secretName := "webex_bot"

    svc, err := SecretManagerFunc()
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
    config := &models.Config{}

    err = json.Unmarshal([]byte(secretString), config)
    if err != nil {
        return nil, fmt.Errorf("failed to unmarshal secret string: %w", err)
    }

    return config, nil
}
