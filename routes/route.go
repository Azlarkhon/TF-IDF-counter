package routes

import (
	"tfidf-app/controllers"

	"github.com/gin-gonic/gin"
)

func Routes(r *gin.Engine) {
    r.GET("/", controllers.ShowUploadForm)
    r.POST("/upload", controllers.HandleFileUpload)

	r.GET("/status", controllers.GetStatus)
	r.GET("/version", controllers.GetVersion)
}
