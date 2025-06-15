package routes

import (
	"tfidf-app/internal/controllers"
	"tfidf-app/internal/middleware"

	"github.com/gin-gonic/gin"
)

func UploadRoute(r *gin.Engine) {
	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware)
	{
		protected.POST("/upload", controllers.HandleFileUpload)
	}
}
