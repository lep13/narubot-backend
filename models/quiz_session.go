package models

type QuizSession struct {
	UserID      string         `bson:"user_id"`
	CurrentQNo  int            `bson:"current_q_no"`
	Scores      map[string]int `bson:"scores"`
	IsCompleted bool           `bson:"is_completed"`
	LastUpdated int64          `bson:"last_updated"`
}

type CharacterInfo struct{
	Description string `json:"description"`
    Image       string `json:"image"`
}

// UpdateScore updates the score based on the selected answer.
func (qs *QuizSession) UpdateScore(answer string) {
	// You can map answers to specific character scores here.
	// For example:
	answerMap := map[string]string{
		"Talk it out diplomatically": "Sakura",
		"Charge in headfirst!":       "Naruto",
		"Stay calm and think":        "Kakashi",
		"Manipulate events":          "Itachi",
	}

	character := answerMap[answer]
	if _, exists := qs.Scores[character]; exists {
		qs.Scores[character] += 1
	} else {
		qs.Scores[character] = 1
	}

	qs.CurrentQNo += 1
}
