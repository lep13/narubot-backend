package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type VertexRequest struct {
	Instances []map[string]interface{} `json:"instances"`
}

type VertexResponse struct {
	Candidates []map[string]interface{} `json:"candidates"`
}

// GenerateVertexAIResponse sends a request to the Vertex AI model using the Generative AI API
func GenerateVertexAIResponse(prompt, serviceAccountKey, projectID, modelID, region string) (string, error) {
	// Use the Generative AI API endpoint
	url := fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:generateText", region, projectID, region, modelID)

	// Create the request payload with the text prompt
	reqBody := map[string]interface{}{
		"prompt": map[string]interface{}{
			"text": prompt,
		},
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	// Create an OAuth2 token from the service account key
	ctx := context.Background()
	credentials, err := google.CredentialsFromJSON(ctx, []byte(serviceAccountKey), "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return "", fmt.Errorf("failed to generate credentials: %v", err)
	}

	// Get the token
	token, err := credentials.TokenSource.Token()
	if err != nil {
		return "", fmt.Errorf("failed to retrieve token: %v", err)
	}

	// Create an HTTP client with the token
	client := oauth2.NewClient(ctx, credentials.TokenSource)

	// Create a POST request to send the prompt
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	// Check if the response was successful
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("non-OK HTTP status: %v", resp.Status)
	}

	// Parse the response body
	var res VertexResponse
	err = json.Unmarshal(body, &res)
	if err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	// Extract the generated text from the response
	if len(res.Candidates) == 0 {
		return "", fmt.Errorf("no candidates found in response")
	}
	responseText, ok := res.Candidates[0]["output"].(string)
	if !ok {
		return "", fmt.Errorf("no output field in response")
	}

	return responseText, nil
}
