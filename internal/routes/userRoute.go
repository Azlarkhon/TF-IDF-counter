package routes

import (
	"tfidf-app/internal/controllers"
	"tfidf-app/internal/database"
	"tfidf-app/internal/middleware"

	"github.com/gin-gonic/gin"
)

func UserRoutes(r *gin.Engine) {
	userController := controllers.NewUserController(database.DB)

	userGroup := r.Group("/users")

	userGroup.POST("/login", userController.Login)
	userGroup.POST("/register", userController.Register)
	userGroup.GET("/logout", userController.Logout)

	userGroup.Use(middleware.AuthMiddleware)
	{
		userGroup.GET("/me", userController.GetMe)
		userGroup.PATCH("/:user_id", userController.UpdateUser)
		userGroup.DELETE("/:user_id", userController.DeleteUser)
	}
}
