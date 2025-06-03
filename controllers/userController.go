package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"tfidf-app/database"
	"tfidf-app/dto"
	"tfidf-app/helper"
	"tfidf-app/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func GetUserByID(c *gin.Context) {
	idStr := c.Param("user_id")
	id, err := strconv.Atoi(idStr)
	fmt.Println(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, helper.NewErrorResponse("Invalid user id format"))
		return
	}

	_, authorized := helper.CheckAuthenticationAndAuthorization(c, id)
	if !authorized {
		return
	}

	var newUser models.User

	result := database.DB.Where("id = ?", id).First(&newUser)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, helper.NewErrorResponse("User not found"))
		} else {
			c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Database error"))
		}
		return
	}

	c.JSON(http.StatusOK, helper.NewSuccessResponse(newUser))
}

func Register(c *gin.Context) {
	var req dto.RegisterUserRequest

	err := c.BindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, helper.NewErrorResponse("Invalid input"))
		return
	}

	var existing models.User
	result := database.DB.Where("email = ?", req.Email).First(&existing)
	if result.RowsAffected > 0 {
		c.JSON(http.StatusBadRequest, helper.NewErrorResponse("Email already registered"))
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Failed to hash password"))
		return
	}

	newUser := models.User{
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	result = database.DB.Create(&newUser)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Failed to create user"))
		return
	}

	c.JSON(http.StatusOK, helper.NewSuccessResponse(newUser))
}

func Login(c *gin.Context) {
	var req dto.LoginRequest
	err := c.BindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, helper.NewErrorResponse("Invalid input"))
		return
	}

	var user models.User
	result := database.DB.Where("email = ?", req.Email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, helper.NewErrorResponse("Invalid email or password"))
		} else {
			c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Database error"))
		}
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, helper.NewErrorResponse("Invalid email or password"))
		return
	}

	token, err := helper.GenerateJWT(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Failed to generate token"))
		return
	}

	c.SetCookie(
		"auth_token", // Имя куки
		token,        // Значение (JWT)
		2592000,      // 30 дней в секундах,
		"/",          // Путь
		"",           // Домен (localhost)
		true,         // Secure (HTTPS-only)
		true,         // HTTP-Only (защита от XSS)
	)

	c.JSON(http.StatusOK, gin.H{
		"id":         user.ID,
		"is_success": true,
	})
}

func Logout(c *gin.Context) {
	c.SetCookie(
		"auth_token",
		"",
		-1, // Удаляем куку
		"/",
		"",
		false,
		true,
	)
	c.JSON(http.StatusOK, helper.NewSuccessResponse(nil))
}

func UpdateUser(c *gin.Context) {
	idStr := c.Param("user_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, helper.NewErrorResponse("Invalid user id format"))
		return
	}

	_, authorized := helper.CheckAuthenticationAndAuthorization(c, id)
	if !authorized {
		return
	}

	var req dto.UpdateUserRequest

	if err := c.BindJSON(&req); err != nil || req.NewPassword == "" {
		c.JSON(http.StatusBadRequest, helper.NewErrorResponse("Invalid or missing new password"))
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Failed to hash password"))
		return
	}

	result := database.DB.Model(&models.User{}).Where("id = ?", id).Update("password", string(hashedPassword))
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Failed to update password"))
		return
	}

	c.JSON(http.StatusOK, helper.NewSuccessResponse("Password updated successfully"))
}

func DeleteUser(c *gin.Context) {
	idStr := c.Param("user_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, helper.NewErrorResponse("Invalid user id format"))
		return
	}

	_, authorized := helper.CheckAuthenticationAndAuthorization(c, id)
	if !authorized {
		return
	}

	result := database.DB.Delete(&models.User{}, id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Failed to delete user"))
		return
	}

	Logout(c)
}
