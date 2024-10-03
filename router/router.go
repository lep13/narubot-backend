package router

import (
    "github.com/lep13/narubot-backend/controllers"
    "github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
    r := gin.Default()

    // Root endpoint for testing
    r.GET("/", func(c *gin.Context) {
        c.String(200, "Welcome to Narubot!")
    })

    // Webhook endpoint
    r.POST("/webhook", controllers.HandleWebhook)
    // r.POST("/webhook", controllers.HandleWebhookTestCard) //for test card

    return r
}
