package controllers

import (
	"net/http"
	"tfidf-app/database"
	"tfidf-app/dto"
	"tfidf-app/helper"
	"tfidf-app/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateCollection(c *gin.Context) {
	userID, err := helper.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, helper.NewErrorResponse("You are not authorized"))
		return
	}

	_, authorized := helper.CheckAuthenticationAndAuthorization(c, userID)
	if !authorized {
		return
	}

	var req dto.CreateCollectionReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, helper.NewErrorResponse("Invalid input"))
		return
	}

	collection := models.Collection{
		Name:   req.Name,
		UserID: userID,
	}

	if err := database.DB.Create(&collection).Error; err != nil {
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Failed to create collection"))
		return
	}

	c.JSON(http.StatusCreated, helper.NewSuccessResponse(collection))
}

func GetCollections(c *gin.Context) {
	userID, err := helper.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, helper.NewErrorResponse("You are not authorized"))
		return
	}

	_, authorized := helper.CheckAuthenticationAndAuthorization(c, userID)
	if !authorized {
		return
	}

	var collections []models.Collection
	if err := database.DB.Where("user_id = ?", userID).Find(&collections).Error; err != nil {
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Failed to get collections"))
		return
	}

	c.JSON(http.StatusOK, helper.NewSuccessResponse(collections))
}

func GetCollectionByID(c *gin.Context) {
	userID, err := helper.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, helper.NewErrorResponse("You are not authorized"))
		return
	}

	_, authorized := helper.CheckAuthenticationAndAuthorization(c, userID)
	if !authorized {
		return
	}

	collectionID := c.Param("collection_id")
	if collectionID == "" {
		c.JSON(http.StatusBadRequest, helper.NewErrorResponse("Collection ID is required"))
		return
	}

	var collection models.Collection
	if err := database.DB.First(&collection, collectionID).Error; err != nil {
		c.JSON(http.StatusNotFound, helper.NewErrorResponse("Collection not found"))
		return
	}

	c.JSON(http.StatusOK, helper.NewSuccessResponse(collection))
}

func UpdateCollection(c *gin.Context) {
	userID, err := helper.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, helper.NewErrorResponse("You are not authorized"))
		return
	}

	_, authorized := helper.CheckAuthenticationAndAuthorization(c, userID)
	if !authorized {
		return
	}

	collectionID := c.Param("collection_id")
	if collectionID == "" {
		c.JSON(http.StatusBadRequest, helper.NewErrorResponse("Collection ID is required"))
		return
	}

	var req dto.UpdateCollectionReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, helper.NewErrorResponse("Invalid input"))
		return
	}

	var collection models.Collection
	if err := database.DB.Where("id = ? AND user_id = ?", collectionID, userID).First(&collection).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, helper.NewErrorResponse("Collection not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Failed to get collection"))
		return
	}

	collection.Name = req.Name
	if err := database.DB.Save(&collection).Error; err != nil {
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Failed to update collection"))
		return
	}

	c.JSON(http.StatusOK, helper.NewSuccessResponse(collection))
}

func DeleteCollection(c *gin.Context) {
	userID, err := helper.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, helper.NewErrorResponse("You are not authorized"))
		return
	}

	_, authorized := helper.CheckAuthenticationAndAuthorization(c, userID)
	if !authorized {
		return
	}

	collectionID := c.Param("collection_id")
	if collectionID == "" {
		c.JSON(http.StatusBadRequest, helper.NewErrorResponse("Collection ID is required"))
		return
	}

	var collection models.Collection
	if err := database.DB.Where("id = ? AND user_id = ?", collectionID, userID).First(&collection).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, helper.NewErrorResponse("Collection not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Failed to get collection"))
		return
	}

	if err := database.DB.Delete(&collection).Error; err != nil {
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Failed to delete collection"))
		return
	}

	c.JSON(http.StatusOK, helper.NewSuccessResponse(nil))
}
