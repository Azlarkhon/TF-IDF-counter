package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"tfidf-app/database"
	"tfidf-app/helper"
	"tfidf-app/models"
	"tfidf-app/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ShowUploadForm(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"words": nil,
	})
}

// HandleFileUpload godoc
// @Summary Upload and process a document
// @Description Uploads a file, processes it for TF-IDF, and saves to database
// @Tags Documents
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Document file to upload"
// @Success 200 {object} helper.Response{data=[]services.WordStat} "TF-IDF statistics"
// @Failure 400 {object} helper.Response
// @Failure 401 {object} helper.Response
// @Failure 409 {object} helper.Response
// @Failure 500 {object} helper.Response
// @Router /upload [post]
func HandleFileUpload(c *gin.Context) {
	startTime := time.Now()

	userID, err := helper.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, helper.NewErrorResponse("Unauthorized"))
		return
	}

	_, authorized := helper.CheckAuthenticationAndAuthorization(c, userID)
	if !authorized {
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, helper.NewErrorResponse("No file provided: "+err.Error()))
		return
	}

	// Проверка: существует ли уже документ с таким именем у этого пользователя
	var existingDoc models.Document
	if err := database.DB.
		Where("name = ? AND user_id = ?", file.Filename, userID).First(&existingDoc).Error; err == nil {
		c.JSON(http.StatusConflict, helper.NewErrorResponse("Document with the same name already exists"))
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

	// Обработка файла (вынесено до транзакции, так как это CPU-bound операция)
	words, err := services.ProcessFile(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Cannot process file: "+err.Error()))
		return
	}

	// Начинаем транзакцию для всех операций с БД
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		// Сохраняем метаинформацию в БД
		document := models.Document{
			Name:     file.Filename,
			FilePath: filePath,
			UserID:   userID,
		}
		if err := tx.Create(&document).Error; err != nil {
			return fmt.Errorf("failed to save document: %w", err)
		}

		// Вычисляем метрики
		processingTime := services.CalculateProcessingTime(startTime)
		fileSizeMB := services.RoundFileSizeMB(file.Size)

		// Обновляем метрики
		metric, err := services.UpdateMetrics(tx, processingTime, fileSizeMB)
		if err != nil {
			return fmt.Errorf("failed to update metrics: %w", err)
		}

		// Сохраняем слова
		if err := services.SaveWords(tx, words, metric.ID); err != nil {
			return fmt.Errorf("failed to save words: %w", err)
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Database error: "+err.Error()))
		return
	}

	// TF-IDF (не требует транзакции, так как это вычисление)
	stats := services.ComputeTFIDFForUpload(words)
	if len(stats) > 50 {
		stats = stats[:50]
	}

	c.JSON(http.StatusOK, gin.H{
		"words": stats,
	})
}
