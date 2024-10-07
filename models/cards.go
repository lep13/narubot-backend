package models

import (
	"fmt"
	"strings"
)

// Card structure
type Card struct {
	RoomId      string           `json:"roomId"`
	Markdown    string           `json:"markdown"`
	Attachments []CardAttachment `json:"attachments"`
}

// CardAttachment structure
type CardAttachment struct {
	ContentType string      `json:"contentType"`
	Content     CardContent `json:"content"`
}

// CardContent structure
type CardContent struct {
	Type    string       `json:"type"`
	Version string       `json:"version"`
	Body    []CardBody   `json:"body"`
	Actions []CardAction `json:"actions"`
}

// CardBody structure
type CardBody struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// CardAction structure
type CardAction struct {
	Type  string            `json:"type"`
	Title string            `json:"title"`
	Data  map[string]string `json:"data"`
}

// CreateQuizCard generates an adaptive card for a quiz question with clickable options
func CreateQuizCard(question string, options []string) (map[string]interface{}, error) {
	choices := make([]string, len(options))
	for i, option := range options {
		choices[i] = fmt.Sprintf("%d. %s", i+1, option)
	}

	cardContent := map[string]interface{}{
		"type":    "AdaptiveCard",
		"version": "1.3",
		"body": []interface{}{
			map[string]interface{}{
				"type":   "TextBlock",
				"text":   question,
				"weight": "Bolder",
				"size":   "Medium",
			},
			map[string]interface{}{
				"type": "TextBlock",
				"text": strings.Join(choices, "\n"),
				"wrap": true,
			},
		},
		"$schema": "http://adaptivecards.io/schemas/adaptive-card.json",
	}

	card := map[string]interface{}{
		"contentType": "application/vnd.microsoft.card.adaptive",
		"content":     cardContent,
	}

	return card, nil
}

// Helper function to convert string options to Webex adaptive card choices
// func convertToChoices(options []string) ([]map[string]interface{}, error) {
// 	if len(options) == 0 {
// 		return nil, fmt.Errorf("no options provided for quiz choices")
// 	}

// 	choices := []map[string]interface{}{}
// 	for _, option := range options {
// 		choices = append(choices, map[string]interface{}{
// 			"title": option,
// 			"value": option,
// 		})
// 	}
// 	return choices, nil
// }
