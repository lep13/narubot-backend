package controllers

import (
	"log"
	"strings"

	"github.com/lep13/narubot-backend/config"
	"github.com/lep13/narubot-backend/models"
	"github.com/lep13/narubot-backend/services"

	"github.com/gin-gonic/gin"
)

// HandleWebhook processes Webex messages and interactive card actions
func HandleWebhook(c *gin.Context) {
	var payload map[string]interface{}
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(400, gin.H{"status": "bad request", "reason": "invalid JSON payload"})
		return
	}

	// Check the resource type to differentiate message and attachmentActions
	resource, resourceOk := payload["resource"].(string)
	if !resourceOk {
		c.JSON(400, gin.H{"status": "bad request", "reason": "resource field missing"})
		return
	}
	// log.Printf("Received payload: %+v\n", payload)
	log.Printf("Full Payload: %+v\n", payload)

	data, ok := payload["data"].(map[string]interface{})
	if !ok {
		c.JSON(400, gin.H{"status": "bad request", "reason": "no data field"})
		return
	}

	log.Printf("Payload Data Content: %+v\n", data)

	personEmail, emailOk := data["personEmail"].(string)
	roomId, roomOk := data["roomId"].(string)
	messageId, messageOk := data["id"].(string)

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		c.JSON(500, gin.H{"status": "failed to load config"})
		return
	}

	// Ensure the message is not from the bot itself
	if personEmail != cfg.BotEmail {
		// Handle card submission via "attachmentActions"
		if resource == "attachmentActions" {
			if !emailOk || !roomOk || !messageOk {
				c.JSON(400, gin.H{"status": "bad request", "reason": "required fields missing for attachmentActions"})
				return
			}

			// Extract the attachment actions data
			attachmentData, actionOk := data["inputs"].(map[string]interface{})
			if !actionOk {
				c.JSON(400, gin.H{"status": "bad request", "reason": "inputs field missing in data"})
				return
			}

			// Extract the action field
			action, actionExists := attachmentData["action"].(string)
			if !actionExists {
				c.JSON(400, gin.H{"status": "bad request", "reason": "action field missing in inputs"})
				return
			}

			// Handle the action based on the extracted value
			switch action {
			case "AskQuestion":
				vertexResponse, err := services.GenerateVertexAIResponse("Ask me a question", cfg)
				if err != nil {
					services.SendMessageToWebex(roomId, "I'm sorry, I couldn't generate a response.", cfg.WebexAccessToken)
					c.JSON(500, gin.H{"status": "failed to generate Vertex AI response"})
					return
				}

				err = services.SendMessageToWebex(roomId, vertexResponse, cfg.WebexAccessToken)
				if err != nil {
					c.JSON(500, gin.H{"status": "failed to send message"})
					return
				}

			case "StartQuiz":
				err := services.StartQuiz(roomId, cfg.WebexAccessToken)
				if err != nil {
					services.SendMessageToWebex(roomId, "I couldn't start the quiz. Please try again.", cfg.WebexAccessToken)
					c.JSON(500, gin.H{"status": "failed to start quiz"})
					return
				}

			default:
				c.JSON(400, gin.H{"status": "bad request", "reason": "unknown action"})
				return
			}
			c.JSON(200, gin.H{"status": "received"})
			return
		}

		// Handle non-card text messages
		if resource == "messages" {
			if !emailOk || !roomOk || !messageOk {
				c.JSON(400, gin.H{"status": "bad request", "reason": "required fields missing for messages"})
				return
			}

			messageText, err := services.GetMessageContent(messageId, cfg.WebexAccessToken)
			if err != nil {
				services.SendMessageToWebex(roomId, "I'm sorry, I couldn't retrieve the message.", cfg.WebexAccessToken)
				c.JSON(500, gin.H{"status": "failed to get message content"})
				return
			}

			// If the message is a greeting, send a card with two options
			if isGreeting(messageText) {
				card, err := models.CreateGreetingCard()
				if err != nil {
					services.SendMessageToWebex(roomId, "I'm sorry, I couldn't create the card.", cfg.WebexAccessToken)
					c.JSON(500, gin.H{"status": "failed to create card"})
					return
				}

				err = services.SendMessageWithCard(roomId, card, cfg.WebexAccessToken)
				if err != nil {
					c.JSON(500, gin.H{"status": "failed to send card"})
					return
				}
				c.JSON(200, gin.H{"status": "card sent"})
				return
			} else {
				vertexResponse, err := services.GenerateVertexAIResponse(messageText, cfg)
				if err != nil {
					services.SendMessageToWebex(roomId, "I'm sorry, I couldn't generate a response.", cfg.WebexAccessToken)
					c.JSON(500, gin.H{"status": "failed to generate Vertex AI response"})
					return
				}

				err = services.SendMessageToWebex(roomId, vertexResponse, cfg.WebexAccessToken)
				if err != nil {
					c.JSON(500, gin.H{"status": "failed to send message"})
					return
				}
			}
		}
		c.JSON(200, gin.H{"status": "received"})
	} else {
		log.Println("Ignoring message from bot or missing email field")
		c.JSON(200, gin.H{"status": "ignored"})
	}
}

// Helper function to check if the message is a greeting
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

// HandleWebhookTestCard - FOR CHECKING POST
// func HandleWebhookTestCard(c *gin.Context) {
//     var payload map[string]interface{}

//     // Parse the incoming JSON payload
//     if err := c.BindJSON(&payload); err != nil {
//         c.JSON(400, gin.H{"status": "bad request", "reason": "failed to parse JSON payload"})
//         return
//     }

//     // Log the incoming payload for debugging
//     log.Printf("Received Test Payload: %+v\n", payload)

//     // Check the resource type to avoid unnecessary responses
//     resource, resourceOk := payload["resource"].(string)
//     if !resourceOk {
//         c.JSON(400, gin.H{"status": "bad request", "reason": "resource field missing"})
//         return
//     }

//     // Only respond to messages or attachment actions
//     if resource != "messages" && resource != "attachmentActions" {
//         log.Println("Ignoring message from bot or invalid resource type.")
//         c.JSON(200, gin.H{"status": "ignored"})
//         return
//     }

//     // Define a single test card with actions
//     testCard := map[string]interface{}{
//         "contentType": "application/vnd.microsoft.card.adaptive",
//         "content": map[string]interface{}{
//             "type":    "AdaptiveCard",
//             "version": "1.3",
//             "$schema": "http://adaptivecards.io/schemas/adaptive-card.json",
//             "body": []map[string]interface{}{
//                 {
//                     "type":   "TextBlock",
//                     "text":   "Test card for debugging actions.",
//                     "weight": "Bolder",
//                     "size":   "Medium",
//                 },
//             },
//             "actions": []map[string]interface{}{
//                 {
//                     "type":  "Action.Submit",
//                     "title": "Test Action 1",
//                     "data": map[string]interface{}{
//                         "inputs": map[string]string{
//                             "action": "TestAction1",
//                         },
//                     },
//                 },
//                 {
//                     "type":  "Action.Submit",
//                     "title": "Test Action 2",
//                     "data": map[string]interface{}{
//                         "inputs": map[string]string{
//                             "action": "TestAction2",
//                         },
//                     },
//                 },
//             },
//         },
//     }

// 	roomId := "Y2lzY29zcGFyazovL3VybjpURUFNOnVzLXdlc3QtMl9yL1JPT00vOTc2MWViMTAtNzY3Ni0xMWVmLWExYTctNzEzNjI0YmM1MDJk"
// 	accessToken := "OGUyZjA1NzMtNTA2MS00MWUxLWJjNjAtYjlhNTlkMDkxYTUwNGZmOGZhNmItZTRl_P0A1_955b0110-97b0-4717-a284-c57ffbc138a4"

//     // Send the card once
//     err := services.SendCardToRoom(roomId, testCard, accessToken)
//     if err != nil {
//         c.JSON(500, gin.H{"status": "error", "reason": err.Error()})
//         return
//     }

//     // Respond with success after sending one card
//     c.JSON(200, gin.H{"status": "test card sent"})
// }
