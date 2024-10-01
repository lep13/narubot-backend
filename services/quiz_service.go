package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/lep13/narubot-backend/db"
	"github.com/lep13/narubot-backend/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

///////////////////////Quiz Session Management/////////////////////////////
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

// SaveUserQuizSession saves the updated quiz session to MongoDB
func SaveUserQuizSession(roomId string, session *models.QuizSession) error {
	collection := db.MongoClient.Database("narubot").Collection("quiz_sessions")

	filter := bson.M{"user_id": roomId}
	update := bson.M{
		"$set": bson.M{
			"current_q_no": session.CurrentQNo,
			"scores":       session.Scores,
			"last_updated": time.Now().Unix(),
		},
	}

	_, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Printf("Error saving session for user %s: %v", roomId, err)
		return err
	}

	return nil
}

// FinalizeQuizAndResetSession calculates the final result, displays it, and resets the session
func FinalizeQuizAndResetSession(roomId string, session *models.QuizSession, accessToken string) error {
	// Calculate the final character
	finalCharacter, err := CalculateQuizResult(session)
	if err != nil {
		return fmt.Errorf("failed to calculate final result: %v", err)
	}

	// Load character descriptions
	descriptions, err := LoadCharacterDescriptions("quiz_questions.json")
	if err != nil {
		return fmt.Errorf("failed to load character descriptions: %v", err)
	}

	// Get the description of the final character
	characterInfo, exists := descriptions[finalCharacter]
	if !exists {
		return fmt.Errorf("character description not found for: %s", finalCharacter)
	}

	// Format the final result message
	finalMessage := fmt.Sprintf("You are most like %s! %s", finalCharacter, characterInfo.Description)

	// Send the result message to Webex
	err = SendMessageToWebex(roomId, finalMessage, accessToken)
	if err != nil {
		return fmt.Errorf("failed to send final result to Webex: %v", err)
	}

	// Reset the session for the user
	return ResetQuizSession(roomId)
}

// ResetQuizSession clears the session so the user can take the quiz again
func ResetQuizSession(roomId string) error {
	collection := db.MongoClient.Database("narubot").Collection("quiz_sessions")
	_, err := collection.DeleteOne(context.Background(), bson.M{"user_id": roomId})
	if err != nil {
		log.Printf("Error resetting session for user %s: %v", roomId, err)
		return err
	}
	return nil
}
///////////////////////Quiz Session Management/////////////////////////////


///////////////////Quiz Progress and Response Handling/////////////////////
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

// LoadQuizQuestions loads the quiz questions and options from a JSON file
func LoadQuizQuestions(filename string) ([]models.QuizQuestion, error) {
	return nil, nil // You can fill in the implementation here
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

// ContinueQuiz moves the user to the next question or calculates the result if the quiz is finished
func ContinueQuiz(roomId, accessToken, userAnswer string) error {
	// Track the user's response and update session
	err := TrackQuizResponse(roomId, userAnswer)
	if err != nil {
		return fmt.Errorf("failed to track user response: %v", err)
	}

	// Load user session to check the current question number
	session, err := GetUserQuizSession(roomId)
	if err != nil {
		return fmt.Errorf("failed to retrieve user session: %v", err)
	}

	// Load quiz questions to check if we are at the last question
	questions, err := LoadQuizQuestions("quiz_questions.json")
	if err != nil {
		return fmt.Errorf("failed to load quiz questions: %v", err)
	}

	// Check if the user has answered all the questions
	if session.CurrentQNo >= len(questions) {
		// All questions answered, calculate the result
		result, err := CalculateQuizResult(session)
		if err != nil {
			return fmt.Errorf("failed to calculate quiz result: %v", err)
		}
		return SendMessageToWebex(roomId, result, accessToken)
	}

	// Move to the next question
	nextQuestion := questions[session.CurrentQNo]
	card, err := models.CreateQuizCard(nextQuestion.Question, nextQuestion.Options)
	if err != nil {
		return fmt.Errorf("failed to create quiz card: %v", err)
	}

	// Send the next question card to Webex
	return SendMessageWithCard(roomId, card, accessToken)
}

// TrackQuizResponse updates the session based on the user's answer
func TrackQuizResponse(roomId string, answer string) error {
	// Load user session
	session, err := GetUserQuizSession(roomId)
	if err != nil {
		return fmt.Errorf("failed to retrieve user session: %v", err)
	}

	// Load quiz questions from JSON file
	questions, err := LoadQuizQuestions("quiz_questions.json")
	if err != nil {
		return fmt.Errorf("failed to load quiz questions: %v", err)
	}

	// Get the current question number
	currentQuestionNo := session.CurrentQNo

	// Make sure we are within the question range
	if currentQuestionNo >= len(questions) {
		return fmt.Errorf("no more questions available")
	}

	// Get the current question and the associated scores for the selected answer
	currentQuestion := questions[currentQuestionNo]
	scoreMapping, exists := currentQuestion.Scores[answer]
	if !exists {
		return fmt.Errorf("invalid answer: %s", answer)
	}

	// Increment the scores for the characters mapped to the selected answer
	for _, character := range scoreMapping {
		session.Scores[character]++
	}

	// Increment the current question number
	session.CurrentQNo++

	// Save the updated session
	return SaveUserQuizSession(roomId, session)
}

// LoadCharacterDescriptions loads the character descriptions from the JSON file
func LoadCharacterDescriptions(filename string) (map[string]models.CharacterInfo, error) {
	jsonFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	var data struct {
		Characters map[string]models.CharacterInfo `json:"characters"`
	}
	err = json.Unmarshal(byteValue, &data)
	if err != nil {
		return nil, err
	}

	return data.Characters, nil
}
///////////////////Quiz Progress and Response Handling/////////////////////

///////////////////////////Result Calculation//////////////////////////////
// CalculateQuizResult determines which character has the highest score, handles ties, and includes character description
func CalculateQuizResult(session *models.QuizSession) (string, error) {
	highestScore := 0
	resultCharacters := []string{}

	// Loop through the scores map to find the highest score and any ties
	for character, score := range session.Scores {
		if score > highestScore {
			highestScore = score
			resultCharacters = []string{character} // Reset to the new highest scoring character
		} else if score == highestScore {
			resultCharacters = append(resultCharacters, character) // Add to the tie
		}
	}

	// If we have a tie, randomly select one of the tied characters
	var finalCharacter string
	if len(resultCharacters) > 1 {
		selectedIndex := rand.Intn(len(resultCharacters)) // Select a random index
		finalCharacter = resultCharacters[selectedIndex]
	} else {
		finalCharacter = resultCharacters[0]
	}

	// Load character descriptions from JSON
	descriptions, err := LoadCharacterDescriptions("quiz_questions.json")
	if err != nil {
		return "", fmt.Errorf("failed to load character descriptions: %v", err)
	}

	// Get the description and image for the final character
	characterInfo, exists := descriptions[finalCharacter]
	if !exists {
		return "", fmt.Errorf("character description not found")
	}

	// Return the result as a message with description and image (optional)
	return fmt.Sprintf("You are most like %s! %s", finalCharacter, characterInfo.Description), nil

}
///////////////////////////Result Calculation//////////////////////////////

/////////////////////////////Error Handling////////////////////////////////
// HandleErrorAndGuideUser informs the user about an error and suggests next steps
func HandleErrorAndGuideUser(roomId string, accessToken string, errorMessage string) error {
	message := fmt.Sprintf("Oops, something went wrong: %s. Please try again or restart the quiz.", errorMessage)
	return SendMessageToWebex(roomId, message, accessToken)
}

// HandleIncompleteQuizResponse prompts users if they have an incomplete quiz and asks them to continue or restart
func HandleIncompleteQuizResponse(roomId, action, accessToken string) error {
	switch action {
	case "ContinueQuiz":
		// Continue the quiz from where they left off
		return ContinueQuiz(roomId, accessToken, "")
	case "RestartQuiz":
		// Reset session and start a new quiz
		err := ResetQuizSession(roomId)
		if err != nil {
			return err
		}
		return StartQuiz(roomId, accessToken)
	default:
		return fmt.Errorf("invalid action for quiz: %s", action)
	}
}

// CheckAndPromptIncompleteSession checks for incomplete sessions and prompts the user to continue or restart
func CheckAndPromptIncompleteSession(roomId, accessToken string) error {
	session, err := GetUserQuizSession(roomId)
	if err != nil {
		if err.Error() == "new session created" {
			// If a new session is created, prompt to start the quiz
			return SendMessageToWebex(roomId, "You have a new quiz session. Would you like to start the quiz?", accessToken)
		}
		return HandleErrorAndGuideUser(roomId, accessToken, "Error retrieving session")
	}

	// If there is an incomplete session, prompt the user to continue or restart
	if !session.IsCompleted {
		return SendMessageToWebex(roomId, "You have an incomplete quiz. Would you like to continue or restart?", accessToken)
	}

	return nil
}
/////////////////////////////Error Handling////////////////////////////////