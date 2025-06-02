package main

import (
	"log"

	"tfidf-app/config"
	"tfidf-app/database"
	"tfidf-app/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	database.ConnectDatabase()

	router := gin.Default()

	router.LoadHTMLGlob("templates/*")

	routes.Routes(router)
	routes.UserRoutes(router)

	if err := router.Run(":" + config.Init.Port); err != nil {
		log.Fatal("Failed to run server: ", err)
	}
}
