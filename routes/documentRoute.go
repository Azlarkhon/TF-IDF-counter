package routes

import (
	"tfidf-app/controllers"
	"tfidf-app/database"
	"tfidf-app/middleware"

	"github.com/gin-gonic/gin"
)

func DocumentRoute(r *gin.Engine) {
	documentController := controllers.NewDocumentController(database.DB)

	protected := r.Group("/documents")
	protected.Use(middleware.AuthMiddleware)
	{
		protected.GET("/", documentController.GetDocuments)
		protected.GET("/:document_id", documentController.GetDocumentByID)
		protected.DELETE("/:document_id", documentController.DeleteDocument)
		protected.GET("/:document_id/statistics", documentController.GetDocumentStatistics)
		protected.GET("/:document_id/huffman", documentController.GetDocumentHuffman)
	}
}
