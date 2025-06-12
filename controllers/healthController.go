package controllers

import (
	"net/http"
	"tfidf-app/helper"
	"tfidf-app/version"

	"github.com/gin-gonic/gin"
)

type HealthController interface {
	GetStatus(c *gin.Context)
	GetVersion(c *gin.Context)
}

type healthController struct{}

func NewHealthController() HealthController {
	return &healthController{}
}

// GetStatus godoc
// @Summary Get API status
// @Description Provides the current status of the API
// @Tags Health
// @Produce json
// @Success 200 {object} helper.Response{data=object{status=string}} "API status"
// @Router /status [get]
func (h *healthController) GetStatus(c *gin.Context) {
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
func (h *healthController) GetVersion(c *gin.Context) {
	c.JSON(http.StatusOK, helper.NewSuccessResponse(gin.H{
		"version": version.AppVersion,
	}))
}
