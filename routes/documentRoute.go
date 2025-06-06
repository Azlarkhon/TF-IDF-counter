package routes

import (
	"tfidf-app/controllers"
	"tfidf-app/middleware"

	"github.com/gin-gonic/gin"
)

func DocumentRoute(r *gin.Engine) {
	protected := r.Group("/documents")
	protected.Use(middleware.AuthMiddleware)
	{
		protected.GET("/", controllers.GetDocuments)
		protected.GET("/:document_id", controllers.GetDocumentByID)
		protected.DELETE("/:document_id", controllers.DeleteDocument)
		protected.GET("/:document_id/statistics", controllers.GetDocumentStatistics)
	}
}
