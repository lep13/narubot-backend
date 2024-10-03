package models

import "fmt"

// quiz question and its options
// type QuizQuestion struct {
//     Question string               `json:"question"`
//     Options  []string             `json:"options"`
//     Scores   map[string][]string   `json:"scores"`  // Maps option to characters
// }
type QuizOption struct {
    Text      string `json:"text"`
    Character string `json:"character"`
    Score     int    `json:"score"`
}

type QuizQuestion struct {
    Question string       `json:"question"`
    Options  []QuizOption `json:"options"`
}

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
	Type    string      `json:"type"`
	Version string      `json:"version"`
	Body    []CardBody  `json:"body"`
	Actions []CardAction `json:"actions"`
}

// CardBody structure
type CardBody struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// CardAction structure
// type CardAction struct {
// 	Type  string            `json:"type"`
// 	Title string            `json:"title"`
// 	Value string            `json:"value"`
// 	Action string           `json:"action"` // for action handling
// }
// CardAction structure
type CardAction struct {
	Type  string            `json:"type"`  // Action.Submit etc.
	Title string            `json:"title"` // Label displayed on the button
	Data  map[string]string `json:"data"`  // Data payload, stores the action identifier
}

// CreateGreetingCard generates a Webex adaptive card for the greeting with clickable options
func CreateGreetingCard() (map[string]interface{}, error) {
	card := map[string]interface{}{
		"contentType": "application/vnd.microsoft.card.adaptive",
		"content": map[string]interface{}{
			"type":    "AdaptiveCard",
			"version": "1.0",
			"body": []map[string]interface{}{
				{
					"type": "TextBlock",
					"text": "Narubot is here, dattabayo! What would you like to do?",
					"weight": "Bolder",
					"size": "Medium",
				},
			},
			"actions": []map[string]interface{}{
				{
					"type":  "Action.Submit",
					"title": "Ask me a question about Naruto",
					"data": map[string]string{
						"action": "AskQuestion",
					},
				},
				{
					"type":  "Action.Submit",
					"title": "Take a personality quiz",
					"data": map[string]string{
						"action": "StartQuiz",
					},
				},
			},
		},
	}
	return card, nil
}

// CreateQuizCard generates a Webex adaptive card for a quiz question
func CreateQuizCard(question string, options []string) (map[string]interface{}, error) {
	if len(options) == 0 {
		return nil, fmt.Errorf("no options provided for the quiz question")
	}
	
	choices, err := convertToChoices(options)  // Handle the second return value (error)
	if err != nil {
		return nil, err  // Return error if converting choices fails
	}

	card := map[string]interface{}{
		"type": "AdaptiveCard",
		"body": []interface{}{
			map[string]interface{}{
				"type":  "TextBlock",
				"text":  question,
				"weight": "Bolder",
				"size":  "Medium",
			},
			map[string]interface{}{
				"type":  "Input.ChoiceSet",
				"id":    "quizAnswer",
				"style": "expanded",
				"choices": choices,  // Use the converted choices
			},
		},
		"actions": []interface{}{
			map[string]interface{}{
				"type":  "Action.Submit",
				"title": "Submit Answer",
			},
		},
		"$schema":  "http://adaptivecards.io/schemas/adaptive-card.json",
		"version":  "1.2",
	}
	return card, nil
}

// Helper function to convert string options to Webex adaptive card choices
func convertToChoices(options []string) ([]map[string]interface{}, error) {
	if len(options) == 0 {
		return nil, fmt.Errorf("no options provided for quiz choices")
	}

	choices := []map[string]interface{}{}
	for _, option := range options {
		choices = append(choices, map[string]interface{}{
			"title": option,
			"value": option,
		})
	}
	return choices, nil
}
