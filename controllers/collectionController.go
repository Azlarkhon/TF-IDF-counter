package controllers

import (
	"math"
	"net/http"
	"tfidf-app/dto"
	"tfidf-app/helper"
	"tfidf-app/models"
	"tfidf-app/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CollectionController interface {
	CreateCollection(c *gin.Context)
	AddDocumentToCollection(c *gin.Context)
	AddDocumentToCollections(c *gin.Context)
	DeleteDocumentFromCollection(c *gin.Context)
	GetCollections(c *gin.Context)
	GetCollectionByID(c *gin.Context)
	UpdateCollection(c *gin.Context)
	DeleteCollection(c *gin.Context)
	GetCollectionStatistics(c *gin.Context)
}

type collectionController struct {
	DB *gorm.DB
}

func NewCollectionController(db *gorm.DB) CollectionController {
	return &collectionController{DB: db}
}

// CreateCollection godoc
// @Summary Create a new collection
// @Description Creates a new document collection
// @Tags Collections
// @Accept json
// @Produce json
// @Param collection body dto.CreateCollectionReq true "Collection details"
// @Success 201 {object} helper.Response{data=models.Collection} "Created collection"
// @Failure 400 {object} helper.Response
// @Failure 401 {object} helper.Response
// @Failure 500 {object} helper.Response
// @Router /collections [post]
func (col *collectionController) CreateCollection(c *gin.Context) {
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

	if err := col.DB.Create(&collection).Error; err != nil {
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Failed to create collection"))
		return
	}

	c.JSON(http.StatusCreated, helper.NewSuccessResponse(collection))
}

// AddDocumentToCollection godoc
// @Summary Add document to collection
// @Description Adds an existing document to a collection
// @Tags Collections
// @Produce json
// @Param collection_id path string true "Collection ID"
// @Param document_id path string true "Document ID"
// @Success 200 {object} helper.Response
// @Failure 400 {object} helper.Response
// @Failure 401 {object} helper.Response
// @Failure 404 {object} helper.Response
// @Failure 500 {object} helper.Response
// @Router /collections/{collection_id}/{document_id} [post]
func (col *collectionController) AddDocumentToCollection(c *gin.Context) {
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
	if err := col.DB.
		Where("id = ? AND user_id = ?", collectionID, userID).First(&collection).Error; err != nil {
		c.JSON(http.StatusNotFound, helper.NewErrorResponse("Collection not found"))
		return
	}

	var document models.Document
	if err := col.DB.
		Where("id = ? AND user_id = ?", documentID, userID).First(&document).Error; err != nil {
		c.JSON(http.StatusNotFound, helper.NewErrorResponse("Document not found"))
		return
	}

	// Добавление документа в коллекцию через связь many2many
	if err := col.DB.Model(&collection).Association("Documents").Append(&document); err != nil {
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Failed to add document to collection"))
		return
	}

	c.JSON(http.StatusOK, helper.NewSuccessResponse("Document successfully added to collection"))
}

// AddDocumentToCollections godoc
// @Summary Add document to multiple collections
// @Description Adds a document to several collections at once
// @Tags Collections
// @Accept json
// @Produce json
// @Param request body dto.AddDocumentToCollectionsReq true "Collection IDs"
// @Success 200 {object} helper.Response
// @Failure 400 {object} helper.Response
// @Failure 401 {object} helper.Response
// @Failure 404 {object} helper.Response
// @Failure 500 {object} helper.Response
// @Router /collections/add-many [post]
func (col *collectionController) AddDocumentToCollections(c *gin.Context) {
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
	if err := col.DB.Where("id = ? AND user_id = ?", req.DocumentID, userID).First(&document).Error; err != nil {
		c.JSON(http.StatusNotFound, helper.NewErrorResponse("Document not found"))
		return
	}

	err = col.DB.Transaction(func(tx *gorm.DB) error {
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

// DeleteDocumentFromCollection godoc
// @Summary Remove document from collection
// @Description Removes a document from a collection
// @Tags Collections
// @Produce json
// @Param collection_id path string true "Collection ID"
// @Param document_id path string true "Document ID"
// @Success 200 {object} helper.Response
// @Failure 400 {object} helper.Response
// @Failure 401 {object} helper.Response
// @Failure 404 {object} helper.Response
// @Failure 500 {object} helper.Response
// @Router /collections/{collection_id}/{document_id} [delete]
func (col *collectionController) DeleteDocumentFromCollection(c *gin.Context) {
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
	if err := col.DB.Where("id = ? AND user_id = ?", collectionID, userID).First(&collection).Error; err != nil {
		c.JSON(http.StatusNotFound, helper.NewErrorResponse("Collection not found"))
		return
	}

	var document models.Document
	if err := col.DB.Where("id = ? AND user_id = ?", documentID, userID).First(&document).Error; err != nil {
		c.JSON(http.StatusNotFound, helper.NewErrorResponse("Document not found"))
		return
	}

	if err := col.DB.Model(&collection).Association("Documents").Delete(&document); err != nil {
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Failed to remove document from collection"))
		return
	}

	c.JSON(http.StatusOK, helper.NewSuccessResponse("Document successfully deleted from collection"))
}

// GetCollections godoc
// @Summary Get all collections
// @Description Returns all collections belonging to the authenticated user
// @Tags Collections
// @Produce json
// @Success 200 {object} helper.Response{data=[]models.Collection} "List of collections"
// @Failure 401 {object} helper.Response
// @Failure 500 {object} helper.Response
// @Router /collections [get]
func (col *collectionController) GetCollections(c *gin.Context) {
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
	if err := col.DB.Preload("Documents").Where("user_id = ?", userID).Find(&collections).Error; err != nil {
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Failed to get collections"))
		return
	}

	c.JSON(http.StatusOK, helper.NewSuccessResponse(collections))
}

// GetCollectionByID godoc
// @Summary Get collection by ID
// @Description Returns a specific collection with its documents
// @Tags Collections
// @Produce json
// @Param collection_id path string true "Collection ID"
// @Success 200 {object} helper.Response{data=models.Collection} "Collection details"
// @Failure 400 {object} helper.Response
// @Failure 401 {object} helper.Response
// @Failure 404 {object} helper.Response
// @Failure 500 {object} helper.Response
// @Router /collections/{collection_id} [get]
func (col *collectionController) GetCollectionByID(c *gin.Context) {
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
	if err := col.DB.Preload("Documents").First(&collection, collectionID).Error; err != nil {
		c.JSON(http.StatusNotFound, helper.NewErrorResponse("Collection not found"))
		return
	}

	c.JSON(http.StatusOK, helper.NewSuccessResponse(collection))
}

// UpdateCollection godoc
// @Summary Update collection
// @Description Updates collection name
// @Tags Collections
// @Accept json
// @Produce json
// @Param collection_id path string true "Collection ID"
// @Param collection body dto.UpdateCollectionReq true "New collection name"
// @Success 200 {object} helper.Response{data=models.Collection} "Updated collection"
// @Failure 400 {object} helper.Response
// @Failure 401 {object} helper.Response
// @Failure 404 {object} helper.Response
// @Failure 500 {object} helper.Response
// @Router /collections/{collection_id} [put]
func (col *collectionController) UpdateCollection(c *gin.Context) {
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
	if err := col.DB.Where("id = ? AND user_id = ?", collectionID, userID).First(&collection).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, helper.NewErrorResponse("Collection not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Failed to get collection"))
		return
	}

	collection.Name = req.Name
	if err := col.DB.Save(&collection).Error; err != nil {
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Failed to update collection"))
		return
	}

	c.JSON(http.StatusOK, helper.NewSuccessResponse(collection))
}

// DeleteCollection godoc
// @Summary Delete collection
// @Description Deletes a collection (does not delete documents)
// @Tags Collections
// @Produce json
// @Param collection_id path string true "Collection ID"
// @Success 200 {object} helper.Response
// @Failure 400 {object} helper.Response
// @Failure 401 {object} helper.Response
// @Failure 404 {object} helper.Response
// @Failure 500 {object} helper.Response
// @Router /collections/{collection_id} [delete]
func (col *collectionController) DeleteCollection(c *gin.Context) {
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
	if err := col.DB.Where("id = ? AND user_id = ?", collectionID, userID).First(&collection).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, helper.NewErrorResponse("Collection not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Failed to get collection"))
		return
	}

	if err := col.DB.Delete(&collection).Error; err != nil {
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Failed to delete collection"))
		return
	}

	c.JSON(http.StatusOK, helper.NewSuccessResponse("Collection deleted successfully"))
}

// GetCollectionStatistics godoc
// @Summary Get collection statistics
// @Description Gets statistics for the collection: TF is calculated as if all documents in the collection were one document, IDF unchanged (gives top 50 most frequent words and their idf)
// @Tags Collections
// @Produce json
// @Param collection_id path string true "Collection ID"
// @Success 200 {object} helper.Response{data=object} "Collection statistics"
// @Failure 400 {object} helper.Response
// @Failure 401 {object} helper.Response
// @Failure 404 {object} helper.Response
// @Failure 500 {object} helper.Response
// @Router /collections/{collection_id}/statistics [get]
func (col *collectionController) GetCollectionStatistics(c *gin.Context) {
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
	if err := col.DB.Preload("Documents").Where("id = ? AND user_id = ?", collectionID, userID).First(&collection).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, helper.NewErrorResponse("Collection not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Failed to get collection"))
		return
	}

	var collectionDocuments []map[string]int
	var tfWords []string
	for _, doc := range collection.Documents {
		words, err := services.ProcessFile(doc.FilePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Cannot process file: "+err.Error()))
			return
		}
		collectionDocuments = append(collectionDocuments, services.CountWords(words))
		tfWords = append(tfWords, words...)
	}

	wordCount := services.CountWords(tfWords)
	tf := services.CalculateTF(wordCount, len(tfWords))

	idf := services.CalculateIDF(collectionDocuments)

	rareWords := services.GetRarestWords(wordCount, 50)

	for i := range rareWords {
		rareWords[i].TF = tf[rareWords[i].Word]
		if idfValue, exists := idf[rareWords[i].Word]; exists {
			rareWords[i].IDF = idfValue
		} else {
			rareWords[i].IDF = math.Log(float64(len(collectionDocuments) + 1))
		}
		rareWords[i].Count = wordCount[rareWords[i].Word]
	}

	c.JSON(http.StatusOK, helper.NewSuccessResponse(gin.H{
		"statistics": rareWords,
		"meta": gin.H{
			"total_documents": len(collection.Documents),
		},
	}))
}
