package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"tfidf-app/database"
	"tfidf-app/helper"
	"tfidf-app/models"
	"tfidf-app/services"

	"github.com/gin-gonic/gin"
)

func ShowUploadForm(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"words": nil,
	})
}

func HandleFileUpload(c *gin.Context) {
	startTime := time.Now()

	userID, err := helper.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, helper.NewErrorResponse("Unauthorized"))
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, helper.NewErrorResponse("No file provided: "+err.Error()))
		return
	}

	// Путь до папки пользователя
	userDir := fmt.Sprintf("documents/user_%d", userID)
	fullPath := filepath.Join("/app", userDir)

	// Создаём папку, если не существует
	if err := os.MkdirAll(fullPath, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Cannot create user folder: "+err.Error()))
		return
	}

	// Сохраняем файл
	filePath := filepath.Join(fullPath, file.Filename)
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Cannot save file: "+err.Error()))
		return
	}

	// Сохраняем метаинформацию в БД
	document := models.Document{
		Name:     file.Filename,
		FilePath: filePath,
		UserID:   userID,
	}

	if err := database.DB.Create(&document).Error; err != nil {
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Failed to save document: "+err.Error()))
		return
	}

	words, err := processFile(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Cannot process file: "+err.Error()))
		return
	}

	processingTime := services.CalculateProcessingTime(startTime)
	fileSizeMB := services.RoundFileSizeMB(file.Size)

	stats := services.ComputeTFIDF(words)

	metric, err := services.UpdateMetrics(processingTime, fileSizeMB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Database error: "+err.Error()))
		return
	}

	err = services.SaveWords(words, metric.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Database error saving words: "+err.Error()))
		return
	}

	if len(stats) > 50 {
		stats = stats[:50]
	}

	c.JSON(http.StatusOK, gin.H{
		"words": stats,
	})
}

func processFile(filePath string) ([]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	text := string(data)
	text = strings.ToLower(text)

	reg := regexp.MustCompile(`[^a-zA-Zа-яА-Я]+`)
	cleaned := reg.ReplaceAllString(text, " ")

	words := strings.Fields(cleaned)
	return words, nil
}
