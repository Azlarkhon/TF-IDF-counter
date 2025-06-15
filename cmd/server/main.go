package main

import (
	"log"

	_ "tfidf-app/docs"
	"tfidf-app/internal/config"
	"tfidf-app/internal/database"
	"tfidf-app/internal/middleware"
	"tfidf-app/internal/routes"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title TF-IDF counter API
// @version 1.2.0
// @description API for document processing using TF-IDF algorithm
// @BasePath /

// @tag.name Upload document
// @tag.name Users
// @tag.name Collections
// @tag.name Documents
// @tag.name Metrics
// @tag.name Health
func main() {
	database.ConnectDatabase()

	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()
	router.Use(middleware.CorsMiddleware)

	routes.UploadRoute(router)
	routes.HealthRoute(router)
	routes.MetricsRoute(router)
	routes.UserRoutes(router)
	routes.DocumentRoute(router)
	routes.CollectionRoute(router)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	if err := router.Run(":" + config.Init.Port); err != nil {
		log.Fatal("Failed to run server: ", err)
	}
}
