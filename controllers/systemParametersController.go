package controllers

import (
	"net/http"
	"tfidf-app/helper"
	"tfidf-app/version"

	"github.com/gin-gonic/gin"
)

func GetStatus(c *gin.Context) {
	c.JSON(http.StatusOK, helper.NewSuccessResponse(gin.H{
		"status" : "OK",
	}))
}

func GetVersion(c *gin.Context) {
	c.JSON(http.StatusOK, helper.NewSuccessResponse(gin.H{
		"version" : version.AppVersion,
	}))
}
