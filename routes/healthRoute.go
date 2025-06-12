package routes

import (
	"tfidf-app/controllers"

	"github.com/gin-gonic/gin"
)

func HealthRoute(r *gin.Engine) {
	healthController := controllers.NewHealthController()

	r.GET("/status", healthController.GetStatus)
	r.GET("/version", healthController.GetVersion)
}
