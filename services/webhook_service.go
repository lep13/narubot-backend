package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// SendMessageWithCard sends a card message to Webex
func SendMessageWithCard(userID string, card map[string]interface{}, accessToken string) error {
	messageData := map[string]interface{}{
		"roomId":      userID,
		"markdown":    "Here's your quiz result!",
		"attachments": []interface{}{card},
	}

	jsonData, err := json.Marshal(messageData)
	if err != nil {
		return fmt.Errorf("failed to marshal card: %v", err)
	}

	req, err := http.NewRequest("POST", "https://webexapis.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send card: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var responseBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&responseBody)
		return fmt.Errorf("non-OK HTTP status: %s", resp.Status)
	}

	return nil
}

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

	var messageData map[string]interface{}
	if err := json.Unmarshal(body, &messageData); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %v", err)
	}

	// Extract the "text" field with logging in case it's missing
	text, ok := messageData["text"].(string)
	if !ok {
		log.Printf("Warning: 'text' field is missing from Webex message data: %+v", messageData)
		return "", fmt.Errorf("no 'text' field found in the message response")
	}

	return text, nil
}

// SendMessageToWebex sends a message back to Webex
func SendMessageToWebex(roomId, message, accessToken string) error {
	client := &http.Client{}
	messageData := map[string]string{
		"roomId": roomId,
		"text":   message,
	}
	jsonData, err := json.Marshal(messageData)
	if err != nil {
		return fmt.Errorf("failed to marshal message data: %v", err)
	}

	req, err := http.NewRequest("POST", "https://webexapis.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message to Webex: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("non-OK HTTP status: %s, response: %s", resp.Status, body)
	}

	return nil
}
