package services

import (
	"math"
	"sort"
	"time"
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
		return stats[i].TF > stats[j].TF
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
