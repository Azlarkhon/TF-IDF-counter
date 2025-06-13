package routes

import (
	"tfidf-app/controllers"
	"tfidf-app/database"

	"github.com/gin-gonic/gin"
)

func MetricsRoute(r *gin.Engine) {
	metricsController := controllers.NewMetricsController(database.DB)

	r.GET("/metrics", metricsController.GetMetrics)
}
