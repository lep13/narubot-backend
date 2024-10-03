package models

import "fmt"

type QuizSession struct {
	UserID      string         `bson:"user_id"`
	CurrentQNo  int            `bson:"current_q_no"`
	Scores      map[string]int `bson:"scores"`
	IsCompleted bool           `bson:"is_completed"`
	LastUpdated int64          `bson:"last_updated"`
}

type CharacterInfo struct {
	Description string `json:"description"`
	Image       string `json:"image"`
}

// UpdateScore updates the score based on the selected answer.
func (qs *QuizSession) UpdateScore(answer string, currentQuestion QuizQuestion) error {
    // Find the matching option based on the user's answer
    var selectedOption *QuizOption
    for _, option := range currentQuestion.Options {
        if option.Text == answer {
            selectedOption = &option
            break
        }
    }

    if selectedOption == nil {
        return fmt.Errorf("invalid answer: %s", answer)
    }

    // Increment the score for the selected character
    if _, exists := qs.Scores[selectedOption.Character]; exists {
        qs.Scores[selectedOption.Character] += selectedOption.Score
    } else {
        qs.Scores[selectedOption.Character] = selectedOption.Score
    }

    // Proceed to the next question
    qs.CurrentQNo++

    return nil
}

