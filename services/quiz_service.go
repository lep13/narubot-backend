package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"narubot-backend/db"
	"narubot-backend/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// CreateQuizSession initializes a quiz session for a user
func CreateQuizSession(userID string) (*mongo.InsertOneResult, error) {
	collection := db.Client.Database("narubot").Collection("quiz_sessions")
	session := models.QuizSession{
		UserID:      userID,
		CurrentQNo:  0,                       // Start from question 0
		Scores:      make(map[string]int),    // Initialize empty score map
		IsCompleted: false,
		LastUpdated: time.Now().Unix(),
	}

	result, err := collection.InsertOne(context.Background(), session)
	if err != nil {
		log.Printf("Error creating session for user %s: %v", userID, err)
		return nil, err
	}

	return result, nil
}

// GetQuizSession retrieves the current session for the user
func GetQuizSession(userID string) (*models.QuizSession, error) {
	collection := db.Client.Database("narubot").Collection("quiz_sessions")
	filter := bson.M{"user_id": userID}

	var session models.QuizSession
	err := collection.FindOne(context.Background(), filter).Decode(&session)
	if err == mongo.ErrNoDocuments {
		// If no session exists, create one
		CreateQuizSession(userID)
		return nil, fmt.Errorf("new session created")
	} else if err != nil {
		log.Printf("Error retrieving session for user %s: %v", userID, err)
		return nil, err
	}

	return &session, nil
}

// UpdateQuizSession updates the session after each answer
func UpdateQuizSession(userID string, questionNo int, updatedScores map[string]int) error {
	collection := db.Client.Database("narubot").Collection("quiz_sessions")

	filter := bson.M{"user_id": userID}
	update := bson.M{
		"$set": bson.M{
			"current_q_no": questionNo,
			"scores":       updatedScores,
			"last_updated": time.Now().Unix(),
		},
	}

	_, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Printf("Error updating session for user %s: %v", userID, err)
		return err
	}

	return nil
}

// CompleteQuizSession marks the quiz as completed
func CompleteQuizSession(userID string) error {
	collection := db.Client.Database("narubot").Collection("quiz_sessions")

	filter := bson.M{"user_id": userID}
	update := bson.M{
		"$set": bson.M{
			"is_completed": true,
			"last_updated": time.Now().Unix(),
		},
	}

	_, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Printf("Error marking session as completed for user %s: %v", userID, err)
		return err
	}

	return nil
}
