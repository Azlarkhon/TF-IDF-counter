package routes

import (
	"tfidf-app/controllers"
	"tfidf-app/middleware"

	"github.com/gin-gonic/gin"
)

func UserRoutes(r *gin.Engine) {
	userGroup := r.Group("/users")

	userGroup.POST("/login", controllers.Login)
	userGroup.POST("/register", controllers.Register)
	userGroup.GET("/logout", controllers.Logout)

	userGroup.Use(middleware.AuthMiddleware)
	{
		userGroup.GET("/:user_id", controllers.GetUserByID)
		userGroup.PATCH("/:user_id", controllers.UpdateUser)
		userGroup.DELETE("/:user_id", controllers.DeleteUser)
	}
}
