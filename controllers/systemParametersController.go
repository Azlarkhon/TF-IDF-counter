package controllers

import (
	"errors"
	"net/http"
	"tfidf-app/database"
	"tfidf-app/helper"
	"tfidf-app/models"
	"tfidf-app/version"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetStatus(c *gin.Context) {
	c.JSON(http.StatusOK, helper.NewSuccessResponse(gin.H{
		"status": "OK",
	}))
}

func GetVersion(c *gin.Context) {
	c.JSON(http.StatusOK, helper.NewSuccessResponse(gin.H{
		"version": version.AppVersion,
	}))
}

func GetMetrics(c *gin.Context) {
	var metrics models.Metric

	result := database.DB.First(&metrics)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, helper.NewErrorResponse("Metrics not found"))
		} else {
			c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Failed to retrieve metrics"))
		}
		return
	}

	c.JSON(http.StatusOK, helper.NewSuccessResponse(metrics))
}
