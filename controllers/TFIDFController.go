package controllers

import (
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"tfidf-app/helper"
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

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, helper.NewErrorResponse("Bad request: " + err.Error()))
		return
	}

	err = os.MkdirAll("./samples", os.ModePerm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Cannot create folder: " + err.Error()))
		return
	}

	filePath := "./samples/" + file.Filename
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Cannot save file: " + err.Error()))
		return
	}

	words, err := processFile(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Cannot process file: " + err.Error()))
		return
	}

	processingTime := services.CalculateProcessingTime(startTime)
	fileSizeMB := services.RoundFileSizeMB(file.Size)

	stats := services.ComputeTFIDF(words)

	metric, err := services.UpdateMetrics(processingTime, fileSizeMB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Database error: " + err.Error()))
		return
	}

	err = services.SaveWords(words, metric.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Database error saving words: " + err.Error()))
		return
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
