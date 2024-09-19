package router

import (
    "narubot-backend/controllers"
    "github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
    r := gin.Default()

    r.GET("/", func(c *gin.Context) {
        c.String(200, "Welcome to Narubot!")
    })

    r.POST("/webhook", controllers.HandleWebhook)

    return r
}
