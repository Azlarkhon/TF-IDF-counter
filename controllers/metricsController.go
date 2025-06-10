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

// GetStatus godoc
// @Summary Get API status
// @Description Provides the current status of the API
// @Tags Health
// @Produce json
// @Success 200 {object} helper.Response{data=object{status=string}} "API status"
// @Router /status [get]
func GetStatus(c *gin.Context) {
	c.JSON(http.StatusOK, helper.NewSuccessResponse(gin.H{
		"status": "OK",
	}))
}

// GetVersion godoc
// @Summary Get application version
// @Description Provides the current version of the application
// @Tags Health
// @Produce json
// @Success 200 {object} helper.Response{data=object{version=string}} "Application version"
// @Router /version [get]
func GetVersion(c *gin.Context) {
	c.JSON(http.StatusOK, helper.NewSuccessResponse(gin.H{
		"version": version.AppVersion,
	}))
}

// GetMetrics godoc
// @Summary Get application metrics
// @Description Retrieves aggregated metrics including processing time, file size, and top words.
// @Tags Metrics
// @Produce json
// @Success 200 {object} helper.Response{data=models.Metric} "Application metrics"
// @Failure 404 {object} helper.Response
// @Failure 500 {object} helper.Response
// @Router /metrics [get]
func GetMetrics(c *gin.Context) {
	var metric models.Metric

	result := database.DB.
		Preload("Words", func(db *gorm.DB) *gorm.DB {
			return db.Order("count DESC").Limit(10)
		}).
		First(&metric)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, helper.NewErrorResponse("Metrics not found"))
		} else {
			c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Failed to retrieve metrics"))
		}
		return
	}

	c.JSON(http.StatusOK, helper.NewSuccessResponse(metric))
}
