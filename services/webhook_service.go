package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/lep13/narubot-backend/models"
)

// SendGreetingWithOptions sends a greeting with interactive options
func SendGreetingWithOptions(roomId, accessToken string) error {
	card := map[string]interface{}{
		"roomId":   roomId,
		"markdown": "Narubot is here, dattabayo!",
		"attachments": []map[string]interface{}{
			{
				"contentType": "application/vnd.microsoft.card.adaptive",
				"content": map[string]interface{}{
					"type":    "AdaptiveCard",
					"version": "1.3",
					"$schema": "http://adaptivecards.io/schemas/adaptive-card.json",
					"body": []map[string]interface{}{
						{
							"type":   "TextBlock",
							"text":   "Narubot is here, dattabayo! What would you like to do?",
							"weight": "Bolder",
							"size":   "Medium",
						},
					},
					"actions": []map[string]interface{}{
						{
							"type":  "Action.Submit",
							"title": "Ask me a question about Naruto",
							"id":    "ask_question_action",
							"data": map[string]interface{}{
								"inputs": map[string]string{
									"action": "AskQuestion",
								},
							},
						},
						{
							"type":  "Action.Submit",
							"title": "Take a personality quiz",
							"id":    "start_quiz_action",
							"data": map[string]interface{}{
								"inputs": map[string]string{
									"action": "StartQuiz",
								},
							},
						},
					},
				},
			},
		},
	}

	cardData, err := json.Marshal(card)
	if err != nil {
		return fmt.Errorf("failed to marshal card: %v", err)
	}

	req, err := http.NewRequest("POST", "https://webexapis.com/v1/messages", bytes.NewBuffer(cardData))
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
		return fmt.Errorf("non-OK HTTP status: %v", resp.Status)
	}

	return nil
}

// ParseCardActionUsingAttachment extracts card actions from Webex's "attachmentActions" payload
func ParseCardActionUsingAttachment(payload map[string]interface{}) (*models.CardAction, error) {
    // Check for data field in payload
    data, ok := payload["data"].(map[string]interface{})
    if !ok {
        return nil, fmt.Errorf("no 'data' field found in the payload")
    }

    // Check for inputs within data
    inputs, ok := data["inputs"].(map[string]interface{})
    if !ok {
        return nil, fmt.Errorf("no 'inputs' field found in 'data'")
    }

	log.Printf("Data Content: %+v\n", data)
	log.Printf("Inputs Content: %+v\n", inputs)

    // Check for action within inputs
    action, actionOk := inputs["action"].(string)
    if !actionOk {
        return nil, fmt.Errorf("no 'action' field found in 'inputs'")
    }

    return &models.CardAction{
        Data: map[string]string{
            "action": action,
        },
    }, nil
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
