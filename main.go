package main

import (
    "narubot-backend/router"
)

func main() {
    r := router.SetupRouter()
    r.Run(":8081")
}
