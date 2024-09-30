package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// QuizSession struct for tracking quiz progress for each user
type QuizSession struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"` // MongoDB unique ID
	UserID       string             `bson:"user_id"`       // Webex User ID (e.g., person's email or ID)
	CurrentQNo   int                `bson:"current_q_no"`  // The current question number user is on
	Scores       map[string]int     `bson:"scores"`        // Cumulative scores for each character
	IsCompleted  bool               `bson:"is_completed"`  // Whether the quiz is finished
	LastUpdated  int64              `bson:"last_updated"`  // Timestamp for session management
}
