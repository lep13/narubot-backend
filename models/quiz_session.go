package models

import "fmt"

// QuizOption and QuizQuestion structs define the structure of a quiz question and its options
type QuizOption struct {
    Text      string `json:"text"`
    Character string `json:"character"`
    Score     int    `json:"score"`
}

type QuizQuestion struct {
    Question string       `json:"question"`
    Options  []QuizOption `json:"options"`
}

// QuizSession manages the state of a user's quiz
type QuizSession struct {
    UserID      string         `bson:"user_id"`
    CurrentQNo  int            `bson:"current_q_no"`
    Scores      map[string]int `bson:"scores"`
    IsCompleted bool           `bson:"is_completed"`
    LastUpdated int64          `bson:"last_updated"`
}

// CharacterInfo stores description details of each character
type CharacterInfo struct {
    Description string `json:"description"`
    Image       string `json:"image"`
}

// UpdateScore updates the score based on the selected answer
func (qs *QuizSession) UpdateScore(answer string, currentQuestion QuizQuestion) error {
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

    if _, exists := qs.Scores[selectedOption.Character]; exists {
        qs.Scores[selectedOption.Character] += selectedOption.Score
    } else {
        qs.Scores[selectedOption.Character] = selectedOption.Score
    }

    qs.CurrentQNo++
    return nil
}
