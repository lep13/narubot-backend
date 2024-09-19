package controllers

import (
    "fmt"
    "narubot-backend/services"
    "narubot-backend/config"
    "github.com/gin-gonic/gin"
)

func HandleWebhook(c *gin.Context) {
    var payload map[string]interface{}
    if err := c.BindJSON(&payload); err == nil {
        fmt.Printf("Received payload: %+v\n", payload)  // Log the entire payload

        // Check if the resource is "messages" and event is "created"
        resource := payload["resource"].(string)
        event := payload["event"].(string)

        if resource != "messages" || event != "created" {
            fmt.Println("Not a message creation event, ignoring.")
            c.JSON(200, gin.H{"status": "ignored"})
            return
        }

        // Extract relevant data
        data := payload["data"].(map[string]interface{})
        personEmail := data["personEmail"].(string)
        roomId := data["roomId"].(string)

        // Check if the "text" field exists in the message payload
        text, ok := data["text"].(string)
        if !ok {
            fmt.Println("No text field in the message payload.")
            c.JSON(200, gin.H{"status": "no text found"})
            return
        }

        // Load configuration (secrets)
        cfg := config.LoadConfig()

        // Ensure the bot is not responding to its own messages
        if personEmail != cfg.BotEmail {
            // Call ChatGPT API using the message as input
            gptResponse, err := services.GetGPTResponse("Respond as Naruto bot: "+text, cfg.GPTAPIKey)
            if err != nil {
                fmt.Println("Error getting GPT response:", err)
                services.SendMessageToWebex(roomId, "Oops! Something went wrong.", cfg.WebexAccessToken) // Include access token
            } else {
                services.SendMessageToWebex(roomId, gptResponse, cfg.WebexAccessToken) // Include access token
            }
        }

        c.JSON(200, gin.H{"status": "received"})
    } else {
        c.JSON(400, gin.H{"status": "bad request"})
    }
}
