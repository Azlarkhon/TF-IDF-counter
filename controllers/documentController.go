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
		tfidf := services.CalculateTFIDF(tf, idf)
		rareWords := services.GetRarestWords(tfidf, wordCount, 50)

		// Заполняем TF, Count и IDF для каждого слова
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
			TFIDF: tfValue,
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
