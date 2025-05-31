package controllers

import (
	"math"
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
	startTime := time.Now()

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

	processingTime := services.CalculateProcessingTime(startTime)
	fileSizeMB := services.RoundFileSizeMB(file.Size)

	stats := services.ComputeTFIDF(words)

	// Работа с метрикой
	var metric models.Metric
	result := database.DB.Preload("Words").First(&metric)
	currentTime := time.Now()

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			metric = models.Metric{
				FilesProcessed:               1,
				LatestFileProcessedTimestamp: currentTime,
				MinTimeProcessed:             processingTime,
				AvgTimeProcessed:             processingTime,
				MaxTimeProcessed:             processingTime,
				TotalFileSizeMB:              fileSizeMB,
				AvgFileSizeMB:                fileSizeMB,
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
		// нью мин и макс
		newMin := math.Min(metric.MinTimeProcessed, processingTime)
		newMax := math.Max(metric.MaxTimeProcessed, processingTime)

		// нью эвередж
		totalTime := metric.AvgTimeProcessed * float64(metric.FilesProcessed)
		newAvg := (totalTime + processingTime) / float64(metric.FilesProcessed+1)
		newAvg = math.Round(newAvg*1000) / 1000

		metric.FilesProcessed++
		metric.LatestFileProcessedTimestamp = currentTime
		metric.MinTimeProcessed = newMin
		metric.AvgTimeProcessed = newAvg
		metric.MaxTimeProcessed = newMax
		metric.TotalFileSizeMB += fileSizeMB
		metric.AvgFileSizeMB = metric.TotalFileSizeMB / float64(metric.FilesProcessed)

		if err := database.DB.Save(&metric).Error; err != nil {
			c.String(http.StatusInternalServerError, "Database error: %s", err.Error())
			return
		}
	}

	wordCount := make(map[string]int)
	for _, w := range words {
		wordCount[w]++
	}

	wordList := make([]string, 0, len(wordCount))
	for word := range wordCount {
		wordList = append(wordList, word)
	}

	var existingWords []models.Word
	if err := database.DB.
		Where("word IN ? AND metric_id = ?", wordList, metric.ID).
		Find(&existingWords).Error; err != nil {
		c.String(http.StatusInternalServerError, "Database error finding existing words: %s", err.Error())
		return
	}

	// Юзаю хэшмэп чтобы найти слова быстрее
	existingWordMap := make(map[string]*models.Word)
	for i := range existingWords {
		existingWordMap[existingWords[i].Word] = &existingWords[i]
	}

	newWords := make([]models.Word, 0)
	for word, count := range wordCount {
		if existing, found := existingWordMap[word]; found {
			existing.Count += count
		} else {
			newWords = append(newWords, models.Word{
				MetricID:  metric.ID,
				Word:      word,
				Count:     count,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			})
		}
	}

	if len(existingWords) > 0 {
		if err := database.DB.Save(&existingWords).Error; err != nil {
			c.String(http.StatusInternalServerError, "Database error updating words: %s", err.Error())
			return
		}
	}

	if len(newWords) > 0 {
		if err := database.DB.Create(&newWords).Error; err != nil {
			c.String(http.StatusInternalServerError, "Database error creating words: %s", err.Error())
			return
		}
	}

	if len(stats) > 50 {
		stats = stats[:50]
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
