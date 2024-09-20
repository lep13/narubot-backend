package controllers

import (
	"fmt"
	"narubot-backend/config"
	"narubot-backend/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HandleWebhook(c *gin.Context) {
	var payload map[string]interface{}
	if err := c.BindJSON(&payload); err == nil {
		fmt.Printf("Received payload: %+v\n", payload)

		// Check if the resource is "messages" and event is "created"
		resource, ok := payload["resource"].(string)
		if !ok || resource != "messages" {
			fmt.Println("Not a message resource, ignoring.")
			c.JSON(http.StatusOK, gin.H{"status": "ignored"})
			return
		}

		event, ok := payload["event"].(string)
		if !ok || event != "created" {
			fmt.Println("Not a message creation event, ignoring.")
			c.JSON(http.StatusOK, gin.H{"status": "ignored"})
			return
		}

		// Extract relevant data safely
		data, ok := payload["data"].(map[string]interface{})
		if !ok {
			fmt.Println("No data field in the webhook payload.")
			c.JSON(http.StatusBadRequest, gin.H{"status": "bad request", "reason": "no data field"})
			return
		}

		personEmail, ok := data["personEmail"].(string)
		roomId, roomOk := data["roomId"].(string)
		messageId, messageOk := data["id"].(string)

		if !ok || !roomOk || !messageOk {
			fmt.Println("Required fields are missing in the webhook payload.")
			c.JSON(http.StatusBadRequest, gin.H{"status": "bad request", "reason": "required fields missing"})
			return
		}

		// Load configuration (secrets)
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Println("Error loading configuration:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"status": "failed to load configuration"})
			return
		}

		// Ensure the bot is not responding to its own messages
		if personEmail != cfg.BotEmail {
			// Retrieve the message content from Webex using messageId
			messageText, err := services.GetMessageContent(messageId, cfg.WebexAccessToken)
			if err != nil {
				fmt.Println("Error fetching message content:", err)
				// Fallback message
				err = services.SendMessageToWebex(roomId, "I'm sorry, I couldn't retrieve the message content.", cfg.WebexAccessToken)
				if err != nil {
					fmt.Println("Error sending fallback message:", err)
				}
				c.JSON(http.StatusInternalServerError, gin.H{"status": "failed to get message content"})
				return
			}

			// Log the messageText received
			fmt.Printf("Message Text: %s\n", messageText)

			// Respond with a simple message for debugging
			err = services.SendMessageToWebex(roomId, "Narubot received your message!", cfg.WebexAccessToken)
			if err != nil {
				fmt.Println("Error sending message:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"status": "failed to send message"})
				return
			}
			fmt.Println("Message successfully sent!")
		}

		c.JSON(http.StatusOK, gin.H{"status": "received"})
	} else {
		fmt.Println("Invalid webhook request format")
		c.JSON(http.StatusBadRequest, gin.H{"status": "bad request"})
	}
}
