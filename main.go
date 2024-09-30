package main

import (
    "github.com/lep13/narubot-backend/router"
)

func main() {
    r := router.SetupRouter()
    r.Run(":8081")
}
