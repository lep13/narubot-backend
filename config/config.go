package config

import (
    "os"
)

// Struct to store the secrets
type Config struct {
    WebexAccessToken string
    BotEmail         string
    GPTAPIKey        string
}

// Function to load secrets from environment variables
func LoadConfig() Config {
    return Config{
        WebexAccessToken: os.Getenv("WEBEX_ACCESS_TOKEN"),
        BotEmail:         os.Getenv("BOT_EMAIL"),
        GPTAPIKey:        os.Getenv("GPT_API_KEY"),
    }
}
