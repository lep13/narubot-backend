package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/lep13/narubot-backend/db"
	"github.com/lep13/narubot-backend/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CreateQuizSession initializes a quiz session for a user
func CreateQuizSession(userID string) (*models.QuizSession, error) {
	collection := db.GetCollection()
	session := models.QuizSession{
		UserID:      userID,
		CurrentQNo:  0,
		Scores:      make(map[string]int),
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
	collection := db.GetCollection()
	filter := bson.M{"user_id": userID}

	var session models.QuizSession
	err := collection.FindOne(context.Background(), filter).Decode(&session)
	if err == mongo.ErrNoDocuments {
		return CreateQuizSession(userID)
	}
	return &session, err
}

// SaveUserQuizSession saves the updated quiz session to MongoDB
func SaveUserQuizSession(userID string, session *models.QuizSession) error {
	collection := db.GetCollection()
	filter := bson.M{"user_id": userID}
	update := bson.M{
		"$set": bson.M{
			"current_q_no": session.CurrentQNo,
			"scores":       session.Scores,
			"is_completed": session.IsCompleted,
			"last_updated": time.Now().Unix(),
		},
	}

	_, err := collection.UpdateOne(context.Background(), filter, update, options.Update())
	if err != nil {
		log.Printf("Error saving session for user %s: %v", userID, err)
		return err
	}
	return nil
}

// ResetQuizSession clears the session so the user can take the quiz again
func ResetQuizSession(userID string) error {
	collection := db.GetCollection()
	_, err := collection.DeleteOne(context.Background(), bson.M{"user_id": userID})
	if err != nil {
		log.Printf("Error resetting session for user %s: %v", userID, err)
		return err
	}
	return nil
}

// TrackQuizResponse updates the user's score based on the selected answer
func TrackQuizResponse(userID string, answer string) error {
	session, err := GetUserQuizSession(userID)
	if err != nil {
		return fmt.Errorf("failed to retrieve user session: %v", err)
	}

	questions, err := LoadQuizQuestions("quiz_questions.json")
	if err != nil {
		return fmt.Errorf("failed to load quiz questions: %v", err)
	}

	if session.CurrentQNo-1 >= len(questions) {
		return fmt.Errorf("no more questions available")
	}

	currentQuestion := questions[session.CurrentQNo-1]
	answerIndex, err := strconv.Atoi(answer)
	if err != nil || answerIndex < 1 || answerIndex > len(currentQuestion.Options) {
		return fmt.Errorf("invalid answer format or out of range: %v", err)
	}

	selectedOption := currentQuestion.Options[answerIndex-1]
	session.Scores[selectedOption.Character] += selectedOption.Score

	return SaveUserQuizSession(userID, session)
}

// LoadQuizQuestions loads the quiz questions and options from a JSON file
func LoadQuizQuestions(filename string) ([]models.QuizQuestion, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("could not open quiz file: %v", err)
	}
	defer file.Close()

	var quizData struct {
		Questions []models.QuizQuestion `json:"questions"`
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&quizData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode quiz questions: %v", err)
	}

	return quizData.Questions, nil
}

// sendQuizQuestion sends the current question in the session to the user with numbering
func sendQuizQuestion(userID string, session *models.QuizSession, questions []models.QuizQuestion, accessToken string) error {
	if session.CurrentQNo >= len(questions) {
		return FinalizeQuizAndResetSession(userID, session, accessToken)
	}

	currentQuestion := questions[session.CurrentQNo]
	questionText := fmt.Sprintf("Question %d: %s", session.CurrentQNo+1, currentQuestion.Question) // Number the question
	options := extractOptions(currentQuestion.Options)
	formattedOptions := ""
	for i, option := range options {
		formattedOptions += fmt.Sprintf("%d. %s\n", i+1, option)
	}

	message := fmt.Sprintf("%s\n%s", questionText, formattedOptions)
	err := SendMessageToWebex(userID, message, accessToken)
	if err != nil {
		return fmt.Errorf("failed to send quiz question: %v", err)
	}

	session.CurrentQNo++
	return SaveUserQuizSession(userID, session)
}

// StartQuiz initializes and sends the first quiz card
func StartQuiz(userID, accessToken string) error {
	err := ResetQuizSession(userID)
	if err != nil {
		return fmt.Errorf("failed to reset session: %v", err)
	}

	session, err := CreateQuizSession(userID)
	if err != nil {
		return fmt.Errorf("failed to start quiz session: %v", err)
	}

	questions, err := LoadQuizQuestions("quiz_questions.json")
	if err != nil {
		return fmt.Errorf("failed to load quiz questions: %v", err)
	}

	return sendQuizQuestion(userID, session, questions, accessToken)
}

// ContinueQuiz sends the next question in the quiz sequence
func ContinueQuiz(userID string, userAnswer string, accessToken string) error {
	if userAnswer == "quit" {
		return HandleQuit(userID, accessToken)
	}

	err := TrackQuizResponse(userID, userAnswer)
	if err != nil {
		return fmt.Errorf("failed to track user response: %v", err)
	}

	session, err := GetUserQuizSession(userID)
	if err != nil {
		return fmt.Errorf("failed to retrieve user session: %v", err)
	}

	questions, err := LoadQuizQuestions("quiz_questions.json")
	if err != nil {
		return fmt.Errorf("failed to load quiz questions: %v", err)
	}

	if session.CurrentQNo >= len(questions) {
		return FinalizeQuizAndResetSession(userID, session, accessToken)
	}

	return sendQuizQuestion(userID, session, questions, accessToken)
}

// HandleQuit allows the user to quit the quiz session.
func HandleQuit(userID string, accessToken string) error {
	quitMessage := "You have quit the quiz session. If you'd like to start again, just say 'start'."
	err := SendMessageToWebex(userID, quitMessage, accessToken)
	if err != nil {
		return fmt.Errorf("failed to send quit message: %v", err)
	}

	return ResetQuizSession(userID)
}

func FinalizeQuizAndResetSession(userID string, session *models.QuizSession, accessToken string) error {
	character, err := CalculateQuizResult(session)
	if err != nil {
		return fmt.Errorf("failed to calculate quiz result: %v", err)
	}

	descriptions, err := LoadCharacterDescriptions("character_description.json")
	if err != nil {
		return fmt.Errorf("failed to load character descriptions: %v", err)
	}

	characterInfo, exists := descriptions[character]
	if !exists {
		characterInfo.Description = "Description not found."
	}

	finalMessage := fmt.Sprintf("You are most like %s! %s\n\n![Image](%s)", character, characterInfo.Description, characterInfo.Image)
	SendMessageToWebex(userID, finalMessage, accessToken)

	return ResetQuizSession(userID)
}

// Utility function to extract option texts
func extractOptions(options []models.QuizOption) []string {
	var opts []string
	for _, option := range options {
		opts = append(opts, option.Text)
	}
	return opts
}

// CalculateQuizResult determines which character has the highest score
func CalculateQuizResult(session *models.QuizSession) (string, error) {
	highestScore := 0
	resultCharacters := []string{}

	for character, score := range session.Scores {
		if score > highestScore {
			highestScore = score
			resultCharacters = []string{character}
		} else if score == highestScore {
			resultCharacters = append(resultCharacters, character)
		}
	}

	var finalCharacter string
	if len(resultCharacters) > 1 {
		selectedIndex := rand.Intn(len(resultCharacters))
		finalCharacter = resultCharacters[selectedIndex]
	} else {
		finalCharacter = resultCharacters[0]
	}

	return finalCharacter, nil
}

// LoadCharacterDescriptions loads character descriptions from a JSON file
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
