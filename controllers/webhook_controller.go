package controllers

import (
    "fmt"
    "narubot-backend/services"
    "github.com/gin-gonic/gin"
    "os"
)

func HandleWebhook(c *gin.Context) {
    var payload map[string]interface{}
    if err := c.BindJSON(&payload); err == nil {
        data := payload["data"].(map[string]interface{})
        personEmail := data["personEmail"].(string)
        roomId := data["roomId"].(string)
        text := data["text"].(string)

        if personEmail != os.Getenv("BOT_EMAIL") {
            gptResponse, err := services.GetGPTResponse("Respond to this as a Naruto bot: "+text, os.Getenv("GPT_API_KEY"))
            if err != nil {
                fmt.Println("Error getting GPT response:", err)
            }

            err = services.SendMessageToWebex(roomId, gptResponse)
            if err != nil {
                fmt.Println("Error sending message:", err)
            }
        }
        c.JSON(200, gin.H{"status": "received"})
    } else {
        c.JSON(400, gin.H{"status": "bad request"})
    }
}
