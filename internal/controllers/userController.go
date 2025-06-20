package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"tfidf-app/internal/dto"
	"tfidf-app/internal/helper"
	"tfidf-app/internal/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserController interface {
	GetMe(c *gin.Context)
	Register(c *gin.Context)
	Login(c *gin.Context)
	Logout(c *gin.Context)
	UpdateUser(c *gin.Context)
	DeleteUser(c *gin.Context)
}

type userController struct {
	DB *gorm.DB
}

func NewUserController(db *gorm.DB) UserController {
	return &userController{DB: db}
}

// GetMe godoc
// @Summary Get information about the current user
// @Tags Users
// @Success 200 {object} helper.Response{data=models.User}
// @Failure 401 {object} helper.Response
// @Failure 404 {object} helper.Response
// @Failure 500 {object} helper.Response
// @Router /users/me [get]
func (u *userController) GetMe(c *gin.Context) {
	userID, err := helper.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, helper.NewErrorResponse("You are not authorized"))
		return
	}

	var me models.User

	result := u.DB.Where("id = ?", userID).First(&me)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, helper.NewErrorResponse("You are not authorized"))
		} else {
			c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Database error"))
		}
		return
	}

	c.JSON(http.StatusOK, helper.NewSuccessResponse(me))
}

// Register godoc
// @Summary New user registration
// @Tags Users
// @Accept json
// @Produce json
// @Param user body dto.RegisterUserRequest true "Регистрационные данные"
// @Success 200 {object} helper.Response{data=models.User}
// @Failure 400 {object} helper.Response
// @Failure 500 {object} helper.Response
// @Router /users/register [post]
func (u *userController) Register(c *gin.Context) {
	var req dto.RegisterUserRequest

	err := c.BindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, helper.NewErrorResponse("Invalid input"))
		return
	}

	var existing models.User
	result := u.DB.Where("email = ?", req.Email).First(&existing)
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

	result = u.DB.Create(&newUser)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Failed to create user"))
		return
	}

	c.JSON(http.StatusOK, helper.NewSuccessResponse(newUser))
}

// Login godoc
// @Summary User login and setting JWT cookie
// @Tags Users
// @Accept json
// @Produce json
// @Param credentials body dto.LoginRequest true "Данные для входа"
// @Success 200 {object} helper.Response
// @Failure 400 {object} helper.Response
// @Failure 401 {object} helper.Response
// @Failure 500 {object} helper.Response
// @Router /users/login [post]
func (u *userController) Login(c *gin.Context) {
	var req dto.LoginRequest
	err := c.BindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, helper.NewErrorResponse("Invalid input"))
		return
	}

	var user models.User
	result := u.DB.Where("email = ?", req.Email).First(&user)
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
		false,        // Secure (HTTPS-only)
		true,        // HTTP-Only (защита от XSS)
	)

	c.JSON(http.StatusOK, helper.NewSuccessResponse(gin.H{
		"id":    user.ID,
		"token": token,
	}))
}

// Logout godoc
// @Summary User logout (delete cookies)
// @Tags Users
// @Produce json
// @Success 200 {object} helper.Response
// @Router /users/logout [get]
func (u *userController) Logout(c *gin.Context) {
	c.SetCookie(
		"auth_token",
		"",
		-1, // Удаляем куку
		"/",
		"",
		false,
		true,
	)
	c.JSON(http.StatusOK, helper.NewSuccessResponse("You have successfully logged out"))
}

// UpdateUser godoc
// @Summary Update user password
// @Tags Users
// @Accept json
// @Produce json
// @Param user_id path int true "ID пользователя"
// @Param update body dto.UpdateUserRequest true "Новый пароль"
// @Success 200 {object} helper.Response
// @Failure 400 {object} helper.Response
// @Failure 401 {object} helper.Response
// @Failure 500 {object} helper.Response
// @Router /users/{user_id} [patch]
func (u *userController) UpdateUser(c *gin.Context) {
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

	result := u.DB.Model(&models.User{}).Where("id = ?", id).Update("password", string(hashedPassword))
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Failed to update password"))
		return
	}

	c.JSON(http.StatusOK, helper.NewSuccessResponse("Password updated successfully"))
}

// DeleteUser godoc
// @Summary Delete user (everything related)
// @Tags Users
// @Produce json
// @Param user_id path int true "ID пользователя"
// @Success 200 {object} helper.Response
// @Failure 400 {object} helper.Response
// @Failure 401 {object} helper.Response
// @Failure 500 {object} helper.Response
// @Router /users/{user_id} [delete]
func (u *userController) DeleteUser(c *gin.Context) {
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

	// Удаление папки пользователя со всеми файлами
	userDir := fmt.Sprintf("/app/documents/user_%d", id)
	if err := os.RemoveAll(userDir); err != nil && !os.IsNotExist(err) {
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Failed to delete user's folder: "+err.Error()))
		return
	}

	// Удаление пользователя из базы данных
	result := u.DB.Delete(&models.User{}, id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Failed to delete user"))
		return
	}

	c.SetCookie("auth_token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, helper.NewSuccessResponse("User deleted and logged out successfully"))
}
