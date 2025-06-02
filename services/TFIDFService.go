package services

import (
	"errors"
	"math"
	"sort"
	"tfidf-app/database"
	"tfidf-app/models"
	"time"

	"gorm.io/gorm"
)

type WordStat struct {
	Word  string
	TF    float64
	Count int
	IDF   float64
	TFIDF float64
}

func ComputeTFIDF(words []string) []WordStat {
	wordCount := make(map[string]int)
	for _, w := range words {
		wordCount[w]++
	}

	totalWords := len(words)

	stats := make([]WordStat, 0, len(wordCount))
	for w, count := range wordCount {
		tf := float64(count) / float64(totalWords)

		idf := math.Log(float64(totalWords) / float64(count))

		stats = append(stats, WordStat{
			Word:  w,
			TF:    tf,
			Count: count,
			IDF:   idf,
			TFIDF: tf * idf,
		})
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].IDF > stats[j].IDF
	})

	return stats
}

func CalculateProcessingTime(start time.Time) float64 {
	seconds := time.Since(start).Seconds()
	return math.Round(seconds*1000) / 1000
}

func RoundFileSizeMB(size int64) float64 {
	mb := float64(size) / (1024 * 1024)
	return math.Round(mb*1000) / 1000
}

func UpdateMetrics(processingTime float64, fileSizeMB float64) (models.Metric, error) {
	var metric models.Metric
	currentTime := time.Now()
	result := database.DB.Preload("Words").First(&metric)

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
			err := database.DB.Create(&metric).Error
			return metric, err
		}
		return metric, result.Error
	}

	metric.FilesProcessed++
	metric.LatestFileProcessedTimestamp = currentTime
	metric.MinTimeProcessed = math.Min(metric.MinTimeProcessed, processingTime)
	metric.MaxTimeProcessed = math.Max(metric.MaxTimeProcessed, processingTime)
	totalTime := metric.AvgTimeProcessed * float64(metric.FilesProcessed-1)
	metric.AvgTimeProcessed = (totalTime + processingTime) / float64(metric.FilesProcessed)
	metric.TotalFileSizeMB += fileSizeMB
	metric.AvgFileSizeMB = metric.TotalFileSizeMB / float64(metric.FilesProcessed)

	err := database.DB.Save(&metric).Error
	return metric, err
}

func SaveWords(words []string, metricID uint) error {
	wordCount := make(map[string]int)
	for _, word := range words {
		wordCount[word]++
	}

	wordList := make([]string, 0, len(wordCount))
	for word := range wordCount {
		wordList = append(wordList, word)
	}

	var existingWords []models.Word
	err := database.DB.
		Where("word IN ? AND metric_id = ?", wordList, metricID).
		Find(&existingWords).Error
	if err != nil {
		return err
	}

	existingWordMap := make(map[string]*models.Word)
	for i := range existingWords {
		existingWordMap[existingWords[i].Word] = &existingWords[i]
	}

	newWords := make([]models.Word, 0)
	for word, count := range wordCount {
		if existing, found := existingWordMap[word]; found {
			existing.Count += count
		} else {
			newWords = append(newWords, models.Word{
				MetricID:  metricID,
				Word:      word,
				Count:     count,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			})
		}
	}

	if len(existingWords) > 0 {
		if err := database.DB.Save(&existingWords).Error; err != nil {
			return err
		}
	}

	if len(newWords) > 0 {
		if err := database.DB.Create(&newWords).Error; err != nil {
			return err
		}
	}

	return nil
}
