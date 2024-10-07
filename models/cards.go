package models

// Card structure for sending adaptive cards
type Card struct {
	RoomId      string           `json:"roomId"`
	Markdown    string           `json:"markdown"`
	Attachments []CardAttachment `json:"attachments"`
}

// CardAttachment structure for adaptive card attachments
type CardAttachment struct {
	ContentType string      `json:"contentType"`
	Content     CardContent `json:"content"`
}

// CardContent structure represents the main content of an adaptive card
type CardContent struct {
	Type    string     `json:"type"`
	Version string     `json:"version"`
	Body    []CardBody `json:"body"`
}

// CardBody structure defines a single element within the card body
type CardBody struct {
	Type      string `json:"type"`
	Text      string `json:"text"`
	Wrap      bool   `json:"wrap"`
	Size      string `json:"size,omitempty"`
	Weight    string `json:"weight,omitempty"`
}

// CreateTextCard generates an adaptive card for text-only messages
func CreateTextCard(text string) map[string]interface{} {
	cardContent := map[string]interface{}{
		"type":    "AdaptiveCard",
		"version": "1.3",
		"body": []interface{}{
			map[string]interface{}{
				"type":   "TextBlock",
				"text":   text,
				"wrap":   true,
				"size":   "Medium",
				"weight": "Bolder",
			},
		},
		"$schema": "http://adaptivecards.io/schemas/adaptive-card.json",
	}

	card := map[string]interface{}{
		"contentType": "application/vnd.microsoft.card.adaptive",
		"content":     cardContent,
	}

	return card
}

// CreateImageCard generates an adaptive card that includes an image and descriptive text
func CreateImageCard(imageURL, title, description string) map[string]interface{} {
	cardContent := map[string]interface{}{
		"type":    "AdaptiveCard",
		"version": "1.3",
		"body": []interface{}{
			map[string]interface{}{
				"type": "Image",
				"url":  imageURL,
				"size": "Large",
				"altText": title,
			},
			map[string]interface{}{
				"type": "TextBlock",
				"text": title,
				"weight": "Bolder",
				"size": "Medium",
			},
			map[string]interface{}{
				"type": "TextBlock",
				"text": description,
				"wrap": true,
			},
		},
		"$schema": "http://adaptivecards.io/schemas/adaptive-card.json",
	}

	card := map[string]interface{}{
		"contentType": "application/vnd.microsoft.card.adaptive",
		"content":     cardContent,
	}

	return card
}
