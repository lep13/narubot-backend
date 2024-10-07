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
    }
    return card, nil
}


func CreateQuizCard(question string, options []string) (map[string]interface{}, error) {
    if len(options) == 0 {
        return nil, fmt.Errorf("no options provided for the quiz question")
    }

    choices, err := convertToChoices(options)  // Convert options to the choice format
    if err != nil {
        return nil, err  // Handle error in conversion
    }

    cardContent := map[string]interface{}{
        "type":    "AdaptiveCard",
        "version": "1.3",
        "body": []interface{}{
            map[string]interface{}{
                "type":  "TextBlock",
                "text":  question,
                "weight": "Bolder",
                "size":  "Medium",
            },
            map[string]interface{}{
                "type":   "Input.ChoiceSet",
                "id":     "quizAnswer",  // Add unique id for capturing input
                "style":  "expanded",
                "choices": choices,
            },
        },
        "actions": []interface{}{
            map[string]interface{}{
                "type":  "Action.Submit",
                "title": "Submit Answer",
                "data": map[string]string{
                    "action": "SubmitQuizAnswer",
                },
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
