package main

import (
    "log"

    "github.com/gin-gonic/gin"
    "tfidf-app/routes"
)

func main() {
    router := gin.Default()

    router.LoadHTMLGlob("templates/*")

    routes.Routes(router)
	
    if err := router.Run(":8080"); err != nil {
        log.Fatal("Failed to run server: ", err)
    }
}
