package routes

import (
	"tfidf-app/controllers"
	"tfidf-app/database"
	"tfidf-app/middleware"

	"github.com/gin-gonic/gin"
)

func MetricsRoute(r *gin.Engine) {
	metricsController := controllers.NewMetricsController(database.DB)

	r.GET("/", controllers.ShowUploadForm)
	r.GET("/metrics", metricsController.GetMetrics)

	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware)
	{
		protected.POST("/upload", controllers.HandleFileUpload)
	}
}
