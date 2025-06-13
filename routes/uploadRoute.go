package routes

import (
	"tfidf-app/controllers"
	"tfidf-app/middleware"

	"github.com/gin-gonic/gin"
)

func UploadRoute(r *gin.Engine) {
	r.GET("/", controllers.ShowUploadForm)

	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware)
	{
		protected.POST("/upload", controllers.HandleFileUpload)
	}
}
