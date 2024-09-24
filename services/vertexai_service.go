package services

import (
	"context"
	"fmt"

	aiplatform "cloud.google.com/go/aiplatform/apiv1"
	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/structpb"
	"narubot-backend/config"
)

// GetGenAIResponse queries Google Gen AI (Vertex AI) for a completion based on the given prompt
func GetGenAIResponse(prompt string) (string, error) {
	ctx := context.Background()

	cfg, err := config.LoadConfig()
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}

	client, err := aiplatform.NewPredictionClient(ctx, option.WithCredentialsFile(cfg.ServiceAccountKey))
	if err != nil {
		return "", fmt.Errorf("failed to create GenAI client: %w", err)
	}
	defer client.Close()

	// Endpoint for Vertex AI prediction
	endpoint := fmt.Sprintf("projects/%s/locations/us-central1/publishers/google/models/text-bison", cfg.GoogleProjectID)

	// Properly initializing the structpb.Value with the prompt
	structVal, err := structpb.NewValue(prompt)
	if err != nil {
		return "", fmt.Errorf("failed to create structpb.Value: %w", err)
	}

	req := &aiplatformpb.PredictRequest{
		Endpoint:  endpoint,
		Instances: []*structpb.Value{structVal},
	}

	resp, err := client.Predict(ctx, req)
	if err != nil {
		return "", fmt.Errorf("GenAI request failed: %w", err)
	}

	if len(resp.Predictions) > 0 {
		// Assuming the response text is located in the first prediction's string value
		return resp.Predictions[0].GetStringValue(), nil
	}

	return "", fmt.Errorf("no response from GenAI")
}
