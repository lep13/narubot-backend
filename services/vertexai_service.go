package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"github.com/lep13/narubot-backend/models"  
	"net/http"
)

type VertexRequest struct {
	Instances []map[string]interface{} `json:"instances"`
}

type VertexResponse struct {
	Predictions []map[string]interface{} `json:"predictions"`
}

// GenerateVertexAIResponse sends a request to the Vertex AI model using the Generative AI API
func GenerateVertexAIResponse(prompt string, cfg *models.Config) (string, error) {  // Updated to models.Config
	// Use the Generative AI API endpoint
	url := fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:predict",
		cfg.GoogleRegion, cfg.GoogleProjectID, cfg.GoogleRegion, cfg.GoogleModelID)

	// Create the request payload with the text prompt
	reqBody := map[string]interface{}{
		"instances": []map[string]interface{}{
			{
				"messages": []map[string]interface{}{
					{
						"content": prompt,
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	// Use the GenAI Access Token from Secrets Manager
	token := cfg.GenAIAccessToken

	// Create an HTTP client and make the POST request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	// Set the headers for authorization and content-type
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
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
	if len(res.Predictions) == 0 {
		return "", fmt.Errorf("no predictions found in response")
	}

	candidates := res.Predictions[0]["candidates"].([]interface{})
	firstCandidate := candidates[0].(map[string]interface{})
	responseText, ok := firstCandidate["content"].(string)
	if !ok {
		return "", fmt.Errorf("no content field in response")
	}

	return responseText, nil
}
