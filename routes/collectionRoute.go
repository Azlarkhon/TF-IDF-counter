package routes

import (
	"tfidf-app/controllers"
	"tfidf-app/database"
	"tfidf-app/middleware"

	"github.com/gin-gonic/gin"
)

func CollectionRoute(r *gin.Engine) {
	collectionController := controllers.NewCollectionController(database.DB)

	protected := r.Group("/collections")
	protected.Use(middleware.AuthMiddleware)
	{
		protected.POST("/", collectionController.CreateCollection)
		protected.GET("/", collectionController.GetCollections)
		protected.GET("/:collection_id", collectionController.GetCollectionByID)
		protected.PUT("/:collection_id", collectionController.UpdateCollection)
		protected.DELETE("/:collection_id", collectionController.DeleteCollection)
		protected.POST("/:collection_id/:document_id", collectionController.AddDocumentToCollection)
		protected.DELETE("/:collection_id/:document_id", collectionController.DeleteDocumentFromCollection)
		protected.POST("/add-many", collectionController.AddDocumentToCollections)
		protected.GET("/:collection_id/statistics", collectionController.GetCollectionStatistics)
	}
}
