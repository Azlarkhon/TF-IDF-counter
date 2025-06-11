package routes

import (
	"tfidf-app/controllers"
	"tfidf-app/middleware"

	"github.com/gin-gonic/gin"
)

func MetricsRoute(r *gin.Engine) {
	r.GET("/", controllers.ShowUploadForm)

	r.GET("/status", controllers.GetStatus)
	r.GET("/version", controllers.GetVersion)
	r.GET("/metrics", controllers.GetMetrics)

	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware)
	{
		protected.POST("/upload", controllers.HandleFileUpload)
	}
}
