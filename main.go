package main

import (
	"log"

	"tfidf-app/config"
	"tfidf-app/database"
	_ "tfidf-app/docs"
	"tfidf-app/routes"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title TF-IDF counter API
// @version 1.2.0
// @description API for document processing using TF-IDF algorithm
// @host localhost
// @BasePath /
func main() {
	database.ConnectDatabase()

	router := gin.Default()

	router.LoadHTMLGlob("templates/*")

	routes.Routes(router)
	routes.UserRoutes(router)
	routes.DocumentRoute(router)
	routes.CollectionRoute(router)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	if err := router.Run(":" + config.Init.Port); err != nil {
		log.Fatal("Failed to run server: ", err)
	}
}
