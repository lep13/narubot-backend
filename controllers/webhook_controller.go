package controllers

import (
	"fmt"
	"strings"

	"github.com/lep13/narubot-backend/config"
	"github.com/lep13/narubot-backend/services"
	"github.com/lep13/narubot-backend/models"

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

			// Handle card submission via "attachmentActions"
			if attachmentActions, ok := payload["attachmentActions"].(map[string]interface{}); ok {
				actionData, err := services.ParseCardActionUsingAttachment(attachmentActions)
				if err == nil {
					switch actionData.Data["action"] {
					case "AskQuestion":
						// Forward the message to the Vertex AI chatbot for questions
						vertexResponse, err := services.GenerateVertexAIResponse("Ask me a question", cfg)
						if err != nil {
							fmt.Println("Error generating Vertex AI response:", err)
							services.SendMessageToWebex(roomId, "I'm sorry, I couldn't generate a response.", cfg.WebexAccessToken)
							c.JSON(500, gin.H{"status": "failed to generate Vertex AI response"})
							return
						}

						// Send the Vertex AI response back to Webex
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

			// Handle non-card text messages (like hello, hey, hi) and send a card in response
			messageText, err := services.GetMessageContent(messageId, cfg.WebexAccessToken)
			if err != nil {
				services.SendMessageToWebex(roomId, "I'm sorry, I couldn't retrieve the message.", cfg.WebexAccessToken)
				c.JSON(500, gin.H{"status": "failed to get message content"})
				return
			}

			// If the message is a greeting, send a card with two options (quiz or question)
			if isGreeting(messageText) {
				card, err := models.CreateGreetingCard()
				if err != nil {
					fmt.Println("Error creating greeting card:", err)
					services.SendMessageToWebex(roomId, "I'm sorry, I couldn't create the card.", cfg.WebexAccessToken)
					c.JSON(500, gin.H{"status": "failed to create card"})
					return
				}

				err = services.SendMessageWithCard(roomId, card, cfg.WebexAccessToken)
				if err != nil {
					fmt.Println("Error sending card:", err)
					c.JSON(500, gin.H{"status": "failed to send card"})
					return
				}
				c.JSON(200, gin.H{"status": "card sent"})
				return
			} else {
				// If it's not a greeting, use Vertex AI to generate a response
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
		}

		c.JSON(200, gin.H{"status": "received"})
	} else {
		c.JSON(400, gin.H{"status": "bad request"})
	}
}

// for greeting card
func isGreeting(message string) bool {
	greetings := []string{"hello", "hey", "hi", "question", "quiz"}
	message = strings.ToLower(message)
	for _, greeting := range greetings {
		if message == greeting {
			return true
		}
	}
	return false
}

// func HandleWebhookTestCard(c *gin.Context) {
//     // Test by sending simple card when any webhook is received
//     roomId := "Y2lzY29zcGFyazovL3VybjpURUFNOnVzLXdlc3QtMl9yL1JPT00vOTc2MWViMTAtNzY3Ni0xMWVmLWExYTctNzEzNjI0YmM1MDJk"
//     accessToken := "OGUyZjA1NzMtNTA2MS00MWUxLWJjNjAtYjlhNTlkMDkxYTUwNGZmOGZhNmItZTRl_P0A1_955b0110-97b0-4717-a284-c57ffbc138a4" 

//     card := map[string]interface{}{
//         "type": "AdaptiveCard",
//         "body": []interface{}{
//             map[string]interface{}{
//                 "type":  "TextBlock",
//                 "text":  "This is a test card!",
//                 "weight": "Bolder",
//                 "size":  "Medium",
//             },
//         },
//         "actions": []interface{}{
//             map[string]interface{}{
//                 "type":  "Action.Submit",
//                 "title": "Test Button",
//                 "data":  map[string]string{"action": "TestAction"},
//             },
//         },
//         "$schema":  "http://adaptivecards.io/schemas/adaptive-card.json",
//         "version":  "1.2",
//     }

//     // Send the card to Webex
//     err := services.SendMessageWithCard(roomId, card, accessToken)
//     if err != nil {
//         fmt.Println("Error sending test card:", err)
//         c.JSON(500, gin.H{"status": "error", "message": err.Error()})
//         return
//     }

//     c.JSON(200, gin.H{"status": "success", "message": "Test card sent successfully!"})
// }
