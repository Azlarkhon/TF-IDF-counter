package middleware

import (
	"net/http"
	"tfidf-app/helper"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(c *gin.Context) {
	// Получаем токен из куки
	token, err := c.Cookie("auth_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, helper.NewErrorResponse("Missing auth cookie"))
		c.Abort()
		return
	}

	// Проверяем JWT
	_, claims, err := helper.VerifyJWT(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, helper.NewErrorResponse(err.Error()))
		c.Abort()
		return
	}

	c.Set("user_id", claims.UserID)
	c.Next()
}
