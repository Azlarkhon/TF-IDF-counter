package controllers

import (
	"net/http"
	"os"
	"tfidf-app/database"
	"tfidf-app/dto"
	"tfidf-app/helper"
	"tfidf-app/models"

	"github.com/gin-gonic/gin"
)

func GetDocuments(c *gin.Context) {
	userID, err := helper.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, helper.NewErrorResponse("You are not authorized"))
		return
	}

	_, authorized := helper.CheckAuthenticationAndAuthorization(c, userID)
	if !authorized {
		return
	}

	var documents []models.Document
	if err := database.DB.Where("user_id = ?", userID).Find(&documents).Error; err != nil {
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Failed to fetch documents"))
		return
	}

	c.JSON(http.StatusOK, helper.NewSuccessResponse(documents))
}

func GetDocumentByID(c *gin.Context) {
	userID, err := helper.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, helper.NewErrorResponse("You are not authorized"))
		return
	}

	_, authorized := helper.CheckAuthenticationAndAuthorization(c, userID)
	if !authorized {
		return
	}

	docID := c.Param("document_id")
	if docID == "" {
		c.JSON(http.StatusBadRequest, helper.NewErrorResponse("Document ID is required"))
		return
	}

	var document models.Document
	if err := database.DB.First(&document, docID).Error; err != nil {
		c.JSON(http.StatusNotFound, helper.NewErrorResponse("Document not found"))
		return
	}

	if document.UserID != userID {
		c.JSON(http.StatusForbidden, helper.NewErrorResponse("Access denied"))
		return
	}

	content, err := os.ReadFile(document.FilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Failed to read document content"))
		return
	}

	response := dto.DocumentResponse{
		ID:          document.ID,
		Name:        document.Name,
		Content:     string(content),
		UplodadedAt: document.CreatedAt,
	}

	c.JSON(http.StatusOK, helper.NewSuccessResponse(response))
}
