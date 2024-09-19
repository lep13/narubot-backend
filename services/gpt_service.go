package services

import (
	"bytes"
	"encoding/json"
	"net/http"
)

const GPTModel = "text-davinci-003"

// OpenAI request payload structure
type GPTRequest struct {
	Model     string `json:"model"`
	Prompt    string `json:"prompt"`
	MaxTokens int    `json:"max_tokens"`
}

// OpenAI response structure
type GPTResponse struct {
	Choices []struct {
		Text string `json:"text"`
	} `json:"choices"`
}

// Function to call OpenAI's GPT API
func GetGPTResponse(prompt string, apiKey string) (string, error) {
	client := &http.Client{}
	reqBody := GPTRequest{
		Model:     GPTModel,
		Prompt:    prompt,
		MaxTokens: 150,
	}

	jsonData, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var gptResponse GPTResponse
	if err := json.NewDecoder(resp.Body).Decode(&gptResponse); err != nil {
		return "", err
	}

	return gptResponse.Choices[0].Text, nil
}
