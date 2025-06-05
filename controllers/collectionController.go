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

func AddDocumentToCollection(c *gin.Context) {
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
	documentID := c.Param("document_id")
	if collectionID == "" || documentID == "" {
		c.JSON(http.StatusBadRequest, helper.NewErrorResponse("Collection ID and Document ID are required"))
		return
	}

	var collection models.Collection
	if err := database.DB.
		Where("id = ? AND user_id = ?", collectionID, userID).First(&collection).Error; err != nil {
		c.JSON(http.StatusNotFound, helper.NewErrorResponse("Collection not found"))
		return
	}

	var document models.Document
	if err := database.DB.
		Where("id = ? AND user_id = ?", documentID, userID).First(&document).Error; err != nil {
		c.JSON(http.StatusNotFound, helper.NewErrorResponse("Document not found"))
		return
	}

	// Добавление документа в коллекцию через связь many2many
	if err := database.DB.Model(&collection).Association("Documents").Append(&document); err != nil {
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Failed to add document to collection"))
		return
	}

	c.JSON(http.StatusOK, helper.NewSuccessResponse("Document successfully added to collection"))
}

func AddDocumentToCollections(c *gin.Context) {
	userID, err := helper.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, helper.NewErrorResponse("You are not authorized"))
		return
	}

	_, authorized := helper.CheckAuthenticationAndAuthorization(c, userID)
	if !authorized {
		return
	}

	var req dto.AddDocumentToCollectionsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, helper.NewErrorResponse("Invalid input"))
		return
	}

	if len(req.CollectionIDs) == 0 {
		c.JSON(http.StatusBadRequest, helper.NewErrorResponse("Collection IDs are required"))
		return
	}

	var document models.Document
	if err := database.DB.Where("id = ? AND user_id = ?", req.DocumentID, userID).First(&document).Error; err != nil {
		c.JSON(http.StatusNotFound, helper.NewErrorResponse("Document not found"))
		return
	}

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		for _, collectionID := range req.CollectionIDs {
			var collection models.Collection
			if err := tx.Where("id = ? AND user_id = ?", collectionID, userID).First(&collection).Error; err != nil {
				return err
			}

			if err := tx.Model(&collection).Association("Documents").Append(&document); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Failed to add document to one or more collections"))
		return
	}

	c.JSON(http.StatusOK, helper.NewSuccessResponse("Document successfully added to all specified collections"))
}

func DeleteDocumentFromCollection(c *gin.Context) {
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
	documentID := c.Param("document_id")
	if collectionID == "" || documentID == "" {
		c.JSON(http.StatusBadRequest, helper.NewErrorResponse("Collection ID and Document ID are required"))
		return
	}

	var collection models.Collection
	if err := database.DB.Where("id = ? AND user_id = ?", collectionID, userID).First(&collection).Error; err != nil {
		c.JSON(http.StatusNotFound, helper.NewErrorResponse("Collection not found"))
		return
	}

	var document models.Document
	if err := database.DB.Where("id = ? AND user_id = ?", documentID, userID).First(&document).Error; err != nil {
		c.JSON(http.StatusNotFound, helper.NewErrorResponse("Document not found"))
		return
	}

	if err := database.DB.Model(&collection).Association("Documents").Delete(&document); err != nil {
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Failed to remove document from collection"))
		return
	}

	c.JSON(http.StatusOK, helper.NewSuccessResponse(nil))
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
	if err := database.DB.Preload("Documents").Where("user_id = ?", userID).Find(&collections).Error; err != nil {
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
	if err := database.DB.Preload("Documents").First(&collection, collectionID).Error; err != nil {
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
