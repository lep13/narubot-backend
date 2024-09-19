package router

import (
    "narubot-backend/controllers"
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

    return r
}
