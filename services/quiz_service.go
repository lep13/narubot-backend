package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/lep13/narubot-backend/db"
	"github.com/lep13/narubot-backend/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// CreateQuizSession initializes a quiz session for a user
func CreateQuizSession(userID string) (*models.QuizSession, error) {
	collection := db.MongoClient.Database("narubot").Collection("quiz_sessions")
	session := models.QuizSession{
		UserID:      userID,
		CurrentQNo:  0,                    // Start from question 0
		Scores:      make(map[string]int), // Initialize empty score map
		IsCompleted: false,
		LastUpdated: time.Now().Unix(),
	}

	_, err := collection.InsertOne(context.Background(), session)
	if err != nil {
		log.Printf("Error creating session for user %s: %v", userID, err)
		return nil, err
	}

	return &session, nil
}

// StartQuiz initializes the personality quiz and sends the first question
func StartQuiz(roomId, accessToken string) error {
	questions, err := LoadQuizQuestions("quiz_questions.json")
	if err != nil {
		return fmt.Errorf("failed to load quiz questions: %v", err)
	}

	firstQuestion := questions[0].Question
	options := questions[0].Options

	card, err := models.CreateQuizCard(firstQuestion, options)
	if err != nil {
		return fmt.Errorf("failed to create quiz card: %v", err)
	}

	return SendMessageWithCard(roomId, card, accessToken)
}

// ContinueQuiz moves the user to the next question
func ContinueQuiz(roomId, accessToken, userAnswer string) error {
	session, err := GetUserQuizSession(roomId)
	if err != nil {
		return fmt.Errorf("failed to retrieve user session: %v", err)
	}

	err = TrackQuizResponse(roomId, userAnswer)
	if err != nil {
		return fmt.Errorf("failed to track user response: %v", err)
	}

	questions, err := LoadQuizQuestions("quiz_questions.json")
	if err != nil {
		return fmt.Errorf("failed to load quiz questions: %v", err)
	}

	if session.CurrentQNo >= len(questions) {
		result := CalculateQuizResult(session)
		return SendMessageToWebex(roomId, result, accessToken)
	}

	nextQuestion := questions[session.CurrentQNo]
	card, err := models.CreateQuizCard(nextQuestion.Question, nextQuestion.Options)
	if err != nil {
		return fmt.Errorf("failed to create quiz card: %v", err)
	}

	return SendMessageWithCard(roomId, card, accessToken)
}

// LoadQuizQuestions loads the quiz questions and options from a JSON file
func LoadQuizQuestions(filename string) ([]models.QuizQuestion, error) {
	return nil, nil // You can fill in the implementation here
}

// TrackQuizResponse updates the session based on the user's answer
func TrackQuizResponse(roomId string, answer string) error {
	// Load user session
	session, err := GetUserQuizSession(roomId)
	if err != nil {
		return fmt.Errorf("failed to retrieve user session: %v", err)
	}

	// Load the current question
	questions, err := LoadQuizQuestions("quiz_questions.json")
	if err != nil {
		return fmt.Errorf("failed to load quiz questions: %v", err)
	}
	currentQuestion := questions[session.CurrentQNo]

	// Update scores based on the user's answer
	if characters, ok := currentQuestion.Scores[answer]; ok {
		for _, character := range characters {
			session.Scores[character] += 1 // Increment the character's score
		}
	}

	// Move to the next question
	session.CurrentQNo += 1

	// Save the updated session back to the database
	return SaveUserQuizSession(roomId, session)
}

// CalculateQuizResult determines the quiz result based on the user's answers
func CalculateQuizResult(session *models.QuizSession) string {
	// Implement logic to calculate the quiz result
	return "Your result"
}

// GetUserQuizSession retrieves the current session for the user
func GetUserQuizSession(userID string) (*models.QuizSession, error) {
	collection := db.MongoClient.Database("narubot").Collection("quiz_sessions")
	filter := bson.M{"user_id": userID}

	var session models.QuizSession
	err := collection.FindOne(context.Background(), filter).Decode(&session)
	if err == mongo.ErrNoDocuments {
		return CreateQuizSession(userID)
	}
	return &session, err
}

// SaveUserQuizSession saves the updated session to the database
func SaveUserQuizSession(userID string, session *models.QuizSession) error {
	collection := db.MongoClient.Database("narubot").Collection("quiz_sessions")
	filter := bson.M{"user_id": userID}
	update := bson.M{"$set": session}

	_, err := collection.UpdateOne(context.Background(), filter, update)
	return err
}

// SendMessageWithCard sends an adaptive card to Webex
func SendMessageWithCard(roomId string, card map[string]interface{}, accessToken string) error {
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
