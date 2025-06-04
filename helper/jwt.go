package helper

import (
	"errors"
	"net/http"
	"strconv"
	"tfidf-app/config"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

type CustomClaims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

func GenerateJWT(userID int) (string, error) {
	claims := CustomClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)), // 1 месяц
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   strconv.Itoa(userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.Init.JWTSecret))
}

func VerifyJWT(tokenString string) (*jwt.Token, *CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(config.Init.JWTSecret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, nil, errors.New("token has expired")
		}
		return nil, nil, errors.New("failed to parse token: " + err.Error())
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, nil, errors.New("invalid token claims")
	}

	return token, claims, nil
}

func CheckAuthenticationAndAuthorization(c *gin.Context, userID int) (int, bool) {
	loggedInUserIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, NewErrorResponse("User is not authenticated"))
		return 0, false
	}

	loggedInUserID, ok := loggedInUserIDVal.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, NewErrorResponse("Invalid user ID type in context"))
		return 0, false
	}

	if loggedInUserID != userID {
		c.JSON(http.StatusForbidden, NewErrorResponse("You are not authorized to access this resource"))
		return 0, false
	}

	return loggedInUserID, true
}

func GetUserIDFromContext(c *gin.Context) (int, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, errors.New("user_id not found in context")
	}

	id, ok := userID.(int)
	if !ok {
		return 0, errors.New("invalid user_id type in context")
	}

	return id, nil
}
