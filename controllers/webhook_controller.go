package controllers

import (
	"fmt"
	"github.com/lep13/narubot-backend/config"
	"github.com/lep13/narubot-backend/services"

	"github.com/gin-gonic/gin"
)

// HandleWebhook processes Webex messages and interactive card actions
func HandleWebhook(c *gin.Context) {
	var payload map[string]interface{}
	if err := c.BindJSON(&payload); err == nil {
		data, ok := payload["data"].(map[string]interface{})
		if !ok {
			c.JSON(400, gin.H{"status": "bad request", "reason": "no data field"})
			return
		}

		personEmail, ok := data["personEmail"].(string)
		roomId, roomOk := data["roomId"].(string)
		messageId, messageOk := data["id"].(string)

		if !ok || !roomOk || !messageOk {
			c.JSON(400, gin.H{"status": "bad request", "reason": "required fields missing"})
			return
		}

		// Load configuration
		cfg, err := config.LoadConfig()
		if err != nil {
			c.JSON(500, gin.H{"status": "failed to load config"})
			return
		}

		// Ensure the message is not from the bot itself
		if personEmail != cfg.BotEmail {
			// Check if this is a card submission via "attachmentActions"
			if attachmentActions, ok := payload["attachmentActions"].(map[string]interface{}); ok {
				// Handle card submission
				actionData, err := services.ParseCardActionUsingAttachment(attachmentActions)
				if err == nil {
					switch actionData.Data["action"] {
					case "AskQuestion":
						// Forward message to the existing Vertex AI chatbot
						vertexResponse, err := services.GenerateVertexAIResponse("Ask me a question", cfg)
						if err != nil {
							fmt.Println("Error generating Vertex AI response:", err)
							services.SendMessageToWebex(roomId, "I'm sorry, I couldn't generate a response.", cfg.WebexAccessToken)
							c.JSON(500, gin.H{"status": "failed to generate Vertex AI response"})
							return
						}

						// Send Vertex AI response back to Webex
						err = services.SendMessageToWebex(roomId, vertexResponse, cfg.WebexAccessToken)
						if err != nil {
							c.JSON(500, gin.H{"status": "failed to send message"})
							return
						}

					case "StartQuiz":
						// Start the personality quiz
						err := services.StartQuiz(roomId, cfg.WebexAccessToken)
						if err != nil {
							fmt.Println("Error starting the quiz:", err)
							services.SendMessageToWebex(roomId, "I couldn't start the quiz. Please try again.", cfg.WebexAccessToken)
							c.JSON(500, gin.H{"status": "failed to start quiz"})
							return
						}

					default:
						// Continue the quiz if any other action is selected
						err := services.ContinueQuiz(roomId, cfg.WebexAccessToken, actionData.Data["action"])
						if err != nil {
							fmt.Println("Error continuing the quiz:", err)
							services.SendMessageToWebex(roomId, "I couldn't proceed with the quiz. Please try again.", cfg.WebexAccessToken)
							c.JSON(500, gin.H{"status": "failed to continue quiz"})
							return
						}
					}
					c.JSON(200, gin.H{"status": "received"})
					return
				} else {
					c.JSON(400, gin.H{"status": "bad request", "reason": "invalid card action"})
					return
				}
			}

			// Handle non-card text messages via Vertex AI
			messageText, err := services.GetMessageContent(messageId, cfg.WebexAccessToken)
			if err != nil {
				services.SendMessageToWebex(roomId, "I'm sorry, I couldn't retrieve the message.", cfg.WebexAccessToken)
				c.JSON(500, gin.H{"status": "failed to get message content"})
				return
			}

			vertexResponse, err := services.GenerateVertexAIResponse(messageText, cfg)
			if err != nil {
				fmt.Println("Error generating Vertex AI response:", err)
				services.SendMessageToWebex(roomId, "I'm sorry, I couldn't generate a response.", cfg.WebexAccessToken)
				c.JSON(500, gin.H{"status": "failed to generate Vertex AI response"})
				return
			}

			// Send Vertex AI response back to Webex
			err = services.SendMessageToWebex(roomId, vertexResponse, cfg.WebexAccessToken)
			if err != nil {
				c.JSON(500, gin.H{"status": "failed to send message"})
				return
			}
		}

		c.JSON(200, gin.H{"status": "received"})
	} else {
		c.JSON(400, gin.H{"status": "bad request"})
	}
}
