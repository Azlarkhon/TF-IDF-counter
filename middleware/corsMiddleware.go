package middleware

import (
	"github.com/gin-gonic/gin"
)

func CorsMiddleware(c *gin.Context) {
	origin := c.Request.Header.Get("Origin")

	allowedOrigins := map[string]bool{
		"http://37.9.53.45:8080": true,
		"http://localhost:8080":  true,
	}

	if allowedOrigins[origin] {
		c.Header("Access-Control-Allow-Origin", origin)
	}

	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
	c.Header("Access-Control-Allow-Credentials", "true")

	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(200)
		return
	}

	c.Next()
}
