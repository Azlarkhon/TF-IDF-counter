package controllers

import (
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"tfidf-app/database"
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

func HandleFileUpload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.String(http.StatusBadRequest, "Bad request: %s", err.Error())
		return
	}

	err = os.MkdirAll("./samples", os.ModePerm)
	if err != nil {
		c.String(http.StatusInternalServerError, "Cannot create folder: %s", err.Error())
		return
	}

	filePath := "./samples/" + file.Filename
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.String(http.StatusInternalServerError, "Cannot save file: %s", err.Error())
		return
	}

	words, err := processFile(filePath)
	if err != nil {
		c.String(http.StatusInternalServerError, "Cannot process file: %s", err.Error())
		return
	}

	stats := services.ComputeTFIDF(words)

	if len(stats) > 50 {
		stats = stats[:50]
	}

	// Работа с метрикой
	var metric models.Metric
	result := database.DB.First(&metric)
	currentTime := time.Now()

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			metric = models.Metric{
				FilesProcessed:               1,
				LatestFileProcessedTimestamp: currentTime,
			}
			if err := database.DB.Create(&metric).Error; err != nil {
				c.String(http.StatusInternalServerError, "Database error: %s", err.Error())
				return
			}
		} else {
			c.String(http.StatusInternalServerError, "Database error: %s", result.Error.Error())
			return
		}
	} else {
		metric.FilesProcessed++
		metric.LatestFileProcessedTimestamp = currentTime
		if err := database.DB.Save(&metric).Error; err != nil {
			c.String(http.StatusInternalServerError, "Database error: %s", err.Error())
			return
		}
	}

	c.HTML(http.StatusOK, "index.tmpl", gin.H{
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
