package services

import (
	"math"
	"os"
	"regexp"
	"sort"
	"strings"
	"tfidf-app/database"
	"tfidf-app/models"
)

type WordStatForUpload struct {
	Word  string
	TF    float64
	Count int
	IDF   float64
	TFIDF float64
}

type WordStat struct {
	Word  string
	TF    float64
	Count int
	IDF   float64
}

func ComputeTFIDFForUpload(words []string) []WordStatForUpload {
	wordCount := make(map[string]int)
	for _, w := range words {
		wordCount[w]++
	}

	totalWords := len(words)

	stats := make([]WordStatForUpload, 0, len(wordCount))
	for w, count := range wordCount {
		tf := float64(count) / float64(totalWords)

		idf := math.Log(float64(totalWords) / float64(count))

		stats = append(stats, WordStatForUpload{
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

func ProcessFile(filePath string) ([]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	text := string(data)
	text = strings.ToLower(text)

	reg := regexp.MustCompile(`[^a-zA-Zа-яА-Я]+`)
	cleaned := reg.ReplaceAllString(text, " ")

	words := strings.Fields(cleaned)
	return words, nil
}

func GetAllCollectionDocuments(collections []*models.Collection) ([]models.Document, error) {
	var allDocs []models.Document
	seenDocs := make(map[uint]bool)

	for _, coll := range collections {
		var docs []models.Document
		if err := database.DB.Model(&coll).Association("Documents").Find(&docs); err != nil {
			return nil, err
		}

		for _, doc := range docs {
			if !seenDocs[doc.ID] {
				allDocs = append(allDocs, doc)
				seenDocs[doc.ID] = true
			}
		}
	}

	return allDocs, nil
}

func CountWords(words []string) map[string]int {
	wordCount := make(map[string]int)
	for _, word := range words {
		wordCount[word]++
	}
	return wordCount
}

func CalculateTF(wordCount map[string]int, totalWords int) map[string]float64 {
	tf := make(map[string]float64)

	if totalWords == 0 {
		return tf
	}

	for word, count := range wordCount {
		tf[word] = float64(count) / float64(totalWords)
	}
	return tf
}

func CalculateIDF(documents []map[string]int) map[string]float64 {
	idf := make(map[string]float64)
	totalDocs := len(documents)
	docFrequency := make(map[string]int) // Сколько документов содержат каждое слово

	for _, doc := range documents {
		for word := range doc {
			docFrequency[word]++
		}
	}

	for word, freq := range docFrequency {
		if freq == 0 {
			idf[word] = 0
		} else {
			idf[word] = math.Log(float64(totalDocs) / float64(freq))
		}
	}

	return idf
}

func GetRarestWords(wordCount map[string]int, limit int) []WordStat {
	var stats []WordStat
	for word, count := range wordCount {
		stats = append(stats, WordStat{
			Word:  word,
			Count: count,
		})
	}
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Count < stats[j].Count
	})
	if len(stats) > limit {
		stats = stats[:limit]
	}
	return stats
}
