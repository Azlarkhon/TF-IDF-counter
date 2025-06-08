package routes

import (
	"tfidf-app/controllers"
	"tfidf-app/middleware"

	"github.com/gin-gonic/gin"
)

func CollectionRoute(r *gin.Engine) {
	protected := r.Group("/collections")
	protected.Use(middleware.AuthMiddleware)
	{
		protected.POST("/", controllers.CreateCollection)
		protected.GET("/", controllers.GetCollections)
		protected.GET("/:collection_id", controllers.GetCollectionByID)
		protected.PUT("/:collection_id", controllers.UpdateCollection)
		protected.DELETE("/:collection_id", controllers.DeleteCollection)
		protected.POST("/:collection_id/:document_id", controllers.AddDocumentToCollection)
		protected.DELETE("/:collection_id/:document_id", controllers.DeleteDocumentFromCollection)
		protected.POST("/add-many", controllers.AddDocumentToCollections)
		protected.GET("/:collection_id/statistics", controllers.GetCollectionStatistics)
	}
}
