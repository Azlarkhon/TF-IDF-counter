package controllers

import (
	"errors"
	"net/http"
	"tfidf-app/internal/helper"
	"tfidf-app/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MetricsController interface {
	GetMetrics(c *gin.Context)
}

// metricsController implements MetricsController
type metricsController struct {
	DB *gorm.DB
}

// NewMetricsController returns an instance of MetricsController
func NewMetricsController(db *gorm.DB) MetricsController {
	return &metricsController{DB: db}
}

// GetMetrics godoc
// @Summary Get application metrics
// @Description Retrieves aggregated metrics including processing time, file size, and top 10 most seen words
// @Tags Metrics
// @Produce json
// @Success 200 {object} helper.Response{data=models.Metric} "Application metrics"
// @Failure 404 {object} helper.Response
// @Failure 500 {object} helper.Response
// @Router /metrics [get]
func (m *metricsController) GetMetrics(c *gin.Context) {
	var metric models.Metric

	result := m.DB.
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
