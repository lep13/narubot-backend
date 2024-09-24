package controllers

import (
	"narubot-backend/config"
	"narubot-backend/services"

	"github.com/gin-gonic/gin"
)

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

		if personEmail != cfg.BotEmail {
			// Retrieve the message content from Webex
			messageText, err := services.GetMessageContent(messageId, cfg.WebexAccessToken)
			if err != nil {
				services.SendMessageToWebex(roomId, "I'm sorry, I couldn't retrieve the message.", cfg.WebexAccessToken)
				c.JSON(500, gin.H{"status": "failed to get message content"})
				return
			}

			// Get GenAI response
			genAIResponse, err := services.GetGenAIResponse(messageText)
			if err != nil {
				services.SendMessageToWebex(roomId, "I'm sorry, I couldn't generate a response.", cfg.WebexAccessToken)
				c.JSON(500, gin.H{"status": "failed to get GenAI response"})
				return
			}

			// Send GenAI response back to Webex
			err = services.SendMessageToWebex(roomId, genAIResponse, cfg.WebexAccessToken)
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
