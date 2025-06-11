package controllers

import (
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"tfidf-app/database"
	"tfidf-app/dto"
	"tfidf-app/helper"
	"tfidf-app/models"
	"tfidf-app/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetDocuments godoc
// @Summary Get all user documents
// @Description Returns a list of all documents belonging to the authenticated user
// @Tags Documents
// @Produce json
// @Success 200 {object} helper.Response{data=[]models.Document} "List of documents"
// @Failure 401 {object} helper.Response
// @Failure 500 {object} helper.Response
// @Router /documents [get]
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

// GetDocumentByID godoc
// @Summary Get a specific document
// @Description Returns document details and content by ID
// @Tags Documents
// @Produce json
// @Param document_id path string true "Document ID"
// @Success 200 {object} helper.Response{data=dto.DocumentResponse} "Document details"
// @Failure 400 {object} helper.Response
// @Failure 401 {object} helper.Response
// @Failure 403 {object} helper.Response
// @Failure 404 {object} helper.Response
// @Failure 500 {object} helper.Response
// @Router /documents/{document_id} [get]
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

// DeleteDocument godoc
// @Summary Delete a document
// @Description Deletes a document by ID (both file and database record)
// @Tags Documents
// @Produce json
// @Param document_id path string true "Document ID"
// @Success 200 {object} helper.Response
// @Failure 400 {object} helper.Response
// @Failure 401 {object} helper.Response
// @Failure 404 {object} helper.Response
// @Failure 500 {object} helper.Response
// @Router /documents/{document_id} [delete]
func DeleteDocument(c *gin.Context) {
	userID, err := helper.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, helper.NewErrorResponse("You are not authorized"))
		return
	}

	_, authorized := helper.CheckAuthenticationAndAuthorization(c, userID)
	if !authorized {
		return
	}

	documentID := c.Param("document_id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, helper.NewErrorResponse("Document ID is required"))
		return
	}

	var document models.Document
	if err := database.DB.Where("id = ? AND user_id = ?", documentID, userID).First(&document).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, helper.NewErrorResponse("Document not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Failed to get document"))
		return
	}

	// Удаление файла с диска
	if err := os.Remove(document.FilePath); err != nil && !os.IsNotExist(err) {
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Failed to delete file from disk: "+err.Error()))
		return
	}

	// Удаление записи из базы данных
	if err := database.DB.Delete(&document).Error; err != nil {
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Failed to delete document from database"))
		return
	}

	c.JSON(http.StatusOK, helper.NewSuccessResponse(nil))
}

// GetDocumentStatistics godoc
// @Summary Get document statistics
// @Description Calculates TF statistics for a given document, and IDF calculated as if all documents in collections, where the document we specified is, is in one collection
// @Tags Documents
// @Produce json
// @Param document_id path string true "Document ID"
// @Success 200 {object} helper.Response{data=object} "Document statistics"
// @Failure 400 {object} helper.Response
// @Failure 401 {object} helper.Response
// @Failure 404 {object} helper.Response
// @Failure 500 {object} helper.Response
// @Router /documents/{document_id}/statistics [get]
func GetDocumentStatistics(c *gin.Context) {
	userID, err := helper.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, helper.NewErrorResponse("You are not authorized"))
		return
	}

	_, authorized := helper.CheckAuthenticationAndAuthorization(c, userID)
	if !authorized {
		return
	}

	documentID := c.Param("document_id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, helper.NewErrorResponse("Document ID is required"))
		return
	}

	var document models.Document
	if err := database.DB.Preload("Collections").Where("id = ? AND user_id = ?", documentID, userID).First(&document).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, helper.NewErrorResponse("Document not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Failed to get document"))
		return
	}

	words, err := services.ProcessFile(document.FilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Cannot process file: "+err.Error()))
		return
	}

	// 4. Расчет TF для текущего документа
	wordCount := services.CountWords(words)
	tf := services.CalculateTF(wordCount, len(words))

	// 5. Обработка случая с коллекциями
	if len(document.Collections) > 0 {
		allDocs, err := services.GetAllCollectionDocuments(document.Collections)
		if err != nil {
			c.JSON(http.StatusInternalServerError, helper.NewErrorResponse("Failed to get collection documents"))
			return
		}

		// Подготовка данных для IDF
		var collectionWords []map[string]int
		for _, doc := range allDocs {
			words, err := services.ProcessFile(doc.FilePath)
			if err != nil {
				// эррор будет нужжен только мне
				log.Printf("Failed to process file %s: %v", doc.FilePath, err)
				continue
			}
			collectionWords = append(collectionWords, services.CountWords(words))
		}

		// Расчет статистики
		idf := services.CalculateIDF(collectionWords)
		rareWords := services.GetRarestWords(wordCount, 50)

		for i := range rareWords {
			rareWords[i].TF = tf[rareWords[i].Word]
			if idfValue, exists := idf[rareWords[i].Word]; exists {
				rareWords[i].IDF = idfValue
			} else {
				rareWords[i].IDF = math.Log(float64(len(collectionWords) + 1))
			}
			rareWords[i].Count = wordCount[rareWords[i].Word]
		}

		c.JSON(http.StatusOK, helper.NewSuccessResponse(gin.H{
			"statistics": rareWords,
			"meta": gin.H{
				"total_collections": len(document.Collections),
				"total_documents":   len(allDocs),
			},
		}))
		return
	}

	// 6. Документ не в коллекциях - возвращаем TF и Count
	tfOnlyStats := make([]services.WordStat, 0)
	for word, tfValue := range tf {
		tfOnlyStats = append(tfOnlyStats, services.WordStat{
			Word:  word,
			TF:    tfValue,
			Count: wordCount[word],
			IDF:   0,
		})
	}

	// Сортировка по TF (самые редкие сначала)
	sort.Slice(tfOnlyStats, func(i, j int) bool {
		return tfOnlyStats[i].TF < tfOnlyStats[j].TF
	})

	if len(tfOnlyStats) > 50 {
		tfOnlyStats = tfOnlyStats[:50]
	}

	c.JSON(http.StatusOK, helper.NewSuccessResponse(gin.H{
		"meta": gin.H{
			"message": "Document is not in any collections - showing TF only",
		},
		"statistics": tfOnlyStats,
	}),
	)
}
