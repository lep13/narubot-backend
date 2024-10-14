package models

// Config structure to hold secrets from AWS Secrets Manager
type Config struct {
    BotEmail          string `json:"BOT_EMAIL"`
    WebexAccessToken  string `json:"WEBEX_ACCESS_TOKEN"`
    ServiceAccountKey string `json:"SERVICE_ACCOUNT_KEY"`
    GoogleProjectID   string `json:"GOOGLE_PROJECT_ID"`
    GoogleModelID     string `json:"GOOGLE_MODEL_ID"`
    GoogleRegion      string `json:"GOOGLE_REGION"`
    GenAIAccessToken  string `json:"GENAI_ACCESS_TOKEN"`  // GenAI OAuth token
    MongoURI          string `json:"MONGO_URI"`           
}
