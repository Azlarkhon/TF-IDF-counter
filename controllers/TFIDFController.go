package controllers

import (
	"net/http"
	"os"
	"regexp"
	"strings"

	"tfidf-app/services"

	"github.com/gin-gonic/gin"
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

	err = os.MkdirAll("./sample", os.ModePerm)
	if err != nil {
		c.String(http.StatusInternalServerError, "Cannot create folder: %s", err.Error())
		return
	}

	filePath := "./sample/" + file.Filename
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
