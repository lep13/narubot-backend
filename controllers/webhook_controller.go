package controllers

import (
	"log"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lep13/narubot-backend/config"
	"github.com/lep13/narubot-backend/models"
	"github.com/lep13/narubot-backend/services"
)

// HandleWebhook processes Webex messages and handles quiz-related actions
func HandleWebhook(c *gin.Context) {
	var payload map[string]interface{}
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(400, gin.H{"status": "bad request", "reason": "invalid JSON payload"})
		return
	}

	resource, resourceOk := payload["resource"].(string)
	if !resourceOk {
		c.JSON(400, gin.H{"status": "bad request", "reason": "resource field missing"})
		return
	}

	data, ok := payload["data"].(map[string]interface{})
	if !ok {
		c.JSON(400, gin.H{"status": "bad request", "reason": "no data field"})
		return
	}

	personEmail, emailOk := data["personEmail"].(string)
	roomId, roomOk := data["roomId"].(string)
	messageId, messageOk := data["id"].(string)

	cfg, err := config.LoadConfig()
	if err != nil {
		c.JSON(500, gin.H{"status": "failed to load config"})
		return
	}

	if personEmail == cfg.BotEmail {
		log.Println("Ignoring message from bot itself")
		c.JSON(200, gin.H{"status": "ignored"})
		return
	}

	// Handle regular messages, e.g., starting or continuing the quiz
	if resource == "messages" && emailOk && roomOk && messageOk {
		messageText, err := services.GetMessageContent(messageId, cfg.WebexAccessToken)
		if err != nil {
			services.SendMessageToWebex(roomId, "I'm sorry, I couldn't retrieve the message.", cfg.WebexAccessToken)
			c.JSON(500, gin.H{"status": "failed to get message content"})
			return
		}

		lowerMessage := strings.ToLower(messageText)
		session, sessionErr := services.GetUserQuizSession(roomId)

		switch {
		case lowerMessage == "quit":
			// Quit the quiz if in session
			if sessionErr == nil && !session.IsCompleted {
				err := services.HandleQuit(roomId, cfg.WebexAccessToken)
				if err != nil {
					c.JSON(500, gin.H{"status": "failed to quit quiz"})
					return
				}
				services.SendMessageToWebex(roomId, "You have exited the quiz. Feel free to start over by saying 'quiz' anytime!", cfg.WebexAccessToken)
				c.JSON(200, gin.H{"status": "quiz quit"})
			} else {
				services.SendMessageToWebex(roomId, "No active quiz to quit. Say 'quiz' to start one!", cfg.WebexAccessToken)
				c.JSON(200, gin.H{"status": "no quiz to quit"})
			}
			return

		case lowerMessage == "quiz":
			// // Send quiz instructions
			// services.SendMessageToWebex(roomId, "I’ll ask you 10 questions. Reply with the number of your choice, and I’ll reveal which Naruto character you are! Say 'start' when you’re ready. You may say 'quit' anytime you want to exit the quiz.", cfg.WebexAccessToken)
			// c.JSON(200, gin.H{"status": "quiz instructions sent"})
			// return
			err := sendQuizInstructionsCard(roomId, cfg.WebexAccessToken)
			if err != nil {
				c.JSON(500, gin.H{"status": "failed to send quiz instructions"})
				return
			}
			c.JSON(200, gin.H{"status": "quiz instructions sent"})

		case lowerMessage == "start":
			// Start the quiz
			err := services.StartQuiz(roomId, cfg.WebexAccessToken)
			if err != nil {
				services.SendMessageToWebex(roomId, "I couldn't start the quiz. Please try again.", cfg.WebexAccessToken)
				c.JSON(500, gin.H{"status": "failed to start quiz"})
				return
			}
			c.JSON(200, gin.H{"status": "quiz started"})
			return

		case isNumeric(lowerMessage):
			// Continue the quiz if in session
			if sessionErr == nil && !session.IsCompleted {
				err := services.ContinueQuiz(roomId, lowerMessage, cfg.WebexAccessToken)
				if err != nil {
					services.SendMessageToWebex(roomId, "There was an issue with your answer. Please reply with '1', '2', '3', or '4', or type 'quit' to exit.", cfg.WebexAccessToken)
					c.JSON(500, gin.H{"status": "failed to continue quiz"})
					return
				}
				c.JSON(200, gin.H{"status": "quiz continued"})
				return
			}
			services.SendMessageToWebex(roomId, "No active quiz. Say 'quiz' to start one!", cfg.WebexAccessToken)
			c.JSON(200, gin.H{"status": "no active quiz"})
			return

		case isGreeting(lowerMessage):
			// Respond to greeting
			// services.SendMessageToWebex(roomId, "Narubot is here, dattabayo! Let's chat! Or if you want to take a quiz about Naruto, say 'quiz'!", cfg.WebexAccessToken)
			// c.JSON(200, gin.H{"status": "greeting sent"})
			// return
			err := sendGreetingCard(roomId, cfg.WebexAccessToken)
			if err != nil {
				c.JSON(500, gin.H{"status": "failed to send greeting card"})
				return
			}
			c.JSON(200, gin.H{"status": "greeting sent"})

		default:
			// If quiz is ongoing, send specific instructions
			if sessionErr == nil && !session.IsCompleted {
				services.SendMessageToWebex(roomId, "Please respond with '1', '2', '3', or '4' for your answer, or type 'quit' to exit the quiz.", cfg.WebexAccessToken)
				c.JSON(200, gin.H{"status": "invalid input during quiz"})
			} else {
				// If no quiz is active, use Vertex AI for other inputs
				vertexResponse, err := services.GenerateVertexAIResponse(messageText, cfg)
				if err != nil {
					services.SendMessageToWebex(roomId, "I'm sorry, I couldn't generate a response.", cfg.WebexAccessToken)
					c.JSON(500, gin.H{"status": "failed to generate Vertex AI response"})
					return
				}
				services.SendMessageToWebex(roomId, vertexResponse, cfg.WebexAccessToken)
				c.JSON(200, gin.H{"status": "response sent"})
			}
		}
	}
}

// Helper function to check if the message is a greeting
func isGreeting(message string) bool {
	greetings := []string{"hello", "hey", "hi"}
	message = strings.ToLower(message)
	for _, greeting := range greetings {
		if strings.Contains(message, greeting) {
			return true
		}
	}
	return false
}

func sendQuizInstructionsCard(userID, accessToken string) error {
	instructionsText := "I’ll ask you 10 questions. Reply with the number of your choice, and I’ll reveal which Naruto character you are! Say 'start' when you’re ready. You may say 'quit' anytime you want to exit the quiz."
	card := models.CreateTextCard(instructionsText)
	return services.SendMessageWithCard(userID, card, accessToken)
}

func sendGreetingCard(userID, accessToken string) error {
	greetingText := "Narubot is here, dattabayo! Let's chat! Or if you want to take a quiz about Naruto, say 'quiz'!"
	card := models.CreateTextCard(greetingText)
	return services.SendMessageWithCard(userID, card, accessToken)
}

// helper function to check if a string is numeric and within the quiz answer range
func isNumeric(s string) bool {
	answer, err := strconv.Atoi(s)
	return err == nil && answer >= 1 && answer <= 4
}
