package main

import (
    "log"
    "github.com/lep13/narubot-backend/db"
    "github.com/lep13/narubot-backend/config"
    "github.com/lep13/narubot-backend/router"
)

func main() {
    // Load configuration with MongoDB URI and other settings
    config, err := config.LoadConfig()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // Initialize MongoDB connection
    err = db.InitializeMongoDB(config)
    if err != nil {
        log.Fatalf("Failed to initialize MongoDB: %v", err)
    }

    // Set up and run the router
    r := router.SetupRouter()
    r.Run(":8081")
}
