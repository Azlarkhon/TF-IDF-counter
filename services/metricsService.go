package services

import (
	"errors"
	"fmt"
	"math"
	"tfidf-app/models"
	"time"

	"gorm.io/gorm"
)

func CalculateProcessingTime(start time.Time) float64 {
	seconds := time.Since(start).Seconds()
	return math.Round(seconds*1000) / 1000
}

func RoundFileSizeMB(size int64) float64 {
	mb := float64(size) / (1024 * 1024)
	return math.Round(mb*1000) / 1000
}

func UpdateMetrics(tx *gorm.DB, processingTime float64, fileSizeMB float64) (models.Metric, error) {
	var metric models.Metric
	currentTime := time.Now()

	result := tx.Preload("Words").First(&metric)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			metric = models.Metric{
				FilesProcessed:               1,
				LatestFileProcessedTimestamp: currentTime,
				MinTimeProcessed:             processingTime,
				AvgTimeProcessed:             processingTime,
				MaxTimeProcessed:             processingTime,
				TotalFileSizeMB:              fileSizeMB,
				AvgFileSizeMB:                fileSizeMB,
			}
			if err := tx.Create(&metric).Error; err != nil {
				return models.Metric{}, fmt.Errorf("failed to create metric: %w", err)
			}
			return metric, nil
		}
		return models.Metric{}, fmt.Errorf("failed to get metric: %w", result.Error)
	}

	metric.FilesProcessed++
	metric.LatestFileProcessedTimestamp = currentTime
	metric.MinTimeProcessed = math.Min(metric.MinTimeProcessed, processingTime)
	metric.MaxTimeProcessed = math.Max(metric.MaxTimeProcessed, processingTime)

	totalTime := metric.AvgTimeProcessed * float64(metric.FilesProcessed-1)
	metric.AvgTimeProcessed = (totalTime + processingTime) / float64(metric.FilesProcessed)

	metric.TotalFileSizeMB += fileSizeMB
	metric.AvgFileSizeMB = metric.TotalFileSizeMB / float64(metric.FilesProcessed)

	if err := tx.Save(&metric).Error; err != nil {
		return models.Metric{}, fmt.Errorf("failed to update metric: %w", err)
	}

	return metric, nil
}

func SaveWords(tx *gorm.DB, words []string, metricID uint) error {
	wordCount := make(map[string]int)
	for _, word := range words {
		wordCount[word]++
	}

	wordList := make([]string, 0, len(wordCount))
	for word := range wordCount {
		wordList = append(wordList, word)
	}

	var existingWords []models.Word
	if err := tx.Where("word IN ? AND metric_id = ?", wordList, metricID).Find(&existingWords).Error; err != nil {
		return fmt.Errorf("failed to find existing words: %w", err)
	}

	existingWordMap := make(map[string]*models.Word)
	for i := range existingWords {
		existingWordMap[existingWords[i].Word] = &existingWords[i]
	}

	var wordsToUpdate []*models.Word
	var wordsToCreate []models.Word

	for word, count := range wordCount {
		if existing, found := existingWordMap[word]; found {
			existing.Count += count
			wordsToUpdate = append(wordsToUpdate, existing)
		} else {
			wordsToCreate = append(wordsToCreate, models.Word{
				MetricID:  metricID,
				Word:      word,
				Count:     count,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			})
		}
	}

	// Обновляем существующие слова
	if len(wordsToUpdate) > 0 {
		if err := tx.Save(wordsToUpdate).Error; err != nil {
			return fmt.Errorf("failed to update words: %w", err)
		}
	}

	// Создаем новые слова
	if len(wordsToCreate) > 0 {
		if err := tx.Create(&wordsToCreate).Error; err != nil {
			return fmt.Errorf("failed to create words: %w", err)
		}
	}

	return nil
}
