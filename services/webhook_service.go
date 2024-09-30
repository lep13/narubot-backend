package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// GetMessageContent fetches the message content using the Webex message ID
func GetMessageContent(messageId, accessToken string) (string, error) {
	client := &http.Client{}
	url := fmt.Sprintf("https://webexapis.com/v1/messages/%s", messageId)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch message content from Webex, status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Print the raw response body for debugging purposes
	fmt.Printf("Webex API Response: %s\n", string(body))

	var messageData map[string]interface{}
	if err := json.Unmarshal(body, &messageData); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %v", err)
	}

	// Log the entire parsed response for debugging
	fmt.Printf("Parsed Webex Response: %+v\n", messageData)

	// Extract the "text" field
	text, ok := messageData["text"].(string)
	if !ok {
		return "", fmt.Errorf("no 'text' field found in the message response")
	}

	return text, nil
}

// Function to send a message back to Webex
func SendMessageToWebex(roomId, message, accessToken string) error {
	client := &http.Client{}
	messageData := map[string]string{
		"roomId": roomId,
		"text":   message,
	}
	jsonData, _ := json.Marshal(messageData)

	req, err := http.NewRequest("POST", "https://webexapis.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	_, err = client.Do(req)
	return err
}
